package watchers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	capiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	minSizeAnnotation = "cluster.x-k8s.io/cluster-api-autoscaler-node-group-min-size"
	maxSizeAnnotation = "cluster.x-k8s.io/cluster-api-autoscaler-node-group-max-size"
)

type Watcher struct {
	managementClient client.Client
	workloadClient   client.Client
	interval         int
}

func NewWatcher(mClient, wClient client.Client, interval int) Watcher {
	return Watcher{
		managementClient: mClient,
		workloadClient:   wClient,
		interval:         interval,
	}
}

func (w Watcher) Start(ctx context.Context) {
	log := logf.Log.WithName("watcher")
	loopDelay := time.Duration(w.interval) * time.Second

	for {
		loopStart := time.Now()

		select {
		case <-ctx.Done():
			return
		default:
			w.harvestAndLogData(ctx, log)
		}

		loopEnd := time.Now()
		loopElapsed := loopEnd.Sub(loopStart)
		if loopElapsed > loopDelay {
			log.Info("watcher loop took longer than expected", "elapsed time", loopElapsed)
		} else {
			time.Sleep(loopDelay - loopElapsed)
		}
	}
}

func (w Watcher) harvestAndLogData(ctx context.Context, log logr.Logger) {
	machineDeploymentList := &capiv1beta1.MachineDeploymentList{}
	if err := w.managementClient.List(ctx, machineDeploymentList); err != nil {
		log.Error(err, "unable to list machinedeployments")
	}

	annotatedMachineDeployments := []string{}
	for _, md := range machineDeploymentList.Items {
		_, minOk := md.Annotations[minSizeAnnotation]
		_, maxOk := md.Annotations[maxSizeAnnotation]
		if minOk && maxOk {
			annotatedMachineDeployments = append(annotatedMachineDeployments, fmt.Sprintf("%s/%d", md.Name, *md.Spec.Replicas))
		}
	}
	log.V(0).Info("observed MachineDeployments with scaling annotations",
		"count", len(annotatedMachineDeployments),
		"names", strings.Join(annotatedMachineDeployments, ","))

	nodeList := &corev1.NodeList{}
	if err := w.workloadClient.List(ctx, nodeList); err != nil {
		log.Error(err, "unable to list nodes")
	}

	readyNodes := []string{}
	notReadyNodes := []string{}
	deletingNodes := []string{}
	for _, n := range nodeList.Items {
		if n.DeletionTimestamp != nil {
			deletingNodes = append(deletingNodes, n.Name)
		}
		for _, c := range n.Status.Conditions {
			if c.Type == corev1.NodeReady {
				if c.Status == corev1.ConditionTrue {
					readyNodes = append(readyNodes, n.Name)
				} else {
					notReadyNodes = append(notReadyNodes, n.Name)
				}
			}
		}
	}
	log.V(0).Info("observed Nodes with true ready condition",
		"count", len(readyNodes),
		"names", strings.Join(readyNodes, ","))
	log.V(0).Info("observed Nodes with false or unknown ready condition",
		"count", len(notReadyNodes),
		"names", strings.Join(notReadyNodes, ","))
	log.V(0).Info("observed Nodes with non-zero deletion timestamp",
		"count", len(deletingNodes),
		"names", strings.Join(deletingNodes, ","))

	podList := &corev1.PodList{}
	if err := w.workloadClient.List(ctx, podList); err != nil {
		log.Error(err, "unable to list pods")
	}

	pendingPods := []string{}
	for _, p := range podList.Items {
		if p.Status.Phase == corev1.PodPending {
			pendingPods = append(pendingPods, fmt.Sprintf("%s/%s", p.Namespace, p.Name))
		}
	}
	log.V(0).Info("observed pending Pods",
		"count", len(pendingPods),
		"names", strings.Join(pendingPods, ","))
}
