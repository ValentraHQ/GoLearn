# syntax=docker/dockerfile:1
# GoLearn API + static frontend. Multi-stage: the final image is distroless and non-root.

FROM node:22-alpine AS web
WORKDIR /web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/golearn ./cmd/golearn \
 && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/runner-daemon ./cmd/runner-daemon

FROM gcr.io/distroless/static-debian12:nonroot AS api
COPY --from=build /out/golearn /golearn
COPY --from=web /web/dist /web/dist
ENV GOLEARN_ADDR=:8080 GOLEARN_STATIC_DIR=/web/dist
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/golearn"]

# The runner daemon needs the docker CLI to start sandbox containers.
FROM docker:27-cli AS runner-daemon
COPY --from=build /out/runner-daemon /usr/local/bin/runner-daemon
EXPOSE 9090
ENTRYPOINT ["/usr/local/bin/runner-daemon"]
