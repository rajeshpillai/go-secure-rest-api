package main

import (
	"crypto/rand"
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
	u, ok := a.users.Add(strings.ToLower(in.Email), hashPassword(in.Password))
	if !ok {
		writeError(w, http.StatusConflict, "email is already registered")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": u.ID, "email": u.Email})
}

// dummyHash lets a login for an unknown email do the same slow work,
// so the response time doesn't reveal which emails exist.
var dummyHash = hashPassword("not-a-real-password-at-all")

func (a *app) login(w http.ResponseWriter, r *http.Request) {
	var in credentials
	if err := readJSON(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	u, found := a.users.Find(strings.ToLower(in.Email))
	if !found {
		checkPassword(in.Password, dummyHash)
	}
	if !found || !checkPassword(in.Password, u.PasswordHash) {
		writeError(w, http.StatusUnauthorized, "wrong email or password")
		return
	}
	a.sendTokens(w, u.ID, rand.Text())
}

// sendTokens answers with a fresh pair: a 15-minute access token
// and a 7-day refresh token in the same family.
func (a *app) sendTokens(w http.ResponseWriter, userID int, family string) {
	access, err := a.newAccessToken(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not sign token")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"access_token":  access,
		"refresh_token": a.sessions.New(userID, family),
		"token_type":    "Bearer",
	})
}

type refreshInput struct {
	RefreshToken string `json:"refresh_token"`
}

func (a *app) refresh(w http.ResponseWriter, r *http.Request) {
	var in refreshInput
	if err := readJSON(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	sess, ok := a.sessions.Use(in.RefreshToken)
	if !ok {
		writeError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}
	a.sendTokens(w, sess.userID, sess.family)
}

func (a *app) logout(w http.ResponseWriter, r *http.Request) {
	var in refreshInput
	if err := readJSON(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	a.sessions.Logout(in.RefreshToken)
	w.WriteHeader(http.StatusNoContent)
}
