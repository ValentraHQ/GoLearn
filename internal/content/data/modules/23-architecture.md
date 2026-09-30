# Go Architecture
id: architecture
number: 23
track: architecture
paths: pro
skill: architecture
requires: rest-api
summary: Practical architecture: layers, clean and hexagonal design, dependency injection and boundaries.

## Layered Architecture
slug: layered-architecture
minutes: 8
objectives: Describe handler, service, repository and domain layers; Keep each layer's responsibility narrow; Lay out a project accordingly
takeaways: Layers: transport (handler) → service (use cases) → repository (storage), with domain types shared; Each layer depends only on the layer below, through interfaces; Keep it as simple as the problem allows — don't add layers for their own sake

### Concept
```
cmd/
  api/main.go            # wiring only
internal/
  domain/                # entities, errors: no imports of other layers
  handler/               # HTTP: decode, call service, encode
  service/               # business rules, orchestration
  repository/            # storage implementations (postgres, memory)
  middleware/
pkg/                     # optional: reusable public packages
migrations/
tests/
```
- **domain** — plain types and rules (`Order`, `ErrOutOfStock`).
- **service** — use cases (`PlaceOrder`), depends on repository *interfaces* (defined by the service).
- **handler** — the delivery mechanism; translates HTTP ⇄ service calls and errors ⇄ status codes.
- **repository** — implements those interfaces with SQL, an API, or memory.

Practical architecture is about **dependency direction and testability**, not ceremony. A 300-line service needs two packages, not five.

### Example
```go
package main

import (
	"errors"
	"fmt"
)

// domain
type Order struct {
	ID    int
	Total int
}

var ErrEmpty = errors.New("order total must be positive")

// service depends on an interface it defines
type Store interface{ Save(Order) error }

type Service struct{ store Store }

func (s Service) Place(total int) (Order, error) {
	if total <= 0 {
		return Order{}, ErrEmpty
	}
	o := Order{ID: 1, Total: total}
	return o, s.store.Save(o)
}

// repository (in-memory)
type memStore struct{ saved []Order }

func (m *memStore) Save(o Order) error { m.saved = append(m.saved, o); return nil }

func main() {
	st := &memStore{}
	svc := Service{st}
	fmt.Println(svc.Place(10))
	fmt.Println(svc.Place(0))
	fmt.Println(len(st.saved))
}
```

### Exercise
Add `Get(id int) (Order, error)` to `Store`, implement it on `memStore` (return `ErrNotFound` when missing) and a `Service.Find(id)` that calls it. Print the found total and the not-found error.
```text expect
25 <nil>
{0 0} not found
```
```go solution
package main

import (
	"errors"
	"fmt"
)

type Order struct {
	ID    int
	Total int
}

var ErrNotFound = errors.New("not found")

// BEGIN
type Store interface {
	Save(Order) error
	Get(id int) (Order, error)
}

type Service struct{ store Store }

func (s Service) Find(id int) (Order, error) { return s.store.Get(id) }

// END

type memStore struct{ byID map[int]Order }

func (m *memStore) Save(o Order) error {
	if m.byID == nil {
		m.byID = map[int]Order{}
	}
	m.byID[o.ID] = o
	return nil
}

// BEGIN
func (m *memStore) Get(id int) (Order, error) {
	if o, ok := m.byID[id]; ok {
		return o, nil
	}
	return Order{}, ErrNotFound
}

// END

func main() {
	st := &memStore{}
	_ = st.Save(Order{ID: 7, Total: 25})
	svc := Service{st}
	o, err := svc.Find(7)
	fmt.Println(o.Total, err)
	fmt.Println(svc.Find(9))
}
```

### Check
Q: In a layered design, which way do dependencies point?
T: mcq
- [ ] Repository → handler
- [x] Handler → service → repository (each depends on the layer below, via interfaces)
- [ ] Every layer depends on every other
- [ ] Domain → handler
E: Higher-level policy shouldn't depend on lower-level details.

Q: A tiny service should be split into as many layers as possible.
T: tf
A: false
E: Add structure when complexity demands it; extra layers add indirection without benefit.

## Clean Architecture
slug: clean-architecture
status: planned

## Hexagonal Architecture
slug: hexagonal-architecture
status: planned

## Dependency Injection
slug: dependency-injection
minutes: 8
objectives: Pass dependencies into constructors; Wire everything in main (composition root); Know when a DI framework is (not) needed
takeaways: DI in Go is just passing dependencies as constructor arguments — no framework required; Build the object graph once in main and hand it out; Prefer small interfaces defined by the consumer

