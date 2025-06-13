package main

import (
	"context"
	"flag"
	"os"

	"github.com/elmiko/cas-capi-monitor/controllers"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	capiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

func main() {
	logf.SetLogger(zap.New())
	log := logf.Log.WithName("cas-capi-monitor")

	var nodeKCFile string

	flag.StringVar(&nodeKCFile, "nk", "", "path to kubeconfig for the node resources")
	flag.Parse()

	capimgr, err := manager.New(config.GetConfigOrDie(), manager.Options{})
	if err != nil {
		log.Error(err, "could not create cluster api resource manager")
		os.Exit(1)
	}

	err = capiv1beta1.AddToScheme(capimgr.GetScheme())
	if err != nil {
		log.Error(err, "unable to add scheme")
		os.Exit(1)
	}

	err = builder.
		ControllerManagedBy(capimgr).
		For(&capiv1beta1.Machine{}).
		Complete(&controllers.MachineReconciler{
			Client: capimgr.GetClient(),
		})
	if err != nil {
		log.Error(err, "could not create machine controller")
		os.Exit(1)
	}

	nodemgr, err := manager.New(getNodeConfigOrDie(nodeKCFile, log), manager.Options{})
	if err != nil {
		log.Error(err, "could not create node resource manager")
		os.Exit(1)
	}
	err = builder.
		ControllerManagedBy(nodemgr).
		For(&corev1.Node{}).
		Complete(&controllers.NodeReconciler{
			Client: nodemgr.GetClient(),
		})
	if err != nil {
		log.Error(err, "could not create node controller")
		os.Exit(1)
	}

	ctx := signals.SetupSignalHandler()

	go startCapiMgr(ctx, capimgr, log)

	if err := nodemgr.Start(ctx); err != nil {
		log.Error(err, "could not start node resource manager")
		os.Exit(1)
	}
}

func startCapiMgr(ctx context.Context, capimgr manager.Manager, log logr.Logger) {
	if err := capimgr.Start(ctx); err != nil {
		log.Error(err, "could not start cluster api resource manager")
		os.Exit(1)
	}
}

func getNodeConfigOrDie(filename string, log logr.Logger) *rest.Config {
	if len(filename) == 0 {
		log.Error(nil, "no kubeconfig for node resources specified")
		os.Exit(1)
	}
	nodeConfig, err := clientcmd.BuildConfigFromFlags("", filename)
	if err != nil {
		log.Error(err, "unable to build kubeconfig for the node resources")
		os.Exit(1)
	}
	return nodeConfig
}
