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
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"tag-label-sync.io/azure"
	"tag-label-sync.io/azure/scalesets"
	"tag-label-sync.io/azure/scalesetvms"
)

// type ResourceType string

const (
	VM   string = "virtualMachines"
	VMSS string = "virtualMachineScaleSets"
)

func newReconciler(mgr manager.Manager, ctx context.Context) reconcile.Reconciler {
	return &ReconcileTagLabelSync{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		// ctx:    ctx,
	}
}

type ReconcileTagLabelSync struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	recorder record.EventRecorder
	ctx      context.Context
}

func (r *ReconcileTagLabelSync) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	ctx := context.TODO()
	log := r.Log.WithValues("tag-label-sync", request.NamespacedName)

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
		vmssClient, err := scalesets.NewClient(provider.SubscriptionID, provider.ResourceGroup)
		if err != nil {
			log.Error(err, "failed to create VMSS client")
		}
		vmss, err := vmssClient.Get(ctx, provider.ResourceName)
		if err != nil {
			log.Error(err, "failed to get VMSS")
		}
		// does this vmss object have anything useful or is it all empty fields :'(
		log.V(1).Info("printing tags...", "number of tags", len(vmss.Tags))
		for k, v := range vmss.Tags {
			log.V(1).Info("virtual machine scale set", "tag", k, "tag value", *v)
		}

		if err := r.applyVMSSTagsToNodes(request, vmss, &node); err != nil {
			log.Error(err, "failed to apply tags to nodes")
			return reconcile.Result{}, err
		}
	case VM:
		// this needs to change to VMs instead of scaleset VMs!
		vmClient, err := scalesetvms.NewClient(provider.SubscriptionID, provider.ResourceGroup)
		if err != nil {
			log.Error(err, "failed to create VM client")
		}
		// vms, err := vmClient.List(ctx, provider.ResourceName)
		_, err = vmClient.List(ctx, provider.ResourceName)
		if err != nil {
			log.Error(err, "failed to get VMs")
		}
		// log.V(1).Info("virtual machine scale set VM", "tags", vms.Tags)
	}

	log.V(1).Info("Node has provider ID", "provider ID", node.Spec.ProviderID)
	log.V(1).Info("Node has resource type", "resource type", provider.ResourceType)

	// Get ARM VM tags in cluster
	// Get node labels in cluster (I think I can list nodes)
	// check if any differences
	// if different, add VM tag to node as label

	for k, v := range node.Labels {
		log.V(1).Info("Node", "label", k, "value", v)
	}

	log.V(1).Info("reconciled")

	return ctrl.Result{}, nil
}

// pass VM -> tags info and assign to nodes on VMs (unless node already has label)
func (r *ReconcileTagLabelSync) applyVMSSTagsToNodes(request reconcile.Request, vmss *scalesets.Spec, node *corev1.Node) error {
	log := r.Log.WithValues("tag-label-sync", request.NamespacedName)
	// each VMSS may have multiple nodes, but I think each nodes is only in one VMSS
	// whats the fastest way to check if Node already has label? benefit of map
	// assign all tags on VMSS to Node, if not already there
	for tagName, tagVal := range vmss.Tags {
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
	//	_, ok := vmss.Tags[labelName]
	//	if !ok {
	//		// add label as tag
	//		log.V(1).Info("applying labels to VMSS", "labelVal", labelVal)
	//	}
	// }
	return nil
}

func (r *ReconcileTagLabelSync) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Node{}).
		Complete(r)
}
