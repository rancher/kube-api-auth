package service

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rancher/kube-api-auth/pkg/service/handlers"
	"github.com/rancher/norman/pkg/kwrapper/k8s"
	clusterv3 "github.com/rancher/rancher/pkg/generated/norman/cluster.cattle.io/v3"
	corev1 "github.com/rancher/rancher/pkg/generated/norman/core/v1"
	"github.com/rancher/rancher/pkg/wrangler"
	log "github.com/sirupsen/logrus"
)

func Serve(listen, namespace, kubeConfig string) error {
	log.Info("Starting Rancher Kube-API-Auth service on ", listen)

	ctx := context.Background()

	_, clientConfig, err := k8s.GetConfig(ctx, "auto", kubeConfig)
	if err != nil {
		return err
	}

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return err
	}

	wranglerCtx, err := wrangler.NewContext(ctx, clientConfig, restConfig)
	if err != nil {
		return err
	}
	clusterAPI := clusterv3.NewFromControllerFactory(wranglerCtx.ControllerFactory)
	coreAPI := corev1.NewFromControllerFactory(wranglerCtx.ControllerFactory)

	// API framework routes
	kubeAPIHandlers := handlers.NewKubeAPIHandlers(namespace, clusterAPI, coreAPI)
	router := RouteContext(kubeAPIHandlers)

	go func() {
		for {
			if err := wranglerCtx.ControllerFactory.Start(ctx, 5); err != nil {
				log.Error(err)
				time.Sleep(2 * time.Second)
			} else {
				break
			}
		}
	}()

	return http.ListenAndServe(listen, router)
}

func RouteContext(kubeAPIHandlers *handlers.KubeAPIHandlers) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	// Healthcheck endpoint
	router.Methods("GET").Path("/healthcheck").Handler(handlers.HealthcheckHandler())
	// V1 Authenticate endpoint
	router.Methods("POST").Path("/v1/authenticate").Handler(kubeAPIHandlers.V1AuthenticateHandler())

	return router
}
