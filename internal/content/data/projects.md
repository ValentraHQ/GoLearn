# Projects

## Hello Go CLI
id: hello-go-cli
level: beginner
module: fundamentals
skills: fundamentals
hours: 1

### Summary
A tiny command-line greeter that takes a name and prints a message — your first real Go program.

### Description
Build `hello`, a CLI that greets people. It reads a name from the command line (or asks for it on standard input), validates it, and prints a friendly message. This project exercises packages, imports, variables, constants, string handling and basic I/O.

### Requirements
- `hello Ada` prints `Hello, Ada!`
- With no argument, prompt `What is your name?` on stdout and read one line from stdin
- Empty names print an error to stderr and exit with status 1
- Greeting text lives in a named constant

### Tasks
- Create a module with `go mod init` and a `main.go` :: Use a module path like example.com/hello
- Read the name from `os.Args` :: `os.Args[0]` is the program name
- Fall back to reading a line from stdin with `bufio.Scanner`
- Trim whitespace and reject empty names :: Print the error with `fmt.Fprintln(os.Stderr, ...)`
- Build the binary with `go build` and run it :: Try it with and without arguments

### Stretch
- Add a `-shout` flag using the standard `flag` package
- Greet in several languages chosen by an environment variable

## CLI Calculator
id: cli-calculator
level: beginner
module: control-flow
skills: functions
hours: 2

### Summary
A calculator that evaluates expressions like `calc 12 + 30` and supports a REPL mode.

### Description
Build a calculator that parses two numbers and an operator, dispatches to small functions, and reports errors clearly (division by zero, unknown operator, invalid numbers). Add an interactive mode that keeps prompting until the user types `quit`.

### Requirements
- Operators: `+ - * /` and `%` for integers
- One function per operation, chosen with a `switch`
- Functions return `(float64, error)`; `main` prints errors to stderr
- REPL mode when no arguments are given
- A variadic `sum` command: `calc sum 1 2 3 4`

### Tasks
- Parse arguments into two numbers and an operator :: Use `strconv.ParseFloat`
- Implement add, subtract, multiply, divide as functions returning `(float64, error)`
- Dispatch on the operator with `switch`
- Handle divide-by-zero and unknown operators with errors
- Add the variadic `sum` and `avg` commands
- Add a REPL loop with `bufio.Scanner` that exits on `quit`

### Stretch
- Support operator precedence with a shunting-yard parser
- Keep a history of results accessible as `ans`

## Contact Manager CLI
id: contact-manager-cli
level: beginner
module: data-structures
skills: data-structures
hours: 3

### Summary
Store, search and list contacts using slices and maps, persisted to a file.

### Description
Manage a small address book from the terminal. Contacts have a name, email and phone. Keep them in a slice and index them by name in a map for fast lookups. Persist to a plain-text file so data survives restarts.

### Requirements
- Commands: `add`, `list`, `find <name>`, `delete <name>`
- Names are unique (case-insensitive); adding a duplicate is rejected
- Contacts are listed sorted by name
- Data is saved to and loaded from a file on start and exit

### Tasks
- Define the contact representation and an in-memory store :: A slice for order, a map for lookup
- Implement add with duplicate detection
- Implement find and delete using the map
- Implement list sorted with `sort.Slice`
- Load and save contacts to a file, one per line
- Write table-driven tests for add/find/delete

### Stretch
- Support fuzzy search by substring
- Export to CSV with `encoding/csv`

## Employee Management System
id: employee-management-system
level: intermediate
module: structs
skills: structs, pointers
hours: 4

### Summary
Model employees, managers and departments with structs, methods and composition.

### Description
Design a small HR model. `Employee` holds core data; `Manager` embeds `Employee` and adds direct reports. Departments own employees and compute payroll totals. The goal is to practise struct design, value versus pointer receivers, and embedding.

### Requirements
- Employee, Manager (embedding Employee), Department types
- Methods: `Raise(percent)`, `FullName()`, `Department.Payroll()`
- Pointer receivers for mutating methods, value receivers for read-only ones
- A report that prints each department with its headcount and payroll

### Tasks
- Define `Employee`, `Manager` and `Department` structs
- Add constructors that validate input :: e.g. `NewEmployee(name string, salary int) (*Employee, error)`
- Implement mutating methods with pointer receivers
- Implement payroll and headcount calculations
- Print a formatted report with `fmt.Printf` and `text/tabwriter`
- Write tests that confirm mutation happens through pointers

