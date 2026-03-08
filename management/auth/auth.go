package auth

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

type UserInfo struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

/*
Receive the login details from auth.js
Decode them and format them as `UserInfo` to send to the control plane
Control plane validates the user is in the database and returns a cookie
*/
func Login(w http.ResponseWriter, r *http.Request) {
	slog.Info("received login request")
	loginInfo := UserInfo{}

	if r.Method != "POST" {
		slog.Error("method not allowed", "method", r.Method)
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&loginInfo)
	if err != nil {
		slog.Error("error decoding login info", "error", err)
		return
	}

	body, err := json.Marshal(loginInfo)
	if err != nil {
		slog.Error("error encoding login info", "error", err)
		return
	}

	res, err := http.Post(
		"http://control-plane:10000/verify-login",
		"application/json",
		bytes.NewBuffer(body),
	)

	if res.StatusCode != http.StatusOK {
		slog.Error("error verifying login", "error", res.Status)
	} else {
		slog.Info("verified login")
	}

	var resp struct {
		SessionId string `json:"sessionId"`
	}
	err = json.NewDecoder(res.Body).Decode(&resp)
	if err != nil {
		slog.Error("error decoding response", "error", err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    resp.SessionId,
		HttpOnly: true,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
	})

	w.WriteHeader(http.StatusOK)
}

func Signup(w http.ResponseWriter, r *http.Request) {
	slog.Info("received signup request")
	loginInfo := UserInfo{}

	if r.Method != "POST" {
		slog.Error("method not allowed", "method", r.Method)
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&loginInfo)
	if err != nil {
		slog.Error("error decoding signup info", "error", err)
		return
	}

	body, err := json.Marshal(loginInfo)
	if err != nil {
		slog.Error("error encoding signup info", "error", err)
		return
	}

	res, err := http.Post(
		"http://control-plane:10000/verify-signup",
		"application/json",
		bytes.NewBuffer(body),
	)

	if res.StatusCode != http.StatusOK {
		slog.Error("error verifying signup", "error", res.Status)
	} else {
		slog.Info("verified signup")
	}

	w.WriteHeader(http.StatusOK)
}

func RequireSession(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		cookie, err := r.Cookie("session")
		if err != nil {
			slog.Error("error getting cookie", "error", err)
			http.Redirect(w, r, "/auth", http.StatusSeeOther)
			return
		}

		req, _ := http.NewRequest(
			"GET",
			"http://control-plane:10000/validate-session",
			nil,
		)
		req.Header.Set("X-Session-ID", cookie.Value)

		res, err := http.DefaultClient.Do(req)
		if err != nil || res.StatusCode != http.StatusOK {
			slog.Error("error validating session", "error", err, "code", res.StatusCode)
			http.Redirect(w, r, "/auth", http.StatusSeeOther)
			return
		}

		next(w, r)
	}
}
