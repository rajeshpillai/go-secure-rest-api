// mint makes demo tokens the server did not issue: an expired one (signed with
// the real key, 16 minutes old) and an "alg: none" one. Used only by run.py.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	then := time.Now().Add(-16 * time.Minute)
	claims := jwt.RegisteredClaims{Issuer: "bookshelf", Subject: "1",
		IssuedAt: jwt.NewNumericDate(then), ExpiresAt: jwt.NewNumericDate(then.Add(15 * time.Minute))}
	switch os.Args[1] {
	case "expired":
		s, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(os.Getenv("JWT_SECRET")))
		fmt.Println(s)
	case "none":
		claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Hour))
		s, _ := jwt.NewWithClaims(jwt.SigningMethodNone, claims).SignedString(jwt.UnsafeAllowNoneSignatureType)
		fmt.Println(s)
	}
}
