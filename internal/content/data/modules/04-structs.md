# Structs, Methods & Pointers
id: structs
number: 04
track: core
paths: beginner, pro
skill: structs
requires: data-structures
project: employee-management-system
summary: Model data with structs, attach behaviour with methods, and understand pointers.

## Structs
slug: structs
minutes: 6
objectives: Define a struct type; Create struct values; Compare and print structs
takeaways: A struct groups named fields into one type; Structs are values and are copied on assignment; Structs with comparable fields can be compared with ==

### Concept
A **struct** is a collection of named fields. It is Go's main tool for modelling data.

```go norun
type Point struct {
	X, Y int
}
p := Point{X: 3, Y: 4}
fmt.Println(p.X, p) // 3 {3 4}
```
`%v` prints `{3 4}`, `%+v` prints field names (`{X:3 Y:4}`), and `%#v` prints Go syntax. Structs are **values**: `q := p` copies every field.

### Example
```go
package main

import "fmt"

type Book struct {
	Title string
	Pages int
}

func main() {
	b := Book{Title: "The Go Programming Language", Pages: 380}
	c := b
	c.Pages = 400
	fmt.Printf("%+v %+v\n", b, c)
	fmt.Println(b == c)
}
```

### Exercise
Define `type Rect struct { W, H int }` and print the area of `Rect{W: 3, H: 5}`.
```text expect
15
```
```go solution
package main

import "fmt"

// BEGIN
type Rect struct {
	W, H int
}

// END

func main() {
	r := Rect{W: 3, H: 5}
	fmt.Println(r.W * r.H)
}
```

### Check
Q: What does this print?
T: output
```go
type P struct{ X, Y int }
a := P{1, 2}
b := a
b.X = 9
fmt.Println(a.X, b.X)
```
- [ ] 9 9
- [x] 1 9
- [ ] 1 1
- [ ] 9 1
E: Structs are copied on assignment; `b` is independent of `a`.

Q: Which verb prints struct field names?
T: mcq
- [ ] %v
- [x] %+v
- [ ] %s
- [ ] %n
E: `%+v` adds field names, e.g. `{X:1 Y:2}`.

## Struct Fields
slug: struct-fields
minutes: 5
objectives: Access and modify fields; Use exported and unexported fields; Add struct tags
takeaways: Fields are accessed with dot notation; Capitalised fields are exported (visible to other packages); Tags attach metadata used by encoding/json and others

### Concept
```go norun
type User struct {
	Name  string `json:"name"` // exported, with a tag
	email string // unexported: only this package can touch it
}
```
Field names starting with an uppercase letter are **exported**; libraries such as `encoding/json` can only see exported fields. **Struct tags** are string metadata read via reflection — you will use them heavily with JSON and databases.

### Example
```go
package main

import "fmt"

type User struct {
	Name string
	Age  int
}

func main() {
	u := User{"Ada", 36}
	u.Age++
	fmt.Println(u.Name, u.Age)
}
```

### Exercise
Increase `p.Balance` by 50 and print the account with `%+v` (`{Owner:Ada Balance:150}`).
```text expect
{Owner:Ada Balance:150}
```
```go solution
package main

import "fmt"

type Account struct {
	Owner   string
	Balance int
}

func main() {
	p := Account{Owner: "Ada", Balance: 100}
	// BEGIN
	p.Balance += 50
	// END
	fmt.Printf("%+v\n", p)
}
```

### Check
Q: Which field can `encoding/json` see?
T: mcq
- [ ] name string
- [x] Name string
- [ ] _name string
- [ ] both
E: Only exported (capitalised) fields are visible to other packages, including `encoding/json`.

Q: Struct tags are checked by the compiler for correct JSON syntax.
T: tf
A: false
E: Tags are plain strings; `go vet` can flag malformed ones but the compiler does not.

## Struct Initialization
slug: struct-initialization
minutes: 6
objectives: Use keyed and positional literals; Take the address of a literal; Write constructor functions
takeaways: Prefer keyed literals — they survive field changes; &T{} yields a pointer; New... constructor functions enforce invariants

