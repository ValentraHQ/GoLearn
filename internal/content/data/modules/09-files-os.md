# Files & Operating System
id: files-os
number: 09
track: core
paths: beginner, pro
skill: stdlib
requires: stdlib
project: system-information-cli
summary: Read and write files, walk directories, use environment variables, arguments and signals.

## Reading Files
slug: reading-files
minutes: 7
objectives: Read a whole file with os.ReadFile; Stream a file line by line; Always close what you open
takeaways: os.ReadFile is the simplest way to load a small file; For big files open, defer Close and stream with bufio; Check errors with errors.Is(err, fs.ErrNotExist)

### Concept
```go norun
data, err := os.ReadFile("config.json") // whole file into memory
```
For large files stream instead:

```go norun
f, err := os.Open(path)
if err != nil {
	return err
}
defer f.Close()
sc := bufio.NewScanner(f)
for sc.Scan() { /* sc.Text() */ }
return sc.Err()
```
Distinguish "not found" from other failures with `errors.Is(err, fs.ErrNotExist)`. The sandbox in these lessons has a writable `/tmp`, so the exercises create their own files.

### Example
```go
package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	f, _ := os.CreateTemp("", "demo-*.txt")
	defer os.Remove(f.Name())
	f.WriteString("one\ntwo\nthree\n")
	f.Close()

	in, err := os.Open(f.Name())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer in.Close()
	sc := bufio.NewScanner(in)
	n := 0
	for sc.Scan() {
		n++
	}
	fmt.Println("lines:", n)
}
```

### Exercise
Write `countLines(path string) (int, error)` streaming the file with `bufio.Scanner`.
```text expect
3 <nil>
0 file does not exist
```
```go solution
package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
)

// BEGIN
func countLines(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	n := 0
	for sc.Scan() {
		n++
	}
	return n, sc.Err()
}

// END

func main() {
	f, _ := os.CreateTemp("", "lines-*.txt")
	defer os.Remove(f.Name())
	f.WriteString("a\nb\nc\n")
	f.Close()
	fmt.Println(countLines(f.Name()))
	_, err := countLines("/no/such/file")
	if errors.Is(err, fs.ErrNotExist) {
		fmt.Println(0, "file does not exist")
	}
}
```

### Check
Q: Why `defer f.Close()` right after a successful `os.Open`?
T: mcq
- [ ] It is optional and only for style
- [x] It guarantees the file descriptor is released on every return path
- [ ] It flushes the OS cache
- [ ] It speeds up reads
E: Descriptors are a limited resource; `defer` ensures cleanup even if the function returns early or panics.

Q: Which is best for a 5 GB log file?
T: mcq
- [ ] os.ReadFile
- [x] Open and stream with bufio.Scanner
- [ ] io.ReadAll
- [ ] strings.Split on the whole file
E: Streaming keeps memory use constant.

## Writing Files
slug: writing-files
minutes: 7
objectives: Write a file atomically-ish with os.WriteFile; Append with OpenFile flags; Check Close errors on written files
takeaways: os.WriteFile creates or truncates a file in one call; os.OpenFile with O_APPEND adds to the end; Errors from Close on a written file can be the only sign of failure

### Concept
```go norun
err := os.WriteFile("out.txt", []byte("hello\n"), 0o644)

f, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
```
`0o644` is the permission (owner rw, group/others r). For important data write to a temp file then `os.Rename` it over the target — rename is atomic on the same filesystem, so readers never see half a file.

For buffered files `Close` (or `Flush`) may be where a disk-full error surfaces — check it.

### Example
```go
package main

import (
	"fmt"
	"os"
)

func main() {
	path := os.TempDir() + "/greeting.txt"
	defer os.Remove(path)
	if err := os.WriteFile(path, []byte("hi\n"), 0o644); err != nil {
		fmt.Println(err)
		return
	}
	f, _ := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	f.WriteString("there\n")
	f.Close()
	b, _ := os.ReadFile(path)
	fmt.Print(string(b))
}
```

### Exercise
Write `appendLine(path, line string) error` that appends `line` plus a newline, creating the file if needed.
```text expect
first
second
```
```go solution
package main

import (
	"fmt"
	"os"
)

// BEGIN
func appendLine(path, line string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.WriteString(line + "\n"); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// END

func main() {
	path := os.TempDir() + "/append-demo.txt"
	os.Remove(path)
	defer os.Remove(path)
	appendLine(path, "first")
	appendLine(path, "second")
	b, _ := os.ReadFile(path)
	fmt.Print(string(b))
}
```

