# Interfaces
id: interfaces
number: 05
track: core
paths: beginner, pro
skill: interfaces
requires: structs
project: pluggable-storage-system
summary: Design small interfaces, use implicit satisfaction, and decouple code for testing.

## What Are Interfaces?
slug: what-are-interfaces
minutes: 6
objectives: Define an interface type; Explain that an interface describes behaviour; Use an interface value in a function
takeaways: An interface is a set of method signatures; Any type with those methods satisfies it; Functions that accept interfaces work with many concrete types

### Concept
An **interface** describes *what a value can do*, not what it is.

```go norun
type Shape interface {
	Area() float64
}

func Print(s Shape) { fmt.Println(s.Area()) }
```
`Print` accepts a circle, a rectangle, or any future type with an `Area() float64` method. This is Go's main tool for abstraction and decoupling.

### Example
```go
package main

import (
	"fmt"
	"math"
)

type Shape interface{ Area() float64 }

type Rect struct{ W, H float64 }
type Circle struct{ R float64 }

func (r Rect) Area() float64   { return r.W * r.H }
func (c Circle) Area() float64 { return math.Pi * c.R * c.R }

func main() {
	shapes := []Shape{Rect{3, 4}, Circle{1}}
	for _, s := range shapes {
		fmt.Printf("%.2f\n", s.Area())
	}
}
```

### Exercise
Add a `Square` type with an `Area()` method (side × side) so it satisfies `Shape`; the program prints `9`.
```text expect
9
```
```go solution
package main

import "fmt"

type Shape interface{ Area() int }

// BEGIN
type Square struct{ Side int }

func (s Square) Area() int { return s.Side * s.Side }

// END

func main() {
	var s Shape = Square{3}
	fmt.Println(s.Area())
}
```

### Check
Q: What does an interface type define?
T: mcq
- [ ] Fields a struct must have
- [x] A set of method signatures
- [ ] A base class
- [ ] A memory layout
E: Interfaces describe behaviour as a method set.

Q: A function parameter of interface type can receive any value whose type has the required methods.
T: tf
A: true
E: That is what makes interfaces useful for decoupling.

## Interface Methods
slug: interface-methods
minutes: 6
objectives: Declare multiple methods in an interface; Call methods through an interface value; Understand method sets for value and pointer types
takeaways: An interface lists every method a type must have; Calling a method on an interface dispatches dynamically to the concrete type; A pointer-receiver method is only in the method set of the pointer type

### Concept
```go norun
type Stringer interface{ String() string }
type ReadWriter interface {
	Read(p []byte) (int, error)
	Write(p []byte) (int, error)
}
```
**Method sets:** a value of type `T` has only the *value-receiver* methods; a `*T` has both. So if `Save` has a pointer receiver, only `*File` satisfies `Saver`, not `File`.

### Example
```go
package main

import "fmt"

type Counter struct{ n int }

func (c *Counter) Inc()     { c.n++ }
func (c *Counter) Count() int { return c.n }

type CounterI interface {
	Inc()
	Count() int
}

func main() {
	var c CounterI = &Counter{} // must be a pointer
	c.Inc()
	c.Inc()
	fmt.Println(c.Count())
}
```

### Exercise
`Speaker` requires `Speak() string`. Make `Dog` satisfy it and print `Woof`.
```text expect
Woof
```
```go solution
package main

import "fmt"

type Speaker interface{ Speak() string }

type Dog struct{}

// BEGIN
func (Dog) Speak() string { return "Woof" }

// END

func main() {
	var s Speaker = Dog{}
	fmt.Println(s.Speak())
}
```

### Check
Q: If `Inc` has a pointer receiver, which value satisfies `interface{ Inc() }`?
T: mcq
- [ ] Counter{}
- [x] &Counter{}
- [ ] both
- [ ] neither
E: A `Counter` value's method set lacks pointer-receiver methods; `*Counter` includes them.

Q: What is the term for the set of methods callable on a type's value?
T: short
A: method set
E: The method set determines which interfaces a type satisfies.

