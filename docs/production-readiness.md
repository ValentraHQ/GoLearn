# Production readiness report

Scope: validation and hardening pass over the existing `master` MVP. Nothing was rebuilt; fixes are small and each has a regression test.
Date of the pass: 2026-09-30. Commits: `615731f` … `e70d2b5` on top of `86366d2`.

Legend used throughout: **Verified** (run and observed here) · **Partially verified** · **Not tested** · **Not implemented** · **Intentionally deferred**.

## 1. Executive summary

The core learning loop works end to end against the real stack: register → login → dashboard → roadmap → module → lesson → editor → execution in a real Docker sandbox → challenge tests → quiz → completion → progress, skills, achievements, streak. It was exercised three ways: SQLite + local Docker, PostgreSQL 16 + local Docker, and the Compose topology (PostgreSQL 17, distroless API without the Docker socket, runner daemon, per-run sandbox containers).

The pass found and fixed **three high-severity** problems (a Compose image-name collision that silently replaced the sandbox image, a runner that turned any JSON output into a "successful" empty run, and a `go.mod` that allowed a Go toolchain with 21 reachable standard-library vulnerabilities) and several medium/low ones (section 11).

**This is not a claim that GoLearn is production-ready.** It is a validated, hardened MVP for a controlled launch. The main gaps are: no independent penetration test, no stronger sandbox runtime (gVisor/Kata) tested, no Kubernetes cluster run, CI changes not yet run on GitHub, and load tested only up to 50 concurrent runs on one 4-vCPU host. See sections 13–15.

## 2. Current architecture

```
Browser (React SPA)
   │ REST/JSON, HttpOnly session cookie, X-GoLearn-CSRF header
   ▼
golearn API (no Docker access in "remote" mode) ──▶ PostgreSQL | SQLite
   │  bearer token, internal network
   ▼
runner-daemon (only component holding the Docker socket)
   │  docker run --network none --read-only --cap-drop ALL …
   ▼
per-request sandbox container (destroyed after each run)
```

## 3. Environment tested

| Item | Value |
| --- | --- |
| Host | Linux container, 4 vCPU, 16 GB RAM, root |
| Docker | Engine 29.8.1, `dockerd` started in the session; runc, overlayfs, **cgroup v1** |
| Go | 1.26.8 (`go.mod` now requires it) |
| Node | 22.22 |
| PostgreSQL | 16-alpine (test suites, live server), 17-alpine (Compose) |
| Browser | Playwright Chromium |
| Not available | gVisor/Kata/Firecracker, a Kubernetes cluster, GitHub Actions, multiple nodes, TLS termination |

## 4. Test matrix

| Area | Status | Evidence |
| --- | --- | --- |
| `gofmt`, `go vet` | Verified | clean |
| `go test -short -race ./...` | Verified | all packages pass |
| Content execution (`go test ./internal/content`, ~150 s) | Verified | every example, exercise solution, quiz snippet, glossary snippet, challenge solution compiled/ran; starters fail their tests |
| Store + API suites on PostgreSQL 16 | Verified | `GOLEARN_TEST_DATABASE_URL`, one schema per test; a wrong password makes them fail, proving they hit Postgres |
| Live E2E (`TestLiveE2E`) | Verified on 3 stacks | SQLite+Docker, Postgres 16+Docker, Compose (Postgres 17 + daemon) |
| Frontend `tsc`, Vitest (19), build | Verified | pass |
| Browser QA at 5 viewports | Verified (see §8) | 1920×1080, 1440×900, 1280×720, 390×844, 412×915 |
| Docker sandbox attack matrix | Verified (see §5) | real Docker |
| Runner daemon auth/validation | Verified | curl probes (§5) |
| Security probes | Partially verified | lightweight, not a penetration test (§9) |
| Deployment: Compose | Verified | `compose config`, stack brought up, runtime hardening inspected |
| Deployment: Kubernetes | Partially verified | `kubectl kustomize` renders 13 objects; **not applied to a cluster** |
| CI workflow | Not tested | YAML parses; the new jobs (`postgres`, `govulncheck`, `npm audit`) have not run on GitHub |
| `deploy/Dockerfile` build stages | Not tested | see §14 |
| gVisor / Kata | Not tested | flag is passed through and unit-tested only |

## 5. Docker sandbox results

Run against the real Docker daemon through `POST /api/run` (direct driver), and again through Compose (API → daemon → Docker).

