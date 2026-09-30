# Networking & HTTP
id: networking
number: 12
track: advanced
paths: beginner, pro
skill: networking
requires: concurrency, json
project: http-monitoring-tool
summary: TCP, HTTP clients and servers, TLS, routing, middleware and connection management.

## Networking Fundamentals
slug: networking-fundamentals
minutes: 7
objectives: Explain addresses, ports and the client/server model; Parse host:port pairs and IP addresses; Classify private, loopback and public addresses
takeaways: A network endpoint is an IP address plus a port; net.SplitHostPort and net/netip parse and classify addresses; Loopback (127.0.0.1) and private ranges are not reachable from the public internet

### Concept
A **client** connects to a **server** at an address like `example.com:443` — a host (name or IP) and a **port** (0–65535). DNS turns names into IPs. **TCP** gives a reliable ordered byte stream; **UDP** gives independent, unreliable datagrams; **HTTP** is a request/response protocol on top of TCP (or QUIC for HTTP/3).

The standard library covers it: `net.Dial`, `net.Listen`, `net.SplitHostPort`, `net.LookupHost`, and the modern `net/netip` package for IP values.

Knowing **private/loopback/link-local** ranges matters for security (SSRF, Module 24).

### Example
```go
package main

import (
	"fmt"
	"net"
	"net/netip"
)

func main() {
	host, port, _ := net.SplitHostPort("example.com:8080")
	fmt.Println(host, port)
	for _, s := range []string{"127.0.0.1", "10.1.2.3", "8.8.8.8"} {
		ip := netip.MustParseAddr(s)
		fmt.Println(s, ip.IsLoopback(), ip.IsPrivate())
	}
}
```

### Exercise
Write `classify(ip string) string` returning `loopback`, `private` or `public` (use `netip`). Invalid input returns `invalid`.
```text expect
loopback private public invalid
```
```go solution
package main

import (
	"fmt"
	"net/netip"
)

// BEGIN
func classify(s string) string {
	ip, err := netip.ParseAddr(s)
	if err != nil {
		return "invalid"
	}
	switch {
	case ip.IsLoopback():
		return "loopback"
	case ip.IsPrivate():
		return "private"
	}
	return "public"
}

// END

func main() {
	fmt.Println(classify("127.0.0.1"), classify("192.168.1.10"), classify("1.1.1.1"), classify("nope"))
}
```

### Check
Q: What identifies a network endpoint?
T: mcq
- [ ] A hostname only
- [x] An IP address (or name) plus a port
- [ ] A MAC address
- [ ] A URL path
E: Connections are made to host:port pairs.

Q: Which protocol gives a reliable, ordered byte stream?
T: short
A: TCP
E: UDP does not guarantee delivery or order.

## TCP
slug: tcp
minutes: 8
objectives: Listen for and accept TCP connections; Dial and exchange data; Handle each connection in its own goroutine
takeaways: net.Listen("tcp", addr) plus Accept() loops handle incoming connections; Handle each connection in a goroutine and always Close it; Use ":0" to let the OS pick a free port (great for tests)

### Concept
```go norun
ln, err := net.Listen("tcp", "127.0.0.1:0") // port 0 = choose a free port
for {
	conn, err := ln.Accept()
	if err != nil { return }
	go handle(conn)
}
```
`net.Conn` is an `io.ReadWriteCloser`, so `bufio`, `io.Copy` and everything you learned about streams applies. Always `defer conn.Close()` and set deadlines (`conn.SetDeadline`) so a stalled peer can't hold a goroutine forever. TCP is a *stream*: message boundaries are yours to define (newlines, length prefixes).

### Example
```go
package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

func main() {
	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	defer ln.Close()
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		line, _ := bufio.NewReader(conn).ReadString('\n')
		fmt.Fprintln(conn, strings.ToUpper(strings.TrimSpace(line)))
	}()
	conn, _ := net.Dial("tcp", ln.Addr().String())
	defer conn.Close()
	fmt.Fprintln(conn, "hello")
	reply, _ := bufio.NewReader(conn).ReadString('\n')
	fmt.Print(reply)
}
```

### Exercise
Write `reverseServer(ln net.Listener)` that accepts **one** connection, reads one line, writes it back reversed (rune-aware) followed by a newline, and closes. The client in `main` sends `gopher`.
```text expect
rehpog
```
```go solution
package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

// BEGIN
func reverseServer(ln net.Listener) {
	conn, err := ln.Accept()
	if err != nil {
		return
	}
	defer conn.Close()
	line, _ := bufio.NewReader(conn).ReadString('\n')
	r := []rune(strings.TrimSpace(line))
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	fmt.Fprintln(conn, string(r))
}

// END

func main() {
	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	defer ln.Close()
	go reverseServer(ln)
	conn, _ := net.Dial("tcp", ln.Addr().String())
	defer conn.Close()
	fmt.Fprintln(conn, "gopher")
	reply, _ := bufio.NewReader(conn).ReadString('\n')
	fmt.Print(reply)
}
```

