# CLI Development
id: cli
number: 16
track: devops
paths: beginner, pro
skill: cli
requires: stdlib
project: professional-devops-cli
summary: Build professional command-line tools: flags, subcommands, config, logging and exit codes.

## Arguments
slug: arguments
status: planned

## Flags
slug: flags
minutes: 8
objectives: Define typed flags and custom flag types; Parse them in a testable way with FlagSet; Write helpful usage output
takeaways: flag.String/Int/Bool/Duration define typed flags; Implement flag.Value to support repeated or custom-typed flags; Use flag.NewFlagSet(args) so parsing can be unit-tested

### Concept
Example tools: `golearn`, `kubectl-helper`, `cluster-health`, `deployment-check`, `log-analyzer` — all start with argument parsing.

```go norun
fs := flag.NewFlagSet("check", flag.ContinueOnError)
timeout := fs.Duration("timeout", 5*time.Second, "request timeout")
verbose := fs.Bool("v", false, "verbose output")
fs.Var(&tags, "tag", "repeatable: -tag a -tag b")
fs.Parse(os.Args[1:])
```
`-flag=value`, `-flag value` and `--flag` all work; flags must precede positional arguments. A custom type just needs `String() string` and `Set(string) error` (the `flag.Value` interface). For richer UX (grouped flags, aliases, completions) most teams use **Cobra** (see the last lesson).

### Example
```go
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

type list []string

func (l *list) String() string     { return strings.Join(*l, ",") }
func (l *list) Set(v string) error { *l = append(*l, v); return nil }

func main() {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	timeout := fs.Duration("timeout", 5*time.Second, "request timeout")
	var tags list
	fs.Var(&tags, "tag", "repeatable tag")
	_ = fs.Parse([]string{"-timeout", "2s", "-tag", "a", "-tag", "b", "target"})
	fmt.Println(*timeout, tags, fs.Args())
}
```

### Exercise
Implement the `Set` method of `list` so the flag can be repeated, then parse `-tag x -tag y` and print the joined tags.
```text expect
x,y
```
```go solution
package main

import (
	"flag"
	"fmt"
	"strings"
)

type list []string

func (l *list) String() string { return strings.Join(*l, ",") }

// BEGIN
func (l *list) Set(v string) error {
	*l = append(*l, v)
	return nil
}

// END

func main() {
	fs := flag.NewFlagSet("demo", flag.ContinueOnError)
	var tags list
	fs.Var(&tags, "tag", "repeatable")
	_ = fs.Parse([]string{"-tag", "x", "-tag", "y"})
	fmt.Println(tags.String())
}
```

### Check
Q: Which interface makes a custom type usable as a flag?
T: mcq
- [ ] fmt.Stringer only
- [x] flag.Value (String and Set)
- [ ] io.Reader
- [ ] encoding.TextMarshaler
E: `flag.Var` accepts any `flag.Value`.

Q: Why use `flag.NewFlagSet` instead of the global flag functions?
T: mcq
- [ ] It is faster
- [x] Parsing can be tested with explicit args and multiple subcommands can have their own flags
- [ ] Global flags are removed in newer Go versions
- [ ] It supports environment variables
E: A FlagSet is a self-contained parser you can call with any argument slice.

## Subcommands
slug: subcommands
minutes: 8
objectives: Dispatch to subcommands like git and kubectl; Give each subcommand its own FlagSet; Return exit codes from a testable run function
takeaways: Read the subcommand from args[0] and hand the rest to its handler; Structure main as os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) so run is testable; Unknown or missing subcommands print usage and exit with code 2

### Concept
```go norun
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		usage(stderr)
		return 2
	}
	switch args[0] {
	case "check":
		return cmdCheck(args[1:], stdout, stderr)
	case "version":
		fmt.Fprintln(stdout, version)
		return 0
	default:
		fmt.Fprintf(stderr, "unknown command %q\n", args[0])
		return 2
	}
}

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }
```
Passing writers instead of using `os.Stdout` directly lets tests capture output; returning an int instead of calling `os.Exit` deep inside keeps deferred cleanup working.

