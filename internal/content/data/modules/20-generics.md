# Generics
id: generics
number: 20
track: advanced-go
paths: beginner, pro
skill: generics
requires: interfaces
summary: Type parameters, constraints and when generics help — or hurt.

## Generic Functions
slug: generic-functions
minutes: 7
challenges: generic-map-filter
objectives: Write a function with a type parameter; Call it with inferred and explicit type arguments; Replace duplicated code across types
takeaways: func Name[T any](x T) declares a type parameter; The compiler usually infers type arguments from the call; Generics remove copy-paste for algorithms that don't depend on the concrete type

### Concept
Before generics you wrote `MaxInt`, `MaxFloat`, `MaxString`… or used `any` and lost type safety. Now:

```go norun
func Map[T, U any](xs []T, f func(T) U) []U {
	out := make([]U, 0, len(xs))
	for _, x := range xs {
		out = append(out, f(x))
	}
	return out
}

names := Map([]int{1, 2, 3}, strconv.Itoa) // T=int, U=string inferred
```
Type parameters go in square brackets before the parameter list. `any` means "no constraint". Explicit instantiation is sometimes needed: `Map[int, string](...)`. The compiler generates efficient code (GC shape stenciling), so there's no boxing like with `any`.

### Example
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
	lens := Map([]string{"go", "gopher", ""}, func(s string) int { return len(s) })
	fmt.Println(lens)
}
```

### Exercise
Write `Filter[T any](xs []T, keep func(T) bool) []T` and print the even numbers from `1..6`.
```text expect
[2 4 6]
```
```go solution
package main

import "fmt"

// BEGIN
func Filter[T any](xs []T, keep func(T) bool) []T {
	var out []T
	for _, x := range xs {
		if keep(x) {
			out = append(out, x)
		}
	}
	return out
}

// END

func main() {
	fmt.Println(Filter([]int{1, 2, 3, 4, 5, 6}, func(n int) bool { return n%2 == 0 }))
}
```

### Check
Q: Where do type parameters appear in a function declaration?
T: mcq
- [ ] After the return type
- [x] In square brackets after the function name
- [ ] Inside the body
- [ ] In angle brackets
E: `func Map[T, U any](...)`.

Q: What does `any` mean as a constraint?
T: mcq
- [ ] Only numeric types
- [x] Any type is allowed
- [ ] Only interfaces
- [ ] Only pointers
E: `any` is an alias for `interface{}` and places no restriction.

## Type Parameters
slug: type-parameters
minutes: 7
objectives: Declare multiple type parameters; Explain instantiation and inference; Use generic types with type parameters
takeaways: A function or type can have several type parameters: [K comparable, V any]; Instantiation supplies concrete types — explicitly or by inference; Methods on generic types repeat the receiver's type parameters but can't add new ones

### Concept
```go norun
type Pair[K comparable, V any] struct {
	Key K
	Val V
}

func (p Pair[K, V]) String() string { return fmt.Sprintf("%v=%v", p.Key, p.Val) }

p := Pair[string, int]{"a", 1} // explicit instantiation
q := Pair[string, int]{Key: "b", Val: 2}
```
Inference works for **function arguments**, not for generic type literals — you write `Pair[string, int]{...}`, though a constructor function can infer: `func NewPair[K comparable, V any](k K, v V) Pair[K, V]`. Methods can't declare *extra* type parameters (`func (p Pair[K,V]) Map[U any]` is illegal) — use a top-level generic function instead.

### Example
```go
package main

import "fmt"

type Pair[K comparable, V any] struct {
	Key K
	Val V
}

func NewPair[K comparable, V any](k K, v V) Pair[K, V] { return Pair[K, V]{k, v} }

func (p Pair[K, V]) String() string { return fmt.Sprintf("%v=%v", p.Key, p.Val) }

func main() {
	p := NewPair("port", 8080) // K=string, V=int inferred
	fmt.Println(p)
}
```

### Exercise
Write `Keys[K comparable, V any](m map[K]V) []K` returning the keys (sorted by the caller — print with the length only) and print `3`.
```text expect
3
```
```go solution
package main

import "fmt"

