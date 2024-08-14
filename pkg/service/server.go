package service

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rancher/kube-api-auth/pkg/service/controllers"
	"github.com/rancher/kube-api-auth/pkg/service/handlers"
	"github.com/rancher/norman/pkg/kwrapper/k8s"
	"github.com/rancher/rancher/pkg/types/config"
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

	apiContext, err := config.NewUserOnlyContext(wranglerCtx)
	if err != nil {
		return err
	}

	router := RouteContext(mux.NewRouter().StrictSlash(true), namespace, apiContext, wranglerCtx)

	go func() {
		for {
			if err := controllers.Start(ctx, apiContext); err != nil {
				log.Error(err)
				time.Sleep(2 * time.Second)
			} else {
				break
			}
		}
	}()

	return http.ListenAndServe(listen, router)
}

func RouteContext(router *mux.Router, namespace string, apiContext *config.UserOnlyContext, wranglerCtx *wrangler.Context) *mux.Router {
	// API framework routes
	kubeAPIHandlers := handlers.NewKubeAPIHandlers(namespace, apiContext, wranglerCtx)
	// Healthcheck endpoint
	router.Methods("GET").Path("/healthcheck").Handler(handlers.HealthcheckHandler())
	// V1 Authenticate endpoint
	router.Methods("POST").Path("/v1/authenticate").Handler(kubeAPIHandlers.V1AuthenticateHandler())

	return router
}