### Stretch
- Detect reporting cycles in the manager hierarchy
- Persist the model as JSON

## Pluggable Storage System
id: pluggable-storage-system
level: intermediate
module: interfaces
skills: interfaces
hours: 4

### Summary
A key-value store with interchangeable backends behind a small interface.

### Description
Define a `Store` interface, then implement in-memory, file-based and (optionally) SQLite backends. Application code depends only on the interface, so backends can be swapped in configuration and faked in tests.

### Requirements
- `Store` interface: `Get`, `Put`, `Delete`, `List`
- At least two implementations that pass one shared conformance test suite
- Selection of the backend via a flag or environment variable
- Errors: a sentinel `ErrNotFound` shared by all backends

### Tasks
- Design the smallest useful `Store` interface
- Implement the in-memory backend with a mutex
- Implement a file-backed backend
- Write a reusable conformance test that takes any `Store`
- Add a CLI (`kv get|put|delete|list`) that uses the interface only
- Add a decorator that logs every call :: Wrap a `Store` in another `Store`

### Stretch
- Add a caching decorator with expiry
- Add a backend that speaks HTTP to a remote store

## Reliable File Processor
id: reliable-file-processor
level: intermediate
module: errors
skills: errors
hours: 3

### Summary
Process a directory of files, surviving bad input with wrapped errors and a final report.

### Description
Read every `.csv` in a directory, validate rows, and write a summary. One bad file or row must never crash the run: collect errors with context, keep going, and exit non-zero only if something failed. Practice wrapping (`%w`), `errors.Is/As`, custom error types and `defer`.

### Requirements
- Custom `RowError` type carrying file, line and cause
- Errors wrapped with context at every layer
- `errors.Join` (or a custom aggregate) for the final report
- Files always closed via `defer`; panics from parsing are recovered per file
- Exit status 0 only when every file processed cleanly

### Tasks
- Walk the directory and list candidate files
- Parse rows and return typed errors with line numbers
- Wrap errors with `fmt.Errorf("...: %w", err)` as they move up the stack
- Continue after errors, collecting them
- Recover from panics per file so one bad file can't stop the batch
- Print a summary and set the exit code

### Stretch
- Retry transient errors with backoff
- Write rejected rows to a `.rejects` file

## Reusable Go Library
id: reusable-go-library
level: intermediate
module: modules
skills: packages
hours: 4

### Summary
Design, version and publish a small library with a clean public API.

### Description
Create a module (for example a `slug` or `retry` library) with a small exported API, documentation examples, tests, and semantic version tags. Use it from a second module through a `replace` directive, then via a real version tag.

### Requirements
- Module with a valid path and `go.mod`
- Exported API kept minimal; internals in `internal/`
- Godoc comments and runnable `Example` functions
- Tests pass with `go test ./...`
- Tagged `v0.1.0` (and a documented breaking change bumping to `v0.2.0`)

### Tasks
- Choose a focused problem and design the public API on paper
- Create the module and package layout with `internal/`
- Write doc comments and `Example` functions
- Add tests and run `go vet` and `go test`
- Consume the library from another module using `replace`
- Tag a version and consume it by version

### Stretch
- Add a `CHANGELOG.md` and CI that runs tests on every push
- Add generics to the API when it improves clarity

## System Information CLI
id: system-information-cli
level: intermediate
module: files-os
skills: stdlib
hours: 3

### Summary
Report host details, environment and process info from the command line.

### Description
Build `sysinfo`, a CLI that prints OS/arch, hostname, working directory, environment variables (with filtering and secret masking), disk usage for a path, and handles SIGINT gracefully.

### Requirements
- Subcommands: `env`, `disk <path>`, `proc`, `watch`
- Mask values of variables whose names contain `KEY`, `TOKEN` or `SECRET`
- `watch` prints a line every second and exits cleanly on Ctrl+C
- Errors go to stderr with a non-zero exit code

### Tasks
- Print runtime and host information with `runtime` and `os`
- Implement `env` with filtering and masking
- Walk a directory to compute total size for `disk`
- Implement `watch` with a ticker and `signal.NotifyContext`
- Run an external command with `os/exec` and show its exit code
- Add tests for masking and size calculation

