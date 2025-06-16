package controllers

import (
	"context"

	capiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type MachineReconciler struct {
	client.Client
}

func (r *MachineReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	log := logf.Log.WithName("machine-reconciler")

	m := &capiv1beta1.Machine{}
	err := r.Get(ctx, req.NamespacedName, m)
	if client.IgnoreNotFound(err) != nil {
		return reconcile.Result{}, err
	}

	if val, ok := m.Annotations[capiv1beta1.DeleteMachineAnnotation]; ok {
		log.V(0).Info("observed Machine with deletion annotation", "name", m.Name, "value", val)
	}

	return reconcile.Result{}, nil
}
