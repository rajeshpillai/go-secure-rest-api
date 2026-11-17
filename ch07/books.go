package main

import (
	"net/http"
	"strconv"
)

type app struct {
	store *Store
}

type bookInput struct {
	Title  string `json:"title"`
	Author string `json:"author"`
}

func (in bookInput) check() string {
	if in.Title == "" {
		return "title is required"
	}
	if len(in.Title) > 200 {
		return "title is too long"
	}
	return ""
}

func bookID(r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	return id, err == nil
}

func (a *app) listBooks(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.store.List())
}

func (a *app) getBook(w http.ResponseWriter, r *http.Request) {
	id, ok := bookID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "id must be a number")
		return
	}
	b, found := a.store.Get(id)
	if !found {
		writeError(w, http.StatusNotFound, "book not found")
		return
	}
	writeJSON(w, http.StatusOK, b)
}

func (a *app) createBook(w http.ResponseWriter, r *http.Request) {
	var in bookInput
	if err := readJSON(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if msg := in.check(); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}
	b := a.store.Add(Book{Title: in.Title, Author: in.Author})
	writeJSON(w, http.StatusCreated, b)
}

func (a *app) updateBook(w http.ResponseWriter, r *http.Request) {
	id, ok := bookID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "id must be a number")
		return
	}
	var in bookInput
	if err := readJSON(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if msg := in.check(); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}
	b := Book{ID: id, Title: in.Title, Author: in.Author}
	if !a.store.Update(b) {
		writeError(w, http.StatusNotFound, "book not found")
		return
	}
	writeJSON(w, http.StatusOK, b)
}

func (a *app) deleteBook(w http.ResponseWriter, r *http.Request) {
	id, ok := bookID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "id must be a number")
		return
	}
	if !a.store.Delete(id) {
		writeError(w, http.StatusNotFound, "book not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