### Stretch
- Output as JSON with `--json`
- Add a `--interval` flag to `watch`

## JSON Data Processing Tool
id: json-data-processing-tool
level: intermediate
module: json
skills: stdlib
hours: 3

### Summary
Filter, transform and validate large JSON files without loading them entirely.

### Description
Build `jq-lite`: read JSON from a file or stdin, select fields, filter records, and emit JSON or CSV. Stream large arrays with `json.Decoder` and validate input with helpful error positions.

### Requirements
- Stream a top-level array of objects with `json.Decoder`
- `--select a,b.c` picks fields; `--where key=value` filters rows
- Output as JSON lines or CSV
- Report malformed JSON with an offset
- Unknown fields are reported in strict mode

### Tasks
- Define types with struct tags for the sample dataset
- Decode a stream token by token
- Implement `--where` filtering
- Implement `--select` including nested fields
- Emit CSV with `encoding/csv`
- Add strict validation with `DisallowUnknownFields`

### Stretch
- Support jq-style paths like `.items[0].name`
- Add gzip input support

## Concurrent Job Processing System
id: concurrent-job-processing-system
level: advanced
module: concurrency
skills: concurrency
hours: 8

### Summary
A worker pool with queues, timeouts, cancellation and graceful shutdown.

### Description
Build a job runner that accepts jobs, runs them on a bounded pool of workers, enforces per-job timeouts, retries failures with backoff, and shuts down gracefully on SIGTERM without losing in-flight work. Prove correctness with `-race` tests.

### Requirements
- Configurable worker count and queue size
- Per-job `context` timeout and global cancellation
- Results delivered on a channel; metrics via `sync/atomic`
- Graceful shutdown drains in-flight jobs within a deadline
- No goroutine leaks (verified in tests)

### Architecture
```text
Producer → [ Queue ] → Worker 1 ─┐
                     → Worker 2 ─┼→ [ Results ] → Collector
                     → Worker N ─┘
             ↑ context cancel / shutdown signal
```

### Tasks
- Define the `Job` and `Result` types
- Implement a fixed-size worker pool with a job channel
- Add per-job timeouts with `context.WithTimeout`
- Add retries with exponential backoff and jitter
- Implement graceful shutdown with `signal.NotifyContext` and a drain deadline
- Write `-race` tests, including a goroutine-leak check
- Add counters (submitted, succeeded, failed) with `atomic`

### Stretch
- Priority queues
- Rate limiting workers with a token bucket

## HTTP Monitoring Tool
id: http-monitoring-tool
level: advanced
module: networking
skills: http, networking
hours: 5

### Summary
Poll a list of URLs concurrently and report status, latency and TLS expiry.

### Description
Build `uptime`, a tool that reads targets from a config file, checks them on an interval with a shared `http.Client`, records latency and status, warns on slow responses or expiring certificates, and serves a small status page.

### Requirements
- Shared `http.Client` with timeouts and connection reuse
- Concurrent checks with a bounded number of goroutines
- Records status code, latency and TLS certificate expiry
- A `/status` JSON endpoint and a plain-text summary
- Graceful shutdown

### Tasks
- Define the config format and load targets
- Implement a single check with timeouts and redirects policy
- Run checks concurrently on a ticker
- Extract TLS certificate expiry from `resp.TLS`
- Serve `/status` with `net/http`
- Add graceful shutdown and tests using `httptest.Server`

### Stretch
- Send alerts to a webhook after N consecutive failures
- Export Prometheus metrics

## Production REST API
id: production-rest-api
level: advanced
module: rest-api
skills: rest, architecture
hours: 10

### Summary
A layered REST API with validation, pagination, authentication and rate limiting.

### Description
Build a `tasks` API following handler → service → repository layering. Include request validation, consistent JSON errors, pagination and filtering, JWT authentication, role-based authorisation, request logging middleware and per-client rate limiting.

### Requirements
- Resources with CRUD, pagination (`page`, `page_size`) and filtering
- Consistent error envelope and correct status codes
- JWT auth with `admin` and `member` roles
- Middleware: request ID, logging, panic recovery, rate limiting
- Versioned routes under `/api/v1`

### Architecture
```text
Client → Router → Middleware → Handler → Service → Repository → Database
```