### Concept
```go norun
a := User{Name: "Ada", Age: 36}   // keyed (preferred)
b := User{"Bob", 40}              // positional (fragile)
c := &User{Name: "Cy"}            // pointer to a new struct
var d User                        // zero value
```
Omitted fields get zero values. When a type has rules (a non-empty name), provide a constructor:

```go norun
func NewUser(name string) (*User, error) {
	if name == "" {
		return nil, errors.New("name required")
	}
	return &User{Name: name}, nil
}
```

### Example
```go
package main

import (
	"errors"
	"fmt"
)

type User struct {
	Name string
	Age  int
}

func NewUser(name string) (*User, error) {
	if name == "" {
		return nil, errors.New("name required")
	}
	return &User{Name: name}, nil
}

func main() {
	u, _ := NewUser("Ada")
	_, err := NewUser("")
	fmt.Println(*u, err)
}
```

### Exercise
Write `NewCounter(start int) *Counter` for `type Counter struct{ N int }` and print `N` for `NewCounter(5)`.
```text expect
5
```
```go solution
package main

import "fmt"

type Counter struct{ N int }

// BEGIN
func NewCounter(start int) *Counter {
	return &Counter{N: start}
}

// END

func main() {
	fmt.Println(NewCounter(5).N)
}
```

### Check
Q: Why prefer keyed struct literals?
T: mcq
- [ ] They run faster
- [x] They keep working when fields are added or reordered
- [ ] Positional literals are invalid
- [ ] They make fields exported
E: Positional literals break (or silently mis-assign) when the struct definition changes.

Q: What does `&User{}` produce?
T: mcq
- [ ] A User value
- [x] A *User pointing to a new zero-valued User
- [ ] A copy of nil
- [ ] A compile error
E: `&T{...}` allocates a `T` and returns its address.

## Nested Structs
slug: nested-structs
minutes: 5
objectives: Nest one struct inside another; Initialise nested values; Access deep fields
takeaways: Struct fields can themselves be structs; Use named field types for reuse and clarity; Zero values nest — an unset inner struct is the zero value of its type

### Concept
```go norun
type Address struct {
	City, Country string
}
type Person struct {
	Name string
	Home Address
}
p := Person{Name: "Ada", Home: Address{City: "London", Country: "UK"}}
fmt.Println(p.Home.City)
```
Prefer named types over anonymous inline structs except in tests and one-off data.

### Example
```go
package main

import "fmt"

type Address struct{ City, Country string }
type Person struct {
	Name string
	Home Address
}

func main() {
	p := Person{Name: "Ada"}
	p.Home.City = "London"
	fmt.Printf("%+v\n", p)
}
```

### Exercise
Given `Employee` containing a `Manager` field of type `Person`, print the manager's name via `e.Manager.Name`.
```text expect
Grace
```
```go solution
package main

import "fmt"

type Person struct{ Name string }
type Employee struct {
	Person
	Manager Person
}

func main() {
	e := Employee{Person: Person{Name: "Ada"}, Manager: Person{Name: "Grace"}}
	// BEGIN
	fmt.Println(e.Manager.Name)
	// END
}
```

### Check
Q: What is `p.Home.City` for `p := Person{}`?
T: mcq
- [ ] nil
- [x] "" (empty string)
- [ ] A panic
- [ ] undefined
E: The nested struct takes its zero value, so its string field is empty.

Q: Nested struct fields are reached with chained dots.
T: tf
A: true
E: `p.Home.City` accesses the `City` field of `Home`.

## Methods
slug: methods
challenges: bank-account
minutes: 6
objectives: Attach methods to types with receivers; Call methods with dot syntax; Add methods to non-struct types
takeaways: A method is a function with a receiver; Methods can be defined on any named type in the same package; Methods make types self-describing

### Concept
```go norun
type Rect struct{ W, H float64 }

func (r Rect) Area() float64 { return r.W * r.H }

fmt.Println(Rect{3, 4}.Area()) // 12
```
The receiver `(r Rect)` appears between `func` and the name. Methods can be defined on any **named type declared in the same package**, including `type Celsius float64`. There is no `this` or `self` keyword — you name the receiver, using a short abbreviation of the type.

### Example
```go
package main

import "fmt"

type Celsius float64

func (c Celsius) Fahrenheit() float64 { return float64(c)*9/5 + 32 }

func main() {
	fmt.Println(Celsius(100).Fahrenheit())
}
```

