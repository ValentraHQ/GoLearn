# Go Glossary

## Goroutine
related: concurrency/goroutines, concurrency/goroutine-lifecycle

### Simple
A very lightweight task that runs at the same time as other code in your program.

### Technical
A function executing concurrently, scheduled by the Go runtime onto OS threads (M:N scheduling). Goroutines start with a few-KB growable stack, so programs can run hundreds of thousands. Started with the `go` keyword; the program exits when `main` returns regardless of running goroutines.

### Code
```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("hello from a goroutine")
	}()
	wg.Wait()
}
```

## Channel
related: concurrency/channels, concurrency/buffered-channels, concurrency/closing-channels

### Simple
A pipe that goroutines use to send values to each other safely.

### Technical
A typed, optionally buffered conduit created with `make(chan T, n)`. Sends and receives synchronise goroutines; receiving from a closed channel yields zero values with `ok == false`; sending on a closed channel panics. Directional types `chan<- T` and `<-chan T` restrict use.

### Code
```go
package main

import "fmt"

func main() {
	ch := make(chan int)
	go func() { ch <- 42 }()
	fmt.Println(<-ch)
}
```

## Buffered Channel
related: concurrency/buffered-channels, concurrency/unbuffered-channels

### Simple
A channel with room to hold a few values, so senders don't have to wait for a receiver right away.

### Technical
Created with a capacity (`make(chan T, n)`). A send blocks only when the buffer is full, a receive only when it is empty. `len(ch)` is the queued count and `cap(ch)` the capacity. Useful for bursts and semaphores; not a substitute for designing backpressure.

### Code
```go
package main

import "fmt"

func main() {
	ch := make(chan string, 2)
	ch <- "a"
	ch <- "b"
	fmt.Println(len(ch), cap(ch))
}
```

## Select
related: concurrency/select

### Simple
A way to wait on several channel operations at once and act on whichever is ready first.

### Technical
The `select` statement blocks until one of its cases can proceed; if several are ready one is chosen pseudo-randomly; a `default` case makes it non-blocking. Commonly used for timeouts (`time.After`) and cancellation (`ctx.Done()`).

### Code
```go
package main

import (
	"fmt"
	"time"
)

func main() {
	ch := make(chan int)
	select {
	case v := <-ch:
		fmt.Println(v)
	case <-time.After(10 * time.Millisecond):
		fmt.Println("timeout")
	}
}
```

## Interface
related: interfaces/what-are-interfaces, interfaces/implicit-interfaces, interfaces/small-interfaces

### Simple
A description of behaviour: "anything that has these methods can be used here".

### Technical
A type that specifies a method set. A concrete type satisfies an interface implicitly by implementing its methods — no `implements` keyword. Interface values hold a (dynamic type, value) pair; the zero value is `nil`. Idiom: define small interfaces at the point of use.

### Code
```go
package main

import "fmt"

type Shape interface{ Area() float64 }

type Square struct{ S float64 }

func (s Square) Area() float64 { return s.S * s.S }

func main() {
	var s Shape = Square{3}
	fmt.Println(s.Area())
}
```

## Empty Interface
related: interfaces/any, interfaces/type-assertions

### Simple
A type that can hold a value of any type — at the price of losing compile-time checks.

### Technical
`interface{}` (aliased as `any` since Go 1.18) has no methods, so every type satisfies it. To use the underlying value you need a type assertion or type switch. Prefer concrete types or generics when possible.

### Code
```go
package main

import "fmt"

func main() {
	var v any = 42
	if n, ok := v.(int); ok {
		fmt.Println(n + 1)
	}
}
```

## Type Assertion
related: interfaces/type-assertions

### Simple
Asking an interface value "are you really this concrete type?" and getting it back if so.

### Technical
`v.(T)` extracts the dynamic value of type `T` and panics on mismatch; the comma-ok form `t, ok := v.(T)` reports success instead. `T` can also be an interface type to test for extra capabilities.

### Code
```go
package main

import "fmt"

func main() {
	var v any = "go"
	s, ok := v.(string)
	n, ok2 := v.(int)
	fmt.Println(s, ok, n, ok2)
}
```

## Type Switch
related: interfaces/type-switches

### Simple
A switch that chooses a branch based on the concrete type inside an interface value.

### Technical
`switch x := v.(type) { case int: ... case string: ... }`. In each case `x` has that case's type; in multi-type cases it keeps the interface type. Best for closed sets of types such as decoded JSON.

### Code
```go
package main

import "fmt"

func kind(v any) string {
	switch v.(type) {
	case int:
		return "int"
	case string:
		return "string"
	}
	return "other"
}

func main() {
	fmt.Println(kind(1), kind("x"), kind(2.5))
}
```

## Struct
related: structs/structs, structs/struct-initialization

### Simple
A custom type that bundles several named values together.

### Technical
A composite type of named fields. Structs are values (copied on assignment), comparable with `==` when all fields are comparable, and are the primary way to model data in Go. Field visibility follows capitalisation; tags add metadata.

### Code
```go
package main

import "fmt"

type Point struct{ X, Y int }

func main() {
	p := Point{X: 1, Y: 2}
	fmt.Printf("%+v\n", p)
}
```

## Struct Tag
related: json/struct-tags, structs/struct-fields