### Tasks
- Lay out `cmd/`, `internal/handler`, `service`, `repository`, `domain`
- Implement CRUD handlers with `net/http` routing patterns
- Add JSON validation and a consistent error type
- Implement pagination, filtering and sorting
- Add JWT login and an auth middleware with roles
- Add logging, recovery, request-ID and rate-limit middleware
- Document the API with an OpenAPI file

### Stretch
- Idempotency keys for POST
- ETag / conditional requests

## Go + PostgreSQL Backend
id: go-postgres-backend
level: advanced
module: databases
skills: databases
hours: 8

### Summary
A repository layer on PostgreSQL with migrations, transactions and tests.

### Description
Back the tasks service with PostgreSQL. Manage the schema with migrations, use `database/sql` with a tuned connection pool, run multi-statement operations in transactions, and test the repository against a real database.

### Requirements
- Migrations applied by a small runner (or `golang-migrate`)
- Repository interface with a PostgreSQL implementation
- Parameterised queries only; no string-built SQL
- Transactions for operations touching multiple rows
- Integration tests against a disposable database

### Tasks
- Design the schema and write up/down migrations
- Configure `sql.DB` pool size and lifetimes
- Implement the repository with prepared/parameterised queries
- Wrap a multi-step operation in a transaction with rollback on error
- Add indexes for the query patterns you support
- Write integration tests using Docker Compose or testcontainers

### Stretch
- Generate the data layer with `sqlc`
- Add keyset pagination

## Fully Tested REST API
id: fully-tested-rest-api
level: advanced
module: testing
skills: testing
hours: 6

### Summary
Bring an API to high confidence with unit, HTTP, integration and race tests.

### Description
Take your REST API and give it a real safety net: table-driven unit tests for services, `httptest` tests for handlers, fakes for repositories, integration tests for the database layer, benchmarks for hot paths, and CI running `go test -race -cover ./...`.

### Requirements
- Service logic covered by table-driven tests with fakes
- Handlers tested with `httptest.NewRecorder`
- At least one integration test hitting a real database
- `go test -race ./...` passes
- Coverage above 80% on `service` and `handler`

### Tasks
- Introduce interfaces where tests need seams
- Write table-driven tests for services
- Write handler tests for success and each error path
- Add an integration test suite behind a build tag
- Add benchmarks for serialisation or validation
- Wire CI to run tests with `-race` and `-cover`

### Stretch
- Fuzz the request parser with `go test -fuzz`
- Add golden-file tests for responses

## Professional DevOps CLI
id: professional-devops-cli
level: advanced
module: cli
skills: cli
hours: 8

### Summary
A Cobra-based CLI with subcommands, config, structured output and shell completion.

### Description
Build `opsctl`, a CLI with subcommands (`check`, `deploy status`, `logs analyze`), layered configuration (flags → env → file), structured logging, JSON/table output, meaningful exit codes and generated shell completion.

### Requirements
- Subcommands with help text and examples
- Config precedence: flag > environment > config file > default
- `--output json|table` on every command
- Exit codes: 0 ok, 1 failure, 2 usage error
- Shell completion for bash and zsh

### Tasks
- Scaffold the command tree with Cobra
- Implement layered configuration
- Implement `check` (HTTP health checks) with concurrent probes
- Implement a log analyzer that reads files or stdin
- Add table and JSON renderers
- Generate completion scripts and document installation
- Add tests for commands using an in-memory output buffer

### Stretch
- Interactive prompts for missing values
- Self-update command

## Containerized Go API
id: containerized-go-api
level: advanced
module: docker
skills: docker
hours: 4

### Summary
Package an API and PostgreSQL with Docker Compose using a minimal, hardened image.

### Description
Containerise the REST API with a multi-stage Dockerfile producing a small non-root image, wire it to PostgreSQL with Compose, configure it through environment variables and add health checks.

### Requirements
- Multi-stage build; final image under 30 MB
- Runs as a non-root user with a read-only root filesystem
- Configuration only through environment variables
- `docker compose up` starts API + PostgreSQL with health checks
- Graceful shutdown on SIGTERM

### Tasks
- Write a multi-stage Dockerfile with a static binary
- Use a minimal base (distroless or scratch) and a non-root user
- Add a `.dockerignore`
- Write `compose.yaml` with a database, API and health checks
- Run migrations on start
- Scan the image for vulnerabilities

### Stretch
- Multi-arch builds with buildx
- Sign the image and generate an SBOM