### Exercise
Add a method `Perimeter()` to `Rect` returning `2*(W+H)`; print the perimeter of `Rect{3, 4}`.
```text expect
14
```
```go solution
package main

import "fmt"

type Rect struct{ W, H int }

// BEGIN
func (r Rect) Perimeter() int {
	return 2 * (r.W + r.H)
}

// END

func main() {
	fmt.Println(Rect{3, 4}.Perimeter())
}
```

### Check
Q: What is the name of the receiver in `func (r Rect) Area() float64`?
T: short
A: r
E: `r` is the receiver variable; `Rect` is its type.

Q: You can define methods on the built-in type `int` directly.
T: tf
A: false
E: Methods must be declared on types defined in the same package; wrap it: `type MyInt int`.

## Value Receivers
slug: value-receivers
minutes: 5
objectives: Explain that value receivers get a copy; Use value receivers for small, read-only types; Predict when mutation is lost
takeaways: A value receiver operates on a copy; Changes to a value receiver are invisible to the caller; Good for small immutable types (Point, Money, time.Time)

### Concept
With `func (c Counter) Inc()` the method receives a **copy** of the struct. Assignments inside affect only that copy.

```go norun
func (c Counter) Inc() { c.N++ } // BUG: increments the copy
```
Value receivers are ideal for small types that behave like values (numbers, points, `time.Time`) and for methods that only read.

### Example
```go
package main

import "fmt"

type Counter struct{ N int }

func (c Counter) IncBroken() { c.N++ }

func main() {
	c := Counter{}
	c.IncBroken()
	fmt.Println(c.N)
}
```

### Exercise
`Money` has a value-receiver method `Add(other Money) Money` returning a new `Money` (do not mutate). Print `Money{Cents: 350}` after adding 200 to 150.
```text expect
{350}
```
```go solution
package main

import "fmt"

type Money struct{ Cents int }

// BEGIN
func (m Money) Add(other Money) Money {
	return Money{Cents: m.Cents + other.Cents}
}

// END

func main() {
	fmt.Println(Money{150}.Add(Money{200}))
}
```

### Check
Q: What does this print?
T: output
```go
type C struct{ N int }
func (c C) Inc() { c.N++ }
func main() {
	x := C{}
	x.Inc()
	fmt.Println(x.N)
}
```
- [x] 0
- [ ] 1
- [ ] 2
- [ ] compile error
E: `Inc` increments its own copy, so the caller's `x.N` is still 0.

Q: A value receiver is a good fit for a small immutable type such as a point.
T: tf
A: true
E: Copying a few words is cheap and gives value semantics.

## Pointer Receivers
slug: pointer-receivers
challenges: stack
minutes: 6
objectives: Use pointer receivers to modify state; Choose between pointer and value receivers; Keep receiver kinds consistent per type
takeaways: A pointer receiver can modify the original; Use pointer receivers for mutation and for large structs; Don't mix pointer and value receivers on one type

### Concept
```go norun
func (c *Counter) Inc() { c.N++ }

var c Counter
c.Inc() // Go takes &c automatically
```
Use a pointer receiver when the method **modifies** the receiver, when the struct is **large** (avoids copying), or when the type contains something that must not be copied (`sync.Mutex`). Rule of thumb: if any method needs a pointer receiver, give them all pointer receivers.

### Example
```go
package main

import "fmt"

type Counter struct{ N int }

func (c *Counter) Inc() { c.N++ }

func main() {
	var c Counter
	c.Inc()
	c.Inc()
	fmt.Println(c.N)
}
```

### Exercise
Add a pointer-receiver method `Deposit(n int)` to `Account` that adds to `Balance`. After depositing 25 then 75 print `100`.
```text expect
100
```
```go solution
package main

import "fmt"

type Account struct{ Balance int }

// BEGIN
func (a *Account) Deposit(n int) {
	a.Balance += n
}

// END

func main() {
	var a Account
	a.Deposit(25)
	a.Deposit(75)
	fmt.Println(a.Balance)
}
```

