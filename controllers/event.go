package controllers

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type EventReconciler struct {
	client.Client
}

func (r *EventReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	log := logf.Log.WithName("event-reconciler")

	e := &corev1.Event{}
	err := r.Get(ctx, req.NamespacedName, e)
	if client.IgnoreNotFound(err) != nil {
		return reconcile.Result{}, err
	}

	if e.Source.Component == "cluster-autoscaler" {
		log.V(0).Info("observed cluster-autoscaler generated Event", "reason", e.Reason, "message", e.Message)
	}

	return reconcile.Result{}, nil
}
