package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProducts(t *testing.T) {
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

func TestOrders(t *testing.T) {
	var body map[string]string

	req := httptest.NewRequest(http.MethodGet, "/orders", nil)
	w := httptest.NewRecorder()

	orders(w, req)
	res := w.Result()
	require.Equal(t, http.StatusOK, res.StatusCode)

	err := json.NewDecoder(res.Body).Decode(&body)
	require.NoError(t, err)
	require.Equal(t, "orders", body["service"])
	require.Equal(t, "/orders", body["path"])
}

func TestHealth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	health(w, req)
	res := w.Result()
	require.Equal(t, http.StatusOK, res.StatusCode)
}
