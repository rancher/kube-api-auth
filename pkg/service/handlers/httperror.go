package handlers

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

//ReturnHTTPError handles sending out CatalogError response
func ReturnHTTPError(w http.ResponseWriter, _ *http.Request, httpStatus int, errorMessage string) {
	log.Error(errorMessage)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
}