### Simple
A little note attached to a struct field that libraries read, for example to name it in JSON.

### Technical
A raw string literal after a field type, conventionally `key:"value"` pairs read via reflection (`reflect.StructTag`). Used by `encoding/json` (`json:"name,omitempty"`), database and validation libraries. The compiler does not validate them; `go vet` checks common formats.

### Code
```go
package main

import (
	"encoding/json"
	"fmt"
)

type User struct {
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
}

func main() {
	b, _ := json.Marshal(User{Name: "Ada"})
	fmt.Println(string(b))
}
```

## Method
related: structs/methods, structs/value-receivers

### Simple
A function attached to a type so you can call it with dot syntax.

### Technical
A function with a receiver argument declared between `func` and the name. Methods can be defined on any named type declared in the same package. The receiver can be a value or a pointer.

### Code
```go
package main

import "fmt"

type Rect struct{ W, H int }

func (r Rect) Area() int { return r.W * r.H }

func main() {
	fmt.Println(Rect{3, 4}.Area())
}
```

## Receiver
related: structs/value-receivers, structs/pointer-receivers

### Simple
The value a method works on — written before the method name.

### Technical
The parameter in `func (r T) Name()` (value receiver: operates on a copy) or `func (r *T) Name()` (pointer receiver: can mutate the original and avoids copying). A type's method set determines which interfaces it satisfies. Keep receiver kinds consistent per type.

### Code
```go
package main

import "fmt"

type Counter struct{ N int }

func (c *Counter) Inc() { c.N++ }

func main() {
	var c Counter
	c.Inc()
	fmt.Println(c.N)
}
```

## Pointer
related: structs/pointers, structs/addresses, structs/dereferencing

### Simple
A value that remembers where another value lives in memory.

### Technical
A variable holding a memory address; `&x` takes an address and `*p` dereferences it. Go has no pointer arithmetic. The zero value is `nil`, and dereferencing nil panics. Escape analysis moves values whose address escapes to the heap.

### Code
```go
package main

import "fmt"

func main() {
	x := 10
	p := &x
	*p = 20
	fmt.Println(x)
}
```

## Embedding
related: structs/struct-composition, interfaces/interface-composition

### Simple
Including one type inside another so the outer type gets its fields and methods for free.

### Technical
Declaring a field with only a type name promotes the embedded type's fields and methods to the outer type (composition, not inheritance). Outer methods can shadow promoted ones. Interfaces can embed interfaces to compose method sets.

### Code
```go
package main

import "fmt"

type Animal struct{ Name string }

func (a Animal) Speak() string { return a.Name + " makes a sound" }

type Dog struct {
	Animal
	Breed string
}

func main() {
	d := Dog{Animal{"Rex"}, "Lab"}
	fmt.Println(d.Name, "|", d.Speak())
}
```

## Array
related: data-structures/arrays

### Simple
A fixed-length list of values of the same type.

### Technical
`[N]T` — the length is part of the type. Arrays are values: assignment and function calls copy every element. Slices, which reference an array, are used far more often.

### Code
```go
package main

import "fmt"

func main() {
	a := [3]int{1, 2, 3}
	b := a
	b[0] = 9
	fmt.Println(a, b)
}
```

## Slice
related: data-structures/slices, data-structures/slice-capacity, data-structures/append

### Simple
A flexible, resizable view onto a list of values.

### Technical
A descriptor (pointer, length, capacity) referencing a segment of an underlying array. `append` grows it, reallocating when capacity is exceeded; slicing shares memory. Nil slices are valid and empty. Always assign the result of `append`.

### Code
```go
package main

import "fmt"

func main() {
	s := []int{1, 2, 3}
	s = append(s, 4)
	fmt.Println(s, len(s), cap(s) >= 4, s[1:3])
}
```

## Map
related: data-structures/maps, data-structures/map-operations

### Simple
A collection that looks values up by key, like a dictionary.

### Technical
`map[K]V`, a hash table with unordered, randomised iteration. Keys must be comparable. Missing keys yield the zero value (use `v, ok := m[k]`); writing to a nil map panics. Not safe for concurrent writes without synchronisation.

### Code
```go
package main

import "fmt"

func main() {
	ages := map[string]int{"ada": 36}
	ages["grace"] = 45
	v, ok := ages["linus"]
	fmt.Println(len(ages), v, ok)
}
```

## Rune
related: data-structures/runes, data-structures/utf8

### Simple
One character of text, including non-English letters and emoji.

### Technical
An alias for `int32` representing a Unicode code point. Ranging over a string decodes runes from UTF-8; `[]rune(s)` converts for per-character editing. `len(s)` counts bytes, not runes.

### Code
```go
package main

import (
	"fmt"
	"unicode/utf8"
)

func main() {
	s := "héllo"
	fmt.Println(len(s), utf8.RuneCountInString(s))
}
```

## Byte
related: data-structures/bytes

### Simple
A tiny unit of data — a number from 0 to 255.

### Technical
An alias for `uint8`. Strings are immutable sequences of bytes; `[]byte` is the mutable counterpart used for binary data and I/O. Conversions between `string` and `[]byte` copy.

### Code
```go
package main

import "fmt"

func main() {
	b := []byte("Go")
	b[0] = 'g'
	fmt.Println(string(b), b[1])
}
```

