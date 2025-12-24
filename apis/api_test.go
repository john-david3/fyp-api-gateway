package apis

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateAPI_main(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	mainRoute(w, req)
	res := w.Result()
	require.Equal(t, http.StatusOK, res.StatusCode)
}

func TestCreateAPI_products(t *testing.T) {
	var body map[string]string

	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	w := httptest.NewRecorder()

	products(w, req)
	res := w.Result()
	require.Equal(t, http.StatusOK, res.StatusCode)

	err := json.NewDecoder(res.Body).Decode(&body)
	require.NoError(t, err)
	require.Equal(t, "products", body["service"])
	require.Equal(t, "/products", body["path"])
}

func TestCreateAPI_health(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	health(w, req)
	res := w.Result()
	require.Equal(t, http.StatusOK, res.StatusCode)
}
