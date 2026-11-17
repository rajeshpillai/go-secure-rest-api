package main

import (
	"crypto/rand"
	"crypto/sha256"
	"sync"
	"time"
)

// A refresh token is a long random string. The server keeps only its SHA-256.
// Every token from one login shares a family, so one theft can end them all.
type session struct {
	userID  int
	family  string
	expires time.Time
	used    bool
}

type Sessions struct {
	mu     sync.Mutex
	byHash map[[32]byte]session
}

func (s *Sessions) New(userID int, family string) string {
	token := rand.Text()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byHash[sha256.Sum256([]byte(token))] = session{
		userID: userID, family: family, expires: time.Now().Add(7 * 24 * time.Hour)}
	return token
}

// Use spends a refresh token. Each one works once.
func (s *Sessions) Use(token string) (session, bool) {
	h := sha256.Sum256([]byte(token))
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.byHash[h]
	if !ok || time.Now().After(sess.expires) {
		return session{}, false
	}
	if sess.used { // a spent token came back: someone has a copy
		s.endFamily(sess.family)
		return session{}, false
	}
	sess.used = true
	s.byHash[h] = sess
	return sess, true
}

// Logout ends the whole login the token belongs to.
func (s *Sessions) Logout(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sess, ok := s.byHash[sha256.Sum256([]byte(token))]; ok {
		s.endFamily(sess.family)
	}
}

func (s *Sessions) endFamily(family string) {
	for h, x := range s.byHash {
		if x.family == family {
			delete(s.byHash, h)
		}
	}
}
