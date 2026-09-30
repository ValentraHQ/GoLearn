# Error Handling
id: errors
number: 06
track: core
paths: beginner, pro
skill: errors
requires: interfaces
project: reliable-file-processor
summary: Errors as values: create, wrap, inspect, and recover from failures the Go way.

## The Go Error Model
slug: error-model
minutes: 6
objectives: Explain that errors are ordinary values; Read the built-in error interface; Follow the if err != nil convention
takeaways: error is a built-in interface with one method, Error() string; Functions return errors instead of throwing exceptions; Handle every error explicitly — check it, or deliberately ignore it

### Concept
Go has **no exceptions**. A function that can fail returns an `error` as its last result, and the caller decides what to do.

```go norun
type error interface {
	Error() string
}

f, err := os.Open("data.txt")
if err != nil {
	return err
}
defer f.Close()
```
`nil` means success. This is verbose by design: every failure path is visible in the code, and there is no hidden control flow.

### Example
```go
package main

import (
	"fmt"
	"strconv"
)

func main() {
	n, err := strconv.Atoi("12x")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(n)
}
```

### Exercise
Parse `"42"` and `"oops"` with `strconv.Atoi`; print the number when parsing succeeds and `invalid` when it fails.
```text expect
42
invalid
```
```go solution
package main

import (
	"fmt"
	"strconv"
)

func main() {
	for _, s := range []string{"42", "oops"} {
		// BEGIN
		n, err := strconv.Atoi(s)
		if err != nil {
			fmt.Println("invalid")
			continue
		}
		fmt.Println(n)
		// END
	}
}
```

### Check
Q: What is `error` in Go?
T: mcq
- [ ] A struct with a message field
- [x] A built-in interface with an Error() string method
- [ ] A keyword for exceptions
- [ ] A string alias
E: Any type with `Error() string` is an error.

Q: A nil error means the operation succeeded.
T: tf
A: true
E: By convention `err == nil` signals success.

## Returning Errors
slug: returning-errors
minutes: 6
objectives: Return (T, error) from functions; Return early on error; Avoid returning useful values alongside non-nil errors
takeaways: Return the error as the last result; On failure return the zero value and a non-nil error; Return early to keep the success path unindented

### Concept
```go norun
func ReadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return Config{}, err
	}
	return c, nil
}
```
On error return the **zero value** for other results; callers must not use them. Return early so the happy path reads straight down the left edge.

### Example
```go
package main

import (
	"errors"
	"fmt"
)

func sqrtInt(n int) (int, error) {
	if n < 0 {
		return 0, errors.New("negative input")
	}
	r := 0
	for (r+1)*(r+1) <= n {
		r++
	}
	return r, nil
}

func main() {
	fmt.Println(sqrtInt(17))
	fmt.Println(sqrtInt(-1))
}
```

### Exercise
Write `safeDiv(a, b int) (int, error)` returning `errors.New("division by zero")` when `b == 0`.
```text expect
5 <nil>
0 division by zero
```
```go solution
package main

import (
	"errors"
	"fmt"
)

// BEGIN
func safeDiv(a, b int) (int, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	return a / b, nil
}

// END

func main() {
	fmt.Println(safeDiv(10, 2))
	fmt.Println(safeDiv(1, 0))
}
```

### Check
Q: What should a function return for its non-error result when it fails?
T: mcq
- [ ] A partially valid value
- [x] The zero value
- [ ] nil always
- [ ] It should panic
E: Callers should ignore other results when `err != nil`; returning the zero value is the convention.

Q: What does this print?
T: output
```go
func f() error { return nil }

func main() {
	fmt.Println(f() == nil)
}
```
- [x] true
- [ ] false
- [ ] nil
- [ ] error
E: A function returning a nil error yields an interface value equal to `nil`.

## Creating Errors
slug: creating-errors
minutes: 6
objectives: Create errors with errors.New and fmt.Errorf; Define sentinel errors; Compare errors safely
takeaways: errors.New makes a fixed error; fmt.Errorf formats context into one; Sentinel errors are package-level vars named ErrXxx