## Zero Value
related: fundamentals/zero-values

### Simple
The default value every variable gets when you don't set one.

### Technical
Numbers are `0`, booleans `false`, strings `""`, and pointers, slices, maps, channels, functions and interfaces are `nil`. Structs are the zero value of each field. Idiomatic Go designs types so the zero value is useful (`sync.Mutex`, `bytes.Buffer`).

### Code
```go
package main

import "fmt"

func main() {
	var n int
	var s string
	var p *int
	fmt.Printf("%d %q %v\n", n, s, p)
}
```

## Constant
related: fundamentals/constants

### Simple
A named value that never changes.

### Technical
Declared with `const`, evaluated at compile time; only numbers, strings and booleans qualify. Untyped constants adapt to context. `iota` generates incrementing values within a `const` block, commonly used for enumerations.

### Code
```go
package main

import "fmt"

const (
	Low = iota
	Mid
	High
)

func main() {
	fmt.Println(Low, Mid, High)
}
```

## Shadowing
related: fundamentals/short-var-decl

### Simple
When an inner variable has the same name as an outer one and hides it.

### Technical
Using `:=` in a nested scope declares a new variable that shadows the outer identifier until the scope ends. Legal but a frequent source of bugs (for example shadowing `err`). `go vet -vettool=shadow` and linters detect it.

### Code
```go
package main

import "fmt"

func main() {
	x := 1
	if true {
		x := 2
		_ = x
	}
	fmt.Println(x)
}
```

## Closure
related: control-flow/anonymous-functions

### Simple
A function that remembers the variables around it when it was created.

### Technical
A function literal that captures variables from its enclosing scope by reference; the captured variables live as long as the closure does. Used for callbacks, counters, middleware and iterators. Since Go 1.22 each loop iteration has its own copy of the loop variable.

### Code
```go
package main

import "fmt"

func counter() func() int {
	n := 0
	return func() int { n++; return n }
}

func main() {
	c := counter()
	c()
	fmt.Println(c())
}
```

## Variadic Function
related: control-flow/variadic

### Simple
A function that accepts any number of arguments of one type.

### Technical
The last parameter has the form `...T` and is received as `[]T`. Call it with individual values or spread an existing slice with `xs...`. `append` and `fmt.Println` are variadic.

### Code
```go
package main

import "fmt"

func sum(nums ...int) (t int) {
	for _, n := range nums {
		t += n
	}
	return
}

func main() {
	fmt.Println(sum(1, 2, 3), sum([]int{4, 5}...))
}
```

## Defer
related: errors/recover, files-os/reading-files

### Simple
Schedule something to run just before the current function finishes.

### Technical
`defer f()` records a call executed when the surrounding function returns (including on panic), in last-in-first-out order. Arguments are evaluated immediately. Commonly used to close files, unlock mutexes and recover from panics.

### Code
```go
package main

import "fmt"

func main() {
	defer fmt.Println("last")
	defer fmt.Println("second")
	fmt.Println("first")
}
```

## Panic
related: errors/panic, errors/recover

### Simple
An emergency stop for bugs that can't be handled normally.

### Technical
`panic(v)` unwinds the stack running deferred calls, then crashes the program with a stack trace unless recovered. Runtime errors (nil dereference, out-of-range index) panic automatically. Use errors, not panics, for expected failures.

### Code
```go
package main

import "fmt"

func main() {
	defer func() { fmt.Println("recovered:", recover()) }()
	var m map[string]int
	m["a"] = 1
}
```

## Recover
related: errors/recover

### Simple
Catching a panic so the program can carry on.

### Technical
`recover()` returns the panic value and stops unwinding, but only when called directly inside a deferred function. Use it at boundaries (HTTP handlers, worker goroutines) to convert panics into errors. A goroutine cannot recover another goroutine's panic.

### Code
```go
package main

import "fmt"

func safe(f func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered: %v", r)
		}
	}()
	f()
	return nil
}

func main() {
	fmt.Println(safe(func() { panic("boom") }))
}
```

## Error
related: errors/error-model, errors/returning-errors

### Simple
Go's way of reporting that something went wrong — as a normal value you return.

### Technical
The built-in interface `error { Error() string }`. Functions return an `error` as the last result; `nil` means success. There are no exceptions; callers handle failures explicitly.

### Code
```go
package main

import (
	"errors"
	"fmt"
)

func div(a, b int) (int, error) {
	if b == 0 {
		return 0, errors.New("divide by zero")
	}
	return a / b, nil
}

func main() {
	fmt.Println(div(6, 3))
	fmt.Println(div(1, 0))
}
```

## Error Wrapping
related: errors/error-wrapping, errors/errors-is, errors/errors-as

### Simple
Adding context to an error while keeping the original error inside it.

### Technical
`fmt.Errorf("context: %w", err)` wraps `err`; `errors.Unwrap`, `errors.Is` and `errors.As` traverse the chain. Wrap when crossing meaningful boundaries; use `%v` when callers shouldn't depend on the cause.

### Code
```go
package main

import (
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("not found")

func main() {
	err := fmt.Errorf("load user: %w", ErrNotFound)
	fmt.Println(err, errors.Is(err, ErrNotFound))
}
```

