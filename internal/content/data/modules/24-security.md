# Go Security
id: security
number: 24
track: security
paths: pro
skill: security
requires: rest-api
project: authentication-service
summary: Secure coding: authentication, JWT, OAuth2, TLS, hashing and web vulnerabilities.

## Authentication
slug: authentication
minutes: 8
challenges: password-hashing
objectives: Distinguish authentication from authorization; Generate unguessable session tokens; Store only hashes of tokens and compare in constant time
takeaways: Authentication answers "who are you?"; authorization answers "what may you do?"; Session tokens must be long, random (crypto/rand) and stored hashed server-side; Compare secrets in constant time and give the same error for unknown user and wrong password

### Concept
Flow for local accounts:

1. User submits credentials over **HTTPS**.
2. Server verifies the password hash (next lesson) — same generic error for "no such user" and "wrong password" (prevents user enumeration) and similar timing.
3. Server creates a **session token**: 32 random bytes from `crypto/rand`, base64/hex encoded.
4. Store only `SHA-256(token)` (like a password) in the database; send the token in a cookie with `HttpOnly; Secure; SameSite=Lax`.
5. On each request hash the presented token and look it up; expire and rotate sessions; support logout by deleting the row.

Never use `math/rand` for secrets, never put tokens in URLs, and never log them. For SSO prefer OIDC (Keycloak, Google, Auth0) over homegrown auth.

### Example
```go
package main

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
)

func newToken() (token, hash string) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	token = hex.EncodeToString(b)
	sum := sha256.Sum256([]byte(token))
	return token, hex.EncodeToString(sum[:])
}

func matches(token, storedHash string) bool {
	sum := sha256.Sum256([]byte(token))
	got := hex.EncodeToString(sum[:])
	return subtle.ConstantTimeCompare([]byte(got), []byte(storedHash)) == 1
}

func main() {
	tok, h := newToken()
	fmt.Println(len(tok), matches(tok, h), matches("forged", h))
}
```

### Exercise
Write `hashToken(token string) string` returning the hex SHA-256 and `verify(token, stored string) bool` using `subtle.ConstantTimeCompare`.
```text expect
true false
```
```go solution
package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
)

// BEGIN
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func verify(token, stored string) bool {
	return subtle.ConstantTimeCompare([]byte(hashToken(token)), []byte(stored)) == 1
}

// END

func main() {
	stored := hashToken("s3cr3t-token")
	fmt.Println(verify("s3cr3t-token", stored), verify("guess", stored))
}
```

### Check
Q: Which package must generate session tokens?
T: mcq
- [ ] math/rand
- [x] crypto/rand
- [ ] time
- [ ] hash/fnv
E: Only crypto/rand is unpredictable enough for secrets.

Q: Why return the same error for "unknown user" and "wrong password"?
T: mcq
- [ ] It's shorter
- [x] To prevent attackers from discovering which accounts exist
- [ ] HTTP requires it
- [ ] To speed up login
E: Distinct messages (or timings) leak account existence — user enumeration.

## Authorization
slug: authorization
status: planned

## JWT
slug: jwt
status: planned

## OAuth2
slug: oauth2
status: planned

## OIDC
slug: oidc
status: planned

## TLS
slug: tls
minutes: 8
objectives: Configure a modern crypto/tls Config; Serve HTTPS and understand certificates; Know why InsecureSkipVerify is dangerous
takeaways: Set MinVersion to TLS 1.2 (prefer 1.3) and let Go choose safe cipher suites; Certificates prove identity via a chain of trust — obtain them with ACME/Let's Encrypt; never set InsecureSkipVerify: true outside tests

### Concept
```go norun
cfg := &tls.Config{MinVersion: tls.VersionTLS12}
srv := &http.Server{Addr: ":443", Handler: mux, TLSConfig: cfg}
log.Fatal(srv.ListenAndServeTLS("cert.pem", "key.pem"))
```
Go's defaults are sensible (secure cipher suites, TLS 1.3 support); don't hand-pick suites. Clients verify the server certificate chain and host name automatically. **`InsecureSkipVerify: true` disables that** — it makes TLS pointless (anyone can impersonate the server). For tests use `httptest.NewTLSServer` and its `Client()`. In production, terminate TLS at a load balancer/ingress or use `golang.org/x/crypto/acme/autocert` for automatic certificates. Add HSTS. mTLS (client certificates) authenticates services to each other.

