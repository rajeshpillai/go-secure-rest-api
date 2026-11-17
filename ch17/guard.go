package main

import (
	"context"
	"log/slog"
	"net/http"
	"slices"
	"strings"
)

// principal is who is calling, and what they may do.
type principal struct {
	UserID int      `json:"user_id"`
	Scopes []string `json:"scopes"`
}

type principalKey struct{}

// A password login can do everything. An API key only what it was given.
var loginScopes = []string{"books:read", "books:write", "keys:manage"}

func (a *app) auth(scope string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p, ok := a.identify(r)
		if !ok {
			w.Header().Set("WWW-Authenticate", "Bearer")
			writeError(w, http.StatusUnauthorized, "missing or invalid credentials")
			return
		}
		if !slices.Contains(p.Scopes, scope) {
			writeError(w, http.StatusForbidden, "missing scope "+scope)
			return
		}
		next(w, r.WithContext(context.WithValue(r.Context(), principalKey{}, p)))
	}
}

// identify reads an API key, or else a bearer token.
func (a *app) identify(r *http.Request) (principal, bool) {
	if key := r.Header.Get("X-API-Key"); key != "" {
		k, ok := a.keys.Find(key)
		return principal{k.UserID, k.Scopes}, ok
	}
	token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !ok {
		return principal{}, false
	}
	id, err := a.parseAccessToken(token)
	if err != nil {
		slog.Warn("bad token", "err", err)
		return principal{}, false
	}
	return principal{id, loginScopes}, true
}

func current(r *http.Request) principal {
	return r.Context().Value(principalKey{}).(principal)
}

func (a *app) me(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, current(r))
}