### Concept
```go norun
err := errors.New("not found")
err = fmt.Errorf("user %d not found", id)

// A sentinel error other packages can compare against:
var ErrNotFound = errors.New("not found")
```
Sentinel errors are exported variables named `ErrSomething` — like `io.EOF` and `os.ErrNotExist`. Each call to `errors.New` creates a **distinct** value, so two errors with the same text are not equal.

### Example
```go
package main

import (
	"errors"
	"fmt"
)

var ErrEmpty = errors.New("empty name")

func validate(name string) error {
	if name == "" {
		return ErrEmpty
	}
	return nil
}

func main() {
	err := validate("")
	fmt.Println(err, err == ErrEmpty)
	fmt.Println(errors.New("x") == errors.New("x"))
}
```

### Exercise
Declare `ErrTooShort = errors.New("too short")` and make `check("go")` return it (minimum length 3). Print `true`.
```text expect
true
```
```go solution
package main

import (
	"errors"
	"fmt"
)

// BEGIN
var ErrTooShort = errors.New("too short")

// END

func check(s string) error {
	if len(s) < 3 {
		return ErrTooShort
	}
	return nil
}

func main() {
	fmt.Println(check("go") == ErrTooShort)
}
```

### Check
Q: What does this print?
T: output
```go
a := errors.New("boom")
b := errors.New("boom")
fmt.Println(a == b)
```
- [ ] true
- [x] false
- [ ] boom
- [ ] compile error
E: Every `errors.New` call returns a distinct error value, even with identical text.

Q: By convention, sentinel errors are named...
T: mcq
- [ ] errSomething only
- [x] ErrSomething
- [ ] SomethingError
- [ ] E_SOMETHING
E: `ErrNotFound`, `ErrClosed`… exported vars with an `Err` prefix.

## Error Wrapping
slug: error-wrapping
minutes: 7
objectives: Add context with %w; Preserve the original cause; Write useful error messages
takeaways: fmt.Errorf("context: %w", err) wraps an error; Wrapped errors keep the chain for errors.Is/As; Messages read outer-to-inner: "load config: open x: no such file"

### Concept
Bare errors like `no such file` are hard to debug. **Wrap** with context as the error travels up:

```go norun
if err != nil {
	return fmt.Errorf("load config %q: %w", path, err)
}
```
`%w` keeps the original error inside so callers can still test for it with `errors.Is` and `errors.As`. Use `%v` instead of `%w` when you deliberately **don't** want callers depending on the cause. Write messages in lower case without trailing punctuation; the chain reads like a sentence.

### Example
```go
package main

import (
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("not found")

func find(id int) error { return ErrNotFound }

func load(id int) error {
	if err := find(id); err != nil {
		return fmt.Errorf("load user %d: %w", id, err)
	}
	return nil
}

func main() {
	err := load(7)
	fmt.Println(err)
	fmt.Println(errors.Unwrap(err) == ErrNotFound)
}
```

### Exercise
Wrap `base` with the context `open db` using `%w` and print the resulting message.
```text expect
open db: connection refused
```
```go solution
package main

import (
	"errors"
	"fmt"
)

func main() {
	base := errors.New("connection refused")
	// BEGIN
	err := fmt.Errorf("open db: %w", base)
	// END
	fmt.Println(err)
}
```

### Check
Q: Which verb wraps an error so it can be inspected later?
T: short
A: %w
E: `%w` in `fmt.Errorf` stores the wrapped error for `errors.Is/As/Unwrap`.

Q: What does this print?
T: output
```go
inner := errors.New("disk full")
err := fmt.Errorf("save: %w", inner)
fmt.Println(err)
```
- [ ] disk full
- [x] save: disk full
- [ ] save
- [ ] disk full: save
E: The formatted message is the context followed by the wrapped error text.

## errors.Is
slug: errors-is
minutes: 6
objectives: Test whether an error chain contains a target; Prefer errors.Is over ==; Support Is for custom comparison
takeaways: errors.Is walks the wrap chain looking for a match; Never compare wrapped errors with ==; Use it with sentinel errors such as io.EOF and os.ErrNotExist