## Sentinel Error
related: errors/creating-errors, errors/errors-is

### Simple
A well-known error value that callers can compare against, like `io.EOF`.

### Technical
A package-level variable created with `errors.New`, named `ErrXxx`. Because every `errors.New` value is distinct, compare with `errors.Is` (which sees through wrapping). Every exported sentinel becomes part of your API.

### Code
```go
package main

import (
	"errors"
	"fmt"
	"io"
)

func main() {
	err := fmt.Errorf("read: %w", io.EOF)
	fmt.Println(errors.Is(err, io.EOF))
}
```

## Package
related: fundamentals/packages, modules/packages, modules/package-organization

### Simple
A folder of Go files that belong together and can be imported by other code.

### Technical
The unit of code organisation: all `.go` files in one directory share a package name. `package main` builds an executable; other packages are libraries. Exported identifiers start with an uppercase letter. Import cycles are forbidden.

### Code
```go
package main

import (
	"fmt"
	"strings"
)

func main() {
	fmt.Println(strings.ToUpper("package"))
}
```

## Module
related: modules/go-modules, modules/go-mod-file, modules/dependencies

### Simple
A versioned collection of packages — what you download and depend on.

### Technical
A tree of packages with a `go.mod` at its root declaring the module path, Go version and requirements. Modules are versioned with semantic versions (`v1.4.2`); major versions ≥ 2 appear in the import path. `go.sum` records dependency checksums.

### Code
```go
package main

import (
	"fmt"
	"strings"
)

func main() {
	gomod := "module github.com/you/project\n\ngo 1.24\n"
	first, _, _ := strings.Cut(gomod, "\n")
	fmt.Println(first)
}
```

## go.mod
related: modules/go-mod-file, modules/go-modules

### Simple
The file that names your project and lists the libraries it needs.

### Technical
The module definition file with the directives `module`, `go`, `toolchain`, `require`, `replace` and `exclude`. Created with `go mod init`, maintained with `go get` and `go mod tidy`. Commit it together with `go.sum`.

### Code
```go
package main

import "fmt"

func main() {
	fmt.Println("go mod init example.com/hello")
}
```

## Semantic Versioning
related: modules/version-management

### Simple
A version numbering scheme (MAJOR.MINOR.PATCH) that tells you how risky an upgrade is.

### Technical
`vMAJOR.MINOR.PATCH`: patch = bug fixes, minor = backward-compatible features, major = breaking changes. Go's Minimal Version Selection builds with the highest of all required minimums. Major versions ≥ 2 change the import path.

### Code
```go
package main

import "fmt"

func main() {
	var major, minor, patch int
	fmt.Sscanf("v1.4.2", "v%d.%d.%d", &major, &minor, &patch)
	fmt.Println(major, minor, patch)
}
```

## Internal Package
related: modules/package-organization

### Simple
A package that only your own project may use, not outsiders.

### Technical
Any package under a directory named `internal` is importable only by code rooted at the parent of that directory; the compiler enforces it. Use it to keep implementation details private while exporting a small public API.

### Code
```go
package main

import "fmt"

func main() {
	fmt.Println("example.com/app/internal/billing is importable only inside example.com/app")
}
```

## Context
related: concurrency/context, concurrency/cancellation, concurrency/timeouts

### Simple
A carrier for cancellation signals, deadlines and request-scoped values across function calls.

### Technical
`context.Context` forms a tree: cancelling a parent cancels its children. Derive with `WithCancel`, `WithTimeout`, `WithDeadline`, `WithValue`. Pass it as the first parameter, select on `ctx.Done()` in long-running work and always call the returned `cancel`.

### Code
```go
package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	<-ctx.Done()
	fmt.Println(ctx.Err())
}
```

## Mutex
related: concurrency/mutex, concurrency/rwmutex, stdlib/sync

### Simple
A lock that lets only one goroutine at a time touch shared data.

### Technical
`sync.Mutex` with `Lock`/`Unlock`. Hold it briefly, `defer Unlock()`, never copy it after first use, and remember it is not reentrant. `sync.RWMutex` allows many concurrent readers or a single writer.

### Code
```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	var mu sync.Mutex
	n := 0
	mu.Lock()
	n++
	mu.Unlock()
	fmt.Println(n)
}
```

## RWMutex
related: concurrency/rwmutex

### Simple
A lock that allows many readers at once but only one writer.

### Technical
`sync.RWMutex` provides `RLock/RUnlock` for shared read access and `Lock/Unlock` for exclusive writes. It costs more than `Mutex`, so use it only for read-heavy workloads with non-trivial critical sections. Upgrading a read lock to a write lock deadlocks.

### Code
```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	var mu sync.RWMutex
	mu.RLock()
	fmt.Println("reading")
	mu.RUnlock()
}
```

## WaitGroup
related: concurrency/waitgroup

### Simple
A counter that lets you wait until a group of goroutines has finished.

### Technical
`sync.WaitGroup` with `Add(n)`, `Done()` and `Wait()`. Call `Add` before starting the goroutine, `defer Done()` inside it, and don't copy a WaitGroup. It only waits — it doesn't collect results or errors.

### Code
```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() { defer wg.Done() }()
	}
	wg.Wait()
	fmt.Println("all done")
}
```

