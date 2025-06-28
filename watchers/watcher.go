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
	loopDelay         = 5 * time.Second
	minSizeAnnotation = "cluster.x-k8s.io/cluster-api-autoscaler-node-group-min-size"
	maxSizeAnnotation = "cluster.x-k8s.io/cluster-api-autoscaler-node-group-max-size"
)

type Watcher struct {
	managementClient client.Client
	workloadClient   client.Client
}

func NewWatcher(mClient, wClient client.Client) Watcher {
	return Watcher{
		managementClient: mClient,
		workloadClient:   wClient,
	}
}

func (w Watcher) Start(ctx context.Context) {
	log := logf.Log.WithName("watcher")

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
			annotatedMachineDeployments = append(annotatedMachineDeployments, fmt.Sprintf("%s/%d", md.Name, md.Spec.Replicas))
		}
	}
	log.V(0).Info("observed MachineDeployments with scaling annotations",
		"count", len(annotatedMachineDeployments),
		"names", strings.Join(annotatedMachineDeployments, ","))

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
