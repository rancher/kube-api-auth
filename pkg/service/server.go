package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rancher/kube-api-auth/pkg/clients"
	"github.com/rancher/kube-api-auth/pkg/service/handlers"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	readHeaderTimeout = 10 * time.Second
	readTimeout       = 30 * time.Second
	writeTimeout      = 30 * time.Second
	idleTimeout       = 120 * time.Second
	shutdownTimeout   = 15 * time.Second
)

func Serve(ctx context.Context, listen, namespace, kubeConfig string) error {
	log.Info("Starting Rancher Kube-API-Auth service on ", listen)

	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		return err
	}

	c, err := clients.New(ctx, restConfig, namespace)
	if err != nil {
		return err
	}

	log.Info("Starting informers and waiting for caches to sync")
	if err := c.Start(ctx); err != nil {
		return fmt.Errorf("starting cluster clients: %w", err)
	}

	kubeAPIHandlers := handlers.NewKubeAPIHandlers(namespace, c)
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

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
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