### Check
Q: Which receiver lets a method change the caller's struct?
T: mcq
- [ ] A value receiver
- [x] A pointer receiver
- [ ] An interface receiver
- [ ] A const receiver
E: A pointer receiver operates on the original value.

Q: Why give a struct containing a `sync.Mutex` pointer receivers?
T: mcq
- [ ] Mutexes are always pointers
- [x] Copying a mutex breaks its guarantees
- [ ] It is required by the compiler
- [ ] Value receivers cannot read fields
E: A copied mutex is a different lock. `go vet` flags copies of types that contain a `sync.Mutex`.

## Pointers
slug: pointers
minutes: 7
objectives: Explain what a pointer is; Use & and * correctly; Recognise nil pointers and their danger
takeaways: A pointer holds the memory address of a value; & takes an address, * dereferences; Dereferencing a nil pointer panics

### Concept
A pointer stores where a value lives.

```go norun
x := 10
p := &x    // p is *int, pointing at x
*p = 20    // write through the pointer
fmt.Println(x) // 20
var q *int // nil
```
Go has **no pointer arithmetic**. Pointers let functions modify their arguments, avoid copying large values, and represent "absent" (`nil`). Reading or writing through a nil pointer panics — check before dereferencing.

### Example
```go
package main

import "fmt"

func setToTen(p *int) { *p = 10 }

func main() {
	n := 1
	setToTen(&n)
	fmt.Println(n)
}
```

### Exercise
Write `swap(a, b *int)` that swaps the values the pointers refer to. Print `2 1` after swapping `1` and `2`.
```text expect
2 1
```
```go solution
package main

import "fmt"

// BEGIN
func swap(a, b *int) {
	*a, *b = *b, *a
}

// END

func main() {
	x, y := 1, 2
	swap(&x, &y)
	fmt.Println(x, y)
}
```

### Check
Q: What does this print?
T: output
```go
x := 5
p := &x
*p = *p + 1
fmt.Println(x)
```
- [ ] 5
- [x] 6
- [ ] a memory address
- [ ] 1
E: `*p` refers to `x`, so writing through `p` updates `x`.

Q: What happens when you dereference a nil pointer?
T: mcq
- [ ] It yields zero
- [x] A run-time panic
- [ ] A compile error
- [ ] Undefined behaviour
E: Go panics with "invalid memory address or nil pointer dereference".

## Addresses
slug: addresses
minutes: 5
objectives: Take addresses of variables and fields; Print and compare pointers; Know why returning &local is safe
takeaways: &x gives the address of x; Two pointers are equal when they point to the same variable; Returning a pointer to a local variable is safe — Go moves it to the heap

### Concept
You can take the address of variables, struct fields and slice elements (not of map entries or constants).

```go norun
u := User{Name: "Ada"}
p := &u.Name
*p = "Grace"
```
In C, returning the address of a local is a bug. In Go the compiler's **escape analysis** allocates such a value on the heap automatically, so `return &local` is safe. Pointers to the same variable compare equal.

### Example
```go
package main

import "fmt"

type T struct{ N int }

func newT() *T {
	t := T{N: 1}
	return &t // safe: escapes to the heap
}

func main() {
	a := newT()
	b := a
	fmt.Println(a == b, a.N)
}
```

### Exercise
Take the address of the `Age` field of `u` and use it to set the age to 30. Print `30`.
```text expect
30
```
```go solution
package main

import "fmt"

type User struct {
	Name string
	Age  int
}

func main() {
	u := User{Name: "Ada"}
	// BEGIN
	p := &u.Age
	*p = 30
	// END
	fmt.Println(u.Age)
}
```

### Check
Q: Is it safe to return the address of a local variable in Go?
T: tf
A: true
E: The compiler detects that the variable escapes and allocates it on the heap.

Q: Can you take the address of a map element (`&m["k"]`)?
T: mcq
- [ ] Yes
- [x] No — map elements aren't addressable
- [ ] Only for struct values
- [ ] Only for ints
E: Maps may move entries when they grow, so their elements are not addressable.

## Dereferencing
slug: dereferencing
challenges: reverse-linked-list
minutes: 5
objectives: Read and write through pointers; Use automatic dereferencing for struct fields; Guard against nil pointers
takeaways: *p reads or writes the pointed-to value; p.Field is shorthand for (*p).Field; Check for nil before dereferencing when nil is a possible value

