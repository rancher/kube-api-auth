package handlers

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

func Healthcheck(_ http.ResponseWriter, _ *http.Request) {
	log.Info("healthcheck")
}