## Implicit Interface Implementation
slug: implicit-interfaces
minutes: 5
objectives: Explain implicit satisfaction; Define interfaces at the point of use; Verify satisfaction at compile time
takeaways: Types satisfy interfaces without declaring it; Define interfaces in the package that uses them; var _ I = (*T)(nil) asserts satisfaction at compile time

### Concept
Go has no `implements` keyword. A type satisfies an interface **just by having the methods**. That means a package can define an interface that types from other packages already satisfy — even standard-library ones.

```go norun
type Sizer interface{ Size() int64 }
// *os.File has no idea Sizer exists, but…
```
To have the compiler check that a type keeps satisfying an interface, add an assertion:

```go norun
var _ io.Reader = (*MyReader)(nil)
```

### Example
```go
package main

import (
	"fmt"
	"strings"
)

// A tiny interface defined by the consumer.
type Namer interface{ Name() string }

type Cat struct{}

func (Cat) Name() string { return "cat" }

func greet(n Namer) string { return strings.ToUpper(n.Name()) }

var _ Namer = Cat{} // compile-time check

func main() {
	fmt.Println(greet(Cat{}))
}
```

### Exercise
`ConsoleLogger` must satisfy `Logger`. The compile-time assertion is already there — add the missing `Log` method (print `[log] ` followed by the message) so the program compiles and prints `[log] hi`.
```text expect
[log] hi
```
```go solution
package main

import "fmt"

type Logger interface{ Log(string) }

type ConsoleLogger struct{}

var _ Logger = ConsoleLogger{}

// BEGIN
func (ConsoleLogger) Log(m string) { fmt.Println("[log] " + m) }

// END

func main() {
	var l Logger = ConsoleLogger{}
	l.Log("hi")
}
```

### Check
Q: How does a Go type declare that it implements an interface?
T: mcq
- [ ] With the implements keyword
- [ ] By embedding the interface
- [x] It doesn't — having the methods is enough
- [ ] With a struct tag
E: Satisfaction is implicit and structural.

Q: Interfaces are best defined by the package that ...
T: mcq
- [ ] implements them
- [x] consumes them
- [ ] is the oldest
- [ ] is named "interfaces"
E: "Accept interfaces, return structs": the consumer knows the minimal behaviour it needs.

## Small Interfaces
slug: small-interfaces
minutes: 6
objectives: Prefer one- and two-method interfaces; Recognise io.Reader and io.Writer; Explain why small interfaces are powerful
takeaways: The bigger the interface, the weaker the abstraction; io.Reader and io.Writer are one-method interfaces that power all of Go I/O; Split fat interfaces by what each caller needs

### Concept
> The bigger the interface, the weaker the abstraction. — Go Proverb

`io.Reader` (`Read([]byte) (int, error)`) and `io.Writer` (`Write([]byte) (int, error)`) are one-method interfaces, yet files, network connections, HTTP bodies, buffers, gzip streams and strings all plug into the same functions (`io.Copy`, `bufio.Scanner`, `json.NewDecoder`).

Design your own the same way: ask "what is the least this function needs?" and accept exactly that.

### Example
```go
package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	var r io.Reader = strings.NewReader("streams are everywhere\n")
	n, _ := io.Copy(os.Stdout, r)
	fmt.Println(n)
}
```

### Exercise
Write `count(r io.Reader) (int, error)` returning the number of bytes readable from `r`, using `io.Copy(io.Discard, r)`.
```text expect
5 5
```
```go solution
package main

import (
	"fmt"
	"io"
	"strings"
)

// BEGIN
func count(r io.Reader) (int, error) {
	n, err := io.Copy(io.Discard, r)
	return int(n), err
}

// END

func main() {
	a, _ := count(strings.NewReader("hello"))
	b, _ := count(strings.NewReader("world"))
	fmt.Println(a, b)
}
```

### Check
Q: How many methods does `io.Reader` have?
T: short
A: 1
E: `io.Reader` has a single method: `Read(p []byte) (n int, err error)`.

Q: Which is preferable for a function that only reads data?
T: mcq
- [ ] func f(f *os.File)
- [x] func f(r io.Reader)
- [ ] func f(s string)
- [ ] func f(v interface{})
E: `io.Reader` accepts files, strings, network connections and test doubles.

