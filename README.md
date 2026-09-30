# GoLearn — Learn Go. Build Real Software.

GoLearn is an interactive learning platform that takes a learner from *"what is Go?"* to concurrent services, REST APIs and Kubernetes controllers. Lessons are short, most include a runnable exercise, every lesson ends in a knowledge check, and progress, skills and achievements are tracked per learner.

```
Learn → Understand → Code → Practice → Build → Review → Master
```

## What is in the box

| Area | Status |
| --- | --- |
| **Curriculum** | 26 modules (00–25), **241 published lessons**, 67 coding challenges with hidden Go tests, 29 projects, 76 glossary terms. ~75 further lessons are listed on the roadmap as *coming soon* (clearly marked, not faked). |
| **Learning loop** | Roadmap → lesson (objectives, concept, example, exercise, quiz, challenge links, solution, takeaways) → progress, XP, streaks, skills, achievements. |
| **Code editor** | CodeMirror 6 with Go highlighting, Run / Reset, console, build errors, expected-output check, test results for challenges, `Ctrl/⌘+Enter`. |
| **Code execution** | Server-side, in a throw-away Docker container per run (no network, read-only FS, dropped capabilities, CPU/memory/PID/time limits). See [docs/sandbox.md](docs/sandbox.md). |
| **Accounts** | Email + password (bcrypt), server-side sessions in an `HttpOnly` cookie, CSRF header check, rate limits. Auth sits behind an `Authenticator` interface so OIDC can be chained in. |
| **Search** | Global `Ctrl/⌘+K` over lessons, modules, challenges, projects, code examples and glossary. |
| **UI** | React 19 + TypeScript (strict) + Tailwind 4, light/dark, responsive (drawers on mobile), keyboard shortcuts, reduced-motion support, ARIA-labelled controls. |
| **Backend** | Go standard library (`net/http` routing), SQLite or PostgreSQL via one portable SQL migration, Prometheus `/metrics`, structured `slog` logs, health/readiness endpoints, graceful shutdown. |
| **Deploy** | Multi-stage Dockerfiles, Docker Compose, Kubernetes manifests (restricted pod security, network policies, probes). |

### Deliberately *not* implemented (and shown as such rather than faked)

- **Redis cache** — curriculum is served from memory; nothing needs a shared cache yet.
- **OIDC / Google / Keycloak login** — the seam exists (`internal/auth.Authenticator`); the login page says it isn't configured. **Password reset** is likewise not implemented.
- **Admin CMS UI, analytics dashboards** (Phase 4) — content is versioned Markdown in `internal/content/data` (with `draft/review/published/archived/planned` statuses) and validated in CI instead.
- **OpenTelemetry tracing** — metrics and logs are in; tracing is taught in Module 19/25 but not wired into the server.
- **Code execution in the default Kubernetes manifests** — disabled there on purpose; see the sandbox doc for why and how to enable it safely.

### Verification status (what was and wasn't exercised)

- Go backend: unit + HTTP integration tests (SQLite in memory), `go vet` clean. Content is checked by compiling and running **every** example, exercise solution, quiz "predict the output" snippet, glossary snippet and challenge solution (with `-race`), and by asserting each starter fails its tests.
- Frontend: strict `tsc`, 19 Vitest tests, and a Playwright walk-through of register → lesson → quiz → complete → dashboard/skills/achievements/search/projects on desktop and mobile widths.
- Runner: the host↔container protocol, timeouts, output caps and build errors are tested with a shim that stands in for the `docker` CLI. **The real Docker path could not be exercised in the environment this was built in (no Docker daemon)** — build the runner image and try the Playground before relying on it.
- PostgreSQL: the code path shares the SQL and a placeholder rewriter (unit-tested) with SQLite, but no live PostgreSQL instance was available for an integration run.

## Quick start

Requirements: Go 1.26+, Node 22+. Docker is optional (needed only to run learner code).

```bash
make build          # builds the frontend (web/dist) and bin/golearn
make run            # http://localhost:8080  (SQLite file golearn.db, code runner off)
```

Enable code execution locally:

```bash
make runner-image                  # builds golearn-runner:latest
GOLEARN_RUNNER=docker make run     # the API talks to your local Docker daemon
```

Or use Compose (PostgreSQL + API; add the `sandbox` profile for the runner daemon):

```bash
docker compose --profile sandbox --profile build-only build
docker compose --profile sandbox up
```

Frontend development with hot reload: `make dev` (API on :8080) and `cd web && npm run dev` (Vite on :5173, proxies `/api`).

## Architecture

