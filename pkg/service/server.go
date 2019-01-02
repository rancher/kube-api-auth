package service

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rancher/kube-api-auth/pkg/service/controllers"
	"github.com/rancher/kube-api-auth/pkg/service/handlers"
	"github.com/rancher/types/config"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func Serve(listen, namespace, kubeconfig string) error {
	log.Info("Starting Rancher Kube-API-Auth service on ", listen)

	var (
		conf *rest.Config
		err  error
	)
	if kubeconfig != "" {
		conf, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		conf, err = rest.InClusterConfig()
	}
	if err != nil {
		return err
	}

	ctx := context.Background()

	apiContext, err := config.NewUserOnlyContext(*conf)
	if err != nil {
		return err
	}

	go func() {
		for {
			err := controllers.Start(ctx, namespace, apiContext)
			if err != nil {
				log.Error(err)
				time.Sleep(2 * time.Second)
			} else {
				break
			}
		}
	}()

	router := RouteContext(mux.NewRouter().StrictSlash(true), namespace, apiContext)
	return http.ListenAndServe(listen, router)
}

func RouteContext(router *mux.Router, namespace string, apiContext *config.UserOnlyContext) *mux.Router {
	// API framework routes
	kubeAPIHandlers := handlers.NewKubeAPIHandlers(namespace, apiContext)
	// Healthcheck endpoint
	router.Methods("GET").Path("/healthcheck").Handler(handlers.HealthcheckHandler())
	// V1 Authenticate endpoint
	router.Methods("POST").Path("/v1/authenticate").Handler(kubeAPIHandlers.V1AuthenticateHandler())

	return router
}