### Concept
Once errors are wrapped, `err == ErrNotFound` stops working. `errors.Is(err, ErrNotFound)` walks the chain (following `Unwrap`) and reports whether any error in it equals the target.

```go norun
_, err := os.Open("missing.txt")
if errors.Is(err, os.ErrNotExist) {
	fmt.Println("create it first")
}
```
Custom types can implement `Is(target error) bool` for special matching, but sentinel values plus `errors.Is` cover most cases.

### Example
```go
package main

import (
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("not found")

func main() {
	err := fmt.Errorf("layer 2: %w", fmt.Errorf("layer 1: %w", ErrNotFound))
	fmt.Println(err == ErrNotFound, errors.Is(err, ErrNotFound))
}
```

### Exercise
Write `isMissing(err error) bool` using `errors.Is(err, fs.ErrNotExist)`. Opening a missing file should print `true`.
```text expect
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
func isMissing(err error) bool {
	return errors.Is(err, fs.ErrNotExist)
}

// END

func main() {
	_, err := os.Open("/definitely/not/here.txt")
	fmt.Println(isMissing(err))
}
```

### Check
Q: Why use `errors.Is` instead of `==`?
T: mcq
- [ ] It is faster
- [x] It sees through wrapped errors
- [ ] It compares error text
- [ ] It works only on custom errors
E: `errors.Is` follows the `Unwrap` chain, so wrapping doesn't break checks.

Q: What does this print?
T: output
```go
base := errors.New("x")
w := fmt.Errorf("ctx: %w", base)
fmt.Println(errors.Is(w, base), w == base)
```
- [ ] false false
- [ ] true true
- [x] true false
- [ ] false true
E: `errors.Is` finds `base` inside `w`, but `w` itself is a different error value.

## errors.As
slug: errors-as
minutes: 6
objectives: Extract a specific error type from a chain; Use a pointer to the target variable; Choose between Is and As
takeaways: errors.As finds the first error in the chain assignable to the target type; The second argument must be a non-nil pointer; Use Is for values, As for types

### Concept
When a caller needs the **fields** of a custom error, use `errors.As`:

```go norun
var pe *fs.PathError
if errors.As(err, &pe) {
	fmt.Println("path:", pe.Path, "op:", pe.Op)
}
```
`As` walks the chain, and on the first match assigns the value to `pe` and returns `true`. Use `errors.Is` to ask "is it *this* error?" and `errors.As` to ask "is it *this kind* of error, and give it to me".

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
	_, err := os.Open("/no/such/file")
	err = fmt.Errorf("startup: %w", err)
	var pe *fs.PathError
	if errors.As(err, &pe) {
		fmt.Println(pe.Op, pe.Path)
	}
}
```

### Exercise
Write `opOf(err error) string` returning the `Op` of a wrapped `*fs.PathError`, or `""` if there is none.
```text expect
open
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
func opOf(err error) string {
	var pe *fs.PathError
	if errors.As(err, &pe) {
		return pe.Op
	}
	return ""
}

// END

func main() {
	_, err := os.Open("/no/such/file")
	fmt.Println(opOf(fmt.Errorf("wrap: %w", err)))
}
```

### Check
Q: What must the second argument of `errors.As` be?
T: mcq
- [ ] A string
- [ ] The error itself
- [x] A non-nil pointer to a variable of the target type
- [ ] A function
E: `As` assigns into the pointed-to variable when it finds a match.

Q: You want to know whether an error chain contains the sentinel `ErrNotFound`. Which do you use?
T: mcq
- [x] errors.Is
- [ ] errors.As
- [ ] type switch on the outer error
- [ ] strings.Contains(err.Error(), ...)
E: Sentinel values are matched with `errors.Is`; `errors.As` is for extracting typed errors.

## Custom Errors
slug: custom-errors
minutes: 7
objectives: Implement the error interface on your own type; Carry structured data in errors; Add Unwrap to keep the chain
takeaways: A custom error is any type with Error() string; Use struct fields to carry machine-readable details; Implement Unwrap() error to expose an underlying cause

### Concept
```go norun
type ValidationError struct {
	Field string
	Msg   string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Msg
}
```
Return `*ValidationError` and callers can use `errors.As` to read `Field`. If your error wraps another, implement `Unwrap() error` so `errors.Is/As` continue through it.

**Gotcha:** never return a nil `*ValidationError` as an `error` — a non-nil interface holding a nil pointer is not `== nil`.

### Example
```go
package main

