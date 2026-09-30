# Packages & Modules
id: modules
number: 07
track: core
paths: beginner, pro
skill: packages
requires: errors
project: reusable-go-library
summary: Organise code into packages, manage dependencies with modules, and design clean APIs.

## Packages
slug: packages
minutes: 6
objectives: Explain package-level scope; Use init functions and know the initialisation order; Recognise the standard package layout
takeaways: Package-level names are visible to every file in the package; Package variables initialise first, then init() functions, then main; Keep package names short, lowercase and singular

### Concept
Everything declared at package level — functions, types, variables, constants — is visible to **all files** in that package, in any order.

Initialisation order: imported packages first, then package-level variables (in dependency order), then every `init()` function, then `main`. Use `init` sparingly; explicit setup functions are easier to test.

Naming: `package http`, not `package httputils`; callers write `http.Get`, so don't repeat the package name (`http.HTTPClient` stutters).

### Example
```go
package main

import "fmt"

var config = load() // runs before init and main

func load() string {
	fmt.Println("1. package variable initialised")
	return "ready"
}

func init() {
	fmt.Println("2. init runs")
}

func main() {
	fmt.Println("3. main runs with", config)
}
```

### Exercise
Add an `init` function that prints `init` so the output is `init` followed by `main`.
```text expect
init
main
```
```go solution
package main

import "fmt"

// BEGIN
func init() {
	fmt.Println("init")
}

// END

func main() {
	fmt.Println("main")
}
```

### Check
Q: In what order do things run?
T: mcq
- [ ] main, init, package variables
- [x] package variables, init functions, main
- [ ] init, package variables, main
- [ ] main only
E: Package-level variables are initialised first, then `init()` functions, then `main`.

Q: `http.HTTPClient` is a good name for a type in package `http`.
T: tf
A: false
E: It stutters; the caller would write `http.HTTPClient`. Prefer `http.Client`.

## Package Organization
slug: package-organization
minutes: 6
objectives: Lay out a typical Go project; Use internal/ to hide implementation; Avoid over-splitting into tiny packages
takeaways: cmd/ holds executables, internal/ holds private packages; Organise by responsibility, not by type ("models", "utils"); Start with few packages and split when boundaries emerge

### Concept
```
myapp/
├── go.mod
├── cmd/
│   └── myapp/main.go     # thin main
├── internal/             # importable only from inside this module
│   ├── billing/
│   └── storage/
└── pkg/                  # optional: public, reusable packages
```
- Keep `main` thin: parse flags, wire dependencies, call into packages.
- `internal/` is enforced by the compiler.
- Avoid `utils`, `common`, `models` packages: they become dumping grounds with no cohesive purpose. Name packages for what they **provide**.

### Example
```go
package main

import (
	"fmt"
	"sort"
	"strings"
)

func main() {
	files := []string{"cmd/app/main.go", "internal/billing/invoice.go", "go.mod", "internal/storage/db.go"}
	sort.Strings(files)
	for _, f := range files {
		fmt.Println(strings.Repeat("  ", strings.Count(f, "/")) + f[strings.LastIndex(f, "/")+1:])
	}
}
```

### Exercise
Write `isInternal(path string) bool` reporting whether an import path contains an `internal` element (`strings.Contains(path, "/internal/")` or prefix).
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
func isInternal(path string) bool {
	return strings.HasPrefix(path, "internal/") || strings.Contains(path, "/internal/")
}

// END

func main() {
	fmt.Println(isInternal("example.com/app/internal/billing"), isInternal("example.com/app/api"))
}
```

### Check
Q: Who may import a package located in `internal/`?
T: mcq
- [ ] Anyone
- [x] Code inside the parent of the internal directory
- [ ] Only main packages
- [ ] Only tests
E: The compiler restricts `internal` packages to code rooted at the parent of that directory.

Q: `utils` is a well-named package.
T: tf
A: false
E: It says nothing about what the package provides and tends to accumulate unrelated code.

## Exported Names
slug: exported-names
minutes: 5
objectives: Recognise exported identifiers; Understand visibility is per package; Document exported names
takeaways: A name starting with an uppercase letter is exported; Visibility is per package — there is no protected or friend; Every exported name should have a doc comment starting with its name

### Concept
Capitalisation controls visibility: `Println` is exported, `println` is not. This applies to functions, types, fields, methods, constants and variables.

```go norun
// Rectangle describes a rectangle. (Doc comments start with the name.)
type Rectangle struct {
	Width  float64 // exported field
	height float64 // unexported field
}
```
Everything you export is a **promise** to your users. Export the minimum.

### Example
```go
package main