### Check
Q: Which flag combination appends to a file, creating it if missing?
T: mcq
- [ ] O_RDONLY
- [ ] O_TRUNC|O_WRONLY
- [x] O_APPEND|O_CREATE|O_WRONLY
- [ ] O_EXCL
E: `O_APPEND` positions writes at the end and `O_CREATE` creates the file when absent.

Q: Why write to a temp file and rename it into place?
T: mcq
- [ ] It is faster
- [x] Rename is atomic, so readers never see a half-written file
- [ ] It avoids permissions
- [ ] It compresses the file
E: Atomic replacement protects readers (and crash recovery) from partial writes.

## Creating Files
slug: creating-files
minutes: 5
objectives: Create files with os.Create and os.CreateTemp; Avoid clobbering with O_EXCL; Clean up temporary files
takeaways: os.Create truncates an existing file; os.CreateTemp gives a unique file name; O_EXCL makes creation fail if the file already exists

### Concept
`os.Create(path)` opens for read/write, **truncating** anything already there — a surprising way to lose data. To create only if missing:

```go norun
f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
if errors.Is(err, fs.ErrExist) { /* already there */ }
```
`os.CreateTemp("", "prefix-*.txt")` returns a uniquely named file; `defer os.Remove(f.Name())`. For a scratch directory use `os.MkdirTemp` (or `t.TempDir()` in tests).

### Example
```go
package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
)

func main() {
	f, _ := os.CreateTemp("", "x-*.txt")
	name := f.Name()
	f.Close()
	defer os.Remove(name)
	_, err := os.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	fmt.Println(errors.Is(err, fs.ErrExist))
}
```

### Exercise
Write `createNew(path string) error` that creates an empty file but fails (with `fs.ErrExist`) if it already exists.
```text expect
<nil>
true
```
```go solution
package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
)

// BEGIN
func createNew(path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	return f.Close()
}

// END

func main() {
	path := os.TempDir() + "/create-new-demo.txt"
	os.Remove(path)
	defer os.Remove(path)
	fmt.Println(createNew(path))
	fmt.Println(errors.Is(createNew(path), fs.ErrExist))
}
```

### Check
Q: What does `os.Create` do if the file already exists?
T: mcq
- [ ] Returns an error
- [ ] Appends to it
- [x] Truncates it to zero length
- [ ] Opens it read-only
E: `os.Create` uses O_TRUNC, discarding the previous contents.

Q: Which function creates a uniquely named temporary file?
T: short
A: os.CreateTemp
E: `os.CreateTemp(dir, pattern)` — pass "" for the default temp directory.

## Directories
slug: directories
minutes: 7
objectives: List, create and remove directories; Walk a tree with filepath.WalkDir; Read directory entries lazily
takeaways: os.ReadDir returns sorted entries; os.MkdirAll creates nested directories; filepath.WalkDir visits every file under a root

### Concept
```go norun
entries, err := os.ReadDir(".")         // sorted by name
os.MkdirAll("a/b/c", 0o755)             // like mkdir -p
os.RemoveAll("a")                       // like rm -rf — careful!
filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error { ... })
```
`d.IsDir()`, `d.Name()`, `d.Info()` (for size/mode). Return `filepath.SkipDir` from the callback to skip a subtree. `io/fs` and `testing/fstest.MapFS` let you test directory code without touching the disk.

### Example
```go
package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func main() {
	root, _ := os.MkdirTemp("", "tree-")
	defer os.RemoveAll(root)
	os.MkdirAll(filepath.Join(root, "a", "b"), 0o755)
	os.WriteFile(filepath.Join(root, "a", "x.txt"), []byte("12345"), 0o644)
	filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		rel, _ := filepath.Rel(root, p)
		fmt.Println(rel, d.IsDir())
		return nil
	})
}
```

### Exercise
Write `totalSize(fsys fs.FS) (int64, error)` summing the sizes of all regular files using `fs.WalkDir`.
```text expect
12 <nil>
```
```go solution
package main

import (
	"fmt"
	"io/fs"
	"testing/fstest"
)

// BEGIN
func totalSize(fsys fs.FS) (int64, error) {
	var total int64
	err := fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		total += info.Size()
		return nil
	})
	return total, err
}

// END

func main() {
	fsys := fstest.MapFS{
		"a.txt":     {Data: []byte("12345")},
		"dir/b.txt": {Data: []byte("1234567")},
	}
	fmt.Println(totalSize(fsys))
}
```