### Concept
```go norun
func main() {
	cfg := mustLoadConfig()
	db := mustOpenDB(cfg.DatabaseURL)
	repo := postgres.NewOrderRepo(db)
	svc := service.NewOrders(repo, clock.System{})
	h := handler.New(svc)
	log.Fatal(http.ListenAndServe(cfg.Addr, h.Routes()))
}
```
`main` is the **composition root**: the only place that knows concrete types. Everything else receives interfaces via constructors, so tests substitute fakes. Avoid globals, `init()` side effects and service locators. DI frameworks (`wire`, `fx`) exist for very large graphs; most Go services are clearer with hand-written wiring.

### Example
```go
package main

import "fmt"

type Clock interface{ Now() string }
type Greeter struct{ clock Clock }

func NewGreeter(c Clock) *Greeter { return &Greeter{c} }
func (g *Greeter) Greet(name string) string { return "hello " + name + " at " + g.clock.Now() }

type fixedClock string

func (f fixedClock) Now() string { return string(f) }

func main() {
	g := NewGreeter(fixedClock("noon")) // wiring in main
	fmt.Println(g.Greet("ada"))
}
```

### Exercise
Write constructor `NewNotifier(s Sender) *Notifier`, inject a `fakeSender` from `main`, and print what it recorded after `Notify("hi")`.
```text expect
[sent: hi]
```
```go solution
package main

import "fmt"

type Sender interface{ Send(msg string) }

// BEGIN
type Notifier struct{ sender Sender }

func NewNotifier(s Sender) *Notifier { return &Notifier{sender: s} }

func (n *Notifier) Notify(msg string) { n.sender.Send(msg) }

// END

type fakeSender struct{ log []string }

func (f *fakeSender) Send(msg string) { f.log = append(f.log, "sent: "+msg) }

func main() {
	f := &fakeSender{}
	NewNotifier(f).Notify("hi")
	fmt.Println(f.log)
}
```

### Check
Q: Where should concrete implementations be chosen?
T: mcq
- [ ] Inside every service constructor
- [x] In main — the composition root
- [ ] In init functions
- [ ] In interfaces
E: Keeping the choice in one place keeps everything else decoupled.

Q: Go requires a DI framework for good architecture.
T: tf
A: false
E: Constructor injection with hand-written wiring is idiomatic and sufficient for most services.

## Repository Pattern
slug: repository-pattern
minutes: 7
objectives: Define repository interfaces in terms of the domain; Return domain errors and hide storage details; Decide when a repository is worth it
takeaways: A repository is a domain-shaped facade over storage: methods like Save(order) and FindByID(id); Define the interface in the consumer's package, keep implementations in a repository package; Skip the abstraction if there's only one trivial query — introduce it when tests or a second backend need it

### Concept
```go norun
// in package service (the consumer):
type OrderRepository interface {
	Save(ctx context.Context, o *domain.Order) error
	FindByID(ctx context.Context, id domain.OrderID) (*domain.Order, error) // domain.ErrNotFound
}
```
Implementations (`internal/repository/postgres`, `.../memory`) translate between the domain and storage and map errors (`sql.ErrNoRows` → `domain.ErrNotFound`). Benefits: business tests without a database; swap/dual backends; a single place for SQL. Costs: an extra layer — worth it once logic and persistence both have real complexity. Don't build a generic `Repository[T]` with `Find(query any)`: domain-specific methods (`FindOverdueInvoices`) communicate intent.

### Example
```go
package main

import (
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("not found")

type Order struct {
	ID    int
	Total int
}

type OrderRepository interface {
	FindByID(id int) (Order, error)
}

type memRepo map[int]Order

func (m memRepo) FindByID(id int) (Order, error) {
	if o, ok := m[id]; ok {
		return o, nil
	}
	return Order{}, ErrNotFound
}

func main() {
	var repo OrderRepository = memRepo{1: {1, 30}}
	fmt.Println(repo.FindByID(1))
	_, err := repo.FindByID(2)
	fmt.Println(errors.Is(err, ErrNotFound))
}
```

