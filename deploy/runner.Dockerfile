# syntax=docker/dockerfile:1
# Sandbox image used to compile and run learner code. It is started per request
# with --network none, --read-only, --cap-drop ALL, CPU/memory/PID limits, and
# is destroyed afterwards (see internal/runner/docker.go).

FROM golang:1.26-alpine AS entry
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY internal/runner ./internal/runner
COPY cmd/sandbox-entrypoint ./cmd/sandbox-entrypoint
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/sandbox-entrypoint ./cmd/sandbox-entrypoint

FROM golang:1.26-alpine
# Pre-build the standard library packages lessons use so cold starts are fast.
# The entrypoint copies this seed into a tmpfs GOCACHE for each run.
ENV GOCACHE=/opt/gocache GOFLAGS=-buildvcs=false CGO_ENABLED=0 GOTOOLCHAIN=local
RUN mkdir -p /opt/gocache \
 && go build fmt errors strings strconv sort slices maps bytes bufio io os time math math/rand/v2 regexp \
      encoding/json encoding/csv path/filepath context sync sync/atomic unicode unicode/utf8 log/slog \
      net net/http net/http/httptest net/netip net/url html/template crypto/sha256 crypto/hmac crypto/rand \
      crypto/pbkdf2 crypto/subtle crypto/tls encoding/hex encoding/base64 os/exec os/signal syscall runtime \
      runtime/pprof testing testing/fstest testing/iotest container/list cmp flag go/token \
 && (go test -c -o /dev/null fmt >/dev/null 2>&1 || true) \
 && chmod -R a+rX /opt/gocache
COPY --from=entry /out/sandbox-entrypoint /usr/local/bin/sandbox-entrypoint
ENV GOCACHE=/tmp/gocache HOME=/tmp GOPROXY=off
USER 65534:65534
WORKDIR /work
ENTRYPOINT ["/usr/local/bin/sandbox-entrypoint"]