import (
	"fmt"
	"go/token"
)

func main() {
	for _, n := range []string{"Println", "println", "Client", "_x", "X"} {
		fmt.Println(n, token.IsExported(n))
	}
}
```

### Exercise
Write `exported(names []string) []string` returning only the exported names, using `token.IsExported`.
```text expect
[Open Close]
```
```go solution
package main

import (
	"fmt"
	"go/token"
)

// BEGIN
func exported(names []string) []string {
	var out []string
	for _, n := range names {
		if token.IsExported(n) {
			out = append(out, n)
		}
	}
	return out
}

// END

func main() {
	fmt.Println(exported([]string{"Open", "read", "Close", "buf"}))
}
```

### Check
Q: Which identifier is exported?
T: mcq
- [ ] name
- [ ] _Name
- [x] Name
- [ ] nAME
E: Only names beginning with an uppercase Unicode letter are exported.

Q: Go has a `protected` visibility level for subclasses.
T: tf
A: false
E: There is no inheritance and only two levels: exported and unexported.

## Private Names
slug: private-names
minutes: 6
objectives: Hide fields behind methods; Enforce invariants with constructors; Understand that privacy is package-level
takeaways: Unexported fields can only be touched inside the package; Constructors plus unexported fields enforce invariants; Privacy is per package, not per type

### Concept
Keep fields unexported and expose behaviour instead. That lets you enforce rules and change internals freely.

```go norun
type Temperature struct{ celsius float64 }

func NewTemperature(c float64) (Temperature, error) {
	if c < -273.15 {
		return Temperature{}, errors.New("below absolute zero")
	}
	return Temperature{c}, nil
}
func (t Temperature) Celsius() float64 { return t.celsius }
```
Code in other packages cannot build an invalid `Temperature`. (Inside the same package any file can still read `celsius` — privacy is per package.)

### Example
```go
package main

import (
	"errors"
	"fmt"
)

type Percent struct{ v int }

func NewPercent(v int) (Percent, error) {
	if v < 0 || v > 100 {
		return Percent{}, errors.New("out of range")
	}
	return Percent{v}, nil
}

func (p Percent) Value() int { return p.v }

func main() {
	p, _ := NewPercent(40)
	_, err := NewPercent(140)
	fmt.Println(p.Value(), err)
}
```

### Exercise
Write `NewEven(n int) (Even, error)` which fails with `not even` for odd numbers; `Even` has an unexported field `n` and a `Value()` method.
```text expect
8 <nil>
0 not even
```
```go solution
package main

import (
	"errors"
	"fmt"
)

// BEGIN
type Even struct{ n int }

func NewEven(n int) (Even, error) {
	if n%2 != 0 {
		return Even{}, errors.New("not even")
	}
	return Even{n}, nil
}

func (e Even) Value() int { return e.n }

// END

func main() {
	a, err := NewEven(8)
	fmt.Println(a.Value(), err)
	b, err := NewEven(3)
	fmt.Println(b.Value(), err)
}
```

### Check
Q: Why keep a struct's fields unexported?
T: mcq
- [ ] It makes the program faster
- [x] Callers can't create invalid states and you can change internals safely
- [ ] It is required by the compiler
- [ ] It prevents copying
E: Encapsulation behind a small exported API preserves invariants and keeps the API stable.

Q: Code in a different file of the same package can read an unexported field.
T: tf
A: true
E: Visibility boundaries are packages, not files or types.

## Imports
slug: imports
minutes: 5
objectives: Read and write import paths; Group imports by origin; Understand import cycles
takeaways: Import paths combine the module path with the package directory; Group standard library, third-party and local imports separately; Go forbids import cycles

### Concept
For a module `example.com/shop`, the package in `internal/billing` is imported as `"example.com/shop/internal/billing"`; the code refers to it as `billing`.

Conventional grouping (goimports does this):

```go norun
import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"example.com/shop/internal/billing"
)
```
**Import cycles** (A imports B imports A) are compile errors. They signal tangled responsibilities: extract the shared piece into a third package or define an interface in the consumer.

### Example
```go
package main

import (
	"fmt"
	"strings"
)