### Exercise
Add a domain-specific method `Overdue(cutoff int) []Order` to the interface and `memRepo` returning orders whose `Total` exceeds `cutoff`, sorted by ID. Print the IDs.
```text expect
[1 3]
```
```go solution
package main

import (
	"fmt"
	"sort"
)

type Order struct {
	ID    int
	Total int
}

// BEGIN
type OrderRepository interface {
	Overdue(cutoff int) []Order
}

// END

type memRepo map[int]Order

// BEGIN
func (m memRepo) Overdue(cutoff int) []Order {
	var out []Order
	for _, o := range m {
		if o.Total > cutoff {
			out = append(out, o)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// END

func main() {
	var repo OrderRepository = memRepo{1: {1, 90}, 2: {2, 10}, 3: {3, 75}}
	var ids []int
	for _, o := range repo.Overdue(50) {
		ids = append(ids, o.ID)
	}
	fmt.Println(ids)
}
```

### Check
Q: Where should a repository interface usually be defined?
T: mcq
- [ ] Next to the SQL implementation
- [x] In the package that consumes it (e.g. service)
- [ ] In main
- [ ] In a shared "interfaces" package
E: The consumer knows the minimal behaviour it needs.

Q: A generic `Find(query any)` repository method communicates intent well.
T: tf
A: false
E: Domain-specific methods like `FindOverdueInvoices` make intent explicit and queries reviewable.

## Service Layer
slug: service-layer
status: planned

## Domain Layer
slug: domain-layer
status: planned

## Package Design
slug: package-design
status: planned

## Dependency Direction
slug: dependency-direction
minutes: 8
objectives: Apply the dependency rule: source code dependencies point inward toward the domain; Use interfaces to invert dependencies on infrastructure; Detect illegal imports
takeaways: Stable, high-level policy (domain, use cases) must not import volatile details (HTTP, SQL drivers); Invert dependencies with interfaces owned by the inner layer; Enforce with import rules (depguard, go-arch-lint, or tests)

### Concept
```
handler ──▶ service ──▶ domain
   │           ▲
   ▼           │ (implements service.Repository)
repository ────┘
```
The **dependency rule**: imports point toward the domain. The domain imports nothing from your project; `service` imports `domain`; `handler` and `repository` import `service`/`domain`, never the reverse. `service` defines the `Repository` interface it needs; `repository/postgres` *implements* it — the arrow of **source** dependency points inward even though the runtime call flows outward.

Guard it: a test that walks imports with `go/parser`, or linters (`depguard`, `go-arch-lint`). Import cycles are Go's built-in alarm for tangled dependencies.

### Example
```go
package main

import (
	"fmt"
	"strings"
)

// allowed lists which packages each layer may import.
var allowed = map[string][]string{
	"domain":     {},
	"service":    {"domain"},
	"handler":    {"service", "domain"},
	"repository": {"domain"},
}

func violates(from, to string) bool {
	for _, a := range allowed[from] {
		if a == to {
			return false
		}
	}
	return true
}

func main() {
	fmt.Println(violates("service", "domain"), violates("domain", "handler"))
	fmt.Println(strings.Repeat("-", 3))
}
```

### Exercise
Write `check(imports map[string][]string) []string` returning violations formatted `<from> -> <to>` for every import not permitted by `allowed`, sorted alphabetically (by the formatted string).
```text expect
[domain -> handler service -> repository]
```
```go solution
package main

import (
	"fmt"
	"sort"
)

var allowed = map[string][]string{
	"domain":     {},
	"service":    {"domain"},
	"handler":    {"service", "domain"},
	"repository": {"domain"},
}

// BEGIN
func check(imports map[string][]string) []string {
	var out []string
	for from, tos := range imports {
		for _, to := range tos {
			ok := false
			for _, a := range allowed[from] {
				if a == to {
					ok = true
				}
			}
			if !ok {
				out = append(out, from+" -> "+to)
			}
		}
	}
	sort.Strings(out)
	return out
}

// END

func main() {
	fmt.Println(check(map[string][]string{
		"handler":    {"service"},
		"service":    {"domain", "repository"},
		"domain":     {"handler"},
		"repository": {"domain"},
	}))
}
```

### Check
Q: According to the dependency rule, which package should the domain import?
T: mcq
- [ ] handler
- [ ] repository
- [x] None of the project's outer layers
- [ ] All of them
E: The innermost layer is independent of delivery and storage details.

Q: How can a service use a database without importing the SQL package?
T: mcq
- [ ] It can't
- [x] It defines a repository interface, and the SQL package implements it
- [ ] By using reflection
- [ ] By using globals
E: Dependency inversion: the inner layer owns the abstraction.

## Application Boundaries
slug: application-boundaries
status: planned
