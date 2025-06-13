package controllers

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type MachineReconciler struct {
	client.Client
}

func (r *MachineReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	return reconcile.Result{}, nil
}