func group(path string) string {
	switch {
	case !strings.Contains(strings.Split(path, "/")[0], "."):
		return "stdlib"
	case strings.HasPrefix(path, "example.com/"):
		return "local"
	default:
		return "third-party"
	}
}

func main() {
	for _, p := range []string{"fmt", "github.com/google/uuid", "example.com/shop/billing"} {
		fmt.Println(p, "->", group(p))
	}
}
```

### Exercise
Write `pkgName(path string) string` returning the last element of an import path (`"net/http"` → `"http"`), using `strings.LastIndex`.
```text expect
http uuid
```
```go solution
package main

import (
	"fmt"
	"strings"
)

// BEGIN
func pkgName(path string) string {
	return path[strings.LastIndex(path, "/")+1:]
}

// END

func main() {
	fmt.Println(pkgName("net/http"), pkgName("github.com/google/uuid"))
}
```

### Check
Q: What does the Go compiler do with an import cycle?
T: mcq
- [ ] Warns and continues
- [ ] Resolves it automatically
- [x] Refuses to compile
- [ ] Loads packages lazily
E: Cycles are forbidden; restructure your packages.

Q: What is the usual way to break an import cycle between package A and package B?
T: mcq
- [ ] Merge everything into main
- [x] Extract shared code into a third package or invert the dependency with an interface
- [ ] Use init functions
- [ ] Use a build tag
E: Depend on abstractions or move the shared parts out.

## Go Modules
slug: go-modules
minutes: 6
objectives: Explain what a module is; Create a module with go mod init; Understand semantic import versioning
takeaways: A module is a tree of packages versioned together, rooted at go.mod; The module path is the import path prefix; Major versions v2+ are part of the import path

### Concept
A **module** is the unit of versioning and distribution: a directory tree with a `go.mod` at its root. Its **module path** (usually a repository URL like `github.com/you/project`) prefixes the import path of every package inside.

```
go mod init github.com/you/project
```
Versions follow **semver** (`v1.4.2`). For `v2` and above the import path gains a suffix (`github.com/you/lib/v2`), so v1 and v2 can coexist in one build.

### Example
```go
package main

import (
	"fmt"
	"strings"
)

// modulePath extracts the path from the first line of a go.mod.
func modulePath(gomod string) string {
	first, _, _ := strings.Cut(gomod, "\n")
	return strings.TrimSpace(strings.TrimPrefix(first, "module"))
}

func main() {
	fmt.Println(modulePath("module github.com/you/project\n\ngo 1.24\n"))
}
```

### Exercise
Write `majorSuffix(v string) string` returning `"/v2"` for `"v2.3.1"` and `""` for v0 and v1 versions.
```text expect
"/v2" "" ""
```
```go solution
package main

import (
	"fmt"
	"strings"
)

// BEGIN
func majorSuffix(v string) string {
	major, _, _ := strings.Cut(strings.TrimPrefix(v, "v"), ".")
	if major == "0" || major == "1" {
		return ""
	}
	return "/v" + major
}

// END

func main() {
	fmt.Printf("%q %q %q\n", majorSuffix("v2.3.1"), majorSuffix("v1.9.0"), majorSuffix("v0.4.0"))
}
```

### Check
Q: What does the module path in go.mod define?
T: mcq
- [ ] The Go version
- [x] The import path prefix for all packages in the module
- [ ] The file name of the binary
- [ ] The Git branch
E: Every package's import path starts with the module path.

Q: How is module version `v3.0.0` reflected in import paths?
T: mcq
- [ ] It isn't
- [x] The module path ends with /v3
- [ ] It uses a build tag
- [ ] It is a directory named v3 only in GOPATH
E: Semantic import versioning: major versions ≥ 2 are part of the import path.

## go.mod
slug: go-mod-file
minutes: 6
objectives: Read every directive in go.mod; Set the go and toolchain lines; Use replace and exclude appropriately
takeaways: go.mod declares the module path, Go version and requirements; replace redirects a module (useful for local development); Don't edit require lines by hand — use go get and go mod tidy

### Concept
```text
module github.com/you/shop

go 1.24

require (
	github.com/google/uuid v1.6.0
	golang.org/x/sync v0.8.0
)

replace github.com/you/lib => ../lib
```
- `go` — minimum language version.
- `require` — direct and indirect dependencies with **minimum** versions.
- `replace` — point a dependency at a local path or fork.
- `toolchain` — the exact toolchain to use.

### Example
```go
package main

import (
	"fmt"
	"strings"
)

