package main

import (
	"fmt"
	"log"
	"net/http"
)

type book struct{ ID, Title string }

var books = []book{
	{"1", "The Go Programming Language"},
	{"2", "Clean Code"},
	{"3", "The Pragmatic Programmer"},
}

func listBooks(w http.ResponseWriter, r *http.Request) {
	for _, b := range books {
		fmt.Fprintln(w, b.ID, b.Title)
	}
}

func getBook(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	for _, b := range books {
		if b.ID == id {
			fmt.Fprintln(w, b.Title)
			return
		}
	}
	http.NotFound(w, r)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /books", listBooks)
	mux.HandleFunc("GET /books/{id}", getBook)
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
