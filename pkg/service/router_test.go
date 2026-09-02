package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRouter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		wantAllow  string
		wantCalled string
	}{
		{
			name:       "healthcheck",
			method:     http.MethodGet,
			path:       "/healthcheck",
			wantStatus: http.StatusOK,
			wantCalled: "healthcheck",
		},
		{
			name:       "healthcheck answers HEAD",
			method:     http.MethodHead,
			path:       "/healthcheck",
			wantStatus: http.StatusOK,
			wantCalled: "healthcheck",
		},
		{
			name:       "healthcheck rejects POST",
			method:     http.MethodPost,
			path:       "/healthcheck",
			wantStatus: http.StatusMethodNotAllowed,
			wantAllow:  "GET, HEAD",
		},
		{
			name:       "authenticate",
			method:     http.MethodPost,
			path:       "/v1/authenticate",
			wantStatus: http.StatusOK,
			wantCalled: "authenticate",
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

			var called string
			router := NewRouter(Handlers{
				Healthcheck:  recordHandler("healthcheck", &called),
				Authenticate: recordHandler("authenticate", &called),
			})

			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest(test.method, test.path, nil))

			assert.Equal(t, test.wantStatus, w.Code)
			assert.Equal(t, test.wantCalled, called)
			if test.wantAllow != "" {
				assert.Equal(t, test.wantAllow, w.Header().Get("Allow"))
			}
		})
	}
}

// recordHandler returns a handler that records its name when the router
// dispatches to it.
func recordHandler(name string, called *string) http.Handler {
	return http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		*called = name
	})
}
