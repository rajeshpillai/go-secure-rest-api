package main

import (
	"net/http"
	"strconv"
)

type Book struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
}

var books = []Book{
	{1, "The Go Programming Language", "Alan Donovan, Brian Kernighan"},
	{2, "Clean Code", "Robert C. Martin"},
	{3, "The Pragmatic Programmer", "Andrew Hunt, David Thomas"},
}
var nextID = 4

func listBooks(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, books)
}

func getBook(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "id must be a number")
		return
	}
	for _, b := range books {
		if b.ID == id {
			writeJSON(w, http.StatusOK, b)
			return
		}
	}
	writeError(w, http.StatusNotFound, "book not found")
}

// bookInput is what a client may send: no id, the server picks it.
type bookInput struct {
	Title  string `json:"title"`
	Author string `json:"author"`
}

func createBook(w http.ResponseWriter, r *http.Request) {
	var in bookInput
	if err := readJSON(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if in.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	b := Book{ID: nextID, Title: in.Title, Author: in.Author}
	nextID++
	books = append(books, b)
	writeJSON(w, http.StatusCreated, b)
}
