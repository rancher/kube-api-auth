package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/rancher/kube-api-auth/pkg/service/handlers"
)

// TestRouteContext covers routing only, so the cases never reach a handler body
// that would touch the cluster clients a zero-value KubeAPIHandlers lacks.
func TestRouteContext(t *testing.T) {
	t.Parallel()

	router := RouteContext(&handlers.KubeAPIHandlers{})

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		wantAllow  string
	}{
		{
			name:       "healthcheck",
			method:     http.MethodGet,
			path:       "/healthcheck",
			wantStatus: http.StatusOK,
		},
		{
			name:       "healthcheck answers HEAD",
			method:     http.MethodHead,
			path:       "/healthcheck",
			wantStatus: http.StatusOK,
		},
		{
			name:       "healthcheck rejects POST",
			method:     http.MethodPost,
			path:       "/healthcheck",
			wantStatus: http.StatusMethodNotAllowed,
			wantAllow:  "GET, HEAD",
		},
		{
			name:       "authenticate rejects GET",
			method:     http.MethodGet,
			path:       "/v1/authenticate",
			wantStatus: http.StatusMethodNotAllowed,
			wantAllow:  "POST",
		},
		{
			name:       "trailing slash is not routed",
			method:     http.MethodGet,
			path:       "/healthcheck/",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "unknown path",
			method:     http.MethodGet,
			path:       "/nope",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest(test.method, test.path, nil))

			assert.Equal(t, test.wantStatus, w.Code)
			if test.wantAllow != "" {
				assert.Equal(t, test.wantAllow, w.Header().Get("Allow"))
			}
		})
	}
}
