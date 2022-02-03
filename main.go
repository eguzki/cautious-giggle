/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"net/http"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	gorillahandlers "github.com/gorilla/handlers"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	gigglekuadrantiov1alpha1 "github.com/eguzki/cautious-giggle/api/v1alpha1"
	"github.com/eguzki/cautious-giggle/controllers"
	"github.com/eguzki/cautious-giggle/pkg/http/handlers"
	"github.com/eguzki/cautious-giggle/pkg/http/html"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(gigglekuadrantiov1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "5958d8ab.giggle.kuadrant.io",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.ApiReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Api")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	startHTTPService()

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func startHTTPService() {

	k8sClient, err := client.New(ctrl.GetConfigOrDie(), client.Options{Scheme: scheme})
	if err != nil {
		panic(err)
	}

	dashboardHandler := &handlers.DashboardHandler{K8sClient: k8sClient}
	serviceDiscoveryHandler := &handlers.ServiceDiscoveryHandler{K8sClient: k8sClient}
	newAPIHandler := &handlers.NewApiHandler{K8sClient: k8sClient}
	createNewAPIHandler := &handlers.CreateNewAPIHandler{K8sClient: k8sClient}
	aPIHandler := &handlers.APIHandler{K8sClient: k8sClient}
	gatewaysHandler := &handlers.GatewaysHandler{K8sClient: k8sClient}
	createGatewaysHandler := &handlers.CreateGatewaysHandler{K8sClient: k8sClient}
	updateAPIGatewayHandler := &handlers.UpdateAPIGatewayHandler{K8sClient: k8sClient}

	mux := http.NewServeMux()
	mux.Handle("/", http.RedirectHandler("/login.html", 301))
	mux.Handle("/login.html", http.FileServer(http.FS(html.LoginContent)))
	mux.Handle("/dashboard", dashboardHandler)
	mux.Handle("/servicediscovery", serviceDiscoveryHandler)
	mux.Handle("/newapi", newAPIHandler)
	mux.Handle("/createnewapi", createNewAPIHandler)
	mux.Handle("/api", aPIHandler)
	mux.Handle("/gateways", gatewaysHandler)
	mux.Handle("/newgateway.html", http.FileServer(http.FS(html.NewGatewayContent)))
	mux.Handle("/creategateway", createGatewaysHandler)
	mux.Handle("/updateapigateway", updateAPIGatewayHandler)

	loggerRoute := gorillahandlers.LoggingHandler(os.Stdout, mux)

	setupLog.Info("starting HTTP service", "port", 8082)

	go func() {
		err := http.ListenAndServe(":8082", loggerRoute)
		if err != nil {
			setupLog.Error(err, "failed to start server")
			os.Exit(1)
		}
	}()
}
