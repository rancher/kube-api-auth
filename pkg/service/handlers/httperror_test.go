package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReturnHTTPError(t *testing.T) {
	t.Parallel()

	t.Run("sets status code and content type", func(t *testing.T) {
		t.Parallel()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/v1/authenticate", nil)

		ReturnHTTPError(w, r, http.StatusUnauthorized, "invalid credentials")

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("bad request", func(t *testing.T) {
		t.Parallel()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/v1/authenticate", nil)

		ReturnHTTPError(w, r, http.StatusBadRequest, "bad request")

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