import (
	"errors"
	"fmt"
)

type ValidationError struct{ Field, Msg string }

func (e *ValidationError) Error() string { return e.Field + ": " + e.Msg }

func validate(age int) error {
	if age < 0 {
		return &ValidationError{"age", "must not be negative"}
	}
	return nil
}

func main() {
	err := validate(-1)
	var ve *ValidationError
	if errors.As(err, &ve) {
		fmt.Println(ve.Field, "|", err)
	}
}
```

### Exercise
Create `type NotFoundError struct{ ID int }` with an `Error()` method returning `item 7 not found` (use `fmt.Sprintf`).
```text expect
item 7 not found
```
```go solution
package main

import "fmt"

// BEGIN
type NotFoundError struct{ ID int }

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("item %d not found", e.ID)
}

// END

func main() {
	var err error = &NotFoundError{ID: 7}
	fmt.Println(err)
}
```

### Check
Q: What does a custom error type need in order to be used as an `error`?
T: short
A: Error() string
E: Implementing `Error() string` is all the `error` interface requires.

Q: What does this print?
T: output
```go
type E struct{}
func (*E) Error() string { return "e" }
func f() error {
	var p *E
	return p
}
func main() {
	fmt.Println(f() == nil)
}
```
- [ ] true
- [x] false
- [ ] e
- [ ] panic
E: The interface holds a typed nil pointer, so the interface value itself is not nil. Return a literal `nil` for success.

## Panic
slug: panic
minutes: 6
objectives: Explain what panic does; Distinguish panics from errors; Know the few cases where panic is appropriate
takeaways: panic stops normal execution and unwinds the stack running deferred calls; Use errors for expected failures and panic for programmer bugs; Libraries should not panic across their API boundary

### Concept
`panic(v)` aborts the current function, runs its `defer`red calls, and unwinds up the stack. If nothing recovers it, the program crashes with a stack trace.

Runtime panics are triggered by bugs: nil dereference, index out of range, closed-channel send, failed type assertion.

Use `panic` for **impossible states** and startup misconfiguration (`regexp.MustCompile`). Anything the caller can reasonably handle — bad input, missing files, network failure — is an **error**.

### Example
This program panics on purpose, so it is shown for reading — run it locally to see the stack trace.
```go norun
package main

import "fmt"

func main() {
	defer fmt.Println("deferred runs during a panic")
	var m map[string]int
	m["a"] = 1 // panics: assignment to entry in nil map
	fmt.Println("never reached")
}
```

### Exercise
The function `mustPositive` should `panic("not positive")` when `n <= 0`. We recover in `main` to show the message; just implement the panic.
```text expect
recovered: not positive
```
```go solution
package main

import "fmt"

func mustPositive(n int) int {
	// BEGIN
	if n <= 0 {
		panic("not positive")
	}
	// END
	return n
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recovered:", r)
		}
	}()
	mustPositive(-1)
}
```

### Check
Q: Which should be reported with an error rather than a panic?
T: mcq
- [ ] A nil pointer dereference in your own code
- [ ] An impossible internal state
- [x] A user supplying an invalid file name
- [ ] A failed regexp.MustCompile of a constant
E: Expected, recoverable failures are errors; panics are for bugs and impossible states.

Q: Deferred functions still run when a function panics.
T: tf
A: true
E: Panicking unwinds the stack while executing deferred calls, which is what makes `recover` and cleanup work.

## Recover
slug: recover
minutes: 7
objectives: Recover from a panic inside a deferred function; Convert a panic into an error at a boundary; Know what recover cannot do
takeaways: recover only works inside a deferred function; Use it at boundaries (HTTP handlers, worker goroutines) to keep one failure from crashing everything; Recover returns nil when nothing panicked

### Concept
```go norun
func safely(f func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered: %v", r)
		}
	}()
	f()
	return nil
}
```
`recover()` stops the panic and returns its value — but **only when called directly from a deferred function**. Use it at boundaries: `net/http` recovers panics per request; a worker pool should recover per job. A panic in one goroutine cannot be recovered by another.

### Example
```go
package main