### Check
Q: Which call is the equivalent of `mkdir -p`?
T: mcq
- [ ] os.Mkdir
- [x] os.MkdirAll
- [ ] os.Create
- [ ] os.MkdirTemp
E: `MkdirAll` creates all missing parent directories.

Q: What does returning `filepath.SkipDir` from a WalkDir callback do?
T: mcq
- [ ] Stops the entire walk
- [x] Skips the rest of that directory's contents
- [ ] Deletes the directory
- [ ] Skips only the current file
E: `SkipDir` prunes the current directory subtree; `SkipAll` stops everything.

## File Permissions
slug: file-permissions
minutes: 6
objectives: Read Unix permission bits; Choose safe modes for files and directories; Change modes with os.Chmod
takeaways: Modes are octal: 0o644 files, 0o755 directories and executables, 0o600 secrets; The umask removes bits from the mode you request; Never make secrets world-readable

### Concept
Unix permissions are three triplets (owner, group, others) of read=4, write=2, execute=1.

| Mode | Meaning | Use |
| --- | --- | --- |
| `0o644` | rw-r--r-- | regular files |
| `0o755` | rwxr-xr-x | directories, executables |
| `0o600` | rw------- | keys, tokens, credentials |

`info.Mode().Perm()` reads them; `os.Chmod(path, 0o600)` changes them. The process **umask** (commonly `022`) clears bits from the mode you pass at creation, so use `Chmod` when you need an exact result.

### Example
```go
package main

import (
	"fmt"
	"os"
)

func main() {
	f, _ := os.CreateTemp("", "perm-")
	defer os.Remove(f.Name())
	f.Close()
	os.Chmod(f.Name(), 0o640)
	info, _ := os.Stat(f.Name())
	fmt.Println(info.Mode().Perm())
}
```

### Exercise
Write `isWorldReadable(mode fs.FileMode) bool` reporting whether the "others read" bit (`0o004`) is set.
```text expect
true false
```
```go solution
package main

import (
	"fmt"
	"io/fs"
)

// BEGIN
func isWorldReadable(mode fs.FileMode) bool {
	return mode.Perm()&0o004 != 0
}

// END

func main() {
	fmt.Println(isWorldReadable(0o644), isWorldReadable(0o600))
}
```

### Check
Q: What permission should a file holding an API key typically have?
T: mcq
- [ ] 0o777
- [ ] 0o644
- [x] 0o600
- [ ] 0o755
E: Only the owner should be able to read or write secrets.

Q: What does `0o644` mean for a file?
T: mcq
- [ ] Owner read-only
- [x] Owner read/write, group and others read-only
- [ ] Everyone read/write
- [ ] Owner execute
E: 6 = rw for the owner, 4 = r for the group, 4 = r for others.

## Paths
slug: paths
minutes: 5
objectives: Build paths with filepath.Join; Resolve relative paths and check containment; Split into dir, base and extension
takeaways: Always build paths with filepath.Join, never string concatenation; filepath.Abs and filepath.Rel convert between absolute and relative; Validate user-supplied paths against a base directory

### Concept
`filepath.Join("data", name)` uses the right separator and cleans `..`. `filepath.Dir`, `Base`, `Ext` split a path; `filepath.Abs` resolves against the working directory; `filepath.Rel(base, target)` computes the relative path (and reveals `..` escapes).

**Security rule:** if a user controls part of a path, cleaning alone isn't enough — verify the final path stays inside the intended base directory (see the "Safe Path Join" challenge).

### Example
```go
package main

import (
	"fmt"
	"path/filepath"
)

func main() {
	p := filepath.Join("/srv", "app", "..", "data", "file.tar.gz")
	fmt.Println(p)
	fmt.Println(filepath.Dir(p), filepath.Base(p), filepath.Ext(p))
	rel, _ := filepath.Rel("/srv/data", "/srv/data/logs/a.log")
	fmt.Println(rel)
}
```

### Exercise
Write `within(base, target string) bool` using `filepath.Rel`: true when `target` is inside `base` (or equal), false when it escapes.
```text expect
true false
```
```go solution
package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

// BEGIN
func within(base, target string) bool {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// END

func main() {
	fmt.Println(within("/srv/data", "/srv/data/a/b"), within("/srv/data", "/srv/other"))
}
```