## Kubernetes Controller
id: kubernetes-controller
level: advanced
module: kubernetes
skills: kubernetes
hours: 12

### Summary
Build a controller that reconciles a custom resource into a Deployment.

### Description
Define a `WebApp` custom resource. Write a controller that watches `WebApp` objects and keeps a matching Deployment and Service in sync, updates status, handles deletion with finalizers and is safe to run repeatedly.

### Requirements
- CRD with a validation schema
- Controller using controller-runtime (or client-go informers)
- Idempotent reconciliation
- Status subresource reports readiness
- Owner references so children are garbage-collected

### Architecture
```text
Custom Resource → Controller → Kubernetes API → Deployment → Pods
```

### Tasks
- Design the `WebApp` spec and status
- Generate the CRD and Go types
- Implement `Reconcile` to create or update the Deployment
- Set owner references and update status conditions
- Add a finalizer for external cleanup
- Test with envtest and deploy to a kind cluster

### Stretch
- Webhook validation and defaulting
- Leader election and metrics

## Observable Production Go Service
id: observable-go-service
level: advanced
module: observability
skills: observability
hours: 6

### Summary
Add metrics, structured logs, traces and health endpoints to a service.

### Description
Instrument the API for production: Prometheus metrics, `slog` structured logs with correlation IDs, OpenTelemetry traces and separate liveness and readiness endpoints, plus a Grafana dashboard.

### Requirements
- `/metrics` exposes request count, latency histogram and in-flight gauge
- JSON logs include request ID and trace ID
- Spans for HTTP handlers and database calls
- `/livez` and `/readyz` with different semantics
- Dashboard JSON committed to the repo

### Tasks
- Add a metrics middleware and business metrics
- Configure `slog` with request-scoped attributes
- Propagate request IDs and trace context
- Add OpenTelemetry tracing to handlers and DB calls
- Implement liveness and readiness probes
- Build a Grafana dashboard and alert rules

### Stretch
- Exemplars linking metrics to traces
- SLO burn-rate alerts

## File Manager
id: file-manager
level: beginner
module: files-os
skills: stdlib
hours: 3

### Summary
A terminal file manager that lists, copies, moves and deletes files safely.

### Description
Build `fm`: list a directory with sizes and permissions, copy and move files, delete with confirmation, and search by name. Focus on the `os`, `io` and `path/filepath` packages and careful error handling.

### Requirements
- `ls`, `cp`, `mv`, `rm`, `find` subcommands
- Refuse to overwrite without `--force`
- Confirm before deleting directories
- Human-readable sizes and permissions

### Tasks
- Implement `ls` with `os.ReadDir`
- Implement `cp` using `io.Copy` and preserving mode
- Implement `mv` with a rename fallback to copy+delete
- Implement `rm` with confirmation
- Implement `find` with `filepath.WalkDir`
- Add tests using `t.TempDir()`

### Stretch
- Interactive TUI mode
- Trash instead of permanent delete

## Todo CLI
id: todo-cli
level: beginner
module: structs
skills: structs
hours: 2

### Summary
A todo list stored as JSON with add, done, list and remove commands.

### Description
Track tasks from the terminal. Practise structs, slices, methods and JSON persistence in a small, complete program.

### Requirements
- `todo add "text"`, `todo list`, `todo done <id>`, `todo rm <id>`
- Tasks have an id, text, done flag and created time
- Stored in `~/.todo.json` (path overridable by environment variable)
- IDs never get reused

### Tasks
- Define a `Task` struct and a `List` type with methods
- Implement add, done and remove
- Persist and load with `encoding/json`
- Parse subcommands from `os.Args`
- Print a neat list with done markers
- Add tests using a temporary file

### Stretch
- Due dates and sorting
- Tags and filtering

## URL Shortener
id: url-shortener
level: intermediate
module: rest-api
skills: rest, http
hours: 5

### Summary
An HTTP service that shortens URLs, redirects, and counts visits.

### Description
Build a shortener with `POST /shorten`, `GET /{code}` redirects and `GET /stats/{code}`. Practise routing, validation, concurrency-safe storage, and unit tests with `httptest`.

### Requirements
- Validate URLs (scheme and host required)
- Generate unique, short, URL-safe codes
- Redirect with 302 and count visits atomically
- In-memory store first, then an interface allowing SQLite or PostgreSQL