### Check
Q: What does listening on port `0` do?
T: mcq
- [ ] Disables the listener
- [x] Lets the OS pick a free port
- [ ] Listens on all ports
- [ ] Is invalid
E: Port 0 requests an ephemeral port; read the actual address from `ln.Addr()`.

Q: Message boundaries are preserved automatically by TCP.
T: tf
A: false
E: TCP is a byte stream; applications frame messages with delimiters or length prefixes.

## HTTP
slug: http
status: planned

## HTTP Clients
slug: http-clients
challenges: fetch-with-retry
minutes: 8
objectives: Configure an http.Client with timeouts; Build requests with context; Retry safely and close response bodies
takeaways: Never use http.DefaultClient in production — set Timeout; Use http.NewRequestWithContext so calls are cancellable; Always close resp.Body (and drain it to reuse connections)

### Concept
```go norun
client := &http.Client{Timeout: 5 * time.Second}
req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
req.Header.Set("Accept", "application/json")
resp, err := client.Do(req)
if err != nil { return err }
defer resp.Body.Close()
```
Reuse **one** client: it pools connections (keep-alive). A 4xx/5xx status is not an `err`; check `resp.StatusCode`. Retry only **idempotent** requests (GET, PUT, DELETE) or ones with idempotency keys, with backoff and a cap. Exercises here use `httptest` servers, so no real network is needed.

### Example
```go
package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"time"
)

func main() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "hello %s", r.Header.Get("X-Name"))
	}))
	defer srv.Close()

	client := &http.Client{Timeout: 2 * time.Second}
	req, _ := http.NewRequest(http.MethodGet, srv.URL, nil)
	req.Header.Set("X-Name", "gopher")
	resp, _ := client.Do(req)
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	fmt.Println(resp.StatusCode, string(b))
}
```

### Exercise
Write `getWithRetry(client *http.Client, url string, attempts int) (int, error)` that retries while the status is 503, returning the final status code.
```text expect
200 <nil>
```
```go solution
package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"time"
)

// BEGIN
func getWithRetry(client *http.Client, url string, attempts int) (int, error) {
	var status int
	for i := 0; i < attempts; i++ {
		resp, err := client.Get(url)
		if err != nil {
			return 0, err
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		status = resp.StatusCode
		if status != http.StatusServiceUnavailable {
			return status, nil
		}
	}
	return status, nil
}

// END

func main() {
	var n atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if n.Add(1) < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	fmt.Println(getWithRetry(&http.Client{Timeout: time.Second}, srv.URL, 5))
}
```

### Check
Q: Does `client.Do` return an error for a `500 Internal Server Error` response?
T: mcq
- [ ] Yes
- [x] No — check resp.StatusCode
- [ ] Only for 5xx
- [ ] Only if the body is empty
E: Errors indicate transport failures; HTTP statuses are part of a successful exchange.

Q: Why should you drain and close `resp.Body`?
T: mcq
- [ ] To decode JSON
- [x] So the underlying connection can be reused and doesn't leak
- [ ] To read headers
- [ ] It's optional
E: Unclosed bodies leak connections and prevent keep-alive reuse.

## HTTP Servers
slug: http-servers
challenges: json-user-handler
minutes: 8
objectives: Write handlers with http.HandlerFunc; Start a server with sensible timeouts; Test handlers with httptest without a real network
takeaways: A handler is func(http.ResponseWriter, *http.Request); Configure http.Server with Read/Write/Idle timeouts, not bare ListenAndServe; httptest.NewRecorder and NewRequest test handlers in-process

### Concept
```go norun
mux := http.NewServeMux()
mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %s!", r.URL.Query().Get("name"))
})
srv := &http.Server{
	Addr: ":8080", Handler: mux,
	ReadHeaderTimeout: 5 * time.Second,
	ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second,
}
log.Fatal(srv.ListenAndServe())
```
`http.ListenAndServe` has **no timeouts** — a slow client can hold connections open forever. Each request runs in its own goroutine, so handlers must be safe for concurrent use. Test with `httptest`: create a request, a recorder, call `handler.ServeHTTP(rec, req)`, inspect `rec.Code` and `rec.Body`.

### Example
```go
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

func hello(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		name = "world"
	}
	fmt.Fprintf(w, "Hello, %s!", name)
}

func main() {
	rec := httptest.NewRecorder()
	hello(rec, httptest.NewRequest("GET", "/hello?name=Ada", nil))
	fmt.Println(rec.Code, rec.Body.String())
}
```