### Concept
`*p` is the value `p` points to. For struct pointers Go lets you write `p.Name` instead of `(*p).Name`.

```go norun
type Node struct {
	Val  int
	Next *Node
}
func length(n *Node) int {
	if n == nil {
		return 0
	}
	return 1 + length(n.Next)
}
```
A `*Node` naturally represents "maybe absent" — `nil` ends the list.

### Example
```go
package main

import "fmt"

type Node struct {
	Val  int
	Next *Node
}

func sum(n *Node) int {
	if n == nil {
		return 0
	}
	return n.Val + sum(n.Next)
}

func main() {
	list := &Node{1, &Node{2, &Node{3, nil}}}
	fmt.Println(sum(list))
}
```

### Exercise
Write `length(n *Node) int` counting nodes in the linked list (0 for nil).
```text expect
3 0
```
```go solution
package main

import "fmt"

type Node struct {
	Val  int
	Next *Node
}

// BEGIN
func length(n *Node) int {
	count := 0
	for ; n != nil; n = n.Next {
		count++
	}
	return count
}

// END

func main() {
	list := &Node{1, &Node{2, &Node{3, nil}}}
	fmt.Println(length(list), length(nil))
}
```

### Check
Q: Which is equivalent to `(*p).Name` when `p` is a `*User`?
T: mcq
- [ ] *p.Name
- [x] p.Name
- [ ] &p.Name
- [ ] p->Name
E: Go automatically dereferences struct pointers when accessing fields.

Q: What is the value of `n.Next` for the last node in the example list?
T: short
A: nil
E: `nil` marks the end of a linked list.

## Struct Composition
slug: struct-composition
challenges: employee-payroll
minutes: 7
objectives: Embed one struct in another; Use promoted fields and methods; Explain composition over inheritance
takeaways: Embedding promotes the inner type's fields and methods; Go has no inheritance — composition is the tool; The outer type can override a promoted method

### Concept
Embed a type by listing it without a field name:

```go norun
type Animal struct{ Name string }
func (a Animal) Describe() string { return a.Name }

type Dog struct {
	Animal        // embedded
	Breed string
}
d := Dog{Animal{"Rex"}, "Lab"}
d.Name       // promoted field
d.Describe() // promoted method
```
This is **composition**, not inheritance: there is no subtype relationship, and a `Dog` is not an `Animal`. If `Dog` defines its own `Describe`, it shadows the promoted one (the inner is still reachable as `d.Animal.Describe()`).

### Example
```go
package main

import "fmt"

type Logger struct{ Prefix string }

func (l Logger) Log(msg string) { fmt.Println(l.Prefix + msg) }

type Server struct {
	Logger
	Addr string
}

func main() {
	s := Server{Logger{"[srv] "}, ":8080"}
	s.Log("listening on " + s.Addr)
}
```

### Exercise
Embed `Person` in `Employee`; `Employee` adds `Title`. Print `Ada, Engineer` using the promoted `Name` field.
```text expect
Ada, Engineer
```
```go solution
package main

import "fmt"

type Person struct{ Name string }

// BEGIN
type Employee struct {
	Person
	Title string
}

// END

func main() {
	e := Employee{Person: Person{Name: "Ada"}, Title: "Engineer"}
	fmt.Println(e.Name + ", " + e.Title)
}
```

### Check
Q: Does embedding make `Dog` a subtype of `Animal`?
T: mcq
- [ ] Yes, it is inheritance
- [x] No — it is composition; methods are promoted but there is no subtype relation
- [ ] Only for interfaces
- [ ] Only for pointers
E: You can't pass a `Dog` where an `Animal` value is required; interfaces express substitutability in Go.

Q: What does this print?
T: output
```go
type A struct{}
func (A) Hi() string { return "A" }
type B struct{ A }
func (B) Hi() string { return "B" }
func main() {
	fmt.Println(B{}.Hi(), B{}.A.Hi())
}
```
- [ ] A A
- [x] B A
- [ ] A B
- [ ] B B
E: The outer type's method shadows the promoted one; the embedded value is still accessible by its type name.
