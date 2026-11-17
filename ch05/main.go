package main

import (
	"log"
	"net/http"
)

func routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /books", listBooks)
	mux.HandleFunc("GET /books/{id}", getBook)
	mux.HandleFunc("POST /books", createBook)
	return mux
}

func main() {
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", routes()))
}