## Atomic Operation
related: concurrency/atomic-operations

### Simple
A tiny operation on a shared number that is guaranteed to happen in one uninterruptible step.

### Technical
The `sync/atomic` package (and types like `atomic.Int64`, `atomic.Bool`) offers lock-free loads, stores, adds and compare-and-swap. Each operation is atomic on its own, but multi-step invariants still need a mutex.

### Code
```go
package main

import (
	"fmt"
	"sync/atomic"
)

func main() {
	var n atomic.Int64
	n.Add(5)
	fmt.Println(n.Load())
}
```

## Race Condition
related: concurrency/race-conditions

### Simple
A bug where the result depends on the unlucky timing of goroutines touching the same data.

### Technical
A data race is an unsynchronised concurrent access to a memory location where at least one access is a write. Behaviour is undefined. Detect with `go test -race`; fix with mutexes, atomics, channels or by not sharing.

### Code
```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	var mu sync.Mutex
	var wg sync.WaitGroup
	n := 0
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			n++
			mu.Unlock()
		}()
	}
	wg.Wait()
	fmt.Println(n)
}
```

## Deadlock
related: concurrency/unbuffered-channels

### Simple
When goroutines are stuck waiting for each other forever.

### Technical
A state where no goroutine can make progress, typically from circular waits on locks or channel operations with no counterpart. The runtime detects the case where *all* goroutines are blocked and aborts with "all goroutines are asleep - deadlock!"; partial deadlocks in servers simply hang.

### Code
```go
package main

import "fmt"

func main() {
	ch := make(chan int, 1) // buffered, so this does not deadlock
	ch <- 1
	fmt.Println(<-ch)
}
```

## Goroutine Leak
related: concurrency/goroutine-leaks, concurrency/cancellation

### Simple
A goroutine that can never finish and quietly wastes memory.

### Technical
A goroutine blocked forever (for example sending on a channel nobody reads). It retains its stack and everything it references. Every goroutine needs a guaranteed exit path: buffered results, closed channels or context cancellation. Detect with `runtime.NumGoroutine` and the goroutine profile.

### Code
```go
package main

import (
	"fmt"
	"runtime"
)

func main() {
	before := runtime.NumGoroutine()
	ch := make(chan int, 1) // buffered: the sender can always finish
	go func() { ch <- 1 }()
	<-ch
	fmt.Println(runtime.NumGoroutine() >= before)
}
```

## Worker Pool
related: concurrency/worker-pools

### Simple
A fixed team of goroutines that take jobs from a shared queue.

### Technical
N goroutines ranging over a jobs channel, sending results to another channel; the producer closes `jobs`, and a closer goroutine closes `results` after `wg.Wait()`. Bounds concurrency to protect CPUs and downstream services.

### Code
```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	jobs := make(chan int)
	var wg sync.WaitGroup
	for w := 0; w < 2; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range jobs {
			}
		}()
	}
	for i := 0; i < 4; i++ {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	fmt.Println("done")
}
```

## Pipeline
related: concurrency/pipelines, concurrency/fan-in, concurrency/fan-out

### Simple
A chain of steps where each step passes its output to the next.

### Technical
Stages connected by channels: each is a function `func(in <-chan T) <-chan U` that owns and closes its output when its input closes. Backpressure is automatic with unbuffered channels. Combine with fan-out (parallel workers) and fan-in (merge).

### Code
```go
package main

import "fmt"

func gen(n ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, v := range n {
			out <- v
		}
	}()
	return out
}

func sq(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for v := range in {
			out <- v * v
		}
	}()
	return out
}

func main() {
	for v := range sq(gen(1, 2, 3)) {
		fmt.Print(v, " ")
	}
	fmt.Println()
}
```

## Graceful Shutdown
related: concurrency/graceful-shutdown, files-os/os-signals

### Simple
Stopping a program politely: finish what's running, then exit.

### Technical
On SIGTERM/SIGINT stop accepting new work (mark unready), drain in-flight requests and jobs within a deadline (`http.Server.Shutdown(ctx)`), close resources, exit. Kubernetes waits for the grace period before SIGKILL.

### Code
```go
package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer stop()
	fmt.Println("waiting for a signal:", ctx.Err() == nil)
}
```

## Generics
related: generics/generic-functions, generics/type-parameters, generics/when-to-use-generics

### Simple
Writing one function or type that works with many different types.

### Technical
Type parameters (Go 1.18+) in square brackets, constrained by interfaces: `func Map[T, U any](xs []T, f func(T) U) []U`. Type arguments are usually inferred. Best for containers and algorithms over slices, maps and channels — not for replacing interfaces.

### Code
```go
package main

import "fmt"

func Map[T, U any](xs []T, f func(T) U) []U {
	out := make([]U, 0, len(xs))
	for _, x := range xs {
		out = append(out, f(x))
	}
	return out
}

func main() {
	fmt.Println(Map([]int{1, 2}, func(n int) string { return fmt.Sprint(n * 2) }))
}
```

## Type Parameter
related: generics/type-parameters, generics/constraints

### Simple
A placeholder type name in a generic function or type.

### Technical
Declared in brackets (`[T any]`) and replaced by a concrete type argument at instantiation. Each parameter has a constraint. Methods on generic types repeat the receiver's parameters but cannot add new ones.