### Exercise
Write a handler `health` that responds `200` with body `ok` for GET and `405 Method Not Allowed` for other methods.
```text expect
200 ok
405
```
```go solution
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

// BEGIN
func health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	fmt.Fprint(w, "ok")
}

// END

func main() {
	rec := httptest.NewRecorder()
	health(rec, httptest.NewRequest("GET", "/health", nil))
	fmt.Println(rec.Code, rec.Body.String())
	rec = httptest.NewRecorder()
	health(rec, httptest.NewRequest("POST", "/health", nil))
	fmt.Println(rec.Code)
}
```

### Check
Q: Why avoid bare `http.ListenAndServe` in production?
T: mcq
- [ ] It is deprecated
- [x] It sets no timeouts, so slow or stalled clients can tie up resources
- [ ] It only listens on localhost
- [ ] It can't route requests
E: Configure an `http.Server` with Read/Write/Idle timeouts.

Q: Are HTTP handlers called sequentially?
T: tf
A: false
E: Each request is served in its own goroutine, so shared state needs synchronisation.

## Requests
slug: requests
status: planned

## Responses
slug: responses
status: planned

## Headers
slug: headers
minutes: 6
objectives: Read and set request and response headers; Use canonical header names and the Header map API; Set Content-Type and caching headers correctly
takeaways: r.Header.Get and w.Header().Set read/write headers (names are canonicalised); Headers must be set before WriteHeader or the first Write; Always set Content-Type explicitly for non-obvious payloads

### Concept
Headers are metadata: `Content-Type`, `Accept`, `Authorization`, `Cache-Control`, `User-Agent`, `X-Request-ID`. The `http.Header` type is `map[string][]string` with helper methods: `Get`, `Set`, `Add`, `Values`, `Del`. Names are **canonicalised** (`content-type` → `Content-Type`).

**Order matters on the server:** once you call `w.WriteHeader(status)` or write the body, the headers are sent — later `Header().Set` calls are ignored.

### Example
```go
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, `{"ok":true}`)
}

func main() {
	rec := httptest.NewRecorder()
	handler(rec, httptest.NewRequest("POST", "/", nil))
	fmt.Println(rec.Code, rec.Header().Get("content-type"), rec.Header().Get("Cache-Control"))
}
```

### Exercise
Write `wantsJSON(r *http.Request) bool` returning true when the `Accept` header contains `application/json`.
```text expect
true false
```
```go solution
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
)

// BEGIN
func wantsJSON(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept"), "application/json")
}

// END

func main() {
	a := httptest.NewRequest("GET", "/", nil)
	a.Header.Set("Accept", "text/html, application/json;q=0.9")
	b := httptest.NewRequest("GET", "/", nil)
	fmt.Println(wantsJSON(a), wantsJSON(b))
}
```

### Check
Q: What happens to `w.Header().Set(...)` calls made after `w.WriteHeader(200)`?
T: mcq
- [ ] They apply normally
- [x] They are ignored — the headers were already sent
- [ ] They panic
- [ ] They are sent as trailers
E: Set headers before the status line and body are written.

Q: `http.Header` canonicalises `content-type` to...
T: short
A: Content-Type
E: Header names are normalised to canonical MIME form.

## Cookies
slug: cookies
status: planned

## TLS
slug: tls
status: planned

## Routing
slug: routing
minutes: 7
objectives: Route by method and path with http.ServeMux patterns; Extract path parameters with r.PathValue; Understand precedence and wildcards
takeaways: Go 1.22+ patterns look like "GET /items/{id}"; r.PathValue("id") reads the wildcard; Longer, more specific patterns win over general ones

### Concept
Since Go 1.22 the standard `http.ServeMux` supports methods and wildcards, so many services no longer need a router library.

```go norun
mux.HandleFunc("GET /items", listItems)
mux.HandleFunc("POST /items", createItem)
mux.HandleFunc("GET /items/{id}", getItem)   // r.PathValue("id")
mux.HandleFunc("GET /files/{path...}", serve) // matches the rest of the path
mux.HandleFunc("GET /{$}", home)              // exactly "/"
```
A request that matches the path but not the method gets `405` automatically. Conflicting patterns panic at registration time — a good early failure.

### Example
```go
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "user %s", r.PathValue("id"))
	})
	for _, target := range []string{"/users/42", "/users"} {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", target, nil))
		fmt.Println(rec.Code, rec.Body.String())
	}
}
```

### Exercise
Register `GET /items/{id}/tags/{tag}` responding with `item <id> tag <tag>` and test it with the request `/items/7/tags/go`.
```text expect
item 7 tag go
```
```go solution
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

func main() {
	mux := http.NewServeMux()
	// BEGIN
	mux.HandleFunc("GET /items/{id}/tags/{tag}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "item %s tag %s", r.PathValue("id"), r.PathValue("tag"))
	})
	// END
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/items/7/tags/go", nil))
	fmt.Println(rec.Body.String())
}
```