| # | Case | Result |
| --- | --- | --- |
| 1 | Hello World | `Hello, World!`, ~0.8 s wall time |
| 2 | Valid program (sort) | correct output |
| 3 | Compilation error | `buildError: true`, compiler message returned |
| 4 | Runtime panic | exit 2, panic text returned (container-internal path `/work/main.go` only) |
| 5 | Infinite loop | `timedOut: true` at ~10.6 s; container removed |
| 6 | 2 M lines of output | stdout capped at 200 KB, `truncated: true`, run bounded by the timeout |
| 7 | Memory hog (64 MB blocks) | killed by the memory limit (exit -1); now reports *"Process was killed by signal … (often the memory limit)"* |
| 8 | Filesystem | runs as uid/gid 65534; `/`, `/usr` read-only; `/work`, `/tmp` writable tmpfs only; host file `/tmp/host-secret.txt` absent; `/etc/shadow`, `/root` denied |
| 9 | Network | `network is unreachable` for public IPs, metadata IP `169.254.169.254`, RFC1918, `host.docker.internal`, DNS; only `lo` present |
| 10 | Host/privilege | all capability sets `0`, `NoNewPrivs: 1`, seccomp filter active; `mount`, `chroot`, `setuid(0)` → `operation not permitted`; no host mounts |
| 11 | Docker socket / cloud / Kubernetes credentials | `/var/run/docker.sock`, `/run/secrets`, service-account token, `~/.kube`, `~/.aws`, gcloud config: absent or denied; environment contains only image defaults, no host secrets |
| 12 | Fork bomb | stopped at 246 processes (limits enforced), nothing left on the host |
| 13 | Disk fill | stopped at ~263 MB (tmpfs cap) |
| 14 | Cleanup | `docker ps -a` empty after timeout, OOM and normal runs; no stray processes on the host |
| 15 | Consecutive/concurrent runs | see §10 |

**Untrusted code never runs on the API host.** In `remote` mode the API container has zero mounts and no socket; only the daemon container mounts `/var/run/docker.sock` and it publishes no port (reachable only on the `internal: true` `sandbox` network).

**Implementation review** (`internal/runner/docker.go`): non-root, `--network none`, `--read-only`, tmpfs `/work` and `/tmp`, `--cap-drop ALL`, `no-new-privileges`, memory=swap, `--cpus`, `--pids-limit`, `nofile`, `core=0`, no env passthrough, name-randomised containers force-removed after every run including timeouts, optional `--runtime`. Judged adequate; nothing was added blindly. Added: request validation before any container starts, and the protocol-marker check (P1 below).

**Runner daemon** (`cmd/runner-daemon`): bearer token compared in constant time (≥16 chars enforced at startup); missing/incorrect token → 401; malformed JSON, unknown fields, unknown kind, bad file names, oversized body → 400 (now without spawning a container); wrong method → 405; 128 KB body cap; server timeouts set. Wrong token from the API is reported to users generically. *Limits:* no rate limiting or lockout on the daemon (it relies on network isolation and a high-entropy token), no replay protection (requests are idempotent runs, not state changes).

## 6. PostgreSQL results

| Check | Result |
| --- | --- |
| Startup, migration `0001_init.sql`, schema creation | Verified (9 tables + `schema_migrations`, 14 indexes) |
| Foreign keys | Verified: 8 FK constraints; orphan insert rejected |
| Full store + API test suites | Verified (Postgres 16); repeated 5× |
| Register/login/progress/quiz/challenge persistence | Verified via live E2E on Postgres 16 and 17 |
| Restart persistence | Verified: container restart kept data; API and Postgres restarted together, existing session cookie still valid |
| Outage behaviour | Verified: readiness → 503; recovers automatically when the DB returns; public curriculum stays up |
| Query plans, load, backups, pooling limits, HA | **Not tested** |

## 7. Backend results

- Authentication/sessions: cookie is HttpOnly + SameSite=Lax (+Secure when configured); tokens stored hashed; logout invalidates the token server-side (replay → 401); password change revokes other sessions; login/register rate limited (429 after the burst); login errors do not reveal whether an account exists; passwords ≥10 chars.
- Authorization/isolation: verified live with two users — anonymous access → 401; B sees none of A's progress/XP; B cannot read A's earned challenge solution; B's actions do not change A.
- Idempotency: repeated exercise passes, quiz passes, lesson completion (3×) and challenge passes (2×) leave counts at 1 and do not double-award XP.
- Persistence: logout/login and a second concurrent device see identical progress.
- Error handling: malformed JSON → 400 with a generic message; oversized body rejected; no stack traces, paths, SQL or Docker details found in any error body probed.
- Streak: `currentStreak ≥ 1` after a completion verified live; multi-day/timezone logic is covered by unit tests only (**partially verified**).
- Achievements: `first-lesson` verified live; the rest by unit tests.

## 8. Frontend results

Browser walk-through at all five viewports with a signed-in user: 15 routes per viewport, lesson exercise execution against the real sandbox, challenge submission showing test results, search palette (results and empty state), keyboard shortcuts, and forced error/loading states.

