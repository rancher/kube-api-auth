package service

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rancher/kube-api-auth/pkg/clients"
	"github.com/rancher/kube-api-auth/pkg/service/handlers"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/clientcmd"
)

func Serve(listen, namespace, kubeConfig string) error {
	log.Info("Starting Rancher Kube-API-Auth service on ", listen)

	ctx := context.Background()

	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		return err
	}

	c, err := clients.New(ctx, restConfig, namespace)
	if err != nil {
		return err
	}

	kubeAPIHandlers := handlers.NewKubeAPIHandlers(namespace, c)
	router := RouteContext(kubeAPIHandlers)

	go func() {
		for {
			if err := c.Start(ctx); err != nil {
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
	router.Methods("GET").Path("/healthcheck").Handler(handlers.HealthcheckHandler())
	router.Methods("POST").Path("/v1/authenticate").Handler(kubeAPIHandlers.V1AuthenticateHandler())

	return router
}
