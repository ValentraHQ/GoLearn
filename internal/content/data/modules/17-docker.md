# Docker & Go
id: docker
number: 17
track: devops
paths: beginner, pro
skill: docker
requires: rest-api
project: containerized-go-api
summary: Containerise Go services with small, secure images and Docker Compose.

## Go Application Containers
slug: go-application-containers
minutes: 7
objectives: Explain why Go suits containers; Build a static, stripped Linux binary; Cross-compile with GOOS and GOARCH
takeaways: Go compiles to a single static binary, so containers need no runtime; CGO_ENABLED=0 GOOS=linux go build produces a portable binary; -trimpath and -ldflags "-s -w" make it reproducible and smaller

### Concept
A Go service is one executable. With `CGO_ENABLED=0` it links no C libraries, so it runs in an image containing *nothing else* (`scratch`) or a tiny base such as `distroless`.

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -trimpath -ldflags="-s -w" -o app ./cmd/app
```
- `GOOS`/`GOARCH` cross-compile (build on a Mac, run on Linux; add `arm64` for Graviton/Apple silicon).
- `-trimpath` removes local file paths; `-s -w` strip debug symbols (≈30% smaller).
- Embed a version: `-ldflags="-X main.version=$(git describe --tags)"`.

Typical result: a 10–20 MB image, starting in milliseconds.

### Example
```go
package main

import (
	"fmt"
	"runtime"
)

func main() {
	fmt.Println(runtime.GOOS, runtime.GOARCH != "")
	fmt.Println("static binary: CGO_ENABLED=0 means no libc dependency")
}
```

### Exercise
Write `buildCmd(goos, goarch, out, pkg string) []string` returning the shell words: the three environment assignments (`CGO_ENABLED=0`, `GOOS=…`, `GOARCH=…`) followed by `go build -trimpath -ldflags=-s -w -o <out> <pkg>` (as separate words with `-ldflags=-s -w` quoted as one: `-ldflags=-s -w` becomes `"-ldflags=-s -w"`).
```text expect
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath "-ldflags=-s -w" -o app ./cmd/app
```
```go solution
package main

import (
	"fmt"
	"strings"
)

// BEGIN
func buildCmd(goos, goarch, out, pkg string) []string {
	return []string{
		"CGO_ENABLED=0", "GOOS=" + goos, "GOARCH=" + goarch,
		"go", "build", "-trimpath", `"-ldflags=-s -w"`, "-o", out, pkg,
	}
}

// END

func main() {
	fmt.Println(strings.Join(buildCmd("linux", "arm64", "app", "./cmd/app"), " "))
}
```

### Check
Q: What does `CGO_ENABLED=0` achieve?
T: mcq
- [ ] Disables the garbage collector
- [x] Produces a statically linked binary without C library dependencies
- [ ] Enables cross-compilation
- [ ] Removes symbols
E: Without cgo, the binary doesn't need libc, so it runs on minimal images.

Q: Which environment variables select the target platform when cross-compiling?
T: short
A: GOOS and GOARCH
E: For example `GOOS=linux GOARCH=arm64 go build`.

## Dockerfile
slug: dockerfile
minutes: 8
objectives: Read the common Dockerfile instructions; Order instructions to make good use of the build cache; Write a basic Dockerfile for a Go service
takeaways: FROM, WORKDIR, COPY, RUN, ENV, EXPOSE, USER, ENTRYPOINT and CMD are the core instructions; Copy go.mod/go.sum and download modules before copying source so dependency layers are cached; Use a .dockerignore to keep the build context small

### Concept
```dockerfile
FROM golang:1.24 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download          # cached until go.mod/go.sum change
COPY . .
RUN CGO_ENABLED=0 go build -o /out/app ./cmd/app

ENTRYPOINT ["/out/app"]
```
Each instruction creates a **layer**; unchanged layers come from cache. Put things that change rarely (dependencies) **before** things that change often (source). `ENTRYPOINT ["app"]` uses the *exec form* (JSON array) so signals like SIGTERM reach your process directly. Add a `.dockerignore` (`.git`, `bin/`, `*.md`) so `COPY . .` doesn't ship junk.

### Example
```go
package main

import (
	"fmt"
	"strings"
)

const dockerfile = `FROM golang:1.24 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ENTRYPOINT ["/app"]
`

func main() {
	for _, line := range strings.Split(strings.TrimSpace(dockerfile), "\n") {
		instr, _, _ := strings.Cut(line, " ")
		fmt.Print(instr, " ")
	}
	fmt.Println()
}
```

### Exercise
Write `instructions(dockerfile string) []string` returning the instruction keywords (first word of every non-empty, non-comment line) in order.
```text expect
[FROM WORKDIR COPY RUN]
```
```go solution
package main

