package main

import (
	"flag"
	"os"

	"github.com/elmiko/cas-capi-monitor/controllers"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	capiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

func main() {
	logf.SetLogger(zap.New())
	log := logf.Log.WithName("cas-capi-monitor")

	var capiKCFile string
	var nodeKCFile string

	flag.StringVar(&capiKCFile, "ck", "", "path to kubeconfig for the cluster api resources")
	flag.StringVar(&nodeKCFile, "nk", "", "path to kubeconfig for the node resources")
	flag.Parse()

	mgr, err := manager.New(getCapiConfigOrDie(capiKCFile, log), manager.Options{})
	if err != nil {
		log.Error(err, "could not create manager")
		os.Exit(1)
	}

	err = capiv1beta1.AddToScheme(mgr.GetScheme())
	if err != nil {
		log.Error(err, "unable to add scheme")
		os.Exit(1)
	}

	err = builder.
		ControllerManagedBy(mgr).
		For(&capiv1beta1.Machine{}).
		Complete(&controllers.MachineReconciler{
			Client: mgr.GetClient(),
		})
	if err != nil {
		log.Error(err, "could not create machine controller")
		os.Exit(1)
	}

	err = builder.
		ControllerManagedBy(mgr).
		For(&corev1.Node{}).
		Complete(&controllers.NodeReconciler{
			Client: getNodeClientOrDie(nodeKCFile, log),
		})
	if err != nil {
		log.Error(err, "could not create node controller")
		os.Exit(1)
	}

	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "could not start manager")
		os.Exit(1)
	}
}

func getCapiConfigOrDie(filename string, log logr.Logger) *rest.Config {
	if len(filename) == 0 {
		return config.GetConfigOrDie()
	} else {
		capiConfig, err := clientcmd.BuildConfigFromFlags("", filename)
		if err != nil {
			log.Error(err, "unable to build kubeconfig for the cluster api resources")
			os.Exit(1)
		}
		return capiConfig
	}
}

func getNodeClientOrDie(filename string, log logr.Logger) client.Client {
	if len(filename) == 0 {
		log.Error(nil, "no kubeconfig for node resources specified")
		os.Exit(1)
	}
	nodeConfig, err := clientcmd.BuildConfigFromFlags("", filename)
	if err != nil {
		log.Error(err, "unable to build kubeconfig for the node resources")
		os.Exit(1)
	}
	nodeClient, err := client.New(nodeConfig, client.Options{})
	if err != nil {
		log.Error(err, "unable to build client for the node resources")
		os.Exit(1)
	}
	return nodeClient
}
