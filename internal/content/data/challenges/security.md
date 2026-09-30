# Security Challenges

## password-hashing
title: Password Hasher
difficulty: intermediate
module: security
lesson: security/password-hashing
skill: security

### Problem
Implement password hashing with PBKDF2-HMAC-SHA256 (`crypto/pbkdf2`, Go 1.24+):

- `HashPassword(password string) (string, error)` — generates a random 16-byte salt and returns the self-describing string `pbkdf2-sha256$<iterations>$<salt hex>$<hash hex>` using **10 000 iterations** and a 32-byte key
- `VerifyPassword(password, encoded string) bool` — parses the string (the iteration count comes from the string, so it can be raised later), recomputes and compares in **constant time**. Malformed input returns `false` and must not panic.

Two hashes of the same password must differ (different salts).

### Constraints
- Salt from `crypto/rand`
- Constant-time comparison
- Never panic on malformed encodings

### Starter
```go
package main

func HashPassword(password string) (string, error) {
	return "", nil
}

func VerifyPassword(password, encoded string) bool {
	return false
}

func main() {}
```

### Tests
```go
package main

import (
	"strings"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	enc, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(enc, "pbkdf2-sha256$10000$") {
		t.Errorf("unexpected format: %q", enc)
	}
	if parts := strings.Split(enc, "$"); len(parts) != 4 || len(parts[2]) != 32 || len(parts[3]) != 64 {
		t.Errorf("want 4 parts with 16-byte salt and 32-byte hash (hex), got %q", enc)
	}
	if !VerifyPassword("correct horse battery staple", enc) {
		t.Error("correct password rejected")
	}
	if VerifyPassword("correct horse battery stapl", enc) || VerifyPassword("", enc) {
		t.Error("wrong password accepted")
	}
}

func TestSaltsDiffer(t *testing.T) {
	a, _ := HashPassword("same")
	b, _ := HashPassword("same")
	if a == b {
		t.Error("two hashes of the same password must differ (random salt)")
	}
	if !VerifyPassword("same", a) || !VerifyPassword("same", b) {
		t.Error("both must verify")
	}
}

func TestMalformedInputNeverPanics(t *testing.T) {
	enc, _ := HashPassword("pw")
	bad := []string{
		"", "garbage", "pbkdf2-sha256$abc$00$00", "pbkdf2-sha256$10000$zz$00",
		"pbkdf2-sha256$10000$00", "pbkdf2-sha256$-5$00$00", "argon2$10000$00$00",
		enc + "$extra", strings.Replace(enc, "$", "|", 1),
	}
	for _, b := range bad {
		if VerifyPassword("pw", b) {
			t.Errorf("malformed %q verified", b)
		}
	}
}

func TestIterationCountComesFromEncoding(t *testing.T) {
	enc, _ := HashPassword("pw")
	parts := strings.Split(enc, "$")
	parts[1] = "20000" // tampering with the cost must invalidate the hash
	if VerifyPassword("pw", strings.Join(parts, "$")) {
		t.Error("changed iteration count should not verify")
	}
}
```

### Hints
- `pbkdf2.Key(sha256.New, password, salt, iterations, 32)` returns `([]byte, error)`
- `subtle.ConstantTimeCompare` for the final comparison
- Reject iteration counts < 1 (and consider an upper bound to avoid denial of service)

### Solution
```go
package main

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

const defaultIterations = 10000

func derive(password string, salt []byte, iterations int) ([]byte, error) {
	return pbkdf2.Key(sha256.New, password, salt, iterations, 32)
}

func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key, err := derive(password, salt, defaultIterations)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("pbkdf2-sha256$%d$%s$%s", defaultIterations, hex.EncodeToString(salt), hex.EncodeToString(key)), nil
}

func VerifyPassword(password, encoded string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 4 || parts[0] != "pbkdf2-sha256" {
		return false
	}
	iterations, err := strconv.Atoi(parts[1])
	if err != nil || iterations < 1 || iterations > 10_000_000 {
		return false
	}
	salt, err := hex.DecodeString(parts[2])
	if err != nil {
		return false
	}
	want, err := hex.DecodeString(parts[3])
	if err != nil {
		return false
	}
	got, err := derive(password, salt, iterations)
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare(got, want) == 1
}

func main() {}
```

