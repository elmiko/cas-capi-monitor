package controllers

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	// tried to import these from upstream but there is some issue with the dependency imports
	toBeDeletedTaint       = "ToBeDeletedByClusterAutoscaler"
	deletionCandidateTaint = "DeletionCandidateOfClusterAutoscaler"
)

type NodeReconciler struct {
	client.Client
}

func (r *NodeReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	log := logf.Log.WithName("node-reconciler")

	n := &corev1.Node{}
	err := r.Get(ctx, req.NamespacedName, n)
	if client.IgnoreNotFound(err) != nil {
		return reconcile.Result{}, err
	}

	for _, t := range n.Spec.Taints {
		switch t.Key {
		case toBeDeletedTaint:
			log.V(0).Info("observed Node with ToBeDeleted taint", "name", n.Name)
		case deletionCandidateTaint:
			log.V(0).Info("observed Node with DeletionCandidate taint", "name", n.Name)
		default:
			continue
		}
	}

	return reconcile.Result{}, nil
}