- Internal links: 209 distinct links crawled, none broken (HTTP or in-app "not found").
- Console/page errors and 5xx: none.
- Overflow: none at any viewport after the fix below (glossary overflowed by ~860–880 px on phones).
- Error state (503 on dashboard) shows a retry button; loading state visible on a slow API; runner outage shows a clear message in the Playground.
- Mobile execution works: the output panel appears under the editor.
- Small tap targets (<24 px) remain: glossary "related lesson" chips (22 px tall) and the visually hidden skip link (intentional). Low severity.
- Unit tests (19), typecheck and build pass. Screenshots were reviewed only for the layouts named above; this is not an exhaustive visual audit.

## 9. Security findings (lightweight review — not a penetration test)

| Topic | Result |
| --- | --- |
| SQL injection | Probed search, lesson/challenge IDs, login: parameterised, no effect |
| XSS | React escapes; stored profile fields round-trip as inert text; strict CSP (`script-src 'self'`), `nosniff`, frame denial |
| CSRF | Mutations without `X-GoLearn-CSRF` → 403, also with a valid cookie and with `text/plain` |
| CORS | No CORS headers unless `GOLEARN_ALLOW_ORIGIN` matches |
| Path traversal | Static handler: encoded/`..` variants → 400 or SPA index, no file disclosure; API IDs not used as paths |
| Command injection / arbitrary code | Learner code only runs in the sandbox; file names validated on the host and in the container |
| SSRF | No server-side fetch of user-supplied URLs exists |
| Sandbox escape | Container isolation verified (§5); a kernel/runtime 0-day is out of scope of this test |
| Secrets/error leakage | None found; runner infrastructure detail removed from user-visible text (fixed) |
| Dependencies | `govulncheck`: 0 reachable vulnerabilities on Go 1.26.8; `npm audit --omit=dev`: 0. One `golang.org/x/crypto` advisory (GO-2026-5932) has no fix and is not called |
| API abuse | Per-IP auth limiter, per-user run limiter, body/code size caps, concurrency cap with 503 back-pressure |

## 10. Performance smoke test

Single host (4 vCPU / 16 GB), Docker driver with `GOLEARN_RUN_CONCURRENCY=8`, CPU-bound program (50 M loop iterations, ~1.6–3.3 s of execution). Registrations were spread over 6 users because the auth limiter is deliberate.

| Concurrent requests | Success | Wall time | Latency p50 / p95 / max | Notes |
| --- | --- | --- | --- | --- |
| 1 | 100 % | 0.8 s | 0.79 s | includes ~0.6 s container start + compile from the seeded cache |
| 10 | 100 % | 4.1 s | 3.4 / 4.0 / 4.1 s | 8 containers at once |
| 25 | 100 % | 5.7 s | 3.9 / 5.5 / 5.6 s | 8 at a time, queued |
| 50 | 96 % (2 × `503 runner_busy`) | 11.2 s | 6.4 / 11.0 / 11.1 s | back-pressure after the 10 s queue wait |

Host memory rose from ~0.75 GB to ~2.0 GB at 8 containers. CPU was not sampled during the runs (only idle `docker stats`), so CPU figures are **not measured**. No optimisation was attempted. Capacity is bounded by the concurrency setting and CPU count, not by the API.

## 11. Issues found

| Sev | Issue | Status |
| --- | --- | --- |
| **P1** | `docker-compose.yml`: the `runner` service's default image name (`golearn-runner`) is the sandbox image tag; building it replaced the sandbox image with the daemon, so executions ran the wrong image | Fixed (explicit image names) |
| **P1** | Runner accepted *any* JSON on the container's stdout as a result; a wrong image printing a log line produced a successful, empty run (fail-open) | Fixed (protocol marker required); test fails without the fix |
| **P1** | `go.mod` allowed Go 1.26.0: `govulncheck` reports 21 reachable stdlib vulnerabilities; 1.26.8 reports 0 | Fixed (`go 1.26.8`; CI runs `govulncheck`) |
| P2 | Database failure was treated as "not signed in": signed-in users got 401 and `/api/auth/me` returned `null` during an outage | Fixed (503, session kept) |
| P2 | Rate limiting used the *first* `X-Forwarded-For` entry, which a client controls when a proxy appends | Fixed (last entry; documented single-proxy assumption) |
| P2 | "Latest submission" could return the older code when two submissions landed in the same second (existing test was flaky) | Fixed |
| P2 | Glossary cards overflowed on phones | Fixed |
| P2 | Kubernetes example Secret ships a `CHANGE-ME` database password and `sslmode=disable` | Open (documented; replace before deploying) |
| P3 | Killed/OOM runs showed no explanation | Fixed |
| P3 | Runner daemon started containers for invalid file names | Fixed (validated first) |
| P3 | API responses were cacheable | Fixed (`Cache-Control: no-store`) |
| P3 | Runner status/error text exposed infrastructure detail to anonymous users | Fixed |
| P3 | Runner image build ran an unnecessary `go mod download` (failed behind a TLS-intercepting proxy) | Fixed |
| P3 | Stray screenshots (`undefined/`) tracked; `.env.example` missing | Fixed |
| P3 | Registration reveals whether an email exists (409) | Open (common trade-off; note for threat model) |
| P3 | No HSTS header from the app | Open (set at the TLS-terminating proxy) |
| P3 | Login during a DB outage returns 500 rather than 503 | Open |
| P3 | Runner daemon has no request throttling | Open (network isolation + token) |
| P3 | Small tap targets in glossary chips | Open |

