package service

import (
	"context"
	"errors"
	"fmt"
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

const (
	controllerWorkers = 5
	cacheSyncTimeout  = 30 * time.Second

	readHeaderTimeout = 10 * time.Second
	readTimeout       = 30 * time.Second
	writeTimeout      = 30 * time.Second
	idleTimeout       = 120 * time.Second
	shutdownTimeout   = 15 * time.Second
)

func Serve(ctx context.Context, listen, namespace, kubeConfig string) error {
	log.Info("Starting Rancher Kube-API-Auth service on ", listen)

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

	kubeAPIHandlers := handlers.NewKubeAPIHandlers(namespace, clusterAPI, coreAPI)

	log.Info("Starting controllers and waiting for caches to sync")
	if err := wranglerCtx.ControllerFactory.Start(ctx, controllerWorkers); err != nil {
		return fmt.Errorf("starting controllers: %w", err)
	}

	syncCtx, cancel := context.WithTimeout(ctx, cacheSyncTimeout)
	defer cancel()
	for gvk, synced := range wranglerCtx.ControllerFactory.SharedCacheFactory().WaitForCacheSync(syncCtx) {
		if !synced {
			return fmt.Errorf("informer for %s did not sync", gvk)
		}
	}

	srv := &http.Server{
		Addr:              listen,
		Handler:           RouteContext(kubeAPIHandlers),
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		log.Info("Shutdown signal received, draining connections")
	}

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancelShutdown()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("http server shutdown: %w", err)
	}
	return nil
}

func RouteContext(kubeAPIHandlers *handlers.KubeAPIHandlers) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	router.Methods("GET").Path("/healthcheck").Handler(handlers.HealthcheckHandler())
	router.Methods("POST").Path("/v1/authenticate").Handler(kubeAPIHandlers.V1AuthenticateHandler())

	return router
}
