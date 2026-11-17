package main

import (
	"net/http"
	"strings"
	"unicode/utf8"
)

type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (a *app) register(w http.ResponseWriter, r *http.Request) {
	var in credentials
	if err := readJSON(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !strings.Contains(in.Email, "@") {
		writeError(w, http.StatusBadRequest, "email is not valid")
		return
	}
	if n := utf8.RuneCountInString(in.Password); n < 15 || n > 64 {
		writeError(w, http.StatusBadRequest, "password must be 15 to 64 characters")
		return
	}
	u, ok := a.users.Add(strings.ToLower(in.Email), in.Password)
	if !ok {
		writeError(w, http.StatusConflict, "email is already registered")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": u.ID, "email": u.Email})
}

func (a *app) login(w http.ResponseWriter, r *http.Request) {
	var in credentials
	if err := readJSON(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	u, found := a.users.Find(strings.ToLower(in.Email))
	if !found || u.Password != in.Password {
		writeError(w, http.StatusUnauthorized, "wrong email or password")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": u.ID, "message": "welcome back"})
}