### Explanation
A random per-hash salt and a slow KDF make offline guessing expensive; encoding the parameters in the string lets you increase the cost later. Constant-time comparison avoids timing side channels and every parse step fails closed.

## csrf-tokens
title: CSRF Token Service
difficulty: advanced
module: security
lesson: security/csrf
skill: security

### Problem
Implement a stateless CSRF-token service bound to a session ID and an expiry:

- `NewCSRF(secret []byte, ttl time.Duration, now func() time.Time) *CSRF`
- `(c *CSRF) Issue(sessionID string) string` — returns `<expiryUnix>.<hex(HMAC-SHA256(secret, sessionID + "." + expiryUnix))>`
- `(c *CSRF) Valid(sessionID, token string) bool` — true only if the token has the right format, hasn't expired, and the HMAC matches (constant time). Tokens for another session, tampered tokens, expired tokens and garbage are rejected without panicking.

### Constraints
- Use `hmac.Equal` for comparison
- Use the injected clock (`now`)
- Bind the token to the session ID **and** the expiry

### Starter
```go
package main

import "time"

type CSRF struct {
}

func NewCSRF(secret []byte, ttl time.Duration, now func() time.Time) *CSRF {
	return &CSRF{}
}

func (c *CSRF) Issue(sessionID string) string {
	return ""
}

func (c *CSRF) Valid(sessionID, token string) bool {
	return false
}

func main() {}
```

### Tests
```go
package main

import (
	"strings"
	"testing"
	"time"
)

func TestIssueAndValidate(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	c := NewCSRF([]byte("server-secret"), time.Hour, func() time.Time { return now })
	tok := c.Issue("session-1")
	if !c.Valid("session-1", tok) {
		t.Fatal("fresh token rejected")
	}
	if c.Valid("session-2", tok) {
		t.Error("token must be bound to its session")
	}
	if !strings.Contains(tok, ".") {
		t.Errorf("unexpected token format %q", tok)
	}
}

func TestExpiry(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	c := NewCSRF([]byte("k"), time.Minute, func() time.Time { return now })
	tok := c.Issue("s")
	now = now.Add(59 * time.Second)
	if !c.Valid("s", tok) {
		t.Error("token should still be valid before expiry")
	}
	now = now.Add(2 * time.Second)
	if c.Valid("s", tok) {
		t.Error("expired token accepted")
	}
}

func TestTamperingRejected(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	c := NewCSRF([]byte("k"), time.Hour, func() time.Time { return now })
	tok := c.Issue("s")
	exp, mac, _ := strings.Cut(tok, ".")
	// Extending the expiry must break the MAC.
	if c.Valid("s", "9999999999."+mac) {
		t.Error("changing the expiry must invalidate the token")
	}
	flipped := mac[:len(mac)-1] + "0"
	if mac[len(mac)-1] == '0' {
		flipped = mac[:len(mac)-1] + "1"
	}
	if c.Valid("s", exp+"."+flipped) {
		t.Error("modified MAC accepted")
	}
	other := NewCSRF([]byte("different"), time.Hour, func() time.Time { return now })
	if other.Valid("s", tok) {
		t.Error("token signed with another secret accepted")
	}
}

func TestGarbageNeverPanics(t *testing.T) {
	c := NewCSRF([]byte("k"), time.Hour, time.Now)
	for _, g := range []string{"", ".", "abc", "1.2.3", "x.y", "123.", ".abc", "123.zz", strings.Repeat("9", 100) + ".00"} {
		if c.Valid("s", g) {
			t.Errorf("%q accepted", g)
		}
	}
}
```

