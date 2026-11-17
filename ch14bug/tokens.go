package main

import (
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// newAccessToken signs a short-lived pass that says "this is user N".
func (a *app) newAccessToken(userID int) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Issuer:    "bookshelf",
		Subject:   strconv.Itoa(userID),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(a.jwtKey)
}

// parseAccessToken checks the signature, the method, the issuer and the expiry.
func (a *app) parseAccessToken(s string) (int, error) {
	var claims jwt.RegisteredClaims
	_, err := jwt.ParseWithClaims(s, &claims,
		func(*jwt.Token) (any, error) { return a.jwtKey, nil },
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithIssuer("bookshelf"),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(claims.Subject)
}