### Code
```go
package main

import "fmt"

type Box[T any] struct{ V T }

func main() {
	b := Box[string]{"hi"}
	fmt.Println(b.V)
}
```

## Constraint
related: generics/constraints, generics/type-sets

### Simple
A rule saying which types are allowed for a type parameter.

### Technical
An interface used as a type-parameter bound: it can list methods and type sets (`~int | ~float64`). Standard ones: `any`, `comparable`, `cmp.Ordered`. The operations valid on `T` are those supported by every type in its set.

### Code
```go
package main

import (
	"cmp"
	"fmt"
)

func Max[T cmp.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func main() {
	fmt.Println(Max(3, 7), Max("a", "b"))
}
```

## Escape Analysis
related: performance/escape-analysis, performance/heap

### Simple
The compiler's check that decides whether a variable can live on the cheap stack or must go on the heap.

### Technical
A compile-time analysis determining if a value's address may outlive its function; if so the value is heap-allocated. View decisions with `go build -gcflags=-m`. Returning `&x`, storing in globals, escaping closures and interface boxing commonly cause escapes.

### Code
```go
package main

import "fmt"

type P struct{ X int }

func newP() *P { return &P{1} } // escapes to the heap

func main() {
	fmt.Println(newP().X)
}
```

## Garbage Collection
related: performance/garbage-collection, performance/heap

### Simple
Automatic cleanup of memory your program no longer uses.

### Technical
Go's collector is concurrent, non-generational tri-colour mark-and-sweep with short stop-the-world phases. Tune with `GOGC` (heap growth target) and `GOMEMLIMIT` (soft memory limit). Cost scales with live pointer-containing heap, so reducing allocations reduces GC work.

### Code
```go
package main

import (
	"fmt"
	"runtime"
)

func main() {
	var m runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m)
	fmt.Println(m.NumGC > 0)
}
```

## Heap
related: performance/heap, performance/stack

### Simple
The shared memory area for values that need to live longer than a single function call.

### Technical
Memory managed by the allocator and reclaimed by the garbage collector, holding values that escape or whose size isn't known at compile time. Slower than stack allocation and adds GC pressure; measure with `-benchmem`.

### Code
```go
package main

import "fmt"

func main() {
	s := make([]int, 1000) // typically heap-allocated
	fmt.Println(len(s))
}
```

## Stack
related: performance/stack

### Simple
Fast, automatically managed memory for a function's local variables.

### Technical
Each goroutine has its own growable stack (starting around 2 KB). Frames are allocated on call and discarded on return with no GC involvement. Values that don't escape live here.

### Code
```go
package main

import "fmt"

func add(a, b int) int {
	sum := a + b // lives in add's stack frame
	return sum
}

func main() {
	fmt.Println(add(1, 2))
}
```

## Benchmark
related: testing/benchmarks, performance/benchmarking

### Simple
A repeatable timing test that shows how fast a piece of code is.

### Technical
A function `BenchmarkXxx(b *testing.B)` looping `b.N` times, run with `go test -bench . -benchmem`. Compare runs statistically with `benchstat`; avoid dead-code elimination by consuming results; use `b.ResetTimer` after setup.

### Code
```go
package main

import (
	"fmt"
	"testing"
)

func main() {
	r := testing.Benchmark(func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = i * 2
		}
	})
	fmt.Println(r.N > 0)
}
```

## Table-Driven Test
related: testing/table-driven-tests, testing/subtests

### Simple
A test that runs the same check over a list of input/expected-output cases.

### Technical
A slice of named case structs looped over, usually with `t.Run(tc.name, ...)`. Adding a case is one line, failures identify the case, and `-run TestX/name` runs one. The idiomatic Go testing style.

### Code
```go
package main

import "fmt"

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func main() {
	for _, tc := range []struct{ in, want int }{{-2, 2}, {3, 3}} {
		fmt.Println(abs(tc.in) == tc.want)
	}
}
```

## Mock
related: testing/mocks, interfaces/interfaces-for-testing

### Simple
A stand-in object used in tests instead of the real dependency.

### Technical
Test doubles come in kinds: stubs (canned answers), spies (record calls), fakes (simplified working implementations) and mocks (pre-programmed expectations). In Go, implicit interfaces let you hand-write them as small structs.

### Code
```go
package main

import "fmt"

type Sender interface{ Send(string) }

type spy struct{ sent []string }

func (s *spy) Send(m string) { s.sent = append(s.sent, m) }

func main() {
	s := &spy{}
	var snd Sender = s
	snd.Send("hi")
	fmt.Println(s.sent)
}
```

## Dependency Injection
related: architecture/dependency-injection, interfaces/dependency-inversion

### Simple
Handing a component the things it needs instead of letting it create them itself.

### Technical
In Go: pass dependencies (as small interfaces) through constructors, wired in `main` (the composition root). Enables testing with fakes and swapping implementations without frameworks or globals.

### Code
```go
package main

import "fmt"

type Store interface{ Get() string }
type Service struct{ store Store }

type mem struct{}

func (mem) Get() string { return "value" }

func main() {
	svc := Service{store: mem{}}
	fmt.Println(svc.store.Get())
}
```

