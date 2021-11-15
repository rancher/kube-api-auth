package handlers

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

func HealthcheckHandler() http.HandlerFunc {
	return healthcheck
}

func healthcheck(_ http.ResponseWriter, _ *http.Request) {
	log.Info("healthcheck")
}
