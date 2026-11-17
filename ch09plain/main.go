package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"
)

type app struct {
	store *Store
	users *Users
}

func (a *app) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /books", a.listBooks)
	mux.HandleFunc("GET /books/{id}", a.getBook)
	mux.HandleFunc("POST /books", a.createBook)
	mux.HandleFunc("PUT /books/{id}", a.updateBook)
	mux.HandleFunc("DELETE /books/{id}", a.deleteBook)
	mux.HandleFunc("POST /register", a.register)
	mux.HandleFunc("POST /login", a.login)
	return logRequests(recoverPanics(mux))
}

func newApp() *app {
	a := &app{store: NewStore(), users: &Users{file: "users.json"}}
	a.store.Add(Book{Title: "The Go Programming Language", Author: "Alan Donovan, Brian Kernighan"})
	a.store.Add(Book{Title: "Clean Code", Author: "Robert C. Martin"})
	a.store.Add(Book{Title: "The Pragmatic Programmer", Author: "Andrew Hunt, David Thomas"})
	return a
}

func main() {
	srv := &http.Server{
		Addr:              ":8080",
		Handler:           newApp().routes(),
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
