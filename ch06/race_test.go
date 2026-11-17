package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// Fifty people add a book at the same moment.
func TestManyAddsAtOnce(t *testing.T) {
	srv := httptest.NewServer(newApp().routes())
	defer srv.Close()
	var wg sync.WaitGroup
	for range 50 {
		wg.Go(func() {
			body := strings.NewReader(`{"title":"Go in Action"}`)
			resp, err := http.Post(srv.URL+"/books", "application/json", body)
			if err == nil {
				resp.Body.Close()
			}
		})
	}
	wg.Wait()
}
