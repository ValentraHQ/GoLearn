# Development Environment
id: setup
number: 00
track: core
paths: beginner
skill: fundamentals
summary: Install the toolchain, learn the go command, and run your first program.

## What is Go?
slug: what-is-go
minutes: 5
objectives: Describe what Go is and who created it; Explain what "compiled" and "statically typed" mean; Recognise where Go is used in industry
takeaways: Go is a compiled, statically typed, garbage-collected language; It produces a single static binary that is easy to deploy; Go powers much of the cloud-native ecosystem, including Docker and Kubernetes

### Concept
Go (often called Golang) is an open-source programming language created at Google and released in 2009. Its designers wanted a language that compiles fast, runs fast, and stays simple enough to read years later.

- **Compiled** — your code becomes a native executable; there is no interpreter or virtual machine to install.
- **Statically typed** — the compiler checks types before the program runs, catching a whole class of bugs early.
- **Garbage collected** — memory is reclaimed for you, but you still control layout with structs and pointers.
- **Concurrent by design** — goroutines and channels are part of the language.

Go is the language behind Docker, Kubernetes, Terraform, Prometheus and much more, which is why cloud and DevOps engineers learn it.

### Example
A complete Go program. `go run` compiles and runs it in one step.
```go
package main

import "fmt"

func main() {
	fmt.Println("Go is compiled and statically typed")
}
```

### Exercise
Print the three words `compiled`, `typed`, `concurrent` — one per line — using `fmt.Println`.
```text expect
compiled
typed
concurrent
```
```go solution
package main

import "fmt"

func main() {
	// BEGIN
	fmt.Println("compiled")
	fmt.Println("typed")
	fmt.Println("concurrent")
	// END
}
```

### Check
Q: Which statement about Go is true?
T: mcq
- [ ] Go programs need a virtual machine installed on the target machine
- [x] Go compiles to a native executable
- [ ] Go is dynamically typed
- [ ] Go was created to replace SQL
E: Go compiles to a native, statically linked executable. Types are checked at compile time and no VM is required.

Q: Go is statically typed, so most type errors are caught before the program runs.
T: tf
A: true
E: The compiler rejects programs with mismatched types, so many bugs never reach production.

## Why Go?
slug: why-go
minutes: 5
objectives: List the design goals of Go; Compare Go with other languages at a high level; Decide when Go is a good fit
takeaways: Go optimises for simplicity, fast builds and easy concurrency; Single-binary deployment suits containers and CLIs; Go is a poor fit for GUI-heavy or data-science-heavy work

### Concept
Go trades cleverness for clarity.

| Strength | What it means for you |
| --- | --- |
| Fast compile times | Tight feedback loop, even in large projects |
| Simple syntax | About 25 keywords; code reads the same everywhere |
| Built-in concurrency | Goroutines are cheap; thousands are normal |
| Static binaries | Copy one file to a server or a `scratch` container |
| Strong standard library | HTTP servers, JSON, crypto and testing ship in the box |
| Tooling | `gofmt`, `go test`, `go vet` and profilers are built in |

Good fits: network services, CLIs, infrastructure tooling, Kubernetes operators. Weaker fits: desktop GUIs, machine-learning research, tiny embedded targets.

### Example
Go's standard library lets you build a working web server without third-party dependencies.
```go norun
package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello from Go")
	})
	http.ListenAndServe(":8080", nil)
}
```

### Exercise
Print a two-line summary using `fmt.Printf` with the `%d` verb: `keywords: 25` on the first line and `binaries: 1` on the second.
```text expect
keywords: 25
binaries: 1
```
```go solution
package main

import "fmt"

func main() {
	// BEGIN
	fmt.Printf("keywords: %d\n", 25)
	fmt.Printf("binaries: %d\n", 1)
	// END
}
```

### Check
Q: Which project is a poor fit for Go?
T: mcq
- [ ] A command-line tool for cluster operations
- [ ] A network proxy
- [x] A native desktop GUI application with rich widgets
- [ ] A Kubernetes operator
E: Go excels at networked services and tooling. GUI toolkits exist but are not its strength.

## Installing Go
slug: installing-go
minutes: 6
objectives: Install Go from the official downloads; Verify the installation with go version; Understand GOROOT and GOPATH at a basic level
takeaways: Install from go.dev/dl and confirm with go version; The toolchain lives in GOROOT; your downloaded modules live under GOPATH; Recent Go versions can download the toolchain a module requires

### Concept
1. Download the installer for your OS from **go.dev/dl**.
2. Run it (macOS `.pkg`, Windows `.msi`) or extract the archive to `/usr/local/go` on Linux.
3. Make sure `/usr/local/go/bin` is on your `PATH`.
4. Open a **new** terminal and check the version.

Go supports side-by-side toolchains: the `toolchain` line in `go.mod` tells `go` to fetch the version a project needs.

### Example
```bash
go version
go env GOROOT GOPATH
```
The first command prints something like `go version go1.24.0 linux/amd64`.

### Exercise
The `runtime` package reports the toolchain that built the program. Print `true` if `runtime.Version()` starts with `go`, using `strings.HasPrefix`.
```text expect
true
```
```go solution
package main

import (
	"fmt"
	"runtime"
	"strings"
)

func main() {
	// BEGIN
	fmt.Println(strings.HasPrefix(runtime.Version(), "go"))
	// END
}
```

### Check
Q: Which command confirms Go is installed?
T: mcq
- [ ] go check
- [ ] go install
- [x] go version
- [ ] go doctor
E: `go version` prints the installed toolchain and platform.

