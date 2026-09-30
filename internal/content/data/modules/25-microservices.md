# Go Microservices
id: microservices
number: 25
track: microservices
paths: pro
skill: microservices
requires: docker, observability
project: microservice-platform
summary: Service boundaries, gRPC, messaging, idempotency, resilience and tracing.

## Microservice Architecture
slug: microservice-architecture
minutes: 9
objectives: Explain when microservices help and when they hurt; Draw service boundaries around business capabilities; Order service startup with dependency analysis
takeaways: Microservices trade in-process simplicity for independent deployment and scaling — and add network failure modes; Draw boundaries around business capabilities, with each service owning its data; Start with a well-structured monolith unless team or scaling needs justify the split

### Concept
The capstone platform:

```
API Gateway
     ↓
User Service
     ↓
Flight Service
     ↓
Notification Service
     ↓
Database
```
Benefits: independent deploys, scaling hot paths separately, team autonomy, fault isolation. Costs: network latency and failure, distributed transactions, harder debugging, operational overhead (deploys, observability, versioned APIs).

Guidelines:

- One service **owns** its data; others go through its API/events — never share a database.
- Boundaries follow **business capabilities** (billing, catalogue), not technical layers.
- Design every call to **fail**: timeouts, retries, circuit breakers, idempotency.
- Prefer async events for cross-service workflows; keep synchronous chains short.
- "Modular monolith first" is often the right start.

### Example
```go
package main

import "fmt"

func main() {
	deps := map[string][]string{
		"gateway":      {"user", "flight"},
		"flight":       {"notification"},
		"user":         {},
		"notification": {},
	}
	fmt.Println(len(deps), "services;", "gateway depends on", deps["gateway"])
}
```

### Exercise
Write `startOrder(deps map[string][]string) ([]string, error)` returning a startup order in which every service comes after its dependencies (alphabetical among ready services for determinism), or an error `dependency cycle` if none exists.
```text expect
[notification flight user gateway] <nil>
[] dependency cycle
```
```go solution
package main

import (
	"errors"
	"fmt"
	"sort"
)

// BEGIN
func startOrder(deps map[string][]string) ([]string, error) {
	remaining := map[string]int{}
	users := map[string][]string{}
	for svc, ds := range deps {
		remaining[svc] += 0
		for _, d := range ds {
			remaining[svc]++
			users[d] = append(users[d], svc)
			if _, ok := remaining[d]; !ok {
				remaining[d] = 0
			}
		}
	}
	var order []string
	for len(remaining) > 0 {
		var ready []string
		for svc, n := range remaining {
			if n == 0 {
				ready = append(ready, svc)
			}
		}
		if len(ready) == 0 {
			return []string{}, errors.New("dependency cycle")
		}
		sort.Strings(ready)
		svc := ready[0]
		order = append(order, svc)
		delete(remaining, svc)
		for _, u := range users[svc] {
			remaining[u]--
		}
	}
	return order, nil
}

// END

func main() {
	fmt.Println(startOrder(map[string][]string{
		"gateway":      {"user", "flight"},
		"flight":       {"notification"},
		"user":         nil,
		"notification": nil,
	}))
	fmt.Println(startOrder(map[string][]string{"a": {"b"}, "b": {"a"}}))
}
```

### Check
Q: Two services should share one database to keep data consistent.
T: tf
A: false
E: Shared databases couple services tightly; each service should own its data and expose it through an API or events.

Q: Which is the better first step for a new product with a small team?
T: mcq
- [ ] 30 microservices from day one
- [x] A modular monolith that can be split later along clear boundaries
- [ ] One giant unstructured package
- [ ] A service per database table
E: Microservices add operational cost; split when you have a real reason.

## REST Services
slug: rest-services
status: planned

## gRPC
slug: grpc
status: planned

## Service-to-Service Communication
slug: service-to-service-communication
status: planned

## Message Queues
slug: message-queues
minutes: 9
objectives: Describe queues, acknowledgements and redelivery; Explain at-least-once versus at-most-once delivery; Model a queue with visibility timeouts
takeaways: A queue decouples producers from consumers and absorbs bursts; Consumers acknowledge after processing — unacknowledged messages are redelivered (at-least-once); Poison messages need a retry limit and a dead-letter queue