### Hints
- Compute `mac := hmac.New(sha256.New, secret)`; write `sessionID + "." + exp`
- Parse the expiry with `strconv.ParseInt`; compare with `now().Unix()`
- Decode the presented MAC from hex and use `hmac.Equal`

### Solution
```go
package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"
)

type CSRF struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

func NewCSRF(secret []byte, ttl time.Duration, now func() time.Time) *CSRF {
	return &CSRF{secret: secret, ttl: ttl, now: now}
}

func (c *CSRF) sign(sessionID, exp string) []byte {
	m := hmac.New(sha256.New, c.secret)
	m.Write([]byte(sessionID + "." + exp))
	return m.Sum(nil)
}

func (c *CSRF) Issue(sessionID string) string {
	exp := strconv.FormatInt(c.now().Add(c.ttl).Unix(), 10)
	return exp + "." + hex.EncodeToString(c.sign(sessionID, exp))
}

func (c *CSRF) Valid(sessionID, token string) bool {
	exp, macHex, ok := strings.Cut(token, ".")
	if !ok || exp == "" || macHex == "" || strings.Contains(macHex, ".") {
		return false
	}
	expUnix, err := strconv.ParseInt(exp, 10, 64)
	if err != nil || c.now().Unix() > expUnix {
		return false
	}
	got, err := hex.DecodeString(macHex)
	if err != nil {
		return false
	}
	return hmac.Equal(got, c.sign(sessionID, exp))
}

func main() {}
```

### Explanation
The HMAC covers both the session and the expiry, so neither can be changed without the secret. Checking expiry before comparing keeps stale tokens out, and `hmac.Equal` avoids timing leaks.

## ssrf-guard
title: SSRF Guard
difficulty: advanced
module: security
lesson: security/ssrf
skill: security

### Problem
Write `CheckURL(raw string, resolve func(host string) ([]netip.Addr, error)) error` that decides whether an outbound request to `raw` is allowed.

Return an error when:

- the URL cannot be parsed, has no host, or embeds credentials (`user:pass@`)
- the scheme is not `http` or `https`
- the port (if any) is not 80, 443 or 8080–8090
- **any** address the host resolves to (or the IP literal itself) is loopback, private, link-local, unspecified, multicast, or in the carrier-grade NAT range `100.64.0.0/10`
- resolution fails or returns no addresses

`resolve` is injected so the check is testable (use it only for host names, not IP literals). IPv4-mapped IPv6 addresses (`::ffff:10.0.0.1`) must be treated as their IPv4 form.

### Constraints
- Check **every** resolved address
- Unmap IPv4-in-IPv6 addresses before classifying
- Fail closed

### Starter
```go
package main

import "net/netip"

func CheckURL(raw string, resolve func(host string) ([]netip.Addr, error)) error {
	return nil
}

func main() {}
```

