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
	recorder record.EventRecorder
	ctx      context.Context
}

// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=nodes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=nodes/status,verbs=get

func (r *ReconcileTagLabelSync) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("tag-label-sync", request.NamespacedName)
	r.ctx = ctx

	configMap := corev1.ConfigMap{}
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
		// log.V(1).Info("configMap", "label prefix value", configOptions.LabelPrefix)
		// log.V(1).Info("configMap", "sync direction", configOptions.SyncDirection)
		// log.V(1).Info("configMap", "interval", configOptions.Interval)
		// log.V(1).Info("configMap", "resource group", configOptions.ResourceGroupFilter)
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
		vmssClient, err := scalesets.NewClient(provider.SubscriptionID, provider.ResourceGroup)
		if err != nil {
			log.Error(err, "failed to create VMSS client")
		}
		vmss, err := vmssClient.Get(ctx, provider.ResourceName)
		if err != nil {
			log.Error(err, "failed to get VMSS")
		}

		// Add VMSS tags to node
		if err := r.applyVMSSTagsToNodes(request, vmss, &node, vmssClient, configOptions); err != nil {
			log.Error(err, "failed to apply tags to nodes")
			return reconcile.Result{}, err
		}
	case VM:
		// Get VM Client
		vmClient, err := vms.NewClient(provider.SubscriptionID, provider.ResourceGroup)
		if err != nil {
			log.Error(err, "failed to create VM client")
		}
		vm, err := vmClient.Get(ctx, provider.ResourceName)
		if err != nil {
			log.Error(err, "failed to get VM")
		}

		// Add VM tags to node
		if err := r.applyVMTagsToNodes(request, vm, &node, vmClient, configOptions); err != nil {
			log.Error(err, "failed to apply tags to nodes")
		}
	default:
		log.V(1).Info("unrecognized resource type", "resource type", provider.ResourceType)
	}

	return ctrl.Result{}, nil
}

// pass VMSS -> tags info and assign to nodes on VMs (unless node already has label)
func (r *ReconcileTagLabelSync) applyVMSSTagsToNodes(request reconcile.Request, vmss *scalesets.Spec, node *corev1.Node, vmssClient *scalesets.Client, configOptions ConfigOptions) error {
	log := r.Log.WithValues("tag-label-sync", request.NamespacedName)
	// each VMSS may have multiple nodes, but I think each nodes is only in one VMSS
	// whats the fastest way to check if Node already has label? benefit of map

	// assign all tags on VMSS to Node, if not already there
	log.V(0).Info("configOptions", "sync direction", configOptions.SyncDirection)
	if configOptions.SyncDirection == TwoWay || configOptions.SyncDirection == ARMToNode {
		for tagName, tagVal := range vmss.Spec().Tags {
			// what if key exists but different value? what takes priority? currently just going to ignore and only add tags that don't exist
			labelVal, ok := node.Labels[convertTagNameToValidLabelName(tagName, configOptions)]
			if !ok {
				// add tag as label
				log.V(1).Info("applying tags to nodes", "tagName", tagName, "tagVal", *tagVal)

				node.Labels[convertTagNameToValidLabelName(tagName, configOptions)] = *tagVal
				err := r.Update(context.TODO(), node) // should this be a patch?
				if err != nil {
					return err
				}
			} else if labelVal != *tagVal {
				// TODO
				switch configOptions.ConflictPolicy {
				case ARMPrecedence:
					// set label anyway
					node.Labels[convertTagNameToValidLabelName(tagName, configOptions)] = *tagVal
					if err := r.Update(context.TODO(), node); err != nil {
						return err
					}
				case NodePrecedence:
					// do nothing
					log.V(0).Info("name->value conflict found", "label value", labelVal, "tag value", *tagVal)
				case Ignore:
					// raise k8s event
					log.V(0).Info("name->value conflict found, leaving unchanged", "label value", labelVal, "tag value", *tagVal)
				default:
					return errors.New("unrecognized conflict policy")
				}
				return errors.New(fmt.Sprintf("Label already exists on node %s but with different value", node.Name))
			}
		}
	}

	// assign all labels on Node to VMSS, if not already there

	if configOptions.SyncDirection == TwoWay || configOptions.SyncDirection == NodeToARM {
		if len(vmss.Spec().Tags) > 50 {
			// error
			log.V(1).Info("can't add any more tags", "number of tags", len(vmss.Spec().Tags))
			return nil
		}
		for labelName, labelVal := range node.Labels {
			tagVal, ok := vmss.Spec().Tags[convertLabelNameToValidTagName(labelName, configOptions)]
			if !ok {
				// add label as tag
				log.V(1).Info("applying labels to VMSS", "labelVal", labelVal)

				// validTagName := azure.ConvertToValidTagName(labelName)
				// vmss.Spec().Tags[validTagName] = &labelVal
				vmss.Spec().Tags[convertLabelNameToValidTagName(labelName, configOptions)] = &labelVal
				if err := vmssClient.Update(context.TODO(), *vmss.Spec().Name, vmss); err != nil {
					// log.Error(err, "failed to update VMSS", "labelName", validTagName, "labelVal", labelVal)
					log.Error(err, "failed to update VMSS", "labelName", labelName, "labelVal", labelVal)
				}
			} else if labelVal != *tagVal {
				// switch configOptions.ConflictPolicy {
				// case NodePrecedence:
				// 	// set tag anyway
				// case ARMPrecedence:
				// 	// do nothing
				// case Ignore:
				// 	// raise kubernetes event
				// default:
				// 	// error
				// }
				return errors.New(fmt.Sprintf("Tag already exists on VMSS %s but with different value", *vmss.Spec().Name))
			}
		}
	}

	return nil
}

func (r *ReconcileTagLabelSync) applyVMTagsToNodes(request reconcile.Request, vm *vms.Spec, node *corev1.Node, vmClient *vms.Client, configOptions ConfigOptions) error {
	log := r.Log.WithValues("tag-label-sync", request.NamespacedName)

	for tagName, tagVal := range vm.Spec().Tags {
		labelVal, ok := node.Labels[tagName]
		if !ok {
			// add tag as label
			log.V(1).Info("applying tags to nodes", "tagName", tagName, "tagVal", *tagVal)

			node.Labels[tagName] = *tagVal
			err := r.Update(context.TODO(), node) // should this be a patch?
			if err != nil {
				return err
			}
		} else if labelVal != *tagVal {
			// TODO
			return errors.New(fmt.Sprintf("Label already exists on node %s but with different value", node.Name))
		}
	}

	return nil
}

func (r *ReconcileTagLabelSync) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Node{}).
		Complete(r)
}
