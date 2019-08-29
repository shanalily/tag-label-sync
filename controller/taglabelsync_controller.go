package controller

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"tag-label-sync.io/azure"
	"tag-label-sync.io/azure/scalesets"
	"tag-label-sync.io/azure/vms"
)

const (
	VM   string = "virtualMachines"
	VMSS string = "virtualMachineScaleSets"
)

type ReconcileTagLabelSync struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
	ctx      context.Context
}

type ComputeResourceClient interface {
	// how can I make an interface for Spec that allows me to use VM and VMSS with the same function?
	// how am I supposed to do this when different clients are returned?
	NewClient(subscriptionID string, resourceName string) error
	Get(ctx context.Context, name string) error
	Update(ctx context.Context) error
	// Spec() (idk how to get this to work)
	Tags() map[string]*string
	// SetTag(name, value string)
}

type VirtualMachineClient struct {
	client *vms.Client
	vm     *vms.Spec
}

func (m VirtualMachineClient) NewClient(subscriptionID, resourceName string) error {
	var err error
	m.client, err = vms.NewClient(subscriptionID, resourceName)
	if err != nil {
		return err
	}
	return nil
}

func (m VirtualMachineClient) Get(ctx context.Context, name string) error {
	var err error
	m.vm, err = m.client.Get(ctx, name)
	if err != nil {
		return err
	}
	return nil
}

func (m VirtualMachineClient) Update(ctx context.Context) error {
	if err := m.client.Update(ctx, *m.vm.Spec().Name, m.vm); err != nil {
		return err
	}
	return nil
}

func (m VirtualMachineClient) Tags() map[string]*string {
	return m.vm.Spec().Tags
}

func (m VirtualMachineClient) SetTag(name, value string) {

}

// okay so maybe what I should have done is make a virtualmachinescaleset interface and then I can fake it more easily
type VirtualMachineScaleSetClient struct {
	client *scalesets.Client
	vmss   *scalesets.Spec
}

// I'm not sure this is actually modifying the client :(
// maybe I can make this not a receiver method... and the others can be receiver methods? but then
// I can't pass as parameter easily...
func (m VirtualMachineScaleSetClient) NewClient(subscriptionID, resourceName string) error {
	// func NewClient(m *VirtualMachineScaleSet, subscriptionID, resourceName string) error {
	client, err := scalesets.NewClient(subscriptionID, resourceName)
	if err != nil {
		return err
	}
	m.client = client
	return nil
}

func (m VirtualMachineScaleSetClient) Get(ctx context.Context, name string) error {
	// func Get(m *VirtualMachineScaleSet, ctx context.Context, name string) error {
	vmss, err := m.client.Get(ctx, name)
	if err != nil {
		return err
	}
	m.vmss = vmss
	return nil
}

func (m VirtualMachineScaleSetClient) Update(ctx context.Context) error {
	if err := m.client.Update(ctx, *m.vmss.Spec().Name, m.vmss); err != nil {
		return err
	}
	return nil
}

func (m VirtualMachineScaleSetClient) Tags() map[string]*string {
	return m.vmss.Spec().Tags
}

// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=nodes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=nodes/status,verbs=get

