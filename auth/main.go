package main

import (
	"crypto/rsa"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var privateKey *rsa.PrivateKey

func loadPrivateKey(path string) *rsa.PrivateKey {
	keyBytes, err := os.ReadFile(path)
	if err != nil {
		slog.Error("Error loading private key", "error", err)
	}

	key, err := jwt.ParseRSAPrivateKeyFromPEM(keyBytes)
	if err != nil {
		slog.Error("Error loading private key", "error", err)
	}
	return key
}

func issueJWT(w http.ResponseWriter, r *http.Request) {
	// This is a mock login; in prod, validate username/password
	claims := jwt.MapClaims{
		"sub":   "123",
		"roles": []string{"admin"},
		"exp":   time.Now().Add(time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := map[string]string{"token": signedToken}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	privateKey = loadPrivateKey("/etc/keys/private.pem")

	http.HandleFunc("/login", issueJWT)
	slog.Info("auth-issuer running on :8089")
	err := http.ListenAndServe(":8089", nil)
	if err != nil {
		slog.Error("Error starting HTTP server", "error", err)
	}
}