### Example
```go
package main

import (
	"crypto/tls"
	"fmt"
)

func main() {
	cfg := &tls.Config{MinVersion: tls.VersionTLS12}
	fmt.Println(cfg.MinVersion == tls.VersionTLS12, cfg.InsecureSkipVerify)
}
```

### Exercise
Write `serverConfig() *tls.Config` with `MinVersion: TLS 1.2` and `MaxVersion` left at the default. Print the min version name using `tls.VersionName`.
```text expect
TLS 1.2 false
```
```go solution
package main

import (
	"crypto/tls"
	"fmt"
)

// BEGIN
func serverConfig() *tls.Config {
	return &tls.Config{MinVersion: tls.VersionTLS12}
}

// END

func main() {
	c := serverConfig()
	fmt.Println(tls.VersionName(c.MinVersion), c.InsecureSkipVerify)
}
```

### Check
Q: What does `InsecureSkipVerify: true` do?
T: mcq
- [ ] Speeds up the handshake safely
- [x] Disables certificate verification, allowing man-in-the-middle attacks
- [ ] Enables mTLS
- [ ] Forces TLS 1.3
E: Never use it in production code.

Q: What minimum TLS version should a modern server require?
T: mcq
- [ ] SSL 3.0
- [ ] TLS 1.0
- [x] TLS 1.2 (or 1.3)
- [ ] No minimum
E: Older protocols have known weaknesses.

## HTTPS
slug: https
status: planned

## Password Hashing
slug: password-hashing
minutes: 9
challenges: password-hashing
objectives: Explain why passwords need slow, salted hashes; Derive keys with PBKDF2 (and know bcrypt/argon2id); Encode salt and parameters with the hash
takeaways: Never store plain or fast-hashed (MD5/SHA-256) passwords; use a slow KDF: argon2id, bcrypt, scrypt or PBKDF2 with a per-user random salt; Store algorithm, cost and salt with the hash so parameters can be upgraded later

### Concept
Databases leak. A leaked table of `SHA-256(password)` is cracked at billions of guesses per second. A **password hash** is deliberately slow and **salted** (unique random bytes per user, defeating rainbow tables).

Recommended: **argon2id** (`golang.org/x/crypto/argon2`), **bcrypt** (`golang.org/x/crypto/bcrypt`), **scrypt**, or PBKDF2 (in the standard library since Go 1.24: `crypto/pbkdf2`, e.g. 600 000 iterations of HMAC-SHA-256).

Store a self-describing string: `$pbkdf2-sha256$600000$<salt>$<hash>` so you can raise the cost later and re-hash at next login. Compare with `subtle.ConstantTimeCompare`. Add rate limiting and (optionally) a server-side **pepper**.

### Example
```go
package main

import (
	"crypto/pbkdf2"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func main() {
	key, err := pbkdf2.Key(sha256.New, "correct horse", []byte("per-user-salt-16"), 10_000, 32)
	fmt.Println(hex.EncodeToString(key)[:16], err)
}
```

### Exercise
Write `Hash(password string, salt []byte) string` returning `hex(salt)$hex(pbkdf2(password, salt, 1000 iterations, 32 bytes))` and `Verify(password, encoded string) bool`, which splits the string, recomputes and compares in constant time.
```text expect
true false
```
```go solution
package main

import (
	"crypto/pbkdf2"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strings"
)

const iterations = 1000

// BEGIN
func Hash(password string, salt []byte) string {
	key, _ := pbkdf2.Key(sha256.New, password, salt, iterations, 32)
	return hex.EncodeToString(salt) + "$" + hex.EncodeToString(key)
}

func Verify(password, encoded string) bool {
	saltHex, _, ok := strings.Cut(encoded, "$")
	if !ok {
		return false
	}
	salt, err := hex.DecodeString(saltHex)
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(Hash(password, salt)), []byte(encoded)) == 1
}

// END

func main() {
	enc := Hash("hunter2", []byte("0123456789abcdef"))
	fmt.Println(Verify("hunter2", enc), Verify("hunter3", enc))
}
```

