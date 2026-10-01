# Code execution sandbox

GoLearn runs untrusted Go code written by learners. **It must never run on the API host.** This document explains the design, its limits, and how to deploy it safely.

## Threat model

An attacker with an account can submit arbitrary Go source (up to 32 KB). They may try to:

- consume CPU, memory, disk, processes or wall-clock time (denial of service),
- reach the network (internal services, cloud metadata, other tenants, outbound abuse),
- read host files or credentials, or escape the container,
- persist state between runs or affect other users' runs.

## What each run gets

Every request starts a fresh container (`internal/runner/docker.go`, arguments covered by unit tests) and removes it afterwards, including on timeout:

| Control | Setting |
| --- | --- |
| Network | `--network none` (loopback only) |
| Filesystem | `--read-only` root; writable `tmpfs` at `/work` (64 MB) and `/tmp` (384 MB), no bind mounts or volumes |
| Privileges | non-root user `65534`, `--cap-drop ALL`, `--security-opt no-new-privileges` |
| Memory | `--memory` = `--memory-swap` (default 512 MB; tmpfs counts against it) |
| CPU | `--cpus 1` (default) |
| Processes | `--pids-limit 256` (fork-bomb guard), `nofile` 256, core dumps off |
| Time | overall deadline enforced inside (`go build` 40 s, program `GOLEARN_RUN_TIMEOUT` ≤ 30 s) **and** outside (context timeout + `docker rm -f`) |
| Output | 256 KB read back from the container, 64 KB per stream inside; the API returns `truncated: true` |
| Code injection | file names validated (`^[a-z][a-z0-9_]*\.go$`); the API only ever writes `main.go` and (for challenges) the platform's own `main_test.go` |
| Concurrency | at most `GOLEARN_RUN_CONCURRENCY` containers at once, queue wait 10 s then `503 runner_busy` |
| Abuse | per-user rate limit (`GOLEARN_RUNS_PER_MINUTE`), 32 KB code limit, authentication required |

The Go toolchain runs offline (`GOPROXY=off`), so learner programs cannot download modules; only the standard library is available.

## Known limits — read these

1. **A container is not a hard security boundary.** A kernel or runtime vulnerability could allow escape. For a public deployment run the sandbox on **dedicated, disposable nodes** with a stronger runtime: gVisor (`GOLEARN_DOCKER_RUNTIME=runsc`), Kata Containers or Firecracker microVMs.
2. **Access to the Docker socket is root-equivalent.** Whoever can talk to `/var/run/docker.sock` controls the host. That is why the API can run with `GOLEARN_RUNNER=remote`: only the small `runner-daemon` (authenticated with a bearer token, reachable only on an internal network) holds runtime access, and the internet-facing API holds none.
3. **Challenge grading is hardened but not unforgeable.** Learner code runs in the same process as the tests, so no channel inside that process can be perfectly trusted. Grading therefore does not believe anything the program prints. The entrypoint (a separate, trusted process) rewrites the platform's hidden test file so each top-level test records its own outcome, adds a `TestMain` that writes a verdict file after `m.Run()` returns, and authenticates that file with a random per-run key compiled into the test binary (`internal/runner/grader.go`). The server accepts a challenge only from a verified verdict in which every expected test ran and passed and `m.Run()` returned 0; test events in stdout are used only to explain failures. This defeats the practical attacks — printing `--- PASS` lines or JSON events, `os.Exit(0)` from `init()`, filtering the real tests out with flags, and writing a guessed verdict file — each of which is a regression test run through real `go test` (`TestChallengeGradingResistsForgedResults`). **Remaining gap:** a learner who reverse-engineers their own compiled test binary within the run, extracts the per-run key, and signs a verdict can still forge a pass. Closing that needs out-of-process (black-box) tests, which would change the challenge format. Cheating only affects the cheater's own progress. Exercises (stdout must equal the expected text) are not tamper-resistant by design: a learner can simply print the expected text.
4. **Resource accounting is per run, not per user.** Rate limiting bounds overall load; add cgroup-level quotas on the runner nodes if you expect abuse.
5. **Only runc was tested.** The Docker path has been exercised against a real daemon (see [production-readiness.md](production-readiness.md)), but not with gVisor, Kata or Firecracker, and not on a multi-node deployment. Repeat the checks on your own runner nodes before launch.

## Topologies

**Local development** — API and Docker on the same machine: `GOLEARN_RUNNER=docker`. Acceptable for a single-user laptop.

**Docker Compose (provided)** — `api` (no socket) → internal network → `runner` daemon (has the socket) → per-run containers. The `sandbox` network is `internal: true`, so the daemon isn't reachable from outside.

**Production (recommended)**

```
Internet → Ingress → API pods (no runtime access)
                          │  private network, bearer token, NetworkPolicy
                          ▼
                 runner-daemon on a dedicated node pool
                          │  containerd/Docker with gVisor or Kata
                          ▼
                 per-request sandbox containers (destroyed after each run)
```
- Taint the runner nodes and give them no credentials or metadata access (block `169.254.169.254`).
- Set egress `NetworkPolicy` to deny everything from the sandbox namespace.
- Alert on runner saturation (`golearn_runs_total{result="busy"|"timeout"}` at `/metrics`).

If you cannot provide that isolation, leave `GOLEARN_RUNNER=disabled` (the default in the Kubernetes manifests). Lessons remain fully readable, quizzes and progress tracking still work, and the UI states plainly that code execution is unavailable.

## The runner image

`deploy/runner.Dockerfile` builds `golearn-runner`: the Go toolchain plus `sandbox-entrypoint`, with the standard-library build cache pre-seeded at `/opt/gocache` (copied into the `tmpfs` at start so cold compiles take about a second instead of many). The entrypoint reads one JSON request on stdin and prints one JSON result.
