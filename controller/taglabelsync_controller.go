package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	taglabelv1 "api/v1"
)

func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr, context.Background()))
}

func newReconciler(mgr manager.Manager, ctx context.Context) reconcile.Reconciler {
	return &ReconcileTagLabelSync{
		Client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
		// ctx:    ctx,
	}
}

func add(mgr manager.Manager, r reconcile.Reconciler) error {
	c, err := controller.New("tag-label-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// watch changes in VM tags
	err = c.Watch(&source.Kind{Type: taglabelv1.TagLabelSync{}}, nil)
	if err != nil {
		return err
	}

	// watch changes in node labels?

	return nil
}

type ReconcileTagLabelSync struct {
	client.Client
	scheme   *runtime.Scheme
	recorder record.EventRecorder
	ctx      context.Context
}

func (r *ReconcileTagLabelSync) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// ctx := context.Background()

	// Get ARM VM tags in cluster
	// Get node labels in cluster
	// check if any differences
	// if different, add VM tag to node as label

	return ctrl.Result{}, nil
}