const gomod = `module github.com/you/shop

go 1.24

require (
	github.com/google/uuid v1.6.0
	golang.org/x/sync v0.8.0
)
`

func main() {
	for _, line := range strings.Split(gomod, "\n") {
		if strings.HasPrefix(line, "go ") {
			fmt.Println("go version:", strings.TrimPrefix(line, "go "))
		}
	}
}
```

### Exercise
Parse the `require` block and print each module path (first field of each indented line), one per line.
```text expect
github.com/google/uuid
golang.org/x/sync
```
```go solution
package main

import (
	"fmt"
	"strings"
)

const gomod = `module github.com/you/shop

go 1.24

require (
	github.com/google/uuid v1.6.0
	golang.org/x/sync v0.8.0
)
`

func main() {
	// BEGIN
	for _, line := range strings.Split(gomod, "\n") {
		if strings.HasPrefix(line, "\t") {
			fmt.Println(strings.Fields(line)[0])
		}
	}
	// END
}
```

### Check
Q: Which directive points a dependency at a local directory?
T: short
A: replace
E: `replace example.com/lib => ../lib` redirects imports to that path.

Q: What does the `go` directive specify?
T: mcq
- [ ] The exact toolchain
- [x] The minimum Go language version for the module
- [ ] The maximum Go version
- [ ] The Go proxy
E: It sets the language version; `toolchain` can pin a specific toolchain.

## go.sum
slug: go-sum
minutes: 5
objectives: Explain the purpose of go.sum; Understand cryptographic hashes; Commit go.sum to version control
takeaways: go.sum records the expected cryptographic hash of every dependency version; It makes builds reproducible and detects tampering; Always commit go.sum

### Concept
`go.sum` lists hashes such as `github.com/google/uuid v1.6.0 h1:NIvaJDMOsjHA8n1jAhLSgzrAzy1Hgr+hNrb57e+94F0=`. When Go downloads a module it checks the hash; a mismatch stops the build, protecting you from a modified or re-tagged dependency. The checksum database (`sum.golang.org`) provides a second, public source of truth.

It is not a lock file (versions come from `go.mod`), but it is essential — **commit it**.

### Example
```go
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func main() {
	sum := sha256.Sum256([]byte("module content v1"))
	fmt.Println(hex.EncodeToString(sum[:])[:16])
	sum = sha256.Sum256([]byte("module content v1 (tampered)"))
	fmt.Println(hex.EncodeToString(sum[:])[:16])
}
```

### Exercise
Write `verify(content, want string) bool` comparing the hex SHA-256 of `content` with `want`.
```text expect
true false
```
```go solution
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// BEGIN
func verify(content, want string) bool {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:]) == want
}

// END

func main() {
	sum := sha256.Sum256([]byte("hello"))
	good := hex.EncodeToString(sum[:])
	fmt.Println(verify("hello", good), verify("hellO", good))
}
```

### Check
Q: What is `go.sum` for?
T: mcq
- [ ] Listing direct dependencies
- [x] Verifying that downloaded modules match known hashes
- [ ] Caching compiled code
- [ ] Locking the Go version
E: It records expected checksums so tampering or drift is detected.

Q: You should add `go.sum` to `.gitignore`.
T: tf
A: false
E: Commit it so every developer and CI run verifies the same hashes.

## Dependencies
slug: dependencies
minutes: 6
objectives: Add a dependency with go get; Understand direct versus indirect dependencies; Vendor when required
takeaways: Importing a package and running go mod tidy (or go get) adds the dependency; Indirect dependencies are needed by your dependencies; Prefer the standard library — every dependency is a cost

### Concept
```bash
go get github.com/google/uuid@v1.6.0   # add or change a version
go get -u ./...                         # upgrade dependencies
go list -m all                          # every module in the build
go mod why github.com/some/pkg          # why is it needed?
go mod vendor                           # copy deps into vendor/
```
Marked `// indirect` in go.mod means your code doesn't import it directly. Before adding a dependency ask: can the standard library do this? Is it maintained? What is its licence and how many transitive dependencies does it bring?

### Example
```go
package main

import (
	"fmt"
	"sort"
	"strings"
)

func main() {
	lines := []string{
		"github.com/google/uuid v1.6.0",
		"golang.org/x/text v0.14.0 // indirect",
		"golang.org/x/sync v0.8.0",
	}
	var direct []string
	for _, l := range lines {
		if !strings.Contains(l, "// indirect") {
			direct = append(direct, strings.Fields(l)[0])
		}
	}
	sort.Strings(direct)
	fmt.Println(direct)
}
```