### Check
Q: Why is `SHA-256(password)` a bad way to store passwords?
T: mcq
- [ ] SHA-256 is broken
- [x] It's fast, so attackers can test billions of guesses per second
- [ ] It produces too-short output
- [ ] It's not deterministic
E: Password hashes must be slow and salted.

Q: What is a salt?
T: mcq
- [ ] A secret shared by all users
- [x] Random per-user data mixed into the hash so identical passwords produce different hashes
- [ ] A compression setting
- [ ] A TLS extension
E: Salts defeat rainbow tables and prevent cross-user comparison.

## Secret Management
slug: secret-management
status: planned

## Input Validation
slug: input-validation
minutes: 7
objectives: Validate with allow-lists rather than block-lists; Enforce length, type and range limits; Validate on the server regardless of client checks
takeaways: Prefer allow-lists (what's permitted) to deny-lists (what's forbidden); Bound everything: length, size, counts, ranges; Client-side validation is UX, server-side validation is security

### Concept
All input — form fields, headers, query strings, file uploads, JSON — is untrusted.

1. **Type & shape:** decode into typed structs; reject unknown fields.
2. **Allow-list:** `^[a-z0-9_]{3,20}$` for usernames beats trying to strip "bad characters".
3. **Limits:** max body size (`http.MaxBytesReader`), string length, slice length, numeric range, page size.
4. **Canonicalise before validating:** normalise Unicode, trim, lowercase emails — then check.
5. **Validate at the boundary**, then use safe APIs downstream (parameterised SQL, `html/template`, `exec.Command` args). Validation reduces risk; encoding/escaping at the point of use prevents injection.

### Example
```go
package main

import (
	"fmt"
	"regexp"
	"unicode/utf8"
)

var usernameRe = regexp.MustCompile(`^[a-z0-9_]+$`)

func validUsername(s string) bool {
	n := utf8.RuneCountInString(s)
	return n >= 3 && n <= 20 && usernameRe.MatchString(s)
}

func main() {
	fmt.Println(validUsername("ada_99"), validUsername("ad"), validUsername("robert'); DROP--"))
}
```

### Exercise
Write `validEmail(s string) bool`: at most 254 bytes, exactly one `@`, non-empty local part (≤ 64) and a domain containing a `.` that doesn't start or end with `.` or `-`.
```text expect
true false false false
```
```go solution
package main

import (
	"fmt"
	"strings"
)

// BEGIN
func validEmail(s string) bool {
	if len(s) > 254 || strings.Count(s, "@") != 1 {
		return false
	}
	local, domain, _ := strings.Cut(s, "@")
	if local == "" || len(local) > 64 {
		return false
	}
	if !strings.Contains(domain, ".") {
		return false
	}
	for _, edge := range []string{".", "-"} {
		if strings.HasPrefix(domain, edge) || strings.HasSuffix(domain, edge) {
			return false
		}
	}
	return true
}

// END

func main() {
	fmt.Println(validEmail("ada@example.com"), validEmail("ada@@example.com"), validEmail("@example.com"), validEmail("ada@example"))
}
```

### Check
Q: Which validation strategy is safer?
T: mcq
- [ ] Block-list of known bad inputs
- [x] Allow-list of what's known good
- [ ] Trusting the browser
- [ ] Removing characters silently
E: Attackers find inputs your deny-list didn't anticipate.

Q: Client-side validation is enough to protect the server.
T: tf
A: false
E: Attackers bypass browsers entirely; always validate on the server.

## SQL Injection
slug: sql-injection
minutes: 7
objectives: Recognise injectable code; Fix it with parameterised queries; Handle dynamic identifiers and LIKE patterns safely
takeaways: SQL injection happens when input is concatenated into SQL text; Parameterised queries send values separately and eliminate the vulnerability for values; Identifiers must be allow-listed, and LIKE wildcards (% and _) escaped when they are literal

### Concept
```go norun
// VULNERABLE
q := "SELECT * FROM users WHERE name = '" + name + "'"
// name = "x' OR '1'='1"  → returns every user

// SAFE
rows, err := db.QueryContext(ctx, "SELECT * FROM users WHERE name = $1", name)
```
Rules: **always parameters** for values (`$1`/`?`); **allow-list** table/column names and sort directions; for `LIKE` searches escape `%`, `_` and `\` in the user's term if they should match literally, then pass the pattern as a parameter; use a least-privilege database account (no `DROP`).

### Example
```go
package main

