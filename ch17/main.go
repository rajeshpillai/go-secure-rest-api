package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"golang.org/x/time/rate"
)

type app struct {
	store                *Store
	users                *Users
	jwtKey               []byte
	corsOrigin           string
	sessions             *Sessions
	keys                 *APIKeys
	loginLimit, apiLimit *limiter
}

func (a *app) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /books", a.auth("books:read", a.listBooks))
	mux.HandleFunc("GET /books/{id}", a.auth("books:read", a.getBook))
	mux.HandleFunc("POST /books", a.auth("books:write", a.createBook))
	mux.HandleFunc("PUT /books/{id}", a.auth("books:write", a.updateBook))
	mux.HandleFunc("DELETE /books/{id}", a.auth("books:write", a.deleteBook))
	mux.HandleFunc("GET /me", a.auth("books:read", a.me))
	mux.HandleFunc("POST /keys", a.auth("keys:manage", a.createKey))
	mux.HandleFunc("GET /keys", a.auth("keys:manage", a.listKeys))
	mux.HandleFunc("DELETE /keys/{id}", a.auth("keys:manage", a.revokeKey))
	mux.HandleFunc("POST /register", a.register)
	mux.Handle("POST /login", a.loginLimit.wrap(http.HandlerFunc(a.login)))
	mux.HandleFunc("POST /refresh", a.refresh)
	mux.HandleFunc("POST /logout", a.logout)
	return logRequests(recoverPanics(cors(a.corsOrigin, a.apiLimit.wrap(mux))))
}

func newApp(cfg config) *app {
	a := &app{store: NewStore(), users: &Users{file: cfg.usersFile},
		jwtKey: cfg.jwtKey, corsOrigin: cfg.corsOrigin,
		sessions: &Sessions{byHash: map[[32]byte]session{}},
		keys:     &APIKeys{},
		// Five login tries, then one more every 12 seconds. Everything else: 20 a second.
		loginLimit: newLimiter(rate.Every(12*time.Second), 5),
		apiLimit:   newLimiter(20, 40),
	}
	return a
}

func main() {
	cfg, err := configFromEnv()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	srv := &http.Server{
		Addr:              ":8443",
		Handler:           newApp(cfg).routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	slog.Info("listening", "addr", "https://localhost:8443")
	err = srv.ListenAndServeTLS("cert.pem", "key.pem")
	slog.Error("server stopped", "err", err)
	os.Exit(1)
}