### Example
```go
package main

import (
	"fmt"
	"io"
	"os"
)

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage: tool <version|hello>")
		return 2
	}
	switch args[0] {
	case "version":
		fmt.Fprintln(stdout, "v1.0.0")
	case "hello":
		fmt.Fprintln(stdout, "hello")
	default:
		fmt.Fprintf(stderr, "unknown command %q\n", args[0])
		return 2
	}
	return 0
}

func main() {
	code := run([]string{"version"}, os.Stdout, os.Stderr)
	fmt.Println("exit", code)
}
```

### Exercise
Add a `sum` subcommand printing the sum of integer arguments (`sum 1 2 3` → `6`). A non-integer argument writes `error: bad number "x"` to stderr and returns 1. Run the two cases.
```text expect
6 0
error: bad number "x" 1
```
```go solution
package main

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// BEGIN
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage: tool sum <numbers...>")
		return 2
	}
	switch args[0] {
	case "sum":
		total := 0
		for _, a := range args[1:] {
			n, err := strconv.Atoi(a)
			if err != nil {
				fmt.Fprintf(stderr, "error: bad number %q\n", a)
				return 1
			}
			total += n
		}
		fmt.Fprintln(stdout, total)
		return 0
	}
	fmt.Fprintf(stderr, "unknown command %q\n", args[0])
	return 2
}

// END

func main() {
	for _, args := range [][]string{{"sum", "1", "2", "3"}, {"sum", "1", "x"}} {
		var out, errb bytes.Buffer
		code := run(args, &out, &errb)
		fmt.Println(strings.TrimSpace(out.String()+errb.String()), code)
	}
}
```

### Check
Q: Why should `run` return an int instead of calling `os.Exit`?
T: mcq
- [ ] `os.Exit` is deprecated
- [x] Deferred functions still run and the function is testable
- [ ] It is faster
- [ ] Exit codes are ignored
E: `os.Exit` skips deferred calls and can't be observed in tests.

Q: Which exit code conventionally signals incorrect command usage?
T: short
A: 2
E: Many CLIs use 1 for runtime failures and 2 for usage errors.

## Configuration
slug: configuration
minutes: 8
objectives: Layer configuration: defaults → file → environment → flags; Make precedence explicit; Validate configuration early
takeaways: A common precedence is flags > environment > config file > defaults; Load into one typed struct and validate it at startup; Never log secrets

### Concept
Users expect the same setting to be changeable in several ways. The **12-factor** convention, highest priority first:

1. command-line flag (`--port 9000`)
2. environment variable (`APP_PORT=9000`)
3. config file (`~/.config/app/config.yaml`)
4. built-in default

Implement it with a single `Config` struct filled in layers, each layer only overriding fields it actually set. Then `Validate()` and fail fast with a clear message. Libraries like `viper` automate the layering, but the hand-rolled version is short and dependency-free.

### Example
```go
package main

import "fmt"

type Config struct {
	Host string
	Port int
}

func resolve(defaults, file, env, flags map[string]string) Config {
	pick := func(key string) string {
		for _, layer := range []map[string]string{flags, env, file, defaults} {
			if v, ok := layer[key]; ok && v != "" {
				return v
			}
		}
		return ""
	}
	var c Config
	c.Host = pick("host")
	fmt.Sscanf(pick("port"), "%d", &c.Port)
	return c
}

func main() {
	c := resolve(
		map[string]string{"host": "localhost", "port": "8080"},
		map[string]string{"port": "9000"},
		map[string]string{"host": "0.0.0.0"},
		nil,
	)
	fmt.Printf("%+v\n", c)
}
```

### Exercise
Write `lookup(key string, layers ...map[string]string) (string, bool)` returning the value from the **first** layer that has a non-empty entry (layers are ordered highest priority first).
```text expect
9 true
 false
```
```go solution
package main

import "fmt"

// BEGIN
func lookup(key string, layers ...map[string]string) (string, bool) {
	for _, l := range layers {
		if v, ok := l[key]; ok && v != "" {
			return v, true
		}
	}
	return "", false
}

// END

func main() {
	flags := map[string]string{"port": "9"}
	file := map[string]string{"port": "1", "host": "h"}
	fmt.Println(lookup("port", flags, file))
	fmt.Println(lookup("nope", flags, file))
}
```

### Check
Q: In the usual precedence, which source overrides an environment variable?
T: mcq
- [ ] The config file
- [x] A command-line flag
- [ ] Defaults
- [ ] Nothing
E: The most explicit, closest-to-the-user input (flags) wins.

