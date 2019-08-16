package controller

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	taglabelv1 "tag-label-sync.io/api/v1"
	"tag-label-sync.io/azure"
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

func add(mgr manager.Manager, r reconcile.Reconciler) error {
	c, err := controller.New("tag-label-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// watch changes in VM tags... I need to change this, unless I can somehow get tag info through my Kind
	err = c.Watch(&source.Kind{Type: &taglabelv1.TagLabelSync{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// watch changes in node labels? should this come later?
	err = c.Watch(&source.Kind{Type: &corev1.Node{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

type ReconcileTagLabelSync struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	recorder record.EventRecorder
	ctx      context.Context
}

func (r *ReconcileTagLabelSync) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("tag-label-sync", request.NamespacedName)

	var sync taglabelv1.TagLabelSync
	err := r.Get(context.TODO(), request.NamespacedName, &sync)
	if err != nil {
		if azure.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
	}

	var node corev1.Node
	if err := r.Get(ctx, request.NamespacedName, &node); err != nil {
		log.Error(err, "unable to fetch Node")
		return ctrl.Result{}, err // what should I return here?
	}
	provider, err := azure.ParseProviderID(node.Spec.ProviderID)
	if err != nil {
		log.Error(err, "invalid provider ID")
	}

	// hardcoded for vmss right now
	if provider.ResourceType == VMSS {
		vmssClient, err := azure.NewVMSSClient(provider.SubscriptionID, provider.ResourceGroup)
		if err != nil {
			log.Error(err, "failed to create VMSS client")
		}
		vmss, err := vmssClient.List(ctx, provider.ResourceName)
	}

	log.V(1).Info("Node has provider ID", "provider ID", node.Spec.ProviderID)
	log.V(1).Info("Node has resource type", "resource type", provider.ResourceType)

	// Get ARM VM tags in cluster
	// Get node labels in cluster (I think I can list nodes)
	// check if any differences
	// if different, add VM tag to node as label

	if err := r.applyTagsToNodes(); err != nil {
		log.Error(err, "failed to apply tags to nodes")
		return reconcile.Result{}, err
	}

	log.V(1).Info("reconciled")

	return ctrl.Result{}, nil
}

// pass VM -> tags info and assign to nodes on VMs (unless node already has label)
func (r *ReconcileTagLabelSync) applyTagsToNodes() error {
	return nil
}

func (r *ReconcileTagLabelSync) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Node{}).
		Complete(r)
}
