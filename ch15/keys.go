package main

import (
	"net/http"
	"slices"
	"strconv"
)

// Scopes a key may carry. A key can never manage other keys.
var keyScopes = []string{"books:read", "books:write"}

type keyInput struct {
	Name   string   `json:"name"`
	Scopes []string `json:"scopes"`
}

func (a *app) createKey(w http.ResponseWriter, r *http.Request) {
	var in keyInput
	if err := readJSON(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if in.Name == "" || len(in.Scopes) == 0 {
		writeError(w, http.StatusBadRequest, "name and scopes are required")
		return
	}
	for _, s := range in.Scopes {
		if !slices.Contains(keyScopes, s) {
			writeError(w, http.StatusBadRequest, "unknown scope "+s)
			return
		}
	}
	plain, key := a.keys.Create(current(r).UserID, in.Name, in.Scopes)
	writeJSON(w, http.StatusCreated, map[string]any{
		"key":     plain,
		"id":      key.ID,
		"prefix":  key.Prefix,
		"scopes":  key.Scopes,
		"warning": "copy this key now: it will not be shown again",
	})
}

func (a *app) listKeys(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.keys.List(current(r).UserID))
}

func (a *app) revokeKey(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || !a.keys.Revoke(current(r).UserID, id) {
		writeError(w, http.StatusNotFound, "key not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