import "fmt"

func safely(f func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered: %v", r)
		}
	}()
	f()
	return nil
}

func main() {
	fmt.Println(safely(func() { panic("boom") }))
	fmt.Println(safely(func() {}))
}
```

### Exercise
Write `try(f func()) error` that runs `f`, turns a panic into an error `panic: <value>`, and returns nil otherwise.
```text expect
panic: index out of range
<nil>
```
```go solution
package main

import "fmt"

// BEGIN
func try(f func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	f()
	return nil
}

// END

func main() {
	fmt.Println(try(func() { panic("index out of range") }))
	fmt.Println(try(func() {}))
}
```

### Check
Q: Where does `recover()` have an effect?
T: mcq
- [ ] Anywhere in the function
- [x] Directly inside a deferred function
- [ ] Only in main
- [ ] Only in goroutines
E: `recover` returns nil unless called directly by a deferred function during a panic.

Q: A panic in one goroutine can be recovered by a `defer` in a different goroutine.
T: tf
A: false
E: Each goroutine must recover its own panics, otherwise the whole program crashes.

## Error Handling Best Practices
slug: error-best-practices
minutes: 7
objectives: Handle each error once; Add context but don't log-and-return; Design errors callers can act on
takeaways: Handle an error once — either return it or log it, not both; Wrap with context at meaningful boundaries; Expose sentinel values or types only when callers need to branch on them

### Concept
1. **Handle once.** Either return the error (with context) or handle it (log, retry, fall back). Logging *and* returning produces duplicate noise up the stack.
2. **Add context, not noise.** `fmt.Errorf("fetch user %d: %w", id, err)`; don't repeat "failed to" at every layer.
3. **Decide what's API.** Every sentinel/type you export is a promise. Keep internal causes behind `%v` if callers shouldn't depend on them.
4. **Don't ignore errors.** If you must (`defer f.Close()` on a read-only file), make it deliberate; `errcheck`/`golangci-lint` catch the rest.
5. **Panics are for bugs.** Recover at boundaries only.

### Example
```go
package main

import (
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("not found")

type store struct{}

func (store) Get(id int) (string, error) { return "", ErrNotFound }

func profile(s store, id int) (string, error) {
	name, err := s.Get(id)
	if err != nil {
		return "", fmt.Errorf("profile %d: %w", id, err) // add context, return once
	}
	return name, nil
}

func main() {
	_, err := profile(store{}, 3)
	fmt.Println(err, errors.Is(err, ErrNotFound))
}
```

### Exercise
Write `loadAll(ids []int) error` that calls `load(id)` for each id and returns the first error wrapped as `load <id>: <cause>` (stop at the first failure).
```text expect
load 3: not found
```
```go solution
package main

import (
	"errors"
	"fmt"
)

func load(id int) error {
	if id == 3 {
		return errors.New("not found")
	}
	return nil
}

// BEGIN
func loadAll(ids []int) error {
	for _, id := range ids {
		if err := load(id); err != nil {
			return fmt.Errorf("load %d: %w", id, err)
		}
	}
	return nil
}

// END

func main() {
	fmt.Println(loadAll([]int{1, 2, 3, 4}))
}
```

### Check
Q: Which is the better way to handle an error you can't recover from locally?
T: mcq
- [ ] Log it and return it
- [x] Wrap it with context and return it
- [ ] Ignore it
- [ ] Panic
E: Handle once: return it with context and let a higher layer decide (log at the top).

Q: Ignoring errors from `Close` on a file you only read is always a bug.
T: tf
A: false
E: It is often acceptable for read-only files, but do check `Close` errors on files you write to — buffered data may only fail at close.