## Interface Composition
slug: interface-composition
minutes: 5
objectives: Build larger interfaces from smaller ones; Recognise io.ReadWriter and io.ReadCloser; Compose only when callers need it
takeaways: Interfaces can embed other interfaces; io.ReadWriteCloser = Reader + Writer + Closer; Compose at the point of need rather than by default

### Concept
```go norun
type ReadWriter interface {
	Reader
	Writer
}
```
The standard library composes `io.ReadCloser`, `io.WriteCloser`, `io.ReadWriteCloser`. A `*os.File` satisfies them all. Compose interfaces in the function that needs them, rather than building large "kitchen-sink" interfaces up front.

### Example
```go
package main

import "fmt"

type Reader interface{ Read() string }
type Writer interface{ Write(string) }
type ReadWriter interface {
	Reader
	Writer
}

type Buf struct{ data string }

func (b *Buf) Read() string    { return b.data }
func (b *Buf) Write(s string) { b.data += s }

func main() {
	var rw ReadWriter = &Buf{}
	rw.Write("go")
	rw.Write("pher")
	fmt.Println(rw.Read())
}
```

### Exercise
Define `type ReadCloser interface { Reader; Closer }` given both parts, so the program compiles and prints `closed`.
```text expect
closed
```
```go solution
package main

import "fmt"

type Reader interface{ Read() string }
type Closer interface{ Close() string }

// BEGIN
type ReadCloser interface {
	Reader
	Closer
}

// END

type File struct{}

func (File) Read() string  { return "data" }
func (File) Close() string { return "closed" }

func main() {
	var rc ReadCloser = File{}
	fmt.Println(rc.Close())
}
```

### Check
Q: How do you compose two interfaces into one?
T: mcq
- [ ] Use the + operator
- [x] Embed them in a new interface
- [ ] Use inheritance
- [ ] Interfaces can't be combined
E: An interface can embed other interfaces, adding their methods to its method set.

Q: `io.ReadWriteCloser` is built from three smaller interfaces.
T: tf
A: true
E: It embeds `io.Reader`, `io.Writer` and `io.Closer`.

## any
slug: any
minutes: 5
objectives: Explain that any is the empty interface; Know why any loses type safety; Prefer concrete types or generics
takeaways: any is an alias for interface{} — every type satisfies it; A value of type any must be asserted before use; Overusing any throws away compile-time checking

### Concept
`any` (alias of `interface{}`) has no methods, so **every** type satisfies it.

```go norun
var v any = 42
v = "now a string"
```
It shows up in `fmt.Println(a ...any)` and JSON decoding into `map[string]any`. But a value of type `any` can't be used until you recover its type, and the compiler can no longer help you. Prefer concrete types, small interfaces, or (Module 20) generics.

### Example
```go
package main

import "fmt"

func describe(v any) {
	fmt.Printf("%v (%T)\n", v, v)
}

func main() {
	describe(42)
	describe("go")
	describe([]int{1, 2})
}
```

### Exercise
Write `count(items ...any) int` returning how many arguments were passed. Print `3`.
```text expect
3
```
```go solution
package main

import "fmt"

// BEGIN
func count(items ...any) int {
	return len(items)
}

// END

func main() {
	fmt.Println(count(1, "two", 3.0))
}
```

### Check
Q: What is `any`?
T: mcq
- [ ] A keyword that disables type checking
- [x] An alias for interface{}
- [ ] A generic constraint only
- [ ] A pointer type
E: `any` was added in Go 1.18 as a readable alias for the empty interface.

Q: You can call `v + 1` directly on `var v any = 41`.
T: tf
A: false
E: The static type is `any`; assert to `int` first.

## Type Assertions
slug: type-assertions
minutes: 6
objectives: Extract the concrete value from an interface; Use the comma-ok form to avoid panics; Assert to another interface
takeaways: v.(T) asserts the dynamic type; The single-value form panics on failure; Use n, ok := v.(T) for safe checks

### Concept
```go norun
var v any = "hello"
s := v.(string)      // ok
n, ok := v.(int)     // 0, false — no panic
n2 := v.(int)        // PANIC
```
You can also assert to an interface to test for optional behaviour:

```go norun
if s, ok := w.(io.StringWriter); ok {
	s.WriteString("fast path")
}
```

### Example
```go
package main

import "fmt"

func main() {
	var v any = 3.14
	if f, ok := v.(float64); ok {
		fmt.Println("float", f)
	}
	if _, ok := v.(string); !ok {
		fmt.Println("not a string")
	}
}
```

### Exercise
Write `asInt(v any) (int, bool)` that returns the int inside `v` or `0, false`.
```text expect
7 true
0 false
```
```go solution
package main

import "fmt"

// BEGIN
func asInt(v any) (int, bool) {
	n, ok := v.(int)
	return n, ok
}

// END

func main() {
	fmt.Println(asInt(7))
	fmt.Println(asInt("7"))
}
```

### Check
Q: What happens with `n := v.(int)` when `v` holds a string?
T: mcq
- [ ] n is 0
- [ ] n is nil
- [x] A run-time panic
- [ ] A compile error
E: The single-value assertion panics on mismatch; use the comma-ok form.

Q: What does this print?
T: output
```go
var v any = "go"
_, isInt := v.(int)
s, isStr := v.(string)
fmt.Println(isInt, isStr, s)
```
- [ ] true false go
- [x] false true go
- [ ] false false
- [ ] true true go
E: `v` holds a string, so the int assertion fails and the string assertion succeeds.

## Type Switches
slug: type-switches
minutes: 6
objectives: Branch on the dynamic type with a type switch; Bind the typed value in each case; Handle default and nil
takeaways: switch v := x.(type) branches on dynamic type; In each case v has that case's type; Use it for a closed, small set of types — otherwise prefer methods

### Concept
```go norun
switch v := x.(type) {
case int:
	fmt.Println("int", v+1)
case string, []byte:
	fmt.Println("text", v) // v is still 'any' here
case nil:
	fmt.Println("nil")
default:
	fmt.Printf("other %T\n", v)
}
```
A type switch is neat for decoding loosely-typed data (JSON into `any`). When you control the types, add a method to an interface instead of switching.

### Example
```go
package main

import "fmt"

func describe(x any) string {
	switch v := x.(type) {
	case int:
		return fmt.Sprintf("int %d", v)
	case string:
		return "string " + v
	case nil:
		return "nil"
	default:
		return fmt.Sprintf("other %T", v)
	}
}

func main() {
	fmt.Println(describe(1), describe("x"), describe(nil), describe(2.5))
}
```

### Exercise
Write `double(x any) any` returning `2*n` for ints, `s+s` for strings and `nil` otherwise.
```text expect
8 gogo <nil>
```
```go solution
package main

import "fmt"

// BEGIN
func double(x any) any {
	switch v := x.(type) {
	case int:
		return v * 2
	case string:
		return v + v
	}
	return nil
}

// END

func main() {
	fmt.Println(double(4), double("go"), double(1.5))
}
```

### Check
Q: What does this print?
T: output
```go
var x any = 5
switch x.(type) {
case string:
	fmt.Println("string")
case int:
	fmt.Println("int")
}
```
- [ ] string
- [x] int
- [ ] any
- [ ] nothing
E: The dynamic type of `x` is `int`, so the `int` case runs.

Q: When you control all the types involved, what is usually better than a type switch?
T: mcq
- [ ] Reflection
- [x] A method on a shared interface
- [ ] A global map
- [ ] Panic and recover
E: Polymorphism through interface methods is more extensible than a central switch.

## Dependency Inversion
slug: dependency-inversion
minutes: 7
objectives: Depend on abstractions rather than concrete types; Inject dependencies through constructors; Keep interfaces small and consumer-defined
takeaways: High-level code should depend on interfaces it defines; Inject dependencies via constructors — no globals; Concrete wiring lives in main

### Concept
A `Notifier` service that calls `smtp.Send` directly is hard to test and welded to email. Invert the dependency:

```go norun
type Sender interface{ Send(to, msg string) error }

type Notifier struct{ sender Sender }

func NewNotifier(s Sender) *Notifier { return &Notifier{sender: s} }
```
`Notifier` knows only `Sender`. `main` chooses `EmailSender`, `SMSSender`, or a test fake. The **consumer defines the interface**; implementations depend on nothing.