## 12. Fixes applied (with regression tests)

| Fix | Test |
| --- | --- |
| Compose image names | `docker compose config` + full stack run (no unit-testable surface) |
| Protocol marker | `TestWrongImageOutputIsRejected` |
| Toolchain pin, CI scans | `govulncheck` clean |
| DB outage ≠ signed out | `TestDatabaseOutageIsNotSignedOut` |
| Last `X-Forwarded-For` | `TestClientIP` |
| Deterministic latest submission | `TestChallengeStatusesAndProjectsAndAchievements` (100 % stable over 50 runs) |
| Request validation before container start | `TestRequestValidate`, `bad file name` case |
| Signal-kill explanation, timeout | `TestExecuteReportsSignalKill`, `TestExecuteTimeout` |
| No infra detail in messages | `TestRemoteBadTokenIsUnavailable` (leak assertions) |
| `no-store` | `TestSecurityHeadersAndUnknownAPI` |
| Glossary overflow | verified in browser at 390/412/1280 (no automated layout test) |

New test infrastructure: `GOLEARN_TEST_DATABASE_URL` (Postgres for the store/API suites) and `TestLiveE2E` (`GOLEARN_E2E_URL`).

## 13. Known limitations

- A container is not a hard security boundary; a public launch needs gVisor/Kata/Firecracker on dedicated, disposable nodes (`docs/sandbox.md`). Only runc was tested.
- Challenge grading trusts `go test -json` produced inside the untrusted process (documented; affects only the cheater's own progress).
- Capacity is per host and configured by `GOLEARN_RUN_CONCURRENCY`; behaviour beyond 50 concurrent runs or on multiple nodes is unknown.
- Not implemented: Redis, OIDC login, password reset, admin CMS/analytics, OpenTelemetry tracing in the server. ~78 of 319 roadmap lessons are labelled *coming soon* rather than faked.
- Intentionally deferred: everything above; code execution is disabled by default in the Kubernetes manifests.

Curriculum integrity (verified via the API and the content test): 26 modules, 241 published lessons, 78 coming-soon lessons, 67 challenges, 29 projects, 76 glossary terms.

## 14. Deployment requirements

- Compose: verified. `deploy/Dockerfile`'s *build* stages (`npm ci`, `go mod download`) could not run here because the sandbox network intercepts TLS; the runtime stages were reproduced from locally built binaries (distroless non-root, read-only, all caps dropped, no-new-privileges, healthy). Build the API image in an unrestricted environment (CI does) before relying on it.
- Kubernetes: manifests render and use restricted Pod Security, non-root, read-only root FS, probes, NetworkPolicies, and `GOLEARN_RUNNER=disabled` by default (verified in the rendered output). Replace the example Secret, use TLS to the database, and treat enabling the runner as a separate hardening project.
- Set `GOLEARN_COOKIE_SECURE=true` behind HTTPS; `GOLEARN_TRUST_PROXY=true` only behind exactly one proxy you control; add HSTS at the proxy.
- Configuration reference: `.env.example`. Health: `/healthz`, `/readyz` (database ping), `/metrics`.
- Runner: `GOLEARN_RUNNER_TOKEN` ≥16 random characters, daemon on an isolated network/node, `docker` socket only on the daemon.

## 15. Remaining work

1. Run the CI workflow on GitHub (new `postgres`, `govulncheck`, `npm audit` jobs and both image builds) and fix anything it finds.
2. Test the sandbox with gVisor (`runsc`) or Kata on a dedicated node pool; measure the overhead.
3. Apply the manifests to a real cluster (probes, NetworkPolicy, PVC, ingress/TLS), including a restart/upgrade rehearsal.
4. Load-test beyond 50 concurrent runs with CPU sampling; consider per-user concurrency quotas.
5. Independent penetration test before a public launch.
6. Password reset and email verification (Not implemented); decide on the email-enumeration trade-off.
7. Backups, retention, and monitoring/alerts (`golearn_runs_total{result="busy"|"timeout"|"error"}`).
8. Low-severity items listed as *Open* in §11.
