package service

import "net/http"

// Handlers are the handlers NewRouter mounts, one per route.
type Handlers struct {
	Healthcheck  http.Handler
	Authenticate http.Handler
}

func NewRouter(h Handlers) *http.ServeMux {
	router := http.NewServeMux()
	router.Handle("GET /healthcheck", h.Healthcheck)
	router.Handle("POST /v1/authenticate", h.Authenticate)

	return router
}