import (
	"fmt"
	"strings"
)

// BEGIN
func instructions(dockerfile string) []string {
	var out []string
	for _, line := range strings.Split(dockerfile, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		word, _, _ := strings.Cut(line, " ")
		out = append(out, strings.ToUpper(word))
	}
	return out
}

// END

func main() {
	fmt.Println(instructions("# build\nFROM golang:1.24\n\nWORKDIR /src\nCOPY . .\nRUN go build\n"))
}
```

### Check
Q: Why copy `go.mod` and `go.sum` and run `go mod download` before `COPY . .`?
T: mcq
- [ ] It's required syntax
- [x] The dependency layer stays cached until dependencies change, speeding up rebuilds
- [ ] It makes the image smaller
- [ ] It verifies checksums twice
E: Source changes then only invalidate layers from `COPY . .` onward.

Q: Why prefer the exec form `ENTRYPOINT ["/app"]` over `ENTRYPOINT /app`?
T: mcq
- [ ] It is required by Docker
- [x] Your process receives signals (SIGTERM) directly instead of a shell swallowing them
- [ ] It runs faster
- [ ] It enables health checks
E: The shell form wraps the command in `/bin/sh -c`, which may not forward signals.

## Multi-Stage Builds
slug: multi-stage-builds
minutes: 8
objectives: Separate build and runtime stages; Copy only the binary into the final image; Name stages and target them
takeaways: A multi-stage Dockerfile compiles in a full toolchain image and ships a tiny runtime image; COPY --from=build copies artefacts between stages; The final image contains no compiler, source or build cache

### Concept
```dockerfile
# ---- build stage ----
FROM golang:1.24 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/app ./cmd/app

# ---- runtime stage ----
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/app /app
USER nonroot:nonroot
ENTRYPOINT ["/app"]
```
The `golang` image is ~800 MB; the final image here is ~10 MB. Only the last stage is tagged as the image — earlier stages are build-time scaffolding. `docker build --target build .` builds up to a named stage (useful for running tests).

### Example
```go
package main

import (
	"fmt"
	"strings"
)

func stages(dockerfile string) []string {
	var names []string
	for _, line := range strings.Split(dockerfile, "\n") {
		f := strings.Fields(line)
		if len(f) >= 2 && strings.EqualFold(f[0], "FROM") {
			name := "(unnamed)"
			if len(f) == 4 && strings.EqualFold(f[2], "AS") {
				name = f[3]
			}
			names = append(names, name)
		}
	}
	return names
}

func main() {
	fmt.Println(stages("FROM golang:1.24 AS build\nRUN go build\nFROM scratch\nCOPY --from=build /app /app\n"))
}
```

### Exercise
Write `isMultiStage(dockerfile string) bool` returning true when the file has more than one `FROM` line.
```text expect
true false
```
```go solution
package main

import (
	"fmt"
	"strings"
)

// BEGIN
func isMultiStage(dockerfile string) bool {
	n := 0
	for _, line := range strings.Split(dockerfile, "\n") {
		if f := strings.Fields(line); len(f) > 0 && strings.EqualFold(f[0], "FROM") {
			n++
		}
	}
	return n > 1
}

// END

func main() {
	fmt.Println(
		isMultiStage("FROM golang AS b\nRUN x\nFROM scratch\n"),
		isMultiStage("FROM golang\nRUN x\n"),
	)
}
```

### Check
Q: What is the main benefit of a multi-stage build?
T: mcq
- [ ] Faster CPU
- [x] The runtime image contains only what's needed to run — no compiler or source
- [ ] It removes the need for a Dockerfile
- [ ] It enables cgo
E: Smaller images mean faster pulls and a smaller attack surface.

Q: Which flag copies a file from an earlier build stage?
T: short
A: --from
E: `COPY --from=build /out/app /app`.

## Minimal Images
slug: minimal-images
minutes: 7
objectives: Compare scratch, distroless and alpine base images; Include CA certificates and time zone data when needed; Pick a base image deliberately
takeaways: scratch is empty; distroless adds CA certs, tzdata and a non-root user without a shell; alpine is small but uses musl libc and includes a shell and package manager; A static Go binary works on all of them

### Concept
| Base | Size | Shell | Notes |
| --- | --- | --- | --- |
| `scratch` | 0 | no | you supply CA certs/tzdata; no debugging tools |
| `gcr.io/distroless/static` | ~2 MB | no | CA certs, tzdata, `nonroot` user — a great default |
| `alpine` | ~5 MB | yes | handy for debugging; musl libc (cgo can surprise you) |
| `debian:slim` | ~30 MB | yes | glibc compatibility |

A pure-Go program making HTTPS calls needs **CA certificates**; on `scratch` copy them: `COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/`. For time zones embed them with the `time/tzdata` import. Fewer packages → fewer CVEs to patch.

### Example
```go
package main