## Go Toolchain
slug: toolchain
minutes: 7
objectives: Use go run, go build and go install; Format, vet and test code with the standard tools; Read go help output
takeaways: go run compiles and runs; go build produces a binary; gofmt and go vet keep code clean; go test runs tests; go doc reads documentation

### Concept
Everything is one command: `go`.

| Command | Purpose |
| --- | --- |
| `go run .` | Compile to a temp dir and run |
| `go build` | Produce an executable |
| `go install` | Build and place the binary in `$GOPATH/bin` |
| `go fmt ./...` | Format all code with `gofmt` |
| `go vet ./...` | Report suspicious constructs |
| `go test ./...` | Run tests |
| `go doc fmt.Println` | Read documentation in the terminal |
| `go mod tidy` | Sync dependencies |

### Example
```bash
go build -o hello ./cmd/hello
./hello
go vet ./...
go test ./...
```

### Exercise
Loop over the slice of commands and print each on its own line as `go <name>`.
```text expect
go run
go build
go test
```
```go solution
package main

import "fmt"

func main() {
	commands := []string{"run", "build", "test"}
	// BEGIN
	for _, c := range commands {
		fmt.Println("go " + c)
	}
	// END
}
```

### Check
Q: Which command formats your source files?
T: mcq
- [ ] go lint
- [x] go fmt
- [ ] go style
- [ ] go beautify
E: `go fmt` runs `gofmt`, the single canonical Go formatter. Nobody debates style in Go code review.

Q: `go build` writes an executable file to disk, while `go run` compiles to a temporary location.
T: tf
A: true
E: `go run` builds into a temp directory and executes it; `go build` leaves the binary for you.

## Go Workspace
slug: workspace
minutes: 6
objectives: Explain what a Go module is at a basic level; Create a project with go mod init; Recognise a typical project layout
takeaways: A module is a folder tree with a go.mod file at its root; go mod init creates go.mod; You no longer need to work inside GOPATH

### Concept
Modern Go uses **modules**. A module is a directory containing `go.mod`, which names the module and records its dependencies. You can create one anywhere.

```
hello/
├── go.mod
├── main.go
└── internal/
    └── greet/
        └── greet.go
```

`cmd/` holds executables, `internal/` holds code private to your module. You will design layouts properly in Module 07.

### Example
```bash
mkdir hello && cd hello
go mod init example.com/hello
```
```text
module example.com/hello

go 1.24
```

### Exercise
Print this directory tree exactly (use one `fmt.Println` per line).
```text expect
hello/
  go.mod
  main.go
```
```go solution
package main

import "fmt"

func main() {
	// BEGIN
	fmt.Println("hello/")
	fmt.Println("  go.mod")
	fmt.Println("  main.go")
	// END
}
```

### Check
Q: What file marks the root of a Go module?
T: short
A: go.mod
E: `go.mod` declares the module path and dependencies.

## VS Code Setup
slug: vscode-setup
minutes: 6
objectives: Install the Go extension and tools; Enable format on save; Use the debugger and test runner
takeaways: The Go extension installs gopls, the language server; Format on save keeps code gofmt-clean; Run and debug tests from the editor

### Concept
1. Install **VS Code** and the **Go** extension (publisher: Go Team at Google).
2. Open the command palette and run **Go: Install/Update Tools** to install `gopls`, `dlv` and friends.
3. Enable format-on-save.

```json
{
  "editor.formatOnSave": true,
  "[go]": {
    "editor.defaultFormatter": "golang.go"
  },
  "gopls": { "ui.semanticTokens": true }
}
```

You get autocomplete, go-to-definition, inline errors, test running and a debugger (Delve).

### Example
Formatting fixes indentation and spacing automatically:
```go
package main

import "fmt"

func main() {
	x := 1
	fmt.Println(x + 1)
}
```

### Exercise
The code below is valid but unformatted. Print the result of `x + 1` where `x := 41`; the output should be `42`.
```text expect
42
```
```go solution
package main

import "fmt"

func main() {
	// BEGIN
	x := 41
	fmt.Println(x + 1)
	// END
}
```

### Check
Q: Which tool provides autocomplete and diagnostics for Go in VS Code?
T: mcq
- [ ] gofmt
- [x] gopls
- [ ] gocov
- [ ] gomod
E: `gopls` is the official Go language server.

## Running Your First Go Program
slug: first-program
minutes: 6
objectives: Write main.go from scratch; Run it with go run; Build and execute a binary
takeaways: Every executable has package main and func main; go run for quick iteration; go build for a distributable binary

### Concept
1. Make a folder and `go mod init`.
2. Create `main.go`.
3. Run `go run .`.

Every executable program needs `package main` and a `main` function — that is where execution begins.

### Example
```go
package main

import "fmt"

func main() {
	fmt.Println("Hello, GoLearn")
}
```
```bash
go run .
go build -o hello .
./hello
```

### Exercise
Write the classic program: print `Hello, GoLearn`.
```text expect
Hello, GoLearn
```
```go solution
package main

import "fmt"

func main() {
	// BEGIN
	fmt.Println("Hello, GoLearn")
	// END
}
```

### Check
Q: Which function is the entry point of a Go executable?
T: mcq
- [ ] start
- [ ] init
- [x] main
- [ ] run
E: Execution starts at `func main()` in `package main`.

Q: What will this print?
T: output
```go
fmt.Println("Go", "rocks")
```
- [x] Go rocks
- [ ] Gorocks
- [ ] Go, rocks
- [ ] "Go" "rocks"
E: `Println` separates its operands with a space and adds a newline.
