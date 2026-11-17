# BookShelf: a secure REST API in Go, step by step

The code from the video tutorial **"Build a secure REST API in Go: passwords, JWT, refresh tokens, API keys"**.
One small API (a bookshelf) grows chapter by chapter, from a 17-line hello server to an API with password
hashing, JWT access tokens, rotating refresh tokens, scoped API keys, ownership checks, rate limits, HTTPS, CORS
and tests. Each folder is the complete program at the end of that chapter, so you can run any step.

Standard library first (Go 1.22+ routing, `encoding/json`, `log/slog`, `httptest`), plus three Go-team or
well-known modules: `golang.org/x/crypto/argon2`, `github.com/golang-jwt/jwt/v5`, `golang.org/x/time/rate`.
Needs Go 1.26 (as in `go.mod`).

| Folder | Chapter | What it adds |
|---|---|---|
| `ch03` | Your first Go server | `net/http`, one handler |
| `ch04` | Routes | `GET /books/{id}`, `r.PathValue`, 404 and 405 for free |
| `ch05` | JSON in, JSON out | body size limit, unknown fields refused, one error shape (and a data race, on purpose) |
| `ch06` | Keeping the books | a mutex-protected store, PUT and DELETE; `go test -race` passes |
| `ch07` | Middleware | logging, panic recovery, server timeouts |
| `ch09plain` | Passwords (the mistake) | passwords stored in plain text: don't do this |
| `ch09` | Passwords, done right | Argon2id (OWASP minimum settings), salts, constant-time compare, same answer and time for unknown emails |
| `ch11` | JWT | HS256 access tokens, method/issuer/expiry checks, secret from the environment |
| `ch12` | Refresh tokens | random tokens stored as SHA-256, used once, reuse ends the whole login, logout |
| `ch13` | API keys | `bk_` prefix, shown once, stored hashed, scopes, revoke |
| `ch14bug` | Authorization (the bug) | any user can read or delete any book by changing the id (OWASP API1:2023) |
| `ch14` | Authorization | ownership check on every book route |
| `ch15` | Rate limiting | token bucket per IP; 5 login tries, then one per 12 s; 429 + Retry-After |
| `ch16` | HTTPS, CORS, secrets | `ListenAndServeTLS`, one allowed origin, config from the environment |
| `ch17` | Tests | table-driven security tests with `httptest`, the rate limit, refresh-token reuse |

## Run a chapter

```bash
export JWT_SECRET=$(openssl rand -base64 32)   # chapters 11+ refuse to start without a 32-byte secret
go run ./ch14                                   # listens on :8080
go test -race ./ch17
```

Chapter 16 serves HTTPS on :8443 and needs a certificate in `ch16/`:

```bash
cd ch16 && go run $(go env GOROOT)/src/crypto/tls/generate_cert.go --host localhost --ca && cd ..
go run ./ch16   # run from inside ch16 so it finds cert.pem and key.pem
```

## The demos

`run.py` starts each chapter's server and runs the exact `curl` commands from the video; `out/` holds what they
printed (every response, token and test result in the video comes from these files). `run.py` passes a fixed,
demo-only `JWT_SECRET`; never reuse it. `tools/mint` makes the expired and `alg: none` tokens used in the JWT demo.

This is teaching code: storage is in memory (users in a JSON file), and a real service would add a database,
cleanup of old rate-limit buckets, and a check of new passwords against known leaked ones.

## Sources

OWASP (Password Storage, JSON Web Token, REST Security and HTML5 Security cheat sheets; API Security Top 10 2023),
RFC 7519 (JWT), RFC 7518 (JWA), RFC 6750 (Bearer tokens), RFC 6585 (429), RFC 9700 (OAuth 2.0 security BCP),
NIST SP 800-63B-4, MDN (403, CORS), the Go blog ("Routing enhancements"), Go 1.24 release notes, `net/http` docs.

## License

MIT