## Repository Pattern
related: databases/repository-pattern, architecture/repository-pattern

### Simple
A tidy layer that hides how and where data is stored behind simple methods.

### Technical
An interface owned by the domain/service (`Save`, `FindByID`) implemented by SQL, in-memory or remote backends. It translates driver errors to domain errors (`sql.ErrNoRows` → `ErrNotFound`) and lets services be tested without a database.

### Code
```go
package main

import "fmt"

type UserRepo interface{ Name(id int) (string, bool) }

type memRepo map[int]string

func (m memRepo) Name(id int) (string, bool) { n, ok := m[id]; return n, ok }

func main() {
	var r UserRepo = memRepo{1: "ada"}
	fmt.Println(r.Name(1))
}
```

## Middleware
related: networking/middleware, rest-api/handlers

### Simple
Code that wraps every request to add shared behaviour like logging or authentication.

### Technical
A function `func(http.Handler) http.Handler` that runs logic before and/or after the wrapped handler and may short-circuit. Order matters: the first in the chain is outermost. Typical stack: recover, request ID, logging, auth, rate limit.

### Code
```go
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("request:", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func main() {
	h := logging(http.NotFoundHandler())
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/x", nil))
}
```

## Handler
related: networking/http-servers, rest-api/handlers

### Simple
The function that receives an HTTP request and writes the response.

### Technical
The `http.Handler` interface has one method, `ServeHTTP(http.ResponseWriter, *http.Request)`; `http.HandlerFunc` adapts plain functions. Handlers run concurrently, one goroutine per request, so shared state must be synchronised.

### Code
```go
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

func main() {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "hello")
	})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	fmt.Println(rec.Body.String())
}
```

## REST
related: rest-api/rest-architecture, rest-api/http-methods

### Simple
A common style for web APIs where URLs name things and HTTP methods say what to do with them.

### Technical
Representational State Transfer: resources addressed by URLs, manipulated through standard methods (GET, POST, PUT, PATCH, DELETE), stateless requests, representations such as JSON, and meaningful status codes.

### Code
```go
package main

import "fmt"

func main() {
	fmt.Println("GET    /users/42")
	fmt.Println("POST   /users")
	fmt.Println("DELETE /users/42")
}
```

## Idempotency
related: microservices/idempotency, rest-api/http-methods

### Simple
Doing something twice has the same effect as doing it once — so retrying is safe.

### Technical
GET, PUT and DELETE are idempotent by definition; POST is not. Make non-idempotent operations retry-safe with an `Idempotency-Key` stored atomically with the effect. At-least-once message delivery requires idempotent consumers.

### Code
```go
package main

import "fmt"

func main() {
	done := map[string]bool{}
	pay := func(key string) string {
		if done[key] {
			return "already processed"
		}
		done[key] = true
		return "charged"
	}
	fmt.Println(pay("k1"), pay("k1"))
}
```

## JWT
related: rest-api/jwt

### Simple
A signed token a server hands out that proves who you are without a lookup.

### Technical
JSON Web Token: `base64url(header).base64url(payload).signature`. Signed (HS256/RS256), not encrypted — claims are readable. Verify signature, algorithm, expiry and audience; use short lifetimes.

### Code
```go
package main

import (
	"encoding/base64"
	"fmt"
	"strings"
)

func main() {
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"ada"}`))
	tok := "header." + payload + ".signature"
	part := strings.Split(tok, ".")[1]
	b, _ := base64.RawURLEncoding.DecodeString(part)
	fmt.Println(string(b))
}
```

## Rate Limiting
related: rest-api/rate-limiting

### Simple
Capping how many requests a client can make so nobody overloads the service.

### Technical
Algorithms: token bucket (bursts up to capacity, steady refill), leaky bucket, fixed/sliding window. Reply `429 Too Many Requests` with `Retry-After`. In-process limiters are per replica; use a shared store or gateway for global limits.

### Code
```go
package main

import "fmt"

func main() {
	tokens := 2
	for i := 0; i < 3; i++ {
		if tokens > 0 {
			tokens--
			fmt.Println("allowed")
		} else {
			fmt.Println("429 Too Many Requests")
		}
	}
}
```

## SQL Injection
related: security/sql-injection, databases/sql-fundamentals

### Simple
An attack where a user's input sneaks into a database query and changes what it does.

### Technical
Caused by concatenating untrusted input into SQL text. Prevent with parameterised queries (`$1`/`?`), allow-listed identifiers, escaped LIKE wildcards when literal, and least-privilege database accounts.

### Code
```go
package main

import "fmt"

func main() {
	query := "SELECT * FROM users WHERE name = $1" // value travels separately
	args := []any{"x' OR '1'='1"}
	fmt.Println(query, args)
}
```

## Transaction
related: databases/transactions

### Simple
A group of database changes that either all happen or none do.

### Technical
`BeginTx`, then `Commit` or `Rollback`. `defer tx.Rollback()` is safe after commit. Keep transactions short: they hold locks and a pooled connection. Retry on serialisation failures.

### Code
```go
package main

import "fmt"