func (r *ReconcileTagLabelSync) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("tag-label-sync", request.NamespacedName)
	r.ctx = ctx

	var configMap corev1.ConfigMap
	var configOptions ConfigOptions
	optionsNamespacedName := types.NamespacedName{Name: "tag-label-sync", Namespace: "default"} // is this okay
	if err := r.Get(ctx, optionsNamespacedName, &configMap); err != nil {
		log.V(1).Info("unable to fetch ConfigMap, instead using default configuration settings")
		// should I allow this to continue? It would be unfortunate to have things sync and then clean it up
		configOptions = DefaultConfigOptions()
	} else {
		configOptions, err = NewConfigOptions(configMap) // ConfigMap.Data is string -> string but I don't always want that
		if err != nil {
			log.Error(err, "failed to load options from config file")
			return ctrl.Result{}, err
		}
	}

	var node corev1.Node
	if err := r.Get(ctx, request.NamespacedName, &node); err != nil {
		log.Error(err, "unable to fetch Node")
		return ctrl.Result{}, err // what should I return here?
	}
	log.V(1).Info("provider info", "provider ID", node.Spec.ProviderID)
	provider, err := azure.ParseProviderID(node.Spec.ProviderID)
	if err != nil {
		log.Error(err, "invalid provider ID")
	}

	switch provider.ResourceType {
	case VMSS:
		// Get VMSS client
		vmssClient := VirtualMachineScaleSetClient{}
		var err error
		vmssClient.client, err = scalesets.NewClient(provider.SubscriptionID, provider.ResourceGroup)
		if err != nil {
			log.Error(err, "failed to create VMSS client")
			return reconcile.Result{}, err
		}
		// have vmssClient wrapped in something that I pass to apply?

		vmssClient.vmss, err = vmssClient.client.Get(r.ctx, provider.ResourceName)
		if err != nil {
			log.Error(err, "failed to get VMSS")
		}

		// Add VMSS tags to node
		if err := r.reconcileVMSS(request, vmssClient, provider.ResourceName, &node, configOptions); err != nil {
			log.Error(err, "failed to apply tags to nodes")
			return reconcile.Result{}, err
		}
	case VM:
		// Get VM Client
		vmClient := VirtualMachineClient{}
		vmClient.client, err = vms.NewClient(provider.SubscriptionID, provider.ResourceGroup)
		if err != nil {
			log.Error(err, "failed to create VM client")
			return reconcile.Result{}, err
		}

		// Add VM tags to node
		if err := r.reconcileVMs(request, vmClient, provider.ResourceName, &node, configOptions); err != nil {
			log.Error(err, "failed to apply tags to nodes")
			return reconcile.Result{}, err
		}
	default:
		log.V(1).Info("unrecognized resource type", "resource type", provider.ResourceType)
	}

	return ctrl.Result{}, nil
}

// pass VMSS -> tags info and assign to nodes on VMs (unless node already has label)
func (r *ReconcileTagLabelSync) reconcileVMSS(request reconcile.Request, vmssClient VirtualMachineScaleSetClient, resourceName string, node *corev1.Node, configOptions ConfigOptions) error {
	log := r.Log.WithValues("tag-label-sync", request.NamespacedName)
	// each VMSS may have multiple nodes, but I think each nodes is only in one VMSS
	// whats the fastest way to check if Node already has label? benefit of map

	log.V(0).Info("configOptions", "sync direction", configOptions.SyncDirection)

	if err := r.applyTagsToNodes(request, vmssClient, node, configOptions); err != nil {
		return err
	}

	// assign all labels on Node to VMSS, if not already there

	if err := r.applyLabelsToAzureResource(request, vmssClient, node, configOptions); err != nil {
		return err
	}

	return nil
}

// I want to get to the point where this function can be called on either vm or vmss
func (r *ReconcileTagLabelSync) reconcileVMs(request reconcile.Request, vmClient ComputeResourceClient, resourceName string, node *corev1.Node, configOptions ConfigOptions) error {
	log := r.Log.WithValues("tag-label-sync", request.NamespacedName)

	log.V(0).Info("configOptions", "sync direction", configOptions.SyncDirection)

	if err := r.applyTagsToNodes(request, vmClient, node, configOptions); err != nil {
		return err
	}

	if err := r.applyLabelsToAzureResource(request, vmClient, node, configOptions); err != nil {
		return err
	}
	return nil
}

