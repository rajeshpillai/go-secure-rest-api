package main

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"slices"
	"sync"
)

type APIKey struct {
	ID     int      `json:"id"`
	Name   string   `json:"name"`
	Prefix string   `json:"prefix"` // the first characters, to tell keys apart
	Scopes []string `json:"scopes"`
	UserID int      `json:"-"`
	hash   [32]byte // SHA-256 of the key; the key itself is never stored
}

type APIKeys struct {
	mu     sync.Mutex
	list   []APIKey
	nextID int
}

// Create makes a key. The plain key is returned once, and never stored.
func (k *APIKeys) Create(userID int, name string, scopes []string) (string, APIKey) {
	plain := "bk_" + rand.Text()
	k.mu.Lock()
	defer k.mu.Unlock()
	k.nextID++
	key := APIKey{ID: k.nextID, Name: name, Prefix: plain[:7], Scopes: scopes,
		UserID: userID, hash: sha256.Sum256([]byte(plain))}
	k.list = append(k.list, key)
	return plain, key
}

func (k *APIKeys) Find(plain string) (APIKey, bool) {
	h := sha256.Sum256([]byte(plain))
	k.mu.Lock()
	defer k.mu.Unlock()
	for _, key := range k.list {
		if subtle.ConstantTimeCompare(key.hash[:], h[:]) == 1 {
			return key, true
		}
	}
	return APIKey{}, false
}

func (k *APIKeys) List(userID int) []APIKey {
	k.mu.Lock()
	defer k.mu.Unlock()
	mine := []APIKey{}
	for _, key := range k.list {
		if key.UserID == userID {
			mine = append(mine, key)
		}
	}
	return mine
}

func (k *APIKeys) Revoke(userID, id int) bool {
	k.mu.Lock()
	defer k.mu.Unlock()
	n := len(k.list)
	k.list = slices.DeleteFunc(k.list, func(key APIKey) bool {
		return key.ID == id && key.UserID == userID
	})
	return len(k.list) < n
}
