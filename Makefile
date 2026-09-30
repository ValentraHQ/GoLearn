.PHONY: help web build test test-short content-check vet run dev runner-image docker clean

help:            ## show targets
	@grep -E '^[a-z-]+:.*##' $(MAKEFILE_LIST) | sed 's/:.*##/\t/'

web:             ## build the frontend into web/dist
	cd web && npm ci && npm run build

build: web       ## build the server binaries into bin/
	go build -trimpath -o bin/golearn ./cmd/golearn
	go build -trimpath -o bin/runner-daemon ./cmd/runner-daemon

vet:             ## static checks
	go vet ./...
	cd web && npm run typecheck

test-short:      ## fast Go tests (skips executing every lesson snippet)
	go test -short -race ./...

test:            ## all Go tests including content verification (needs a local Go toolchain)
	go test -race ./...
	cd web && npm test

content-check:   ## compile and run every lesson example, exercise solution and challenge
	go test ./internal/content/ -run TestContentRuns -count=1 -timeout 30m

run: build       ## run the server with SQLite (runner disabled unless the image exists)
	GOLEARN_STATIC_DIR=web/dist ./bin/golearn

dev:             ## API on :8080; run `npm run dev` in web/ for the Vite dev server
	GOLEARN_RUNNER=$${GOLEARN_RUNNER:-disabled} go run ./cmd/golearn

runner-image:    ## build the sandbox image used for code execution
	docker build -f deploy/runner.Dockerfile -t golearn-runner:latest .

docker:          ## build the API image
	docker build -f deploy/Dockerfile --target api -t golearn-api:latest .

clean:
	rm -rf bin web/dist