func (r *ReconcileTagLabelSync) applyTagsToNodes(request reconcile.Request, computeClient ComputeResourceClient, node *corev1.Node, configOptions ConfigOptions) error {
	log := r.Log.WithValues("tag-label-sync", request.NamespacedName)

	if configOptions.SyncDirection == TwoWay || configOptions.SyncDirection == ARMToNode {
		for tagName, tagVal := range computeClient.Tags() {
			// what if key exists but different value? what takes priority? currently just going to ignore and only add tags that don't exist
			validLabelName := ConvertTagNameToValidLabelName(tagName, configOptions)
			labelVal, ok := node.Labels[validLabelName]
			if !ok {
				// add tag as lael
				log.V(1).Info("applying tags to nodes", "tagName", tagName, "tagVal", *tagVal)

				node.Labels[validLabelName] = *tagVal
				if err := r.Update(r.ctx, node); err != nil { // should this be a patch?
					return err
				}
			} else if labelVal != *tagVal {
				log.V(0).Info("updating", "using policy", configOptions.ConflictPolicy)
				switch configOptions.ConflictPolicy {
				case ARMPrecedence:
					// set label anyway
					node.Labels[validLabelName] = *tagVal
					if err := r.Update(r.ctx, node); err != nil {
						return err
					}
				case NodePrecedence:
					// do nothing
					log.V(0).Info("name->value conflict found", "node label value", labelVal, "ARM tag value", *tagVal)
				case Ignore:
					// raise k8s event
					r.Recorder.Event(node, "Warning", "ConflictingTagLabelValues",
						fmt.Sprintf("ARM tag was not applied to node because a different value for '%s' already exists (%s != %s).", tagName, *tagVal, labelVal))
					log.V(0).Info("name->value conflict found, leaving unchanged", "label value", labelVal, "tag value", *tagVal)
				default:
					return errors.New("unrecognized conflict policy")
				}
			}
		}
	}
	return nil
}

// I need to make sure I can get update to work with ComputeResource interface! value vs reference issue
func (r *ReconcileTagLabelSync) applyLabelsToAzureResource(request reconcile.Request, computeClient ComputeResourceClient, node *corev1.Node, configOptions ConfigOptions) error {
	log := r.Log.WithValues("tag-label-sync", request.NamespacedName)

	if configOptions.SyncDirection == TwoWay || configOptions.SyncDirection == NodeToARM {
		if len(computeClient.Tags()) > maxNumTags {
			// error
			log.V(1).Info("can't add any more tags", "number of tags", len(computeClient.Tags()))
			return nil
		}
		for labelName, labelVal := range node.Labels {
			if !ValidTagName(labelName, configOptions) {
				// I don't think I want to return yet
				// return errors.New(fmt.Sprintf("invalid tag name: %s", labelName))
				// log.Error(errors.New(fmt.Sprintf("invalid tag name")), fmt.Sprintf("label name: %s", labelName))
				log.V(0).Info("invalid tag name", "label name", labelName)
				continue
			}
			validTagName := ConvertLabelNameToValidTagName(labelName, configOptions)
			tagVal, ok := computeClient.Tags()[validTagName]
			if !ok {
				// add label as tag
				log.V(1).Info("applying labels to VMSS", "labelVal", labelVal, "tagVal", *tagVal)

				computeClient.Tags()[validTagName] = &labelVal // problem!!!
				if err := computeClient.Update(r.ctx); err != nil {
					// log.Error(err, "failed to update VMSS", "labelName", validTagName, "labelVal", labelVal)
					log.Error(err, "failed to update VMSS", "labelName", labelName, "labelVal", labelVal)
				}
			} else if *tagVal != labelVal {
				switch configOptions.ConflictPolicy {
				case NodePrecedence:
					// set tag anyway
					computeClient.Tags()[validTagName] = &labelVal // problem!!!
					if err := computeClient.Update(r.ctx); err != nil {
						// log.Error(err, "failed to update VMSS", "labelName", validTagName, "labelVal", labelVal)
						log.Error(err, "failed to update VMSS", "labelName", labelName, "labelVal", labelVal)
					}
				case ARMPrecedence:
					// do nothing
					log.V(0).Info("name->value conflict found", "node label value", labelVal, "ARM tag value", *tagVal)
				case Ignore:
					// raise kubernetes event
					r.Recorder.Event(node, "Warning", "ConflictingTagLabelValues",
						fmt.Sprintf("node label was not applied to VMSS because a different value for '%s' already exists (%s != %s).", labelName, labelVal, *tagVal))
					log.V(0).Info("name->value conflict found, leaving unchanged", "label value", labelVal, "tag value", *tagVal)
				default:
					return errors.New("unrecognized conflict policy")
				}
			}
		}
	}
	return nil
}

func (r *ReconcileTagLabelSync) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Node{}).
		Complete(r)
}