Q: Configuration should be validated...
T: mcq
- [ ] lazily, when first used
- [x] once at startup, failing fast with a clear message
- [ ] never
- [ ] only in tests
E: Discovering a bad setting an hour into a run is painful.

## Logging
slug: logging
minutes: 8
objectives: Log structured key/value data with log/slog; Configure levels and handlers; Separate diagnostics (stderr) from output (stdout)
takeaways: log/slog (Go 1.21+) provides structured, leveled logging in the standard library; Text and JSON handlers turn the same call into human or machine output; Logs go to stderr, program results to stdout

### Concept
```go norun
logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
slog.SetDefault(logger)

slog.Info("server started", "port", 8080, "tls", false)
slog.Error("request failed", "err", err, "path", r.URL.Path)
l := logger.With("request_id", id) // child logger with fixed attributes
```
Prefer **key/value pairs** over formatted strings: `"user", id` beats `fmt.Sprintf("user %d", id)` because log platforms can index and filter fields. Levels: `Debug < Info < Warn < Error`. For CLIs, print normal output to **stdout** and logs to **stderr** so pipes keep working; add `-v`/`--log-level` flags.

### Example
```go
package main

import (
	"log/slog"
	"os"
)

func main() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{} // drop the timestamp for reproducible output
			}
			return a
		},
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))
	logger.Info("server started", "port", 8080)
	logger.Debug("not shown at info level")
	logger.With("request_id", "r-1").Warn("slow request", "ms", 1200)
}
```

### Exercise
Create a text `slog` logger writing to stdout (timestamp dropped as above) at level **Debug**, and log `Debug("cache miss", "key", "user:7")`.
```text expect
level=DEBUG msg="cache miss" key=user:7
```
```go solution
package main

import (
	"log/slog"
	"os"
)

func main() {
	// BEGIN
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))
	logger.Debug("cache miss", "key", "user:7")
	// END
}
```

### Check
Q: Why prefer `slog.Info("user login", "user_id", id)` over `log.Printf("user %d logged in", id)`?
T: mcq
- [ ] It's shorter
- [x] Fields are structured, so log tools can filter and aggregate by them
- [ ] It prints faster
- [ ] Printf is removed
E: Structured attributes are machine-readable.

Q: Where should a CLI write log messages?
T: mcq
- [ ] stdout
- [x] stderr
- [ ] a file only
- [ ] nowhere
E: stdout is for program output that may be piped; diagnostics belong on stderr.

## Exit Codes
slug: exit-codes
minutes: 6
objectives: Use conventional exit codes; Return errors from run and translate them to codes once in main; Carry a specific code inside an error
takeaways: 0 success, 1 general failure, 2 usage error; 126/127 are reserved by shells (not executable/not found), 130 is SIGINT; Convert errors to exit codes in main only

### Concept
Exit codes are your CLI's API for scripts and CI.

| Code | Meaning |
| --- | --- |
| 0 | success |
| 1 | generic failure |
| 2 | usage / invalid arguments |
| 3+ | your documented codes (e.g. 3 = check failed) |
| 126, 127 | reserved by shells |
| 130 | terminated by Ctrl+C |

```go norun
type ExitError struct {
	Code int
	Err  error
}
func (e *ExitError) Error() string { return e.Err.Error() }
func (e *ExitError) Unwrap() error { return e.Err }

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		var ee *ExitError
		if errors.As(err, &ee) { os.Exit(ee.Code) }
		os.Exit(1)
	}
}
```
Deep code returns errors; only `main` decides the process exit status.

### Example
```go
package main

import (
	"errors"
	"fmt"
)

type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string { return e.Err.Error() }
func (e *ExitError) Unwrap() error { return e.Err }

func codeFor(err error) int {
	if err == nil {
		return 0
	}
	var ee *ExitError
	if errors.As(err, &ee) {
		return ee.Code
	}
	return 1
}

func main() {
	fmt.Println(codeFor(nil), codeFor(errors.New("boom")), codeFor(fmt.Errorf("wrapped: %w", &ExitError{3, errors.New("check failed")})))
}
```