### Exercise
Write `indirect(lines []string) int` counting lines marked `// indirect`.
```text expect
2
```
```go solution
package main

import (
	"fmt"
	"strings"
)

// BEGIN
func indirect(lines []string) int {
	n := 0
	for _, l := range lines {
		if strings.HasSuffix(l, "// indirect") {
			n++
		}
	}
	return n
}

// END

func main() {
	fmt.Println(indirect([]string{"a v1 // indirect", "b v2", "c v3 // indirect"}))
}
```

### Check
Q: What does `// indirect` next to a requirement mean?
T: mcq
- [ ] It is optional
- [x] Your module doesn't import it directly; a dependency needs it
- [ ] It is deprecated
- [ ] It is vendored
E: Indirect requirements are recorded so builds are reproducible.

Q: Which command explains why a module is in your build?
T: short
A: go mod why
E: `go mod why <module>` prints the import chain that requires it.

## Version Management
slug: version-management
minutes: 7
objectives: Read semantic versions; Explain minimal version selection; Tag releases for your own modules
takeaways: Semver: MAJOR.MINOR.PATCH — breaking, feature, fix; Go picks the minimum version that satisfies all requirements (MVS); Tag releases with git tags like v1.2.3

### Concept
Semantic versioning: `v1.4.2` = major 1, minor 4, patch 2. Breaking changes bump the **major** version.

Go uses **Minimal Version Selection**: if A needs `lib v1.2.0` and B needs `lib v1.4.0`, Go picks `v1.4.0` — the *highest of the minimums*, never the newest available. Builds stay reproducible until you explicitly upgrade.

Publish a release by tagging: `git tag v0.1.0 && git push --tags`. Versions before `v1.0.0` make no compatibility promises.

### Example
```go
package main

import (
	"fmt"
	"strconv"
	"strings"
)

type Version struct{ Major, Minor, Patch int }

func Parse(s string) Version {
	parts := strings.Split(strings.TrimPrefix(s, "v"), ".")
	n := func(i int) int { v, _ := strconv.Atoi(parts[i]); return v }
	return Version{n(0), n(1), n(2)}
}

func main() {
	fmt.Printf("%+v\n", Parse("v1.4.2"))
}
```

### Exercise
Write `Less(a, b Version) bool` for semantic version ordering and print the highest of `v1.2.0` and `v1.4.0` (what MVS would choose).
```text expect
v1.4.0
```
```go solution
package main

import (
	"fmt"
	"strconv"
	"strings"
)

type Version struct{ Major, Minor, Patch int }

func Parse(s string) Version {
	parts := strings.Split(strings.TrimPrefix(s, "v"), ".")
	n := func(i int) int { v, _ := strconv.Atoi(parts[i]); return v }
	return Version{n(0), n(1), n(2)}
}

// BEGIN
func Less(a, b Version) bool {
	if a.Major != b.Major {
		return a.Major < b.Major
	}
	if a.Minor != b.Minor {
		return a.Minor < b.Minor
	}
	return a.Patch < b.Patch
}

// END

func main() {
	a, b := "v1.2.0", "v1.4.0"
	if Less(Parse(a), Parse(b)) {
		fmt.Println(b)
	} else {
		fmt.Println(a)
	}
}
```

### Check
Q: Module A requires `lib v1.2.0`, module B requires `lib v1.4.0`. Which version does Go build with?
T: mcq
- [ ] v1.2.0
- [x] v1.4.0
- [ ] The latest published version
- [ ] It's an error
E: MVS chooses the highest of the required minimums.

Q: A breaking API change should bump which part of the version?
T: mcq
- [ ] Patch
- [ ] Minor
- [x] Major
- [ ] None
E: Semver reserves the major version for incompatible changes.

## go mod tidy
slug: go-mod-tidy
minutes: 5
objectives: Explain what go mod tidy adds and removes; Use it before every commit; Know how it interacts with go.sum
takeaways: go mod tidy adds missing requirements and removes unused ones; It also refreshes go.sum; Run it whenever imports change and commit the result

### Concept
`go mod tidy` scans your code (including tests) and makes `go.mod` and `go.sum` match reality:

- adds modules you import but haven't required,
- removes requirements nothing imports,
- adds missing hashes to `go.sum`.

