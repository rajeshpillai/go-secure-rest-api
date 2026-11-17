package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"
)

type app struct {
	store    *Store
	users    *Users
	jwtKey   []byte
	sessions *Sessions
	keys     *APIKeys
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
	mux.HandleFunc("POST /login", a.login)
	mux.HandleFunc("POST /refresh", a.refresh)
	mux.HandleFunc("POST /logout", a.logout)
	return logRequests(recoverPanics(mux))
}

func newApp(jwtKey []byte) *app {
	a := &app{store: NewStore(), users: &Users{file: "users.json"}, jwtKey: jwtKey,
		sessions: &Sessions{byHash: map[[32]byte]session{}},
		keys:     &APIKeys{}}
	a.store.Add(Book{Title: "The Go Programming Language", Author: "Alan Donovan, Brian Kernighan"})
	a.store.Add(Book{Title: "Clean Code", Author: "Robert C. Martin"})
	a.store.Add(Book{Title: "The Pragmatic Programmer", Author: "Andrew Hunt, David Thomas"})
	return a
}

func main() {
	key := os.Getenv("JWT_SECRET")
	if len(key) < 32 {
		slog.Error("JWT_SECRET must be set, at least 32 bytes")
		os.Exit(1)
	}
	srv := &http.Server{
		Addr:              ":8080",
		Handler:           newApp([]byte(key)).routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	slog.Info("listening", "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
