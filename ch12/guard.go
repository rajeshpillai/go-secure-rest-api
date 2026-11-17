package main

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
)

type userKey struct{}

// requireUser lets a request through only with a valid access token.
func (a *app) requireUser(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok {
			w.Header().Set("WWW-Authenticate", "Bearer")
			writeError(w, http.StatusUnauthorized, "missing token")
			return
		}
		id, err := a.parseAccessToken(token)
		if err != nil {
			slog.Warn("bad token", "err", err)
			w.Header().Set("WWW-Authenticate", `Bearer error="invalid_token"`)
			writeError(w, http.StatusUnauthorized, "invalid token")
			return
		}
		ctx := context.WithValue(r.Context(), userKey{}, id)
		next(w, r.WithContext(ctx))
	}
}

func currentUser(r *http.Request) int {
	return r.Context().Value(userKey{}).(int)
}

func (a *app) me(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]int{"user_id": currentUser(r)})
}