func main() {
	fmt.Println("BEGIN; UPDATE accounts SET balance = balance - 10 WHERE id = 1; UPDATE accounts SET balance = balance + 10 WHERE id = 2; COMMIT;")
}
```

## Structured Logging
related: observability/structured-logging, cli/logging

### Simple
Logging events as labelled fields instead of free-form sentences, so tools can search them.

### Technical
`log/slog` emits key/value attributes through text or JSON handlers with levels. Keep messages constant and put variable data in attributes; attach request-scoped attributes with `With`; never log secrets.

### Code
```go
package main

import (
	"log/slog"
	"os"
)

func main() {
	l := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	}))
	l.Info("order placed", "order_id", 42)
}
```

## Health Check
related: observability/health-checks, observability/readiness

### Simple
An endpoint that tells the platform whether your service is alive and ready for traffic.

### Technical
Liveness (`/livez`) reports whether the process is stuck — failure restarts the container; readiness (`/readyz`) reports whether it can serve now — failure removes it from load balancing. Liveness must not depend on external services.

### Code
```go
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

func main() {
	rec := httptest.NewRecorder()
	http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "ok") }).
		ServeHTTP(rec, httptest.NewRequest("GET", "/livez", nil))
	fmt.Println(rec.Code, rec.Body.String())
}
```

## Metrics
related: observability/metrics

### Simple
Numbers your service publishes — request counts, error counts, latencies — for graphs and alerts.

### Technical
Counters (only increase), gauges (up and down) and histograms (bucketed observations). Prometheus scrapes a `/metrics` text endpoint. Use the RED method for services and avoid unbounded label values.

### Code
```go
package main

import "fmt"

func main() {
	fmt.Println("# TYPE http_requests_total counter")
	fmt.Println(`http_requests_total{code="200"} 42`)
}
```

## Circuit Breaker
related: production/resilience

### Simple
A safety switch that stops calling a failing service for a while so it can recover.

### Technical
States: closed (normal), open (fail fast after a failure threshold), half-open (probe after a cool-down). Protects callers' resources and the struggling dependency. Combine with timeouts, bounded retries and bulkheads.

### Code
```go
package main

import (
	"errors"
	"fmt"
)

func main() {
	failures, threshold := 3, 3
	if failures >= threshold {
		fmt.Println(errors.New("circuit open: failing fast"))
	}
}
```

## Reconciliation
related: kubernetes/reconciliation, kubernetes/controllers

### Simple
Repeatedly comparing "what should exist" with "what does exist" and fixing the difference.

### Technical
The core loop of a Kubernetes controller: read desired state (spec) and observed state, compute the minimal changes, apply them, update status. Must be idempotent and level-triggered so missed or duplicate events don't matter.

### Code
```go
package main

import "fmt"

func main() {
	desired, actual := 3, 1
	for actual < desired {
		actual++
		fmt.Println("create pod", actual)
	}
}
```

## Controller
related: kubernetes/controllers, kubernetes/reconciliation

### Simple
A program that watches Kubernetes objects and works to make reality match what they ask for.

### Technical
Informers deliver watch events into a de-duplicating work queue of object keys; workers call `Reconcile(key)`. Failures are re-queued with rate-limited backoff. Built with client-go or controller-runtime.

### Code
```go
package main

import "fmt"

func main() {
	queue := []string{"default/web", "default/api"}
	for _, key := range queue {
		fmt.Println("reconcile", key)
	}
}
```

## CRD
related: kubernetes/crds

### Simple
A way to teach Kubernetes about a new kind of object that you define.

### Technical
A CustomResourceDefinition registers a new resource (group, version, kind) with an OpenAPI schema validating `spec`. With a controller it forms an operator. Generated from Go types by kubebuilder/controller-gen.

### Code
```go
package main

import "fmt"

func main() {
	fmt.Println("apiVersion: example.com/v1alpha1")
	fmt.Println("kind: WebApp")
}
```

## Multi-Stage Build
related: docker/multi-stage-builds

### Simple
Building your program in one Docker image and copying only the finished binary into a tiny one.

### Technical
A Dockerfile with several `FROM` stages; `COPY --from=build` moves artefacts into the final runtime stage, so the shipped image has no compiler, source or caches. Typical Go result: a static binary on `scratch` or distroless.

### Code
```go
package main

import "fmt"

func main() {
	fmt.Println("FROM golang:1.24 AS build")
	fmt.Println("FROM gcr.io/distroless/static")
	fmt.Println("COPY --from=build /out/app /app")
}
```

## Cobra
related: cli/building-clis-with-cobra

### Simple
A popular library for building command-line tools with subcommands and help text.

### Technical
`github.com/spf13/cobra` models a CLI as a tree of `*cobra.Command` values with flags, argument validation, generated help and shell completions. It powers kubectl, the Docker CLI and gh. Often paired with viper for configuration.

### Code
```go
package main

import "fmt"

func main() {
	fmt.Println("opsctl check --timeout 5s")
}
```

## Testing Coverage
related: testing/test-coverage

### Simple
A percentage showing how much of your code the tests actually run.

### Technical
`go test -cover` reports statement coverage; `-coverprofile` plus `go tool cover -html` shows missed lines. Coverage shows what executed, not whether assertions were meaningful — use it to find untested branches.

### Code
```go
package main

import "fmt"

func main() {
	covered, total := 5, 8
	fmt.Printf("%.1f%%\n", float64(covered)/float64(total)*100)
}
```
