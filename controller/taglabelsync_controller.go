package controller

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
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

	log.V(1).Info("request", "NamespacedName", request.NamespacedName)
	// am I going to have to load the config map every time? I don't expect it to change much
	// also, how do I ensure there's only one config map? do I find it by namespace?
	var configMap corev1.ConfigMap // am I going to have to load the config map every time? I don't expect it to change much
	if err := r.Get(ctx, request.NamespacedName, &configMap); err != nil {
		log.Error(err, "unable to fetch ConfigMap, instead using default configuration settings")
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
		if err := r.applyVMSSTagsToNodes(request, vmss, &node, vmssClient); err != nil {
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
		if err := r.applyVMTagsToNodes(request, vm, &node, vmClient); err != nil {
			log.Error(err, "failed to apply tags to nodes")
		}
	default:
		log.V(1).Info("unrecognized resource type", "resource type", provider.ResourceType)
	}

	return ctrl.Result{}, nil
}

// pass VMSS -> tags info and assign to nodes on VMs (unless node already has label)
func (r *ReconcileTagLabelSync) applyVMSSTagsToNodes(request reconcile.Request, vmss *scalesets.Spec, node *corev1.Node, vmssClient *scalesets.Client) error {
	log := r.Log.WithValues("tag-label-sync", request.NamespacedName)
	// each VMSS may have multiple nodes, but I think each nodes is only in one VMSS
	// whats the fastest way to check if Node already has label? benefit of map

	// assign all tags on VMSS to Node, if not already there
	for tagName, tagVal := range vmss.Spec().Tags {
		// what if key exists but different value? what takes priority? currently just going to ignore and only add tags that don't exist
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

	// assign all labels on Node to VMSS, if not already there

	// for labelName, labelVal := range node.Labels {
	// 	_, ok := vmss.Spec().Tags[labelName]
	// 	if !ok {
	// 		// add label as tag
	// 		log.V(1).Info("applying labels to VMSS", "labelVal", labelVal)

	// 		// validTagName := azure.ConvertToValidTagName(labelName)
	// 		// vmss.Spec().Tags[validTagName] = &labelVal
	// 		vmss.Spec().Tags[labelName] = &labelVal
	// 		if err := vmssClient.Update(context.TODO(), *vmss.Spec().Name, vmss); err != nil {
	// 			// log.Error(err, "failed to update VMSS", "labelName", validTagName, "labelVal", labelVal)
	// 			log.Error(err, "failed to update VMSS", "labelName", labelName, "labelVal", labelVal)
	// 		}
	// 	}
	// }
	return nil
}

func (r *ReconcileTagLabelSync) applyVMTagsToNodes(request reconcile.Request, vm *vms.Spec, node *corev1.Node, vmClient *vms.Client) error {
	return nil
}

func (r *ReconcileTagLabelSync) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// For(&corev1.ConfigMap{}).
		For(&corev1.Node{}).
		// Owns(&corev1.ConfigMap{}).
		Complete(r)
}