import (
	"fmt"
	"strings"
)

func escapeLike(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return r.Replace(s)
}

func main() {
	fmt.Println(escapeLike("50%_off"))
}
```

### Exercise
Write `orderBy(col, dir string) (string, error)` returning `ORDER BY <col> <dir>` only if `col` is in the allow-list `id`, `name`, `created_at` and `dir` is `asc` or `desc` (case-insensitive, output uppercase); otherwise return an error.
```text expect
ORDER BY name DESC <nil>
 invalid sort
```
```go solution
package main

import (
	"errors"
	"fmt"
	"strings"
)

var cols = map[string]bool{"id": true, "name": true, "created_at": true}

// BEGIN
func orderBy(col, dir string) (string, error) {
	d := strings.ToUpper(dir)
	if !cols[col] || (d != "ASC" && d != "DESC") {
		return "", errors.New("invalid sort")
	}
	return "ORDER BY " + col + " " + d, nil
}

// END

func main() {
	fmt.Println(orderBy("name", "desc"))
	fmt.Println(orderBy("name; DROP TABLE users", "asc"))
}
```

### Check
Q: What is the correct fix for `"... WHERE name = '" + name + "'"`?
T: mcq
- [ ] Remove quotes from the input
- [x] Use a parameterised query: WHERE name = $1
- [ ] Encode the string as base64
- [ ] Use a longer password
E: Parameters keep data separate from SQL code.

Q: Can a column name in ORDER BY be a query parameter?
T: tf
A: false
E: Parameters only stand for values; allow-list identifiers instead.

## XSS
slug: xss
minutes: 7
objectives: Explain reflected and stored cross-site scripting; Use html/template's contextual auto-escaping; Avoid marking untrusted data as safe
takeaways: XSS lets attacker-controlled data run as script in a victim's browser; html/template escapes output for its HTML, attribute, JS and URL context automatically; text/template does NOT — never use it for HTML — and template.HTML() disables escaping

### Concept
If user input is written into a page unescaped, `<script>steal(document.cookie)</script>` executes. Defences:

1. **Escape on output** — `html/template` does it contextually.
2. **Content Security Policy** header limiting where scripts load from.
3. **`HttpOnly` cookies** so scripts can't read session cookies.
4. **Never** convert untrusted strings to `template.HTML`, `template.JS` or `template.URL`.
5. For JSON APIs set `Content-Type: application/json` and `X-Content-Type-Options: nosniff`.

`html/template` vs `text/template`: same API; only the former escapes.

### Example
```go
package main

import (
	"html/template"
	"os"
)

func main() {
	t := template.Must(template.New("p").Parse(`<p>Hello, {{.}}!</p>` + "\n"))
	t.Execute(os.Stdout, `<script>alert(1)</script>`)
}
```

### Exercise
Render `<a href="{{.URL}}">{{.Label}}</a>` with `html/template` where `URL` is `javascript:alert(1)` and `Label` is `<b>hi</b>`. Print the result (html/template neutralises both).
```text expect
<a href="#ZgotmplZ">&lt;b&gt;hi&lt;/b&gt;</a>
```
```go solution
package main

import (
	"html/template"
	"os"
)