### Tasks
- Define the store interface and in-memory implementation
- Generate codes safely (crypto/rand)
- Implement the three endpoints with `net/http`
- Validate input and return consistent JSON errors
- Add visit counting without data races
- Write `httptest` tests and run with `-race`

### Stretch
- Custom aliases and expiry
- Rate limiting per IP

## Authentication Service
id: authentication-service
level: intermediate
module: security
skills: security
hours: 6

### Summary
Registration, login, password hashing and JWT issuing done securely.

### Description
Build a small auth service with registration, login, refresh tokens and logout. Store password hashes with bcrypt or argon2, issue short-lived JWTs, rotate refresh tokens, rate-limit login attempts and add security headers.

### Requirements
- Passwords hashed with bcrypt or argon2id
- Access tokens expire in minutes; refresh tokens are revocable
- Constant-time credential checks; no user enumeration
- Rate limiting and lockout on repeated failures
- Secure headers and HttpOnly cookies

### Tasks
- Design the user and token tables
- Implement registration with input validation
- Implement login with hashing and constant-time comparison
- Issue and verify JWTs with an expiry and audience
- Implement refresh token rotation and revocation
- Add rate limiting and security headers
- Write negative tests (bad password, expired token, tampering)

### Stretch
- TOTP two-factor authentication
- OIDC login with an external provider

## Log Processor
id: log-processor
level: intermediate
module: concurrency
skills: concurrency, stdlib
hours: 4

### Summary
Parse large log files concurrently and produce aggregated statistics.

### Description
Read one or many log files, parse lines in parallel with a worker pool, aggregate counts by level, status and endpoint, and print the top offenders. Compare a sequential implementation with the concurrent one using benchmarks.

### Requirements
- Streaming reads — never load the whole file
- Worker pool sized to `runtime.NumCPU()`
- Deterministic output despite concurrency
- Benchmarks comparing sequential and parallel versions

### Tasks
- Define the log format and a parser with tests
- Read lines with `bufio.Scanner` (raise the buffer limit)
- Fan out batches of lines to workers, fan in partial results
- Merge maps deterministically
- Print top-N tables
- Benchmark and profile

### Stretch
- Follow a growing file like `tail -f`
- Output histograms of latency percentiles

## Worker Queue
id: worker-queue
level: intermediate
module: concurrency
skills: concurrency
hours: 4

### Summary
A persistent-style job queue with priorities, retries and a dead-letter list.

### Description
Implement an in-process queue that supports enqueueing typed jobs, priorities, delayed retries with backoff and a dead-letter list after N failures. Expose an HTTP API to enqueue and inspect jobs.

### Requirements
- Priority ordering and FIFO within a priority
- Retries with exponential backoff and a maximum attempt count
- Dead-letter list of permanently failed jobs
- Safe under concurrent producers and consumers (`-race`)

### Tasks
- Design the job type and queue interface
- Implement a heap-based priority queue with a mutex
- Add worker goroutines that pull and execute jobs
- Implement retry scheduling with timers
- Add the dead-letter list
- Expose `POST /jobs` and `GET /jobs` over HTTP

### Stretch
- Persist the queue to disk
- Visibility timeouts like SQS

## Monitoring Agent
id: monitoring-agent
level: advanced
module: observability
skills: observability, networking
hours: 8

### Summary
A lightweight agent that collects host metrics and ships them to a collector.

### Description
Build an agent that samples CPU, memory, disk and network from `/proc`, batches samples, retries delivery with backoff, buffers on failure and exposes its own health and metrics endpoints.

### Requirements
- Configurable sampling interval
- Batching with a bounded in-memory buffer
- Delivery with retry and backoff; drop-oldest policy when the buffer is full
- Own `/healthz` and `/metrics`
- Clean shutdown that flushes the buffer

### Tasks
- Parse `/proc` files for CPU and memory
- Design the sample and batch types
- Implement the collector loop with a ticker
- Implement the shipper with retries and a bounded buffer
- Expose health and Prometheus metrics
- Flush on shutdown and test with a fake collector

### Stretch
- Plugin-style collectors behind an interface
- Compression and TLS client authentication

## Microservice Platform
id: microservice-platform
level: advanced
module: microservices
skills: microservices
hours: 16

### Summary
An API gateway plus user, flight and notification services with async events.

