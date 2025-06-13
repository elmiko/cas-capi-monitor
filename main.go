package main

import (
	"os"

	"github.com/elmiko/cas-capi-monitor/controllers"
	capiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

func main() {
	logf.SetLogger(zap.New())

	log := logf.Log.WithName("cas-capi-monitor")

	mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{})
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
		For(&capiv1beta1.Machine{},
			builder.WithPredicates(predicate.AnnotationChangedPredicate{}),
			builder.OnlyMetadata).
		Complete(&controllers.MachineReconciler{
			Client: mgr.GetClient(),
		})
	if err != nil {
		log.Error(err, "could not create controller")
		os.Exit(1)
	}

	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "could not start manager")
		os.Exit(1)
	}
}