// BEGIN
func Keys[K comparable, V any](m map[K]V) []K {
	out := make([]K, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// END

func main() {
	fmt.Println(len(Keys(map[string]int{"a": 1, "b": 2, "c": 3})))
}
```

### Check
Q: Can a method on a generic type declare its own additional type parameters?
T: tf
A: false
E: Methods may only use the type parameters of their receiver; use a generic function for extras.

Q: What does `comparable` allow?
T: mcq
- [ ] Ordering with <
- [x] Comparison with == and !=, e.g. as map keys
- [ ] Only strings
- [ ] Any type at all
E: `comparable` constrains to types that support equality.

## Constraints
slug: constraints
minutes: 8
objectives: Restrict type parameters with interfaces; Use cmp.Ordered and comparable; Write your own constraint with methods
takeaways: A constraint is an interface that limits which types may be used; cmp.Ordered covers types supporting < and >; constraints can require methods — like any interface — and type sets

### Concept
```go norun
import "cmp"

func Max[T cmp.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

type Stringer interface{ String() string }

func Join[T Stringer](xs []T, sep string) string { ... }
```
Built-in/standard constraints: `any`, `comparable`, `cmp.Ordered` (integers, floats, strings). Combine methods and type sets in one constraint interface. Note that since Go 1.21 the standard library already provides generic `min`, `max`, `slices.Max`, `slices.Sort`, `slices.Contains`… — reach for those before writing your own.

### Example
```go
package main

import (
	"cmp"
	"fmt"
)

func Max[T cmp.Ordered](xs ...T) T {
	m := xs[0]
	for _, x := range xs[1:] {
		if x > m {
			m = x
		}
	}
	return m
}

func main() {
	fmt.Println(Max(3, 9, 4), Max("go", "rust", "c"), Max(2.5, 1.5))
}
```

### Exercise
Write `Clamp[T cmp.Ordered](v, lo, hi T) T`. Print `Clamp(15, 0, 10)`, `Clamp(-1.5, 0.0, 1.0)` and `Clamp("m", "a", "f")`.
```text expect
10 0 f
```
```go solution
package main

import (
	"cmp"
	"fmt"
)

// BEGIN
func Clamp[T cmp.Ordered](v, lo, hi T) T {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// END

func main() {
	fmt.Println(Clamp(15, 0, 10), Clamp(-1.5, 0.0, 1.0), Clamp("m", "a", "f"))
}
```

### Check
Q: Which constraint allows the `<` operator on type parameters?
T: short
A: cmp.Ordered
E: `cmp.Ordered` covers all built-in ordered types.

Q: A constraint can require methods.
T: tf
A: true
E: Constraints are ordinary interfaces; `interface{ String() string }` restricts T to types with that method.

## Generic Data Structures
slug: generic-data-structures
minutes: 9
challenges: generic-stack, generic-lru-cache
objectives: Implement a generic Stack and Set; Use zero values with var zero T; Expose a small, ergonomic API
takeaways: Generic containers avoid interface{} casts and duplicated code per element type; Return (T, bool) rather than a magic zero value when an element may be absent; map[T]struct{} implements a Set[T comparable]

### Concept
```go norun
type Stack[T any] struct{ items []T }

func (s *Stack[T]) Push(x T) { s.items = append(s.items, x) }

func (s *Stack[T]) Pop() (T, bool) {
	var zero T
	if len(s.items) == 0 {
		return zero, false
	}
	x := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return x, true
}
```
`var zero T` yields T's zero value whatever T is. A Set:

```go norun
type Set[T comparable] map[T]struct{}
func (s Set[T]) Add(x T)           { s[x] = struct{}{} }
func (s Set[T]) Has(x T) bool      { _, ok := s[x]; return ok }
```
Prefer the standard library where it exists (`slices`, `maps`, `container/list`) and add your own generic types only for real repeated needs.

### Example
```go
package main

import "fmt"

type Stack[T any] struct{ items []T }

func (s *Stack[T]) Push(x T) { s.items = append(s.items, x) }

func (s *Stack[T]) Pop() (T, bool) {
	var zero T
	if len(s.items) == 0 {
		return zero, false
	}
	x := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return x, true
}

func main() {
	var s Stack[string]
	s.Push("a")
	s.Push("b")
	fmt.Println(s.Pop())
	fmt.Println(s.Pop())
	fmt.Println(s.Pop())
}
```

### Exercise
Implement `Set[T comparable]` (a map type) with `Add`, `Has` and `Len`. Add `1, 2, 2, 3` and print `3 true false`.
```text expect
3 true false
```
```go solution
package main

import "fmt"

// BEGIN
type Set[T comparable] map[T]struct{}

func (s Set[T]) Add(x T)      { s[x] = struct{}{} }
func (s Set[T]) Has(x T) bool { _, ok := s[x]; return ok }
func (s Set[T]) Len() int     { return len(s) }

// END

func main() {
	s := Set[int]{}
	for _, n := range []int{1, 2, 2, 3} {
		s.Add(n)
	}
	fmt.Println(s.Len(), s.Has(2), s.Has(9))
}
```

### Check
Q: How do you return the zero value of type parameter `T`?
T: mcq
- [ ] return nil
- [x] var zero T; return zero
- [ ] return T{}
- [ ] return 0
E: The zero value depends on T; declaring `var zero T` works for any type.

Q: Why return `(T, bool)` from Pop instead of just `T`?
T: mcq
- [ ] It's required by generics
- [x] The caller can tell "empty" from a real zero value
- [ ] It's faster
- [ ] To avoid panics in append
E: An empty stack and a stack holding a zero value would otherwise look identical.

## Type Sets
slug: type-sets
minutes: 8
objectives: Write constraints as unions of types; Use ~T to include defined types; Know what operations a type set permits
takeaways: A constraint can list types: interface{ ~int | ~float64 }; ~T means "any type whose underlying type is T"; Operators are allowed on T only if every type in the set supports them

### Concept
```go norun
type Number interface {
	~int | ~int64 | ~float64
}

func Sum[T Number](xs []T) T {
	var total T
	for _, x := range xs {
		total += x // + is valid for every type in the set
	}
	return total
}

type Celsius float64
Sum([]Celsius{20, 5.5}) // works because of ~float64
```
`~float64` includes any **defined type** whose underlying type is `float64`; plain `float64` would exclude `Celsius`. Interfaces with type unions can only be used as constraints, not as regular variable types. `golang.org/x/exp/constraints` (and now `cmp.Ordered`) provide common sets.

### Example
```go
package main

import "fmt"

type Number interface {
	~int | ~int64 | ~float64
}

func Sum[T Number](xs ...T) T {
	var total T
	for _, x := range xs {
		total += x
	}
	return total
}

type Celsius float64

func main() {
	fmt.Println(Sum(1, 2, 3), Sum(1.5, 2.5), Sum(Celsius(20), Celsius(1.5)))
}
```

### Exercise
Write `Average[T Number](xs []T) float64` (0 for empty) using the `Number` constraint and print the average of `[]int{2, 4, 9}` formatted with `%.1f`.
```text expect
5.0
```
```go solution
package main

import "fmt"

type Number interface {
	~int | ~int64 | ~float64
}

// BEGIN
func Average[T Number](xs []T) float64 {
	if len(xs) == 0 {
		return 0
	}
	var total T
	for _, x := range xs {
		total += x
	}
	return float64(total) / float64(len(xs))
}

// END

func main() {
	fmt.Printf("%.1f\n", Average([]int{2, 4, 9}))
}
```

### Check
Q: What does `~int` mean in a constraint?
T: mcq
- [ ] The bitwise negation of int
- [x] Any type whose underlying type is int, including defined types like `type ID int`
- [ ] Only exactly int
- [ ] Pointer to int
E: The tilde broadens the set to types built on int.

Q: Can `T Number` be used with `+` if Number is `~int | ~float64`?
T: tf
A: true
E: All types in the set support addition, so the operator is permitted.

## When to Use Generics
slug: when-to-use-generics
minutes: 7
objectives: Recognise the good use cases for generics; Prefer generics over interface{} and code generation for containers and algorithms; Apply the "three similar copies" rule
takeaways: Good fits: container types, algorithms over slices/maps/channels, and functions that only differ by element type; Generic code should stay simple — if you need reflection or type switches, reconsider; Wait for real duplication before generalising

### Concept
Use generics when you write **the same code for different types**:

- Data structures: `Stack[T]`, `Set[T]`, `Cache[K, V]`, `LinkedList[T]`
- Slice/map/channel helpers: `Map`, `Filter`, `Reduce`, `Keys`, `Uniq`
- Functions constrained by an operator set: `Sum[T Number]`
- Result/Option-like wrappers, thread-safe wrappers (`atomic.Pointer[T]`, `sync.OnceValue`)

Signals you need them: you're copying a function and changing only types; you're using `any` and immediately casting back; you're tempted to generate code.

**Rule of thumb:** write concrete code first; when you've copied it a third time, generalise.

### Example
```go
package main

import (
	"fmt"
	"slices"
)

func main() {
	// The standard library is generic now: no per-type helpers needed.
	ints := []int{3, 1, 2}
	strs := []string{"b", "a"}
	slices.Sort(ints)
	slices.Sort(strs)
	fmt.Println(ints, strs, slices.Contains(strs, "a"))
}
```

### Exercise
Write `Reduce[T, A any](xs []T, init A, f func(A, T) A) A` and use it to sum `1..4` and to join `["a","b"]` as `ab`.
```text expect
10 ab
```
```go solution
package main

import "fmt"

// BEGIN
func Reduce[T, A any](xs []T, init A, f func(A, T) A) A {
	acc := init
	for _, x := range xs {
		acc = f(acc, x)
	}
	return acc
}

// END

func main() {
	sum := Reduce([]int{1, 2, 3, 4}, 0, func(a, x int) int { return a + x })
	joined := Reduce([]string{"a", "b"}, "", func(a string, x string) string { return a + x })
	fmt.Println(sum, joined)
}
```

### Check
Q: Which is a good use of generics?
T: mcq
- [ ] A function that behaves differently for each type via a type switch
- [x] A Set[T comparable] container used with many element types
- [ ] Replacing every interface in your codebase
- [ ] Avoiding writing tests
E: Generics shine where the code is identical across types.

Q: When should you generalise a function?
T: mcq
- [ ] Immediately, before writing it
- [x] After you've written similar code for several concrete types
- [ ] Never
- [ ] Only in libraries
E: Wait for real duplication; premature abstraction adds complexity.

## When Not to Use Generics
slug: when-not-to-use-generics
minutes: 7
objectives: Recognise when an interface is the better abstraction; Avoid generics that add complexity without removing duplication; Know performance and readability trade-offs
takeaways: If different types need different behaviour, use an interface with methods — not generics plus a type switch; Don't use generics just because you can: simple concrete code is easier to read; Generic APIs are harder to evolve — keep exported generic surface small

### Concept
Avoid generics when:

1. **Behaviour differs per type** — that's what interfaces (method sets) are for. `io.Reader` isn't generic and shouldn't be.
2. **There is only one concrete use.** A `Repository[T]` with one implementation for `User` is ceremony.
3. **You'd need reflection or type switches inside the generic function** — it's not really generic.
4. **Readability suffers.** `func Process[K comparable, V any, R Result[K, V]](…)` is harder than two small concrete functions.
5. **Public API stability matters.** Once exported, constraints are hard to change.

Interfaces express *what a value can do*; generics express *the same algorithm for many types*. Ask which question you're answering.

### Example
```go
package main

import "fmt"

// Different behaviour per type: an interface is the right tool, not generics.
type Shape interface{ Area() float64 }

type Square struct{ S float64 }
type Circle struct{ R float64 }

func (s Square) Area() float64 { return s.S * s.S }
func (c Circle) Area() float64 { return 3 * c.R * c.R }

func Total(shapes []Shape) (sum float64) {
	for _, s := range shapes {
		sum += s.Area()
	}
	return
}

func main() {
	fmt.Println(Total([]Shape{Square{2}, Circle{1}}))
}
```

### Exercise
The `Describe` function below is a generic function that switches on the type — a smell. Replace it with the `Namer` interface: implement `Name() string` on `Dog` and `Cat` (returning `dog` and `cat`) and make `Describe` take a `Namer`.
```text expect
I am dog
I am cat
```
```go solution
package main

import "fmt"

// BEGIN
type Namer interface{ Name() string }

type Dog struct{}
type Cat struct{}

func (Dog) Name() string { return "dog" }
func (Cat) Name() string { return "cat" }

func Describe(n Namer) string { return "I am " + n.Name() }

// END

func main() {
	fmt.Println(Describe(Dog{}))
	fmt.Println(Describe(Cat{}))
}
```

### Check
Q: Different types need different behaviour behind one operation. What should you use?
T: mcq
- [ ] Generics with a type switch
- [x] An interface with methods
- [ ] reflect
- [ ] Code generation only
E: Methods let each type provide its own behaviour.

Q: A generic type used with exactly one concrete type is usually a sign of over-abstraction.
T: tf
A: true
E: If there's no duplication to remove, generics only add complexity.