### Check
Q: Why prefer `filepath.Join` over `dir + "/" + name`?
T: mcq
- [ ] It is shorter
- [x] It uses the correct separator and cleans the result
- [ ] It creates the directory
- [ ] It checks the file exists
E: `Join` handles OS differences and redundant separators or `..` elements.

Q: `filepath.Ext("archive.tar.gz")` returns...
T: short
A: .gz
E: `Ext` returns only the final extension.

## Environment Variables
slug: environment-variables
minutes: 6
objectives: Read and set environment variables; Load configuration with defaults; Keep secrets out of source code
takeaways: os.Getenv returns "" for unset; use LookupEnv to distinguish; Twelve-factor apps read configuration from the environment; Never commit secrets — inject them at deploy time

### Concept
```go norun
port := os.Getenv("PORT")
if v, ok := os.LookupEnv("DEBUG"); ok { /* set, maybe empty */ }
os.Setenv("MODE", "test")     // this process (and children) only
os.Environ()                  // []string{"KEY=value", ...}
```
Environment variables are the standard way to configure containers and services. Parse them into a typed `Config` at startup, fail fast on invalid values, and never log secrets. `os.ExpandEnv("$HOME/x")` substitutes variables in a string.

### Example
```go
package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	os.Setenv("APP_NAME", "golearn")
	fmt.Println(os.ExpandEnv("service=$APP_NAME"))
	for _, kv := range os.Environ() {
		if strings.HasPrefix(kv, "APP_") {
			fmt.Println(kv)
		}
	}
}
```

### Exercise
Write `mask(env []string) []string` that keeps `KEY=value` entries but replaces the value with `***` when the key contains `TOKEN`, `SECRET` or `KEY` (case-sensitive, key part only).
```text expect
[USER=ada API_TOKEN=*** PATH=/bin DB_SECRET=***]
```
```go solution
package main

import (
	"fmt"
	"strings"
)

// BEGIN
func mask(env []string) []string {
	out := make([]string, 0, len(env))
	for _, kv := range env {
		k, _, ok := strings.Cut(kv, "=")
		if ok && (strings.Contains(k, "TOKEN") || strings.Contains(k, "SECRET") || strings.Contains(k, "KEY")) {
			kv = k + "=***"
		}
		out = append(out, kv)
	}
	return out
}

// END

func main() {
	fmt.Println(mask([]string{"USER=ada", "API_TOKEN=abc", "PATH=/bin", "DB_SECRET=hunter2"}))
}
```

### Check
Q: How do you distinguish an unset variable from one set to the empty string?
T: short
A: os.LookupEnv
E: `LookupEnv` returns a second `ok` value.

Q: Environment variables set with `os.Setenv` are visible to other processes on the machine.
T: tf
A: false
E: They live in the current process (and are inherited by children it starts), not system-wide.

## Command-Line Arguments
slug: command-line-arguments
minutes: 7
objectives: Read raw arguments from os.Args; Parse flags with the flag package; Print usage and exit codes properly
takeaways: os.Args[0] is the program, the rest are arguments; The flag package parses -name value and --name=value; Usage errors exit with status 2

### Concept
```go norun
name := flag.String("name", "world", "who to greet")
verbose := flag.Bool("v", false, "verbose output")
flag.Parse()
args := flag.Args() // remaining positional arguments
```
`flag.Parse()` reads `os.Args[1:]`. For testable code create a `flag.NewFlagSet("cmd", flag.ContinueOnError)` and call `fs.Parse(args)` with a slice. Print errors to stderr and exit with **2** for usage mistakes and **1** for runtime failures. For subcommands and richer UX see Module 16.

### Example
```go
package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	fs := flag.NewFlagSet("greet", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	name := fs.String("name", "world", "who to greet")
	loud := fs.Bool("loud", false, "shout")
	if err := fs.Parse([]string{"-name", "Ada", "-loud", "extra"}); err != nil {
		return
	}
	fmt.Println(*name, *loud, fs.Args())
}
```

### Exercise
Using a `FlagSet`, parse `-n 3` with `n := fs.Int("n", 1, "count")` and print `hello` that many times, one per line.
```text expect
hello
hello
hello
```
```go solution
package main

import (
	"flag"
	"fmt"
)

func main() {
	fs := flag.NewFlagSet("repeat", flag.ContinueOnError)
	// BEGIN
	n := fs.Int("n", 1, "count")
	_ = fs.Parse([]string{"-n", "3"})
	for i := 0; i < *n; i++ {
		fmt.Println("hello")
	}
	// END
}
```