### Exercise
Implement `codeFor(err error) int` returning 0 for nil, the embedded code for an `*ExitError` (found via `errors.As`), and 1 otherwise.
```text expect
0 2 1
```
```go solution
package main

import (
	"errors"
	"fmt"
)

type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string { return e.Err.Error() }
func (e *ExitError) Unwrap() error { return e.Err }

// BEGIN
func codeFor(err error) int {
	if err == nil {
		return 0
	}
	var ee *ExitError
	if errors.As(err, &ee) {
		return ee.Code
	}
	return 1
}

// END

func main() {
	fmt.Println(codeFor(nil), codeFor(&ExitError{2, errors.New("usage")}), codeFor(errors.New("boom")))
}
```

### Check
Q: Which exit status means success?
T: short
A: 0
E: Any non-zero status is a failure.

Q: Where should the decision to call `os.Exit` normally be made?
T: mcq
- [ ] In every function that detects an error
- [x] In main, after run returns an error
- [ ] In init
- [ ] In deferred functions
E: Central handling keeps cleanup working and error handling consistent.

## Interactive CLI
slug: interactive-cli
status: planned

## Shell Integration
slug: shell-integration
status: planned

## Building CLIs with Cobra
slug: building-clis-with-cobra
minutes: 9
objectives: Recognise Cobra's command tree model; Understand what Cobra gives you over the flag package; Know when to reach for it
takeaways: Cobra models a CLI as a tree of *cobra.Command values with flags, help, aliases and completions; It powers kubectl, docker CLI, hugo and gh; Use it when you have several subcommands and want consistent UX — the standard flag package is enough for small tools

### Concept
```go norun
import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "opsctl",
	Short: "Operate your clusters",
}

var checkCmd = &cobra.Command{
	Use:   "check [url]",
	Short: "Run a health check",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		timeout, _ := cmd.Flags().GetDuration("timeout")
		return check(cmd.Context(), args[0], timeout)
	},
}

func init() {
	checkCmd.Flags().Duration("timeout", 5*time.Second, "request timeout")
	rootCmd.AddCommand(checkCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil { os.Exit(1) }
}
```
You get for free: `--help` for every command, flag inheritance (`PersistentFlags`), argument validators, `completion bash|zsh|fish|powershell` scripts, typo suggestions and man-page generation. Pair it with **viper** for config files and env-var binding.

This lesson's exercise builds a tiny command-tree renderer so you understand the model Cobra implements.

### Example
```go
package main

import (
	"fmt"
	"strings"
)

type Command struct {
	Use   string
	Short string
	Subs  []*Command
}

func (c *Command) Help(indent int) {
	fmt.Printf("%s%-8s %s\n", strings.Repeat("  ", indent), c.Use, c.Short)
	for _, s := range c.Subs {
		s.Help(indent + 1)
	}
}

func main() {
	root := &Command{"opsctl", "operate clusters", []*Command{
		{"check", "run a health check", nil},
		{"deploy", "manage deployments", []*Command{{"status", "show status", nil}}},
	}}
	root.Help(0)
}
```

### Exercise
Write `find(root *Command, path ...string) *Command` that walks the tree by `Use` names (`find(root, "deploy", "status")`) and returns `nil` if any step is missing. Print the short description.
```text expect
show status
<nil>
```
```go solution
package main

import "fmt"

type Command struct {
	Use   string
	Short string
	Subs  []*Command
}

// BEGIN
func find(root *Command, path ...string) *Command {
	cur := root
	for _, name := range path {
		var next *Command
		for _, s := range cur.Subs {
			if s.Use == name {
				next = s
				break
			}
		}
		if next == nil {
			return nil
		}
		cur = next
	}
	return cur
}

// END

func main() {
	root := &Command{"opsctl", "operate clusters", []*Command{
		{"deploy", "manage deployments", []*Command{{"status", "show status", nil}}},
	}}
	fmt.Println(find(root, "deploy", "status").Short)
	fmt.Println(find(root, "deploy", "nope"))
}
```

### Check
Q: Which well-known tools are built with Cobra?
T: mcq
- [ ] Only toy projects
- [x] kubectl, the Docker CLI, gh and hugo
- [ ] git and curl
- [ ] Go itself
E: Cobra is the de-facto standard for Go command-line applications.

Q: For a tiny single-purpose tool with three flags, the standard flag package is usually enough.
T: tf
A: true
E: Reach for Cobra when you have a command tree and want consistent help, completion and flag inheritance.
