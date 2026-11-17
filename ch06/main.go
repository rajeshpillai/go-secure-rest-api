package main

import (
	"log"
	"net/http"
)

func (a *app) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /books", a.listBooks)
	mux.HandleFunc("GET /books/{id}", a.getBook)
	mux.HandleFunc("POST /books", a.createBook)
	mux.HandleFunc("PUT /books/{id}", a.updateBook)
	mux.HandleFunc("DELETE /books/{id}", a.deleteBook)
	return mux
}

func newApp() *app {
	a := &app{store: NewStore()}
	a.store.Add(Book{Title: "The Go Programming Language", Author: "Alan Donovan, Brian Kernighan"})
	a.store.Add(Book{Title: "Clean Code", Author: "Robert C. Martin"})
	a.store.Add(Book{Title: "The Pragmatic Programmer", Author: "Andrew Hunt, David Thomas"})
	return a
}

func main() {
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", newApp().routes()))
}
