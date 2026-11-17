package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// testServer starts the whole API in memory, on a random port.
func testServer(t *testing.T) *httptest.Server {
	cfg := config{
		jwtKey:    []byte("test-key-that-is-at-least-32-bytes!"),
		usersFile: t.TempDir() + "/users.json",
	}
	srv := httptest.NewServer(newApp(cfg).routes())
	t.Cleanup(srv.Close)
	return srv
}

// call sends one request and returns the status code and the JSON reply.
func call(t *testing.T, srv *httptest.Server, method, path, auth, body string) (int, map[string]any) {
	t.Helper()
	req, _ := http.NewRequest(method, srv.URL+path, strings.NewReader(body))
	if strings.HasPrefix(auth, "bk_") {
		req.Header.Set("X-API-Key", auth)
	} else if auth != "" {
		req.Header.Set("Authorization", "Bearer "+auth)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var out map[string]any
	data, _ := io.ReadAll(resp.Body)
	json.Unmarshal(data, &out)
	return resp.StatusCode, out
}

// signUp registers a user and returns an access token.
func signUp(t *testing.T, srv *httptest.Server, email string) string {
	creds := `{"email":"` + email + `","password":"a-long-test-password"}`
	call(t, srv, "POST", "/register", "", creds)
	_, out := call(t, srv, "POST", "/login", "", creds)
	return out["access_token"].(string)
}

func TestSecurity(t *testing.T) {
	srv := testServer(t)
	bella := signUp(t, srv, "bella@example.com")
	mallory := signUp(t, srv, "mallory@example.com")
	call(t, srv, "POST", "/books", bella, `{"title":"Dune"}`) // book 1, Bella's
	_, key := call(t, srv, "POST", "/keys", bella, `{"name":"export","scopes":["books:read"]}`)
	readKey := key["key"].(string)

	tests := []struct {
		name, method, path, auth, body string
		want                           int
	}{
		{"no credentials", "GET", "/books", "", "", 401},
		{"a made-up token", "GET", "/books", "not.a.token", "", 401},
		{"a made-up API key", "GET", "/books", "bk_FAKE", "", 401},
		{"Bella reads her book", "GET", "/books/1", bella, "", 200},
		{"Mallory reads Bella's book", "GET", "/books/1", mallory, "", 403},
		{"a read key can read", "GET", "/books/1", readKey, "", 200},
		{"a read key cannot write", "POST", "/books", readKey, `{"title":"Hack"}`, 403},
		{"a key cannot make keys", "POST", "/keys", readKey, `{"name":"x","scopes":["books:write"]}`, 403},
		{"the client cannot pick the id", "POST", "/books", bella, `{"title":"Hack","id":7}`, 400},
		{"Mallory deletes Bella's book", "DELETE", "/books/1", mallory, "", 403},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, out := call(t, srv, tc.method, tc.path, tc.auth, tc.body)
			if got != tc.want {
				t.Errorf("got %d, want %d (%v)", got, tc.want, out)
			}
		})
	}
}

func TestLoginRateLimit(t *testing.T) {
	srv := testServer(t)
	wrong := `{"email":"bella@example.com","password":"not-her-password-at-all"}`
	for i := 1; i <= 5; i++ {
		if got, _ := call(t, srv, "POST", "/login", "", wrong); got != 401 {
			t.Fatalf("try %d: got %d, want 401", i, got)
		}
	}
	if got, _ := call(t, srv, "POST", "/login", "", wrong); got != 429 {
		t.Fatalf("try 6: got %d, want 429", got)
	}
}

func TestRefreshTokenWorksOnce(t *testing.T) {
	srv := testServer(t)
	creds := `{"email":"bella@example.com","password":"a-long-test-password"}`
	call(t, srv, "POST", "/register", "", creds)
	_, out := call(t, srv, "POST", "/login", "", creds)
	first := `{"refresh_token":"` + out["refresh_token"].(string) + `"}`
	if got, _ := call(t, srv, "POST", "/refresh", "", first); got != 200 {
		t.Fatalf("first use: got %d, want 200", got)
	}
	if got, _ := call(t, srv, "POST", "/refresh", "", first); got != 401 {
		t.Fatalf("second use: got %d, want 401", got)
	}
}
