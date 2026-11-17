package main

import (
	"encoding/json"
	"os"
	"sync"
)

type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Users is our tiny user database: a list, saved to a JSON file.
type Users struct {
	mu   sync.Mutex
	list []User
	file string
}

func (u *Users) Add(email, password string) (User, bool) {
	u.mu.Lock()
	defer u.mu.Unlock()
	for _, x := range u.list {
		if x.Email == email {
			return User{}, false
		}
	}
	user := User{ID: len(u.list) + 1, Email: email, Password: password}
	u.list = append(u.list, user)
	data, _ := json.MarshalIndent(u.list, "", "  ")
	os.WriteFile(u.file, data, 0o600)
	return user, true
}

func (u *Users) Find(email string) (User, bool) {
	u.mu.Lock()
	defer u.mu.Unlock()
	for _, x := range u.list {
		if x.Email == email {
			return x, true
		}
	}
	return User{}, false
}
