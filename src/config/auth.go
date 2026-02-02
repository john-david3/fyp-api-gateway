package config

import (
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

var PUBLIC_KEY interface{}

func keyFunc(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
		return nil, jwt.ErrTokenUnverifiable
	}
	return PUBLIC_KEY, nil
}

func validateJWT(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	token := strings.TrimPrefix(auth, "Bearer ")

	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, claims, keyFunc)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func loadPublicKey(path string) (interface{}, error) {
	keyBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return jwt.ParseRSAPublicKeyFromPEM(keyBytes)
}

func StartAuth() {
	var err error
	PUBLIC_KEY, err = loadPublicKey("/etc/keys/public.pem")
	if err != nil {
		slog.Error("failed to load public key", "error", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/validate", validateJWT)
	slog.Info("server running on port 8088")
	err = http.ListenAndServe(":8088", mux)
	if err != nil {
		slog.Error("error starting validation service", "error", err)
	}
}