func main() {
	// BEGIN
	t := template.Must(template.New("l").Parse(`<a href="{{.URL}}">{{.Label}}</a>` + "\n"))
	t.Execute(os.Stdout, map[string]string{"URL": "javascript:alert(1)", "Label": "<b>hi</b>"})
	// END
}
```

### Check
Q: Which package auto-escapes template output for HTML contexts?
T: mcq
- [ ] text/template
- [x] html/template
- [ ] fmt
- [ ] encoding/xml
E: Use `html/template` for anything that produces HTML.

Q: Converting untrusted input with `template.HTML(input)` is safe.
T: tf
A: false
E: It tells the template engine to skip escaping — exactly what an attacker wants.

## CSRF
slug: csrf
minutes: 8
challenges: csrf-tokens
objectives: Explain how cross-site request forgery abuses ambient credentials; Defend with SameSite cookies and anti-CSRF tokens; Generate and verify HMAC-bound tokens
takeaways: CSRF makes a victim's browser send an authenticated request the user didn't intend; Defences: SameSite=Lax/Strict cookies, anti-CSRF tokens, checking Origin/Referer and custom headers; Tokens should be unguessable and bound to the session (e.g. HMAC(sessionID))

### Concept
Browsers attach cookies to requests to your site *even when triggered from another site*. An attacker's page can auto-submit a form to `POST /transfer` and your server sees a valid session.

Defences (use several):

- **`SameSite=Lax` (or Strict)** on session cookies — the modern baseline.
- **Anti-CSRF token** in a hidden field or header that an attacker's page can't read (same-origin policy); the server verifies it.
- Reject state-changing requests without an expected **custom header** (`X-Requested-With`, this platform's `X-GoLearn-CSRF`) — cross-origin requests with custom headers require CORS preflight.
- Verify **Origin** / `Sec-Fetch-Site`.
- Never change state on `GET`.

A stateless token: `HMAC-SHA256(secret, sessionID)` — verify by recomputing.

### Example
```go
package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func token(secret, session string) string {
	m := hmac.New(sha256.New, []byte(secret))
	m.Write([]byte(session))
	return hex.EncodeToString(m.Sum(nil))
}

func main() {
	t := token("server-secret", "session-1")
	fmt.Println(len(t), t == token("server-secret", "session-1"), t == token("server-secret", "session-2"))
}
```

### Exercise
Write `valid(secret, session, presented string) bool` that recomputes the token and compares it in constant time with `hmac.Equal` (decode hex first; invalid hex is false).
```text expect
true false false
```
```go solution
package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func token(secret, session string) string {
	m := hmac.New(sha256.New, []byte(secret))
	m.Write([]byte(session))
	return hex.EncodeToString(m.Sum(nil))
}

// BEGIN
func valid(secret, session, presented string) bool {
	got, err := hex.DecodeString(presented)
	if err != nil {
		return false
	}
	want, _ := hex.DecodeString(token(secret, session))
	return hmac.Equal(got, want)
}

// END

func main() {
	t := token("k", "s1")
	fmt.Println(valid("k", "s1", t), valid("k", "s2", t), valid("k", "s1", "zz"))
}
```

### Check
Q: Which cookie attribute is the modern baseline defence against CSRF?
T: short
A: SameSite
E: `SameSite=Lax` stops cookies being sent on most cross-site sub-requests.

Q: State-changing operations should be accepted over GET as long as the user is logged in.
T: tf
A: false
E: GET must be safe; links and images can trigger GETs cross-site.

## SSRF
slug: ssrf
minutes: 8
challenges: ssrf-guard
objectives: Explain server-side request forgery and why cloud metadata endpoints are a target; Validate outbound URLs against private and link-local ranges; Understand DNS rebinding and redirect risks
takeaways: SSRF tricks your server into requesting internal resources (localhost, 10.x, 169.254.169.254 cloud metadata); Block non-http(s) schemes and private, loopback, link-local and unspecified addresses — after DNS resolution and on every redirect; Prefer allow-lists of destination hosts

### Concept
If your service fetches a user-supplied URL (webhooks, image proxies, importers), an attacker can point it at `http://169.254.169.254/latest/meta-data/` (cloud credentials!) or `http://localhost:6379`.

Defences:

1. Allow only `http`/`https`.
2. **Resolve** the hostname and reject loopback, private (`10/8`, `172.16/12`, `192.168/16`), link-local (`169.254/16`), unspecified and multicast addresses (IPv4 and IPv6).
3. Connect to the **validated IP** (custom `DialContext`) to defeat **DNS rebinding** (a name that resolves safely at check time and unsafely at connect time).
4. Validate **redirects** the same way (or disable them).
5. Best of all: an **allow-list** of hosts, plus egress network policies.

### Example
```go
package main

import (
	"fmt"
	"net/netip"
)

func blocked(ip netip.Addr) bool {
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsUnspecified() || ip.IsMulticast()
}

func main() {
	for _, s := range []string{"127.0.0.1", "10.0.0.5", "169.254.169.254", "8.8.8.8", "::1"} {
		fmt.Println(s, blocked(netip.MustParseAddr(s)))
	}
}
```