### Check
Q: How do you read the `{id}` wildcard from pattern `GET /users/{id}`?
T: mcq
- [ ] r.URL.Query().Get("id")
- [x] r.PathValue("id")
- [ ] mux.Param("id")
- [ ] r.Form["id"]
E: `Request.PathValue` returns the matched wildcard segment.

Q: What status does the standard mux send when the path matches but the method doesn't?
T: short
A: 405
E: `405 Method Not Allowed` (with an `Allow` header) is generated automatically.

## Middleware
slug: middleware
minutes: 8
challenges: middleware-chain
objectives: Write middleware as func(http.Handler) http.Handler; Chain middleware in a defined order; Implement logging, recovery and request-ID middleware
takeaways: Middleware wraps a handler and runs code before and/or after it; Order matters: the first in the chain is outermost; Recover from panics at the edge so one bad request can't kill the process

### Concept
```go norun
type Middleware func(http.Handler) http.Handler

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}
```
Typical stack, outermost first: **recover → request ID → logging → auth → rate limit → handler**. To capture the status code for logs wrap the `ResponseWriter` in a struct that records `WriteHeader`. Middleware is just function composition — no framework needed.

### Example
```go
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

func tag(name string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Print(name, ">")
			next.ServeHTTP(w, r)
			fmt.Print("<", name, " ")
		})
	}
}

func main() {
	h := tag("a")(tag("b")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Print("handler ")
	})))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
	fmt.Println()
}
```

### Exercise
Write `Recover(next http.Handler) http.Handler` that turns a handler panic into a `500` response with body `internal error`.
```text expect
500 internal error
```
```go solution
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

// BEGIN
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// END

func main() {
	h := Recover(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	fmt.Println(rec.Code, rec.Body.String()[:14])
}
```

### Check
Q: In the chain `a(b(handler))`, which middleware runs first on the way in?
T: short
A: a
E: The outermost wrapper sees the request first and the response last.

Q: Where should panic-recovery middleware sit in the chain?
T: mcq
- [ ] Innermost, next to the handler
- [x] Outermost, so it catches panics from everything inside
- [ ] After the response is written
- [ ] It doesn't matter
E: Outermost placement means panics in any inner middleware or handler are caught.

## Timeouts
slug: timeouts
minutes: 7
objectives: Set client and server timeouts at the right layers; Use http.TimeoutHandler for slow handlers; Propagate request context to downstream calls
takeaways: Configure ReadHeaderTimeout, ReadTimeout, WriteTimeout and IdleTimeout on servers; Set http.Client.Timeout and use per-request contexts; r.Context() is cancelled when the client disconnects — pass it downstream

### Concept
Every network operation needs a bound.

**Server:** `ReadHeaderTimeout` (slowloris defence), `ReadTimeout`, `WriteTimeout`, `IdleTimeout`. `http.TimeoutHandler(h, d, msg)` returns 503 if the handler takes longer than `d`.

**Client:** `http.Client.Timeout` bounds the whole request including reading the body. Finer control: `Transport` timeouts (`DialContext`, `TLSHandshakeTimeout`, `ResponseHeaderTimeout`).

**Context:** `r.Context()` is cancelled when the client goes away or the server shuts down — pass it to database and downstream HTTP calls so abandoned requests stop consuming resources.

### Example
```go
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"
)

func main() {
	slow := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(200 * time.Millisecond):
			fmt.Fprint(w, "done")
		case <-r.Context().Done():
		}
	})
	h := http.TimeoutHandler(slow, 20*time.Millisecond, "too slow")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	fmt.Println(rec.Code, rec.Body.String())
}
```

### Exercise
Wrap `slow` with `http.TimeoutHandler` (30 ms, message `timeout`) so the request returns `503 timeout`.
```text expect
503 timeout
```
```go solution
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"
)

func main() {
	slow := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(300 * time.Millisecond):
			fmt.Fprint(w, "done")
		case <-r.Context().Done():
		}
	})
	// BEGIN
	h := http.TimeoutHandler(slow, 30*time.Millisecond, "timeout")
	// END
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	fmt.Println(rec.Code, rec.Body.String())
}
```

### Check
Q: Which server timeout defends against slow header sending (slowloris)?
T: mcq
- [ ] WriteTimeout
- [ ] IdleTimeout
- [x] ReadHeaderTimeout
- [ ] TimeoutHandler
E: `ReadHeaderTimeout` bounds how long a client may take to send request headers.

Q: What does `r.Context()` tell a handler?
T: mcq
- [ ] The database connection
- [x] Whether the client disconnected or the request was cancelled
- [ ] The HTTP method
- [ ] The response size
E: It is cancelled when the client goes away — pass it to downstream calls.

## Connection Management
slug: connection-management
status: planned