A common CI check: run `go mod tidy` and fail if `git diff` is not empty, so nobody commits a stale `go.mod`.

### Example
```go
package main

import (
	"fmt"
	"sort"
)

func main() {
	required := map[string]bool{"a": true, "b": true, "c": true}
	imported := map[string]bool{"a": true, "c": true, "d": true}
	var add, remove []string
	for m := range imported {
		if !required[m] {
			add = append(add, m)
		}
	}
	for m := range required {
		if !imported[m] {
			remove = append(remove, m)
		}
	}
	sort.Strings(add)
	sort.Strings(remove)
	fmt.Println("add", add, "remove", remove)
}
```

### Exercise
Write `unused(required, imported []string) []string` returning modules that are required but not imported, in the original order (what `tidy` would remove).
```text expect
[b]
```
```go solution
package main

import "fmt"

// BEGIN
func unused(required, imported []string) []string {
	used := make(map[string]bool, len(imported))
	for _, m := range imported {
		used[m] = true
	}
	var out []string
	for _, m := range required {
		if !used[m] {
			out = append(out, m)
		}
	}
	return out
}

// END

func main() {
	fmt.Println(unused([]string{"a", "b", "c"}, []string{"a", "c"}))
}
```

### Check
Q: Which does `go mod tidy` NOT do?
T: mcq
- [ ] Remove unused requirements
- [ ] Add missing requirements
- [x] Upgrade every dependency to the latest version
- [ ] Update go.sum
E: Tidy syncs go.mod with your imports; upgrading is `go get -u`.

Q: A CI job that runs `go mod tidy` and checks for a clean `git diff` catches stale dependency files.
T: tf
A: true
E: If tidy changes anything, the committed files were out of date.

## Package Design
slug: package-design
minutes: 8
objectives: Design a small, coherent public API; Apply "accept interfaces, return structs"; Write example-driven documentation
takeaways: A package should do one thing and its name should say what; Accept interfaces, return concrete types; Document exported names and provide Example functions

### Concept
Good package design in Go:

1. **Cohesion** — one purpose; the name is a noun (`ratelimit`, `slug`).
2. **Small API** — export little. You can always add, never remove.
3. **Accept interfaces, return structs** — take `io.Reader`, return `*Parser`.
4. **Zero value useful** — `var b bytes.Buffer` works.
5. **No global state** — configuration through constructors and options.
6. **Document** — doc comments and runnable `Example` functions (verified by `go test`).

Functional options give extensible constructors without breaking callers: `NewServer(addr, WithTimeout(5*time.Second))`.

### Example
```go
package main

import (
	"fmt"
	"time"
)

type Client struct {
	timeout time.Duration
	retries int
}

type Option func(*Client)

func WithTimeout(d time.Duration) Option { return func(c *Client) { c.timeout = d } }
func WithRetries(n int) Option           { return func(c *Client) { c.retries = n } }

func New(opts ...Option) *Client {
	c := &Client{timeout: 5 * time.Second, retries: 3} // sensible defaults
	for _, o := range opts {
		o(c)
	}
	return c
}

func main() {
	c := New(WithRetries(5))
	fmt.Println(c.timeout, c.retries)
}
```

### Exercise
Add a `WithName(string)` option and print `name=api retries=3` for `New(WithName("api"))`.
```text expect
name=api retries=3
```
```go solution
package main

import "fmt"

type Client struct {
	name    string
	retries int
}

type Option func(*Client)

// BEGIN
func WithName(n string) Option { return func(c *Client) { c.name = n } }

// END

func New(opts ...Option) *Client {
	c := &Client{name: "default", retries: 3}
	for _, o := range opts {
		o(c)
	}
	return c
}

func main() {
	c := New(WithName("api"))
	fmt.Printf("name=%s retries=%d\n", c.name, c.retries)
}
```

### Check
Q: "Accept interfaces, return structs" means...
T: mcq
- [ ] Never return interfaces from any function
- [x] Parameters should be as general as needed; return concrete types so callers get full functionality
- [ ] Structs should have no methods
- [ ] Interfaces should always be exported
E: Flexible inputs, concrete outputs: callers can pass fakes and still use all of a concrete type's methods.

Q: What is the benefit of the functional options pattern?
T: mcq
- [ ] It's faster than struct literals
- [x] New options can be added without breaking existing callers
- [ ] It removes the need for constructors
- [ ] It makes fields exported
E: Variadic option functions keep constructor signatures stable as configuration grows.
