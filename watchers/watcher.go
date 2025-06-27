package watchers

import (
	"context"
	"strings"
	"time"

	"github.com/go-logr/logr"
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
	err := w.managementClient.List(ctx, machineDeploymentList)
	if err != nil {
		log.Error(err, "unable to list machinedeployments")
	}

	annotatedMachineDeployments := []string{}
	for _, md := range machineDeploymentList.Items {
		_, minOk := md.Annotations[minSizeAnnotation]
		_, maxOk := md.Annotations[maxSizeAnnotation]
		if minOk && maxOk {
			annotatedMachineDeployments = append(annotatedMachineDeployments, md.Name)
		}
	}
	log.V(0).Info("observed MachineDeployments with scaling annotations", "length", len(annotatedMachineDeployments), "machinedeployments", strings.Join(annotatedMachineDeployments, ","))
}