### Concept
```
Producer ──▶ [ queue ] ──▶ Consumer (ack) 
                 ▲             │ no ack / crash
                 └─ redeliver ─┘   after a visibility timeout
```
Common brokers: RabbitMQ, NATS JetStream, Kafka (log-based), AWS SQS, Google Pub/Sub, Redis Streams.

**Delivery guarantees:**

- *At-most-once* — fire and forget; may lose messages.
- *At-least-once* — ack after processing; may duplicate → handlers must be **idempotent**. The usual choice.
- *Exactly-once* — end-to-end this is really "at-least-once + idempotent processing".

Robust consumers: bounded retries with backoff, a **dead-letter queue** for poison messages, graceful shutdown that finishes the current message, and metrics for queue depth and age.

### Example
```go
package main

import "fmt"

type Queue struct{ msgs []string }

func (q *Queue) Publish(m string) { q.msgs = append(q.msgs, m) }

func (q *Queue) Receive() (string, bool) {
	if len(q.msgs) == 0 {
		return "", false
	}
	m := q.msgs[0]
	q.msgs = q.msgs[1:]
	return m, true
}

func main() {
	var q Queue
	q.Publish("welcome-email:ada")
	m, _ := q.Receive()
	fmt.Println(m)
}
```

### Exercise
Implement acknowledgements: `Receive()` returns the next message but keeps it as *in flight*; `Ack(m)` removes it; `Nack(m)` puts it back at the **front** of the queue for redelivery. Publish `a, b`, receive `a`, nack it, receive again (still `a`), ack it, then receive `b`.
```text expect
a a b
```
```go solution
package main

import "fmt"

type Queue struct {
	msgs     []string
	inflight map[string]bool
}

func (q *Queue) Publish(m string) { q.msgs = append(q.msgs, m) }

// BEGIN
func (q *Queue) Receive() (string, bool) {
	if len(q.msgs) == 0 {
		return "", false
	}
	if q.inflight == nil {
		q.inflight = map[string]bool{}
	}
	m := q.msgs[0]
	q.msgs = q.msgs[1:]
	q.inflight[m] = true
	return m, true
}

func (q *Queue) Ack(m string) { delete(q.inflight, m) }

func (q *Queue) Nack(m string) {
	if q.inflight[m] {
		delete(q.inflight, m)
		q.msgs = append([]string{m}, q.msgs...)
	}
}

// END

func main() {
	var q Queue
	q.Publish("a")
	q.Publish("b")
	first, _ := q.Receive()
	q.Nack(first)
	again, _ := q.Receive()
	q.Ack(again)
	next, _ := q.Receive()
	fmt.Println(first, again, next)
}
```

### Check
Q: Which delivery guarantee do most practical systems choose?
T: mcq
- [ ] At-most-once
- [x] At-least-once with idempotent consumers
- [ ] Exactly-once at the network level
- [ ] None
E: True network-level exactly-once doesn't exist; idempotency achieves the effect.

Q: What is a dead-letter queue for?
T: mcq
- [ ] Storing old messages forever
- [x] Holding messages that repeatedly fail so they don't block the queue
- [ ] Encrypting messages
- [ ] Ordering messages
E: Poison messages get parked for inspection instead of being retried forever.

## Event-Driven Architecture
slug: event-driven-architecture
minutes: 9
objectives: Describe events versus commands; Publish and subscribe with a simple bus; Design events as immutable facts
takeaways: Events state that something happened (OrderPlaced); commands ask for something to happen (PlaceOrder); Publishers don't know who consumes — services stay decoupled; Events are immutable, versioned facts; consumers must be idempotent and tolerate reordering

### Concept
In an event-driven design the booking service publishes `FlightBooked`; the notification service, analytics and loyalty services each subscribe and react independently. Adding a new consumer requires no change to the publisher.

Design tips:

- Name events in the **past tense**, carry the data consumers need (or an ID to fetch), and include `id`, `type`, `occurredAt`, `version`.
- The **outbox pattern**: write the event to an `outbox` table in the same transaction as the state change, and a relay publishes it — avoiding the "saved but never published" gap.
- Expect **duplicates and reordering**; make handlers idempotent and tolerant.
- Keep an event schema policy (add fields, don't break consumers).

### Example
```go
package main

import "fmt"

type Bus struct{ handlers map[string][]func(payload string) }

func (b *Bus) Subscribe(topic string, h func(string)) {
	if b.handlers == nil {
		b.handlers = map[string][]func(string){}
	}
	b.handlers[topic] = append(b.handlers[topic], h)
}

func (b *Bus) Publish(topic, payload string) {
	for _, h := range b.handlers[topic] {
		h(payload)
	}
}

func main() {
	var bus Bus
	bus.Subscribe("flight.booked", func(p string) { fmt.Println("email:", p) })
	bus.Subscribe("flight.booked", func(p string) { fmt.Println("analytics:", p) })
	bus.Publish("flight.booked", "AB123")
}
```

### Exercise
Add **wildcard** support: `Publish("flight.booked", …)` should also call handlers registered on the topic `flight.*`. Print each handler call.
```text expect
exact: AB123
wildcard: AB123
```
```go solution
package main

import (
	"fmt"
	"strings"
)

type Bus struct{ handlers map[string][]func(payload string) }

func (b *Bus) Subscribe(topic string, h func(string)) {
	if b.handlers == nil {
		b.handlers = map[string][]func(string){}
	}
	b.handlers[topic] = append(b.handlers[topic], h)
}

// BEGIN
func (b *Bus) Publish(topic, payload string) {
	for _, h := range b.handlers[topic] {
		h(payload)
	}
	if prefix, _, ok := strings.Cut(topic, "."); ok {
		for _, h := range b.handlers[prefix+".*"] {
			h(payload)
		}
	}
}

// END

func main() {
	var bus Bus
	bus.Subscribe("flight.booked", func(p string) { fmt.Println("exact:", p) })
	bus.Subscribe("flight.*", func(p string) { fmt.Println("wildcard:", p) })
	bus.Publish("flight.booked", "AB123")
}
```

### Check
Q: Which is a well-formed event name?
T: mcq
- [ ] BookFlight
- [x] FlightBooked
- [ ] doBooking
- [ ] booking
E: Events describe facts that already happened, in the past tense.

Q: What problem does the outbox pattern solve?
T: mcq
- [ ] Slow queries
- [x] A state change being saved but its event never published (or vice versa)
- [ ] Message compression
- [ ] Load balancing
E: Writing state and event in one transaction, then relaying, keeps them consistent.

## Idempotency
slug: idempotency
minutes: 9
objectives: Define idempotency and why retries need it; Implement an idempotency-key store; Handle concurrent duplicate requests
takeaways: An idempotent operation has the same effect whether executed once or many times; Clients send an Idempotency-Key; the server stores the first result and replays it for duplicates; Concurrent duplicates must not both execute — coordinate per key

### Concept
Networks fail *after* the server acted: the client times out, retries, and a naive server charges the card twice. Solution: an **idempotency key**.

```http
POST /payments
Idempotency-Key: 7b1c2f4e-...
```
Server logic:

1. Look up the key. If a stored result exists → return it (same status and body).
2. If the key is **in progress** → wait or return `409`.
3. Otherwise execute, **store the result under the key** (with a TTL, e.g. 24 h), respond.

Store the key atomically with the effect (same DB transaction, unique constraint on the key). Consumers of message queues need the same protection because delivery is *at-least-once*: track processed message IDs.

### Example
```go
package main

import "fmt"

type Store struct{ results map[string]string }

func (s *Store) Do(key string, f func() string) string {
	if r, ok := s.results[key]; ok {
		return r
	}
	r := f()
	s.results[key] = r
	return r
}

func main() {
	s := &Store{results: map[string]string{}}
	charges := 0
	charge := func() string { charges++; return "receipt-1" }
	fmt.Println(s.Do("k1", charge), s.Do("k1", charge), charges)
}
```

### Exercise
Make `Store.Do` safe for concurrent duplicates: 20 goroutines call `Do("k", charge)` at once; `charge` must run exactly once. Print the number of charges. (Use a per-key `sync.Once`-style entry.)
```text expect
1
```
```go solution
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// BEGIN
type entry struct {
	once   sync.Once
	result string
}

type Store struct {
	mu      sync.Mutex
	entries map[string]*entry
}

func (s *Store) Do(key string, f func() string) string {
	s.mu.Lock()
	if s.entries == nil {
		s.entries = map[string]*entry{}
	}
	e, ok := s.entries[key]
	if !ok {
		e = &entry{}
		s.entries[key] = e
	}
	s.mu.Unlock()
	e.once.Do(func() { e.result = f() })
	return e.result
}

// END

func main() {
	var s Store
	var charges atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Do("k", func() string { charges.Add(1); return "ok" })
		}()
	}
	wg.Wait()
	fmt.Println(charges.Load())
}
```

### Check
Q: Why do message consumers need idempotency too?
T: mcq
- [ ] Queues are slow
- [x] Delivery is usually at-least-once, so a message may be processed more than once
- [ ] Messages are encrypted
- [ ] Consumers are single-threaded
E: Duplicates happen after timeouts, rebalances and retries.

Q: Where should the idempotency key be recorded relative to the side effect?
T: mcq
- [ ] Before, in a separate system
- [x] Atomically with the effect (e.g. same transaction with a unique constraint)
- [ ] Never
- [ ] After, best effort
E: Otherwise a crash between the two steps breaks the guarantee.

## Retries
slug: retries
status: planned

## Timeouts
slug: timeouts
status: planned

## Circuit Breakers
slug: circuit-breakers
status: planned

## Distributed Tracing
slug: distributed-tracing
minutes: 9
objectives: Explain traces, spans and context propagation; Parse and create W3C traceparent headers; Propagate trace context across service calls
takeaways: A trace is the journey of one request; spans are timed steps within it; The W3C traceparent header (version-traceid-spanid-flags) carries context between services; Every hop must forward the trace ID and create a child span

### Concept
```
Application ──▶ OpenTelemetry ──▶ Tracing Backend (Jaeger, Tempo)
```
One user request may touch a gateway, three services and two databases. A **trace ID** ties all the work together; each unit of work is a **span** with a parent span ID, start/end times and attributes.

Propagation uses the **W3C `traceparent`** header:

```
traceparent: 00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01
             ver  trace-id (16 bytes hex)          span-id (8 bytes)  flags
```
With OpenTelemetry (`go.opentelemetry.io/otel`): instrument `http.Handler`/`http.Client` with `otelhttp`, start spans with `tracer.Start(ctx, "name")`, and export via OTLP to a collector. The current span travels in the `context.Context` — so always pass `ctx` down.

### Example
```go
package main

import (
	"fmt"
	"strings"
)

func parse(h string) (traceID, spanID string, ok bool) {
	parts := strings.Split(h, "-")
	if len(parts) != 4 || parts[0] != "00" || len(parts[1]) != 32 || len(parts[2]) != 16 {
		return "", "", false
	}
	return parts[1], parts[2], true
}

func main() {
	fmt.Println(parse("00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"))
	fmt.Println(parse("garbage"))
}
```

### Exercise
Write `child(traceparent string, newSpanID string) (string, bool)` returning a new header that keeps the **trace ID**, replaces the span ID with `newSpanID` (must be 16 hex chars) and keeps the flags. Invalid input returns `"", false`.
```text expect
00-4bf92f3577b34da6a3ce929d0e0e4736-aaaaaaaaaaaaaaaa-01 true
 false
```
```go solution
package main

import (
	"fmt"
	"strings"
)

func isHex(s string, n int) bool {
	if len(s) != n {
		return false
	}
	for _, r := range s {
		if !(r >= '0' && r <= '9' || r >= 'a' && r <= 'f') {
			return false
		}
	}
	return true
}

// BEGIN
func child(traceparent, newSpanID string) (string, bool) {
	parts := strings.Split(traceparent, "-")
	if len(parts) != 4 || parts[0] != "00" || !isHex(parts[1], 32) || !isHex(parts[2], 16) || !isHex(parts[3], 2) {
		return "", false
	}
	if !isHex(newSpanID, 16) {
		return "", false
	}
	return strings.Join([]string{parts[0], parts[1], newSpanID, parts[3]}, "-"), true
}

// END

func main() {
	fmt.Println(child("00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01", "aaaaaaaaaaaaaaaa"))
	fmt.Println(child("nope", "aaaaaaaaaaaaaaaa"))
}
```

### Check
Q: Which HTTP header carries the W3C trace context?
T: short
A: traceparent
E: Format: `00-<trace-id>-<span-id>-<flags>`.

Q: What must a service do with an incoming trace ID when it calls another service?
T: mcq
- [ ] Generate a brand new trace ID
- [x] Forward it (with a new child span) so the whole request stays one trace
- [ ] Drop it
- [ ] Encrypt it
E: Propagation is what makes a trace span multiple services.

## Service Discovery
slug: service-discovery
status: planned