```
Browser (React SPA)
   │  REST/JSON  (+ HttpOnly session cookie, X-GoLearn-CSRF header)
   ▼
cmd/golearn ── internal/api ──▶ internal/progress   (module states, skills, streaks, XP, achievements)
                   │      └───▶ internal/content    (embedded Markdown curriculum, validated at startup)
                   │      └───▶ internal/search     (in-memory index)
                   ├────────▶ internal/store        (learner state: PostgreSQL or SQLite)
                   └────────▶ internal/runner       (Docker driver │ remote client │ disabled)
                                     │
                          cmd/runner-daemon ──▶ docker run … golearn-runner  (per-request, destroyed after)
                                                     └─ cmd/sandbox-entrypoint (inside the container)
```

- **Curriculum as code.** Modules, lessons, challenges, projects and glossary are Markdown files embedded in the binary — reviewed in pull requests, versioned with the app, and verified by CI. Learner state (progress, submissions, sessions) lives in the database. See [docs/content-authoring.md](docs/content-authoring.md).
- **Progress is derived**, not stored in aggregate: `internal/progress` computes module states, skill levels, streaks, XP and achievements from raw records with a pure function that has unit tests.
- **API responses** use one envelope: `{"data": …, "meta": …}` or `{"error": {"code", "message"}}`.

### Repository layout

```
cmd/golearn            API server + static frontend host
cmd/runner-daemon      HTTP front for the Docker sandbox (holds the container-runtime access)
cmd/sandbox-entrypoint runs inside the sandbox container
internal/api           handlers, routing, middleware wiring
internal/auth          local accounts, sessions, Authenticator interface
internal/content       curriculum model, Markdown parser, validation, embedded data/
internal/progress      pure progress/skills/achievement engine
internal/runner        Docker driver, remote client, test-JSON parsing
internal/search        in-memory search
internal/store         SQL access (portable between PostgreSQL and SQLite)
migrations/            SQL migrations
web/                   React + TypeScript + Vite + Tailwind frontend
deploy/                Dockerfiles and Kubernetes manifests
docs/                  sandbox and content-authoring guides
```

## Configuration (environment variables)

| Variable | Default | Purpose |
| --- | --- | --- |
| `GOLEARN_ADDR` | `:8080` | listen address |
| `DATABASE_URL` | `sqlite:golearn.db` | `postgres://…` or `sqlite:path` (`sqlite::memory:`) |
| `GOLEARN_STATIC_DIR` | `web/dist` | built frontend (empty disables static serving) |
| `GOLEARN_COOKIE_SECURE` | `false` | set `true` behind HTTPS |
| `GOLEARN_TRUST_PROXY` | `false` | use the last `X-Forwarded-For` entry as the client IP — only behind exactly one proxy you control that appends it |
| `GOLEARN_BCRYPT_COST` | `12` | password hashing cost |
| `GOLEARN_RUNNER` | `docker` | `docker`, `remote` or `disabled` |
| `GOLEARN_RUNNER_IMAGE` | `golearn-runner:latest` | sandbox image |
| `GOLEARN_RUNNER_URL` / `GOLEARN_RUNNER_TOKEN` | – | remote runner daemon (required for `remote`) |
| `GOLEARN_DOCKER_RUNTIME` | – | alternative OCI runtime, e.g. `runsc` (gVisor) |
| `GOLEARN_RUN_TIMEOUT`, `GOLEARN_RUN_MEMORY_MB`, `GOLEARN_RUN_CPUS`, `GOLEARN_RUN_PIDS`, `GOLEARN_RUN_CONCURRENCY`, `GOLEARN_RUNS_PER_MINUTE` | 15s, 512, 1, 256, 4, 30 | sandbox limits and per-user rate limit |

Endpoints for operators: `GET /healthz`, `GET /readyz` (database ping), `GET /metrics` (Prometheus).

## Testing

```bash
make test-short      # Go tests with the race detector (fast)
make content-check   # compile & run every lesson snippet and challenge (≈2 minutes, needs Go)
cd web && npm test   # frontend unit tests
cd web && npm run typecheck
```

While authoring a module, run only its snippets: `CONTENT_ONLY=concurrency/ go test ./internal/content -run TestContentRuns -v`.

## Security notes

- Learner code is **never** compiled or run on the API host — only inside the per-request sandbox container. Read [docs/sandbox.md](docs/sandbox.md) before deploying.
- Passwords are bcrypt-hashed; session tokens are 256-bit random values stored only as SHA-256 hashes; login timing does not reveal whether an account exists.
- State-changing API calls require the `X-GoLearn-CSRF` header (cross-origin requests can't send it without a CORS preflight) on top of `SameSite=Lax` cookies.
- Responses carry a strict CSP (`script-src 'self'`), `nosniff`, frame denial and a restrictive `Permissions-Policy`.
- Quiz answers, challenge tests and challenge solutions are never sent to the browser before they are earned.