import (
	"fmt"
	"time"
	_ "time/tzdata" // embed the tz database so scratch images can load locations
)

func main() {
	loc, err := time.LoadLocation("Europe/Paris")
	fmt.Println(loc, err)
}
```

### Exercise
Write `pickBase(needsShell, needsCGO bool) string` returning `"debian:slim"` when cgo is needed, else `"alpine"` when a shell is needed, else `"distroless"`.
```text expect
distroless alpine debian:slim
```
```go solution
package main

import "fmt"

// BEGIN
func pickBase(needsShell, needsCGO bool) string {
	switch {
	case needsCGO:
		return "debian:slim"
	case needsShell:
		return "alpine"
	}
	return "distroless"
}

// END

func main() {
	fmt.Println(pickBase(false, false), pickBase(true, false), pickBase(true, true))
}
```

### Check
Q: Why does an HTTPS client on a `scratch` image fail with certificate errors?
T: mcq
- [ ] scratch blocks TLS
- [x] There is no CA certificate bundle in the image
- [ ] Go can't do TLS without libc
- [ ] scratch has no network
E: Copy `ca-certificates.crt` into the image (distroless includes it).

Q: A statically linked Go binary (CGO_ENABLED=0) can run on an image with no libc.
T: tf
A: true
E: That's why `scratch` and distroless images work for Go.

## Container Security
slug: container-security
minutes: 8
objectives: Run as a non-root user with a read-only filesystem; Drop capabilities and avoid privileged mode; Scan images and pin base versions
takeaways: Never run as root: add USER and run with --read-only, --cap-drop=ALL and no-new-privileges; Pin base images by digest or specific tag and rebuild regularly; Scan images (Trivy, Grype) and keep secrets out of layers

### Concept
Checklist:

1. **`USER nonroot`** (or numeric `65532`) — a compromised process shouldn't be root.
2. **Read-only root filesystem** (`--read-only`, Kubernetes `readOnlyRootFilesystem: true`) with `tmpfs` for scratch space.
3. **Drop capabilities** (`--cap-drop=ALL`) and set `no-new-privileges`.
4. **No secrets in the image** — `ENV`/`COPY` bake values into layers forever; inject at runtime.
5. **Pin and update** base images (`golang:1.24.3`, or `@sha256:…`); rebuild for patches.
6. **Scan**: `trivy image myapp:latest`.
7. Use minimal bases and `.dockerignore` to avoid shipping `.env` files.

The lesson exercise is a mini Dockerfile linter for a few of these rules.

### Example
```go
package main

import (
	"fmt"
	"strings"
)

func lint(dockerfile string) []string {
	var problems []string
	if !strings.Contains(dockerfile, "\nUSER ") {
		problems = append(problems, "no USER instruction (runs as root)")
	}
	if strings.Contains(dockerfile, ":latest") {
		problems = append(problems, "base image uses :latest")
	}
	return problems
}

func main() {
	fmt.Println(lint("FROM golang:latest\nRUN go build\n"))
}
```

### Exercise
Write `lint(dockerfile string) []string` returning these messages, in this order, when they apply: `runs as root` (no `USER` line, or the last USER is `root`), `unpinned image` (a `FROM` with `:latest` or no tag, ignoring `scratch` and stage references), `secret in ENV` (an `ENV` line whose name contains `PASSWORD`, `SECRET` or `TOKEN`).
```text expect
[runs as root unpinned image secret in ENV]
[]
```
```go solution
package main

import (
	"fmt"
	"strings"
)

// BEGIN
func lint(dockerfile string) []string {
	var problems []string
	user := ""
	unpinned, secret := false, false
	stages := map[string]bool{}
	for _, line := range strings.Split(dockerfile, "\n") {
		f := strings.Fields(line)
		if len(f) < 2 {
			continue
		}
		switch strings.ToUpper(f[0]) {
		case "USER":
			user = f[1]
		case "FROM":
			img := f[1]
			if img != "scratch" && !stages[img] && (!strings.Contains(img, ":") || strings.HasSuffix(img, ":latest")) {
				unpinned = true
			}
			if len(f) == 4 && strings.EqualFold(f[2], "AS") {
				stages[f[3]] = true
			}
		case "ENV":
			key := strings.ToUpper(strings.SplitN(f[1], "=", 2)[0])
			if strings.Contains(key, "PASSWORD") || strings.Contains(key, "SECRET") || strings.Contains(key, "TOKEN") {
				secret = true
			}
		}
	}
	if user == "" || user == "root" {
		problems = append(problems, "runs as root")
	}
	if unpinned {
		problems = append(problems, "unpinned image")
	}
	if secret {
		problems = append(problems, "secret in ENV")
	}
	return problems
}

