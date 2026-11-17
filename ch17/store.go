package main

import (
	"slices"
	"sync"
)

type Book struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Author  string `json:"author"`
	OwnerID int    `json:"owner_id"`
}

// Store keeps the books. The mutex lets one request in at a time.
type Store struct {
	mu     sync.Mutex
	books  map[int]Book
	nextID int
}

func NewStore() *Store {
	return &Store{books: map[int]Book{}, nextID: 1}
}

func (s *Store) Add(b Book) Book {
	s.mu.Lock()
	defer s.mu.Unlock()
	b.ID = s.nextID
	s.nextID++
	s.books[b.ID] = b
	return b
}

func (s *Store) Get(id int) (Book, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, ok := s.books[id]
	return b, ok
}

func (s *Store) List(ownerID int) []Book {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := []Book{}
	for _, b := range s.books {
		if b.OwnerID == ownerID {
			list = append(list, b)
		}
	}
	slices.SortFunc(list, func(a, b Book) int { return a.ID - b.ID })
	return list
}

func (s *Store) Update(b Book) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.books[b.ID]; !ok {
		return false
	}
	s.books[b.ID] = b
	return true
}

func (s *Store) Delete(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.books[id]; !ok {
		return false
	}
	delete(s.books, id)
	return true
}