### Description
Build a small platform: an API gateway routes requests to a user service and a flight service; booking a flight publishes an event that the notification service consumes. Add timeouts, retries, idempotency keys, circuit breaking, tracing and Docker Compose orchestration.

### Requirements
- Gateway with routing, auth and rate limiting
- Services communicate over HTTP/gRPC and a message queue
- Idempotent booking endpoint
- Timeouts, retries and a circuit breaker on outbound calls
- Distributed tracing across services

### Architecture
```text
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

### Tasks
- Define service boundaries and contracts
- Build the gateway with reverse proxying
- Implement the user and flight services with their own databases
- Publish and consume booking events
- Add idempotency keys to booking
- Add timeouts, retries and a circuit breaker
- Add tracing and a Compose file to run everything

### Stretch
- Saga pattern for compensating actions
- Service discovery instead of static addresses

## Kubernetes CLI
id: kubernetes-cli
level: advanced
module: kubernetes
skills: kubernetes, cli
hours: 8

### Summary
A kubectl-style helper that lists workloads and diagnoses common problems.

### Description
Use client-go to build `kubectl-helper`: list pods and deployments with health at a glance, explain why a pod is not ready (image pull errors, crash loops, unschedulable), and stream events.

### Requirements
- Uses kubeconfig and the current context
- `pods`, `deployments`, `why <pod>`, `events --watch`
- Colourised table output with `--output json`
- Handles RBAC "forbidden" errors with a clear message

### Tasks
- Load kubeconfig and construct a clientset
- List pods and deployments in a namespace
- Compute a health summary from status conditions
- Implement `why` by inspecting container states and events
- Stream events with a watch
- Test with the fake clientset

### Stretch
- Plugin packaging for `kubectl krew`
- Cluster-wide summary with concurrency

## Distributed Job System
id: distributed-job-system
level: advanced
module: microservices
skills: microservices, concurrency
hours: 20

### Summary
A scheduler and worker fleet coordinating through a database with leases.

### Description
Build a job system where a scheduler stores jobs in PostgreSQL and workers claim them using leases (`SELECT ... FOR UPDATE SKIP LOCKED`). Handle worker crashes with lease expiry, retries, idempotent handlers and observability.

### Requirements
- Jobs survive process restarts
- At-least-once execution with idempotent handlers
- Leases expire so crashed workers do not block jobs
- Multiple workers run in parallel without double-claiming
- Metrics for queue depth, latency and failures

### Tasks
- Design the jobs table and lease columns
- Implement enqueue and claim with `SKIP LOCKED`
- Implement heartbeats that extend leases
- Reclaim expired leases and retry
- Add idempotency keys for handlers
- Expose metrics and a small admin API
- Chaos-test by killing workers mid-job

### Stretch
- Cron-style recurring jobs
- Job dependencies (DAGs)

## Cloud-Native Go Operations Platform
id: cloud-native-operations-platform
level: capstone
module: microservices
skills: architecture, kubernetes, observability, security
hours: 40

### Summary
The final capstone: an operations platform combining an API, a CLI, agents and a Kubernetes controller.

### Description
Design and build a platform that manages deployments across clusters. A central API stores desired state in PostgreSQL; a CLI talks to the API; agents report health from each cluster; a Kubernetes controller reconciles desired state; everything is observable, secure and deployable with Helm or plain manifests.

### Requirements
- REST API (authenticated, versioned) with PostgreSQL storage
- `opsctl` CLI built on the API
- Agent that reports cluster health and receives commands
- Controller reconciling a custom resource
- Metrics, structured logs and traces across all components
- CI pipeline: lint, test with `-race`, build images, scan, deploy

### Architecture
```text
opsctl ──→ API Gateway ──→ Platform API ──→ PostgreSQL
                              ↑    ↓
                         Agents   Controller ──→ Kubernetes API
```

### Tasks
- Write a design document with boundaries, data model and failure modes
- Build the platform API with authentication and authorisation
- Build the CLI against the API
- Build the cluster agent and its reporting protocol
- Build the Kubernetes controller and CRD
- Add end-to-end observability
- Containerise everything and add Kubernetes manifests
- Add CI with tests, race detector, image scanning and deploy
- Load-test and write a runbook

### Stretch
- Multi-tenancy with per-team quotas
- GitOps delivery with an operator that syncs from a Git repository
