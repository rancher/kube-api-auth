package handlers

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

func HealthcheckHandler() http.HandlerFunc {
	return http.HandlerFunc(healthcheck)
}

func healthcheck(w http.ResponseWriter, r *http.Request) {
	log.Info("healthcheck")
}