### Example
```go
package main

import "fmt"

type Sender interface{ Send(to, msg string) error }

type ConsoleSender struct{}

func (ConsoleSender) Send(to, msg string) error {
	fmt.Printf("to=%s msg=%s\n", to, msg)
	return nil
}

type Notifier struct{ s Sender }

func NewNotifier(s Sender) *Notifier { return &Notifier{s} }

func (n *Notifier) Welcome(user string) error {
	return n.s.Send(user, "welcome!")
}

func main() {
	_ = NewNotifier(ConsoleSender{}).Welcome("ada")
}
```

### Exercise
Implement `UpperSender` whose `Send` prints the message uppercased, and inject it into `Notifier`. Output: `ADA: WELCOME`.
```text expect
ADA: WELCOME
```
```go solution
package main

import (
	"fmt"
	"strings"
)

type Sender interface{ Send(to, msg string) error }

// BEGIN
type UpperSender struct{}

func (UpperSender) Send(to, msg string) error {
	fmt.Printf("%s: %s\n", strings.ToUpper(to), strings.ToUpper(msg))
	return nil
}

// END

type Notifier struct{ s Sender }

func NewNotifier(s Sender) *Notifier { return &Notifier{s} }

func (n *Notifier) Welcome(user string) error { return n.s.Send(user, "welcome") }

func main() {
	_ = NewNotifier(UpperSender{}).Welcome("ada")
}
```

### Check
Q: Where should concrete implementations be chosen and wired together?
T: mcq
- [ ] Inside each service's methods
- [ ] In init functions
- [x] In main (the composition root)
- [ ] In the interface definition
E: Keeping wiring in `main` keeps everything else decoupled and testable.

Q: Global variables are a good way to provide dependencies to services.
T: tf
A: false
E: Explicit constructor injection makes dependencies visible and replaceable in tests.

## Interfaces for Testing
slug: interfaces-for-testing
minutes: 7
objectives: Replace real dependencies with fakes; Write a hand-rolled fake; Keep test seams small
takeaways: Small consumer-defined interfaces make fakes trivial; Hand-written fakes beat heavy mocking frameworks for most cases; Test behaviour, not implementation details

### Concept
When a service depends on an interface, a test passes in a **fake** that records calls or returns canned data:

```go norun
type fakeSender struct{ sent []string }

func (f *fakeSender) Send(to, msg string) error {
	f.sent = append(f.sent, to+":"+msg)
	return nil
}
```
No network, no flakiness, instant. Because Go interfaces are implicit, the fake needs no registration — it just has the right methods. You will write real tests in Module 15; this lesson shows why interfaces make them easy.

### Example
```go
package main

import "fmt"

type Sender interface{ Send(to, msg string) error }

type fake struct{ sent []string }

func (f *fake) Send(to, msg string) error {
	f.sent = append(f.sent, to+":"+msg)
	return nil
}

func welcome(s Sender, user string) error { return s.Send(user, "hi") }

func main() {
	f := &fake{}
	_ = welcome(f, "ada")
	_ = welcome(f, "bob")
	fmt.Println(f.sent)
}
```

### Exercise
Write a fake `failing` whose `Send` always returns an error, so `welcome` prints `send failed`.
```text expect
send failed
```
```go solution
package main

import (
	"errors"
	"fmt"
)

type Sender interface{ Send(to, msg string) error }

// BEGIN
type failing struct{}

func (failing) Send(to, msg string) error { return errors.New("send failed") }

// END

func welcome(s Sender, user string) error { return s.Send(user, "hi") }

func main() {
	if err := welcome(failing{}, "ada"); err != nil {
		fmt.Println(err)
	}
}
```

### Check
Q: Why do small interfaces make testing easier?
T: mcq
- [ ] They compile faster
- [x] A fake only needs to implement a few methods
- [ ] They avoid the need for tests
- [ ] They remove pointers
E: The smaller the interface, the less code a fake needs.

Q: A test double must be registered with the interface it implements.
T: tf
A: false
E: Interface satisfaction is implicit — having the methods is enough.