### Tests
```go
package main

import (
	"errors"
	"net/netip"
	"testing"
)

func fixedResolver(m map[string][]string) func(string) ([]netip.Addr, error) {
	return func(host string) ([]netip.Addr, error) {
		ips, ok := m[host]
		if !ok {
			return nil, errors.New("no such host")
		}
		var out []netip.Addr
		for _, s := range ips {
			out = append(out, netip.MustParseAddr(s))
		}
		return out, nil
	}
}

func TestAllowed(t *testing.T) {
	res := fixedResolver(map[string][]string{
		"example.com": {"93.184.216.34"},
		"api.dual.io": {"93.184.216.34", "2606:2800:220:1:248:1893:25c8:1946"},
	})
	for _, u := range []string{
		"https://example.com/hook",
		"http://example.com:8080/x",
		"https://api.dual.io/",
		"https://93.184.216.34/",
		"http://example.com:8085/",
	} {
		if err := CheckURL(u, res); err != nil {
			t.Errorf("CheckURL(%q) = %v, want nil", u, err)
		}
	}
}

func TestBlocked(t *testing.T) {
	res := fixedResolver(map[string][]string{
		"internal.example": {"10.0.0.7"},
		"mixed.example":    {"93.184.216.34", "127.0.0.1"}, // one bad address is enough
		"meta.example":     {"169.254.169.254"},
		"cgnat.example":    {"100.64.1.2"},
		"empty.example":    {},
	})
	blocked := []string{
		"ftp://example.com/x",
		"file:///etc/passwd",
		"http://localhost/",
		"http://127.0.0.1/",
		"http://[::1]/",
		"http://10.1.2.3/",
		"http://172.16.5.4/",
		"http://192.168.0.1/",
		"http://169.254.169.254/latest/meta-data/",
		"http://[::ffff:10.0.0.1]/",
		"http://0.0.0.0/",
		"http://224.0.0.1/",
		"http://100.100.100.100/",
		"http://internal.example/",
		"http://mixed.example/",
		"http://meta.example/",
		"http://cgnat.example/",
		"http://empty.example/",
		"http://unknown.example/",
		"https://user:pass@93.184.216.34/",
		"https://93.184.216.34:22/",
		"http://93.184.216.34:9000/",
		"not a url",
		"http:///nohost",
		"",
	}
	for _, u := range blocked {
		if err := CheckURL(u, res); err == nil {
			t.Errorf("CheckURL(%q) = nil, want an error", u)
		}
	}
}

func TestResolverNotUsedForIPLiterals(t *testing.T) {
	called := false
	res := func(host string) ([]netip.Addr, error) {
		called = true
		return nil, errors.New("should not be called")
	}
	if err := CheckURL("https://93.184.216.34/", res); err != nil {
		t.Errorf("public literal rejected: %v", err)
	}
	if called {
		t.Error("resolver must not be called for IP literals")
	}
}
```

### Hints
- `net/url.Parse`; `u.Hostname()`, `u.Port()`, `u.User`
- `netip.ParseAddr(host)` succeeds for IP literals; call `.Unmap()` on the result
- `netip.MustParsePrefix("100.64.0.0/10").Contains(ip)` for CGNAT
- Loop over **all** resolved addresses

### Solution
```go
package main

import (
	"errors"
	"fmt"
	"net/netip"
	"net/url"
	"strconv"
)

var cgnat = netip.MustParsePrefix("100.64.0.0/10")

func blockedIP(ip netip.Addr) bool {
	ip = ip.Unmap()
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
		ip.IsUnspecified() || ip.IsMulticast() || cgnat.Contains(ip)
}

func allowedPort(p string) bool {
	if p == "" {
		return true
	}
	n, err := strconv.Atoi(p)
	if err != nil {
		return false
	}
	return n == 80 || n == 443 || (n >= 8080 && n <= 8090)
}

func CheckURL(raw string, resolve func(host string) ([]netip.Addr, error)) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("scheme %q not allowed", u.Scheme)
	}
	host := u.Hostname()
	if host == "" {
		return errors.New("missing host")
	}
	if u.User != nil {
		return errors.New("credentials in URL not allowed")
	}
	if !allowedPort(u.Port()) {
		return fmt.Errorf("port %q not allowed", u.Port())
	}
	if ip, err := netip.ParseAddr(host); err == nil {
		if blockedIP(ip) {
			return fmt.Errorf("address %s is not allowed", ip)
		}
		return nil
	}
	if host == "localhost" {
		return errors.New("localhost is not allowed")
	}
	addrs, err := resolve(host)
	if err != nil {
		return fmt.Errorf("resolve %s: %w", host, err)
	}
	if len(addrs) == 0 {
		return fmt.Errorf("resolve %s: no addresses", host)
	}
	for _, ip := range addrs {
		if blockedIP(ip) {
			return fmt.Errorf("%s resolves to disallowed address %s", host, ip)
		}
	}
	return nil
}

func main() {}
```

### Explanation
Every layer of the check fails closed: scheme, credentials, port, then *all* resolved addresses. Unmapping IPv4-in-IPv6 stops `::ffff:10.0.0.1` from slipping through, and in production you would additionally pin the connection to the validated IP to defeat DNS rebinding.