// END

func main() {
	fmt.Println(lint("FROM golang\nENV DB_PASSWORD=hunter2\nRUN go build\n"))
	fmt.Println(lint("FROM golang:1.24 AS b\nRUN go build\nFROM gcr.io/distroless/static:nonroot\nCOPY --from=b /app /app\nUSER nonroot\n"))
}
```

### Check
Q: Which is the best way to give a container a database password?
T: mcq
- [ ] `ENV DB_PASSWORD=...` in the Dockerfile
- [ ] COPY a .env file into the image
- [x] Inject it at runtime (orchestrator secret, env var from a secret store)
- [ ] Hard-code it in Go
E: Anything baked into an image layer can be extracted by anyone who can pull the image.

Q: Why run a container process as a non-root user?
T: mcq
- [ ] Faster startup
- [x] Limits the damage if the process is compromised
- [ ] Required by Docker
- [ ] It saves memory
E: Least privilege reduces the blast radius of a vulnerability.

## Environment Configuration
slug: environment-configuration
minutes: 7
objectives: Configure containers with environment variables; Provide sane defaults and fail fast on missing required values; Support _FILE variables for secrets
takeaways: Twelve-factor apps read config from environment variables; Required settings must fail at startup with a clear message; A KEY_FILE convention lets Docker/Kubernetes mount secrets as files

### Concept
The same image runs in dev, staging and production; only the environment changes.

```dockerfile
ENV PORT=8080
```
```bash
docker run -e DATABASE_URL=postgres://... -e LOG_LEVEL=debug myapp
docker run --env-file .env myapp        # for local development only
```
In Go, parse into a struct once at startup: defaults for optional values, an error for missing required ones. **Secrets** are better delivered as files (Docker/Kubernetes mount them): support `DB_PASSWORD_FILE=/run/secrets/db_password` — read the file when set.

### Example
```go
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

func get(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func secret(key string) (string, error) {
	if path := os.Getenv(key + "_FILE"); path != "" {
		b, err := os.ReadFile(path)
		return strings.TrimSpace(string(b)), err
	}
	if v := os.Getenv(key); v != "" {
		return v, nil
	}
	return "", errors.New(key + " is required")
}

func main() {
	fmt.Println(get("PORT", "8080"))
	_, err := secret("DB_PASSWORD")
	fmt.Println(err)
}
```

### Exercise
Write `required(getenv func(string) string, keys ...string) error` returning an error listing **all** missing/empty keys (`missing required env: A, C`), or nil when all are set.
```text expect
missing required env: DB_URL, TOKEN
<nil>
```
```go solution
package main

import (
	"fmt"
	"strings"
)

// BEGIN
func required(getenv func(string) string, keys ...string) error {
	var missing []string
	for _, k := range keys {
		if getenv(k) == "" {
			missing = append(missing, k)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required env: %s", strings.Join(missing, ", "))
	}
	return nil
}

// END

func main() {
	env := map[string]string{"PORT": "80"}
	get := func(k string) string { return env[k] }
	fmt.Println(required(get, "PORT", "DB_URL", "TOKEN"))
	env["DB_URL"], env["TOKEN"] = "x", "y"
	fmt.Println(required(get, "PORT", "DB_URL", "TOKEN"))
}
```

### Check
Q: Why report all missing environment variables at once?
T: mcq
- [ ] It is faster
- [x] The operator can fix everything in one deploy
- [ ] Docker requires it
- [ ] os.Getenv panics otherwise
E: Failing on the first missing variable causes a slow fix-redeploy loop.

Q: What does the `DB_PASSWORD_FILE` convention allow?
T: mcq
- [ ] Storing the password in the image
- [x] Reading the secret from a file mounted by the orchestrator instead of an env var
- [ ] Encrypting environment variables
- [ ] Setting defaults
E: Files avoid leaking secrets through process listings and environment dumps.

## Docker Compose
slug: docker-compose
status: planned

## Go + PostgreSQL
slug: go-postgresql
status: planned
