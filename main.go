package main

import (
	"context"
	"flag"
	"os"

	"github.com/elmiko/cas-capi-monitor/controllers"
	"github.com/go-logr/logr"
	"go.uber.org/zap/zapcore"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
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
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

func main() {
	var debugMode bool
	var nodeKCFile string

	flag.BoolVar(&debugMode, "debug", false, "turn on extra debug logging")
	flag.StringVar(&nodeKCFile, "nk", "", "path to kubeconfig for the node resources")
	flag.Parse()

	if debugMode {
		logf.SetLogger(zap.New(zap.Level(zapcore.Level(-5))))
	} else {
		logf.SetLogger(zap.New())
	}
	log := logf.Log.WithName("cas-capi-monitor")

	capimgrScheme := scheme.Scheme
	err := capiv1beta1.AddToScheme(capimgrScheme)
	if err != nil {
		log.Error(err, "unable to add scheme")
		os.Exit(1)
	}

	capimgrOptions := manager.Options{
		Scheme: capimgrScheme,
		Client: client.Options{
			Cache: &client.CacheOptions{
				DisableFor: []client.Object{
					&capiv1beta1.Machine{},
				},
			},
		},
		Metrics: server.Options{BindAddress: "0"},
	}
	capimgr, err := manager.New(config.GetConfigOrDie(), capimgrOptions)
	if err != nil {
		log.Error(err, "could not create cluster api resource manager")
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

	nodemgrOptions := manager.Options{
		Metrics: server.Options{BindAddress: "0"},
		Client: client.Options{
			Cache: &client.CacheOptions{
				DisableFor: []client.Object{
					&corev1.Node{},
				},
			},
		},
	}
	nodemgr, err := manager.New(getNodeConfigOrDie(nodeKCFile, log), nodemgrOptions)
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
