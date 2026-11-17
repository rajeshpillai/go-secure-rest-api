package main

import (
	"crypto/sha256"
	"testing"
)

func BenchmarkSHA256(b *testing.B) {
	for b.Loop() {
		sha256.Sum256([]byte("correct horse battery staple"))
	}
}

func BenchmarkArgon2id(b *testing.B) {
	for b.Loop() {
		hashPassword("correct horse battery staple")
	}
}
