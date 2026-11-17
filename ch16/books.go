package main

import (
	"net/http"
	"strconv"
)

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

// ownBook finds the book in the URL and checks it belongs to the caller.
// It writes the error itself; ok is false if the handler should stop.
func (a *app) ownBook(w http.ResponseWriter, r *http.Request) (Book, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "id must be a number")
		return Book{}, false
	}
	b, found := a.store.Get(id)
	if !found {
		writeError(w, http.StatusNotFound, "book not found")
		return Book{}, false
	}
	if b.OwnerID != current(r).UserID {
		writeError(w, http.StatusForbidden, "not your book")
		return Book{}, false
	}
	return b, true
}

func (a *app) listBooks(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.store.List(current(r).UserID))
}

func (a *app) getBook(w http.ResponseWriter, r *http.Request) {
	if b, ok := a.ownBook(w, r); ok {
		writeJSON(w, http.StatusOK, b)
	}
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
	b := a.store.Add(Book{Title: in.Title, Author: in.Author, OwnerID: current(r).UserID})
	writeJSON(w, http.StatusCreated, b)
}

func (a *app) updateBook(w http.ResponseWriter, r *http.Request) {
	b, ok := a.ownBook(w, r)
	if !ok {
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
	b.Title, b.Author = in.Title, in.Author
	a.store.Update(b)
	writeJSON(w, http.StatusOK, b)
}

func (a *app) deleteBook(w http.ResponseWriter, r *http.Request) {
	if b, ok := a.ownBook(w, r); ok {
		a.store.Delete(b.ID)
		w.WriteHeader(http.StatusNoContent)
	}
}
