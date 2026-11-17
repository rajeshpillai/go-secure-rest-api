package main

import (
	"errors"
	"os"
)

// config comes from the environment, never from the code.
type config struct {
	jwtKey     []byte
	corsOrigin string
	usersFile  string
}

func configFromEnv() (config, error) {
	key := os.Getenv("JWT_SECRET")
	if len(key) < 32 {
		return config{}, errors.New("JWT_SECRET must be set, at least 32 bytes")
	}
	return config{
		jwtKey:     []byte(key),
		corsOrigin: os.Getenv("CORS_ORIGIN"),
		usersFile:  "users.json",
	}, nil
}
