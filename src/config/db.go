package config

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"fyp-api-gateway/src/utils"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	uuid "github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"golang.org/x/crypto/argon2"
)

type Database struct {
	Conn *sql.DB
}

type Server struct {
	DB *Database
}

type LoginInfo struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

func NewDatabase(dsn string) (*Database, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		slog.Error("error opening database connection:", "error", err)
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}

	return &Database{Conn: db}, nil
}

func (d *Database) StartDB(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		slog.Error("error reading database file", "error", err)
		return err
	}

	if _, err = d.Conn.Exec(string(content)); err != nil {
		slog.Error("error executing database statement", "error", err)
		return err
	}

	return nil
}

func hashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)

	encoded := fmt.Sprintf("$argon2id$v=19$m=65536,t=1,p=4$%s$%s",
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)
	return encoded, nil
}

func (s *Server) Signup(w http.ResponseWriter, r *http.Request) {
	slog.Info("attempting to sign up new user...")
	loginInfo := &LoginInfo{}

	if r.Method != http.MethodPost {
		slog.Error("invalid method", "method", r.Method)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&loginInfo); err != nil {
		slog.Error("error decoding loginInfo", "error", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var id string
	err := s.DB.Conn.QueryRow(
		"SELECT id FROM users WHERE username = $1",
		loginInfo.Name,
	).Scan(&id)

	if err == nil || id != "" {
		slog.Error("error querying user", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	hashed_password, err := hashPassword(loginInfo.Password)
	if err != nil {
		slog.Error("error hashing password", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	_, err = s.DB.Conn.Exec(`
		INSERT INTO users (username, password, config_yaml)
		VALUES ($1, $2, $3);`,
		loginInfo.Name, string(hashed_password), utils.DefaultConfigContent,
	)
	if err != nil {
		slog.Error("error inserting user", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	err = InitUserNGINX(loginInfo.Name)
	if err != nil {
		slog.Error("error initializing user", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

/* Code for this method supplied by https://www.alexedwards.net/blog/how-to-hash-and-verify-passwords-with-argon2-in-go */
func verifyPassword(password, hash string) (bool, error) {
	var m, t, p uint32
	var saltB64, hashB64 string
	_, err := fmt.Sscanf(hash, "$argon2id$v=19$m=%d,t=%d,p=%d$%s",
		&m, &t, &p, &saltB64)
	if err != nil {
		return false, err
	}

	parts := strings.Split(hash, "$")
	if len(parts) != 6 {
		return false, fmt.Errorf("invalid hash format")
	}
	saltB64 = parts[4]
	hashB64 = parts[5]

	salt, err := base64.RawStdEncoding.DecodeString(saltB64)
	if err != nil {
		return false, err
	}
	storedHash, err := base64.RawStdEncoding.DecodeString(hashB64)
	if err != nil {
		return false, err
	}

	keyLen := uint32(len(storedHash))
	computed := argon2.IDKey([]byte(password), salt, t, m, uint8(p), keyLen)

	if subtle.ConstantTimeCompare(computed, storedHash) != 1 {
		return false, nil
	}
	return true, nil
}

/*
Receive the login info from the management plane and decode it
Check the password used is the same as the one in the database, also check the username is there
If the user has no session, create one and send it back, otherwise return the existing session
*/
func (s *Server) VerifyLoginInfo(w http.ResponseWriter, r *http.Request) {
	slog.Info("validating login information")
	loginInfo := &LoginInfo{}

	if r.Method != http.MethodPost {
		slog.Error("invalid method", "method", r.Method)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&loginInfo); err != nil {
		slog.Error("error decoding loginInfo", "error", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var storedHash string
	err := s.DB.Conn.QueryRow(
		"SELECT password FROM users WHERE username = $1",
		loginInfo.Name,
	).Scan(&storedHash)

	if errors.Is(err, sql.ErrNoRows) {
		slog.Error("user not found", "username", loginInfo.Name)
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	isCorrect, err := verifyPassword(loginInfo.Password, storedHash)
	if err != nil || !isCorrect {
		slog.Error("error verifying loginInfo", "error", err)
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	// check is the user has a session already
	var sessionId string
	sessionId, isSession := s.sessionExists(loginInfo.Name)

	if !isSession {
		sessionId, err = s.createSession(loginInfo.Name)
		if err != nil {
			slog.Error("error creating session", "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sessionId": sessionId,
	})
}

func (s *Server) sessionExists(username string) (string, bool) {
	var sessionId string
	err := s.DB.Conn.QueryRow(
		"SELECT id FROM sessions WHERE username=$1",
		username,
	).Scan(&sessionId)

	if err != nil {
		return "", false
	}
	return sessionId, true
}

func (s *Server) createSession(name string) (string, error) {
	sessionId := uuid.New().String()
	expires := time.Now().Add(24 * time.Hour)

	_, err := s.DB.Conn.Exec(
		"INSERT INTO sessions(id, username, expires) VALUES ($1, $2, $3)",
		sessionId, name, expires,
	)

	if err != nil {
		slog.Error("error creating session", "error", err)
		return "", err
	}

	return sessionId, nil
}

func (s *Server) ValidateSession(w http.ResponseWriter, r *http.Request) {
	slog.Info("validating user session")
	sessionId := r.Header.Get("X-Session-ID")

	var username string
	err := s.DB.Conn.QueryRow(
		"SELECT username FROM sessions WHERE id=$1 AND expires > NOW()",
		sessionId,
	).Scan(&username)

	if err != nil {
		slog.Error("error querying session", "error", err)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) UserConfig(w http.ResponseWriter, r *http.Request) {
	slog.Info("received request for user config")

	cookie, err := r.Cookie("session")
	if err != nil {
		slog.Error("failed getting session id", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	sessionId := cookie.Value

	var gatewayCfg string
	err = s.DB.Conn.QueryRow(`
		SELECT u.config_yaml
		FROM users AS u 
		JOIN sessions AS s ON u.username = s.username 
		WHERE s.id = $1 AND s.expires > NOW()`,
		sessionId).Scan(&gatewayCfg)

	if err != nil {
		slog.Error("error querying session", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/yaml")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(gatewayCfg))
	if err != nil {
		slog.Error("error writing response", "error", err)
		return
	}
}

func RetrieveUserBySessionId(sessionId string) string {
	dsn := os.Getenv("DATABASE_URL")

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		slog.Error("error opening database connection:", "error", err)
		return ""
	}

	var username string
	err = db.QueryRow(`
		SELECT username
		FROM sessions
		WHERE id = $1`,
		sessionId).Scan(&username)

	return username
}

func RetrieveUserConfig(username string) string {
	dsn := os.Getenv("DATABASE_URL")

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		slog.Error("error opening database connection:", "error", err)
		return ""
	}

	var gatewayCfg string
	err = db.QueryRow(`
		SELECT config_yaml
		FROM users 
		WHERE username = $1`,
		username,
	).Scan(&gatewayCfg)

	return gatewayCfg
}

func InsertNewConfig(sessionId, gatewayCfg string) error {
	dsn := os.Getenv("DATABASE_URL")

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		UPDATE users AS u
		SET config_yaml = $1
		FROM sessions AS s
		WHERE s.id = $2
		AND s.expires > NOW()
		AND u.username = s.username`,
		gatewayCfg, sessionId,
	)
	slog.Info("inserted new config to database")

	if err != nil {
		return err
	}

	return nil
}