### Exercise
Write `safe(raw string) bool`: the URL must parse, use `http` or `https`, and its host must be either a **public IP literal** or a name (assume names are checked elsewhere). IP literals that are loopback, private, link-local, unspecified or multicast are unsafe; the host `localhost` is unsafe too.
```text expect
true false false false false
```
```go solution
package main

import (
	"fmt"
	"net/netip"
	"net/url"
	"strings"
)

// BEGIN
func safe(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return false
	}
	host := u.Hostname()
	if strings.EqualFold(host, "localhost") || host == "" {
		return false
	}
	if ip, err := netip.ParseAddr(host); err == nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsUnspecified() || ip.IsMulticast() {
			return false
		}
	}
	return true
}

// END

func main() {
	fmt.Println(
		safe("https://example.com/hook"),
		safe("http://169.254.169.254/latest/meta-data"),
		safe("http://localhost:8080"),
		safe("file:///etc/passwd"),
		safe("http://192.168.1.10/admin"),
	)
}
```

### Check
Q: Which address is a classic SSRF target on cloud platforms?
T: mcq
- [ ] 8.8.8.8
- [x] 169.254.169.254 (instance metadata service)
- [ ] 1.1.1.1
- [ ] 224.0.0.1
E: The metadata endpoint can expose temporary cloud credentials.

Q: Validating the hostname string once is enough to stop SSRF.
T: tf
A: false
E: DNS can change between check and use (rebinding) and redirects can point inward; validate the resolved IP at connect time and on every redirect.

## Rate Limiting
slug: rate-limiting
status: planned

## Secure Headers
slug: secure-headers
minutes: 7
objectives: Set the essential security headers; Explain what each header prevents; Apply them in middleware
takeaways: Key headers: Content-Security-Policy, Strict-Transport-Security, X-Content-Type-Options: nosniff, X-Frame-Options / frame-ancestors, Referrer-Policy, Permissions-Policy; Set them once in middleware so every response has them; CSP is the strongest XSS mitigation but needs careful rollout (report-only first)

### Concept
| Header | Purpose |
| --- | --- |
| `Content-Security-Policy` | restricts where scripts, styles, frames load from — mitigates XSS |
| `Strict-Transport-Security` | tells browsers to use HTTPS only (`max-age=63072000; includeSubDomains`) |
| `X-Content-Type-Options: nosniff` | stops MIME sniffing |
| `X-Frame-Options: DENY` / CSP `frame-ancestors` | prevents clickjacking |
| `Referrer-Policy: strict-origin-when-cross-origin` | limits referrer leakage |
| `Permissions-Policy` | disables powerful browser features |

```go norun
func SecureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		...
		next.ServeHTTP(w, r)
	})
}
```
GoLearn's own server sets these — look at `internal/httpx` in the repository.

### Example
```go
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

func SecureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	})
}

func main() {
	rec := httptest.NewRecorder()
	SecureHeaders(http.NotFoundHandler()).ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	fmt.Println(rec.Header().Get("X-Frame-Options"), rec.Header().Get("X-Content-Type-Options"))
}
```

### Exercise
Extend `SecureHeaders` to also set `Referrer-Policy: strict-origin-when-cross-origin` and `Content-Security-Policy: default-src 'self'`. Print those two header values.
```text expect
strict-origin-when-cross-origin default-src 'self'
```
```go solution
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

func SecureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		// BEGIN
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Content-Security-Policy", "default-src 'self'")
		// END
		next.ServeHTTP(w, r)
	})
}

func main() {
	rec := httptest.NewRecorder()
	SecureHeaders(http.NotFoundHandler()).ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	fmt.Println(rec.Header().Get("Referrer-Policy"), rec.Header().Get("Content-Security-Policy"))
}
```

### Check
Q: Which header tells browsers to only use HTTPS for your site?
T: mcq
- [ ] X-Frame-Options
- [x] Strict-Transport-Security
- [ ] Referrer-Policy
- [ ] X-Content-Type-Options
E: HSTS pins the site to HTTPS for the given max-age.

Q: Which header primarily helps prevent clickjacking?
T: mcq
- [ ] Cache-Control
- [x] X-Frame-Options (or CSP frame-ancestors)
- [ ] ETag
- [ ] Accept
E: It stops other sites from embedding your pages in frames.

## Dependency Security
slug: dependency-security
status: planned