### Check
Q: What is `os.Args[0]`?
T: mcq
- [ ] The first user argument
- [x] The program name or path
- [ ] The number of arguments
- [ ] The working directory
E: User-supplied arguments start at `os.Args[1]`.

Q: Which exit status conventionally signals a usage error?
T: short
A: 2
E: Go's `flag` package itself exits with 2 on parse errors.

## OS Signals
slug: os-signals
minutes: 8
objectives: Receive signals with signal.Notify and NotifyContext; Shut down gracefully on SIGINT/SIGTERM; Know which signals cannot be caught
takeaways: signal.NotifyContext turns SIGINT/SIGTERM into context cancellation; Containers stop with SIGTERM, then SIGKILL after a grace period; SIGKILL and SIGSTOP cannot be caught

### Concept
```go norun
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()
<-ctx.Done() // Ctrl+C or `kill <pid>`
// clean up, flush, close connections…
```
Kubernetes and Docker send **SIGTERM**, wait the grace period (30 s by default), then **SIGKILL**. A well-behaved service catches SIGTERM, stops accepting work, finishes in-flight requests and exits. `signal.Notify(ch, sig...)` needs a **buffered** channel (size ≥ 1).

### Example
```go
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGUSR1)
	p, _ := os.FindProcess(os.Getpid())
	p.Signal(syscall.SIGUSR1) // send ourselves a signal
	fmt.Println("got:", <-ch)
}
```

### Exercise
Register for `SIGUSR1`, send it to yourself, and print `received` once it arrives.
```text expect
received
```
```go solution
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// BEGIN
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGUSR1)
	p, _ := os.FindProcess(os.Getpid())
	_ = p.Signal(syscall.SIGUSR1)
	<-ch
	fmt.Println("received")
	// END
}
```

### Check
Q: Which signal do Docker and Kubernetes send first to ask a container to stop?
T: short
A: SIGTERM
E: SIGTERM is the polite request; SIGKILL follows after the grace period.

Q: Which of these can a Go program NOT catch?
T: mcq
- [ ] SIGINT
- [ ] SIGTERM
- [x] SIGKILL
- [ ] SIGHUP
E: SIGKILL and SIGSTOP are handled by the kernel and can't be intercepted.

## Process Management
slug: process-management
minutes: 8
objectives: Run external commands with os/exec; Capture output and exit codes; Use context timeouts for child processes
takeaways: exec.Command builds a command; Output/CombinedOutput/Run execute it; Use exec.CommandContext so a timeout kills a runaway child; never build commands with shell string concatenation from user input

### Concept
```go norun
out, err := exec.Command("git", "rev-parse", "HEAD").Output()
var ee *exec.ExitError
if errors.As(err, &ee) { fmt.Println("exit code", ee.ExitCode()) }
```
- `Run()` — wait; `Output()` — capture stdout; `CombinedOutput()` — stdout+stderr.
- `exec.CommandContext(ctx, ...)` kills the process when `ctx` ends.
- Pass arguments as separate strings — **never** `exec.Command("sh", "-c", "ls "+userInput)`; that is command injection.

### Example
```go
package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func main() {
	out, err := exec.Command("echo", "hello", "world").Output()
	fmt.Println(strings.TrimSpace(string(out)), err)
}
```

### Exercise
Write `run(name string, args ...string) (string, int)` returning trimmed stdout and the exit code (0 on success; the process exit code on `*exec.ExitError`).
```text expect
hi 0
 1
```
```go solution
package main

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// BEGIN
func run(name string, args ...string) (string, int) {
	out, err := exec.Command(name, args...).Output()
	code := 0
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		code = ee.ExitCode()
	} else if err != nil {
		code = -1
	}
	return strings.TrimSpace(string(out)), code
}

// END

func main() {
	out, code := run("echo", "hi")
	fmt.Println(out, code)
	out, code = run("false")
	fmt.Println(out, code)
}
```

### Check
Q: Why is `exec.Command("sh", "-c", "ls "+userInput)` dangerous?
T: mcq
- [ ] It is slow
- [x] Shell metacharacters in userInput allow command injection
- [ ] sh isn't installed
- [ ] It ignores exit codes
E: Pass arguments separately so no shell interprets them.

Q: Which constructor kills the child if a deadline passes?
T: mcq
- [ ] exec.Command
- [x] exec.CommandContext
- [ ] exec.LookPath
- [ ] os.StartProcess only
E: `CommandContext` ties the child's lifetime to a `context.Context`.
