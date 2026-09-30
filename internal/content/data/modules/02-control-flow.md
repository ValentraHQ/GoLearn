# Control Flow & Functions
id: control-flow
number: 02
track: core
paths: beginner, pro
skill: functions
requires: fundamentals
project: cli-calculator
summary: Make decisions, repeat work, and package logic into functions with multiple return values.

## If Statements
slug: if-statements
minutes: 5
objectives: Write if conditions without parentheses; Use an init statement in an if; Combine conditions with && and ||
takeaways: Conditions need no parentheses but braces are mandatory; if can start with a short statement scoped to the block; Conditions must be bool — there is no truthiness

### Concept
```go norun
if score >= 50 {
	fmt.Println("pass")
}
```
Braces are required, parentheses are not. The condition must be a real `bool`: `if n {` does not compile for an int.

An `if` may begin with a short statement whose variables live only inside the `if`/`else` chain:

```go norun
if n, err := strconv.Atoi(s); err == nil {
	fmt.Println(n)
}
```

### Example
```go
package main

import "fmt"

func main() {
	age := 20
	if age >= 18 && age < 65 {
		fmt.Println("adult")
	}
	if n := age * 2; n > 30 {
		fmt.Println("double is", n)
	}
}
```

### Exercise
Print `even` if `n` is even (use `n%2 == 0`).
```text expect
even
```
```go solution
package main

import "fmt"

func main() {
	n := 10
	// BEGIN
	if n%2 == 0 {
		fmt.Println("even")
	}
	// END
}
```

### Check
Q: Which is valid Go?
T: mcq
- [ ] if (x > 1) fmt.Println(x)
- [ ] if x { ... } where x is an int
- [x] if x > 1 { fmt.Println(x) }
- [ ] if x > 1: fmt.Println(x)
E: Braces are mandatory and the condition must be a bool expression.

Q: A variable declared in an if's init statement is visible after the if statement ends.
T: tf
A: false
E: Its scope is the if/else chain only.

## Else and Else If
slug: else-if
minutes: 5
objectives: Chain decisions with else if; Prefer early returns over deep nesting; Place else on the same line as the closing brace
takeaways: else must follow the closing brace on the same line; Go style prefers returning early to nesting else blocks; Use switch when an else-if chain gets long

### Concept
```go norun
if n < 0 {
	fmt.Println("negative")
} else if n == 0 {
	fmt.Println("zero")
} else {
	fmt.Println("positive")
}
```
`else` must sit on the same line as `}` (automatic semicolon insertion would otherwise break it). Idiomatic Go avoids `else` after a `return`:

```go norun
if err != nil {
	return err
}
// continue with the happy path, unindented
```

### Example
```go
package main

import "fmt"

func grade(score int) string {
	if score >= 90 {
		return "A"
	} else if score >= 80 {
		return "B"
	}
	return "C"
}

func main() {
	fmt.Println(grade(95), grade(85), grade(10))
}
```

### Exercise
Complete `sign` so it returns `"negative"`, `"zero"` or `"positive"`.
```text expect
negative zero positive
```
```go solution
package main

import "fmt"

func sign(n int) string {
	// BEGIN
	if n < 0 {
		return "negative"
	} else if n == 0 {
		return "zero"
	}
	return "positive"
	// END
}

func main() {
	fmt.Println(sign(-3), sign(0), sign(8))
}
```

### Check
Q: Where must `else` appear?
T: mcq
- [ ] On its own line below }
- [x] On the same line as the closing }
- [ ] Anywhere
- [ ] Before the }
E: A newline after `}` inserts a semicolon that would detach the `else`; gofmt enforces the correct form.

Q: In idiomatic Go, after `if err != nil { return err }` you normally write `else { ... }` for the success path.
T: tf
A: false
E: Return early and keep the success path at the left margin.

## For Loops
slug: for-loops
minutes: 7
objectives: Use the three-part for loop; Use for as a while loop; Range over slices, strings and integers
takeaways: for is Go's only loop keyword; for cond {} replaces while; for i := range n counts from 0 to n-1 (Go 1.22+); range yields index and value

### Concept
```go norun
for i := 0; i < 3; i++ { }   // classic
for i < 10 { i *= 2 }        // while
for i := range 3 { }         // 0, 1, 2  (Go 1.22+)
for i, v := range []string{"a", "b"} { }
```
`range` over a slice gives `(index, value)`, over a string gives `(byte index, rune)`, over a map gives `(key, value)`. Since Go 1.22 each iteration has its own copy of the loop variable, so closures capture what you expect.

### Example
```go
package main

import "fmt"

func main() {
	sum := 0
	for i := 1; i <= 4; i++ {
		sum += i
	}
	fmt.Println(sum)
	for i, c := range "go" {
		fmt.Println(i, string(c))
	}
}
```

### Exercise
Print the numbers 1 to 5 separated by spaces on one line. Use `fmt.Print` inside the loop and end with a newline.
```text expect
1 2 3 4 5
```
```go solution
package main

import "fmt"

func main() {
	// BEGIN
	for i := 1; i <= 5; i++ {
		if i > 1 {
			fmt.Print(" ")
		}
		fmt.Print(i)
	}
	fmt.Println()
	// END
}
```

### Check
Q: What does this print?
T: output
```go
total := 0
for i := range 4 {
	total += i
}
fmt.Println(total)
```
- [ ] 10
- [x] 6
- [ ] 4
- [ ] 3
E: `range 4` yields 0, 1, 2, 3 and their sum is 6.

Q: Which keyword implements a while loop in Go?
T: short
A: for
E: Go has a single loop keyword; `for condition { }` is the while form.

## Infinite Loops
slug: infinite-loops
minutes: 5
objectives: Write an infinite for loop; Exit loops with break or return; Recognise loops that wait for events
takeaways: for { } loops forever until break, return, or panic; Servers and workers are usually infinite loops with a clear exit; Always give a long-running loop a way to stop

### Concept
A `for` with no condition loops forever:

```go norun
for {
	// do work
	if done {
		break
	}
}
```
This is normal for servers, REPLs and workers. What matters is a **guaranteed exit path** — a `break`, a `return`, or (later) a cancelled `context`.

### Example
```go
package main

import "fmt"

func main() {
	n := 1
	for {
		n *= 2
		if n > 100 {
			break
		}
	}
	fmt.Println(n)
}
```

### Exercise
Double `n` starting at 3 inside an infinite loop and stop once it exceeds 50. Print the final value.
```text expect
96
```
```go solution
package main

import "fmt"

func main() {
	n := 3
	// BEGIN
	for {
		n *= 2
		if n > 50 {
			break
		}
	}
	// END
	fmt.Println(n)
}
```

### Check
Q: What does `for { }` do without a break or return?
T: mcq
- [ ] Does not compile
- [ ] Runs once
- [x] Runs forever
- [ ] Runs until a timeout
E: A condition-less `for` is an infinite loop.

Q: An infinite loop is always a bug.
T: tf
A: false
E: Many programs (servers, workers, event loops) intentionally loop forever with a controlled exit.

## Break
slug: break
minutes: 5
objectives: Stop a loop early with break; Break out of an outer loop using a label; Know that break also exits switch and select
takeaways: break exits the innermost for, switch or select; A label lets break exit an outer loop; Prefer extracting a function and returning over labels

### Concept
`break` leaves the innermost `for`, `switch` or `select`. To leave an outer loop, label it:

```go norun
outer:
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if j == 2 {
				break outer
			}
		}
	}
```

### Example
```go
package main

import "fmt"

func main() {
	nums := []int{4, 8, 15, 16, 23}
	for i, n := range nums {
		if n > 15 {
			fmt.Println("first over 15 at index", i)
			break
		}
	}
}
```

### Exercise
Find the first negative number in the slice, print it, and stop.
```text expect
-2
```
```go solution
package main

import "fmt"

func main() {
	nums := []int{5, 3, -2, 8, -9}
	// BEGIN
	for _, n := range nums {
		if n < 0 {
			fmt.Println(n)
			break
		}
	}
	// END
}
```

### Check
Q: What does this print?
T: output
```go
for i := 0; i < 5; i++ {
	if i == 3 {
		break
	}
	fmt.Print(i)
}
```
- [ ] 01234
- [x] 012
- [ ] 0123
- [ ] 12
E: The loop stops when `i` reaches 3, after printing 0, 1 and 2.

Q: Inside a `switch` within a `for`, a bare `break` exits the loop.
T: tf
A: false
E: It exits only the `switch`. Use a label (or return) to leave the loop.

## Continue
slug: continue
minutes: 5
objectives: Skip to the next iteration with continue; Use guard clauses in loops; Combine continue with labels
takeaways: continue jumps to the next iteration; Guard clauses with continue reduce nesting; continue label targets an outer loop

### Concept
`continue` skips the rest of the current iteration:

```go norun
for _, n := range nums {
	if n%2 != 0 {
		continue
	}
	fmt.Println(n) // even numbers only
}
```
It plays the same role in loops that an early `return` plays in functions.

### Example
```go
package main

import "fmt"

func main() {
	for i := 1; i <= 6; i++ {
		if i%3 == 0 {
			continue
		}
		fmt.Print(i, " ")
	}
	fmt.Println()
}
```

### Exercise
Print only the odd numbers between 1 and 7, one per line, using `continue` to skip evens.
```text expect
1
3
5
7
```
```go solution
package main

import "fmt"

func main() {
	// BEGIN
	for i := 1; i <= 7; i++ {
		if i%2 == 0 {
			continue
		}
		fmt.Println(i)
	}
	// END
}
```

### Check
Q: What does this print?
T: output
```go
for i := 0; i < 4; i++ {
	if i == 1 {
		continue
	}
	fmt.Print(i)
}
```
- [ ] 0123
- [x] 023
- [ ] 01
- [ ] 123
E: Iteration 1 is skipped, so the output is 0, 2, 3.

Q: `continue` exits the loop entirely.
T: tf
A: false
E: `continue` only skips to the next iteration; `break` exits.

## Switch
slug: switch
minutes: 7
objectives: Write expression and tagless switches; Use multiple values per case; Know that cases do not fall through
takeaways: Cases do not fall through; use fallthrough explicitly if needed; A tagless switch replaces long if/else chains; Cases can list several values

### Concept
```go norun
switch day {
case "sat", "sun":
	fmt.Println("weekend")
default:
	fmt.Println("weekday")
}

switch { // tagless: each case is a bool
case n < 0:
case n == 0:
default:
}
```
Go breaks automatically after a case. Cases are evaluated top to bottom. A **type switch** (`switch v := x.(type)`) appears in the Interfaces module.

### Example
```go
package main

import "fmt"

func kind(n int) string {
	switch {
	case n < 0:
		return "negative"
	case n == 0:
		return "zero"
	case n < 10:
		return "small"
	default:
		return "large"
	}
}

func main() {
	fmt.Println(kind(-1), kind(0), kind(7), kind(70))
}
```

### Exercise
Implement `days(month)`: return 28 for `"feb"`, 30 for `"apr"`, `"jun"`, `"sep"`, `"nov"`, and 31 otherwise.
```text expect
28 30 31
```
```go solution
package main

import "fmt"

func days(month string) int {
	// BEGIN
	switch month {
	case "feb":
		return 28
	case "apr", "jun", "sep", "nov":
		return 30
	default:
		return 31
	}
	// END
}

func main() {
	fmt.Println(days("feb"), days("jun"), days("jan"))
}
```

### Check
Q: What does this print?
T: output
```go
switch 2 {
case 1:
	fmt.Println("one")
case 2:
	fmt.Println("two")
case 3:
	fmt.Println("three")
}
```
- [ ] two / three
- [x] two
- [ ] one / two
- [ ] one / two / three
E: Go cases don't fall through, so only the matching case runs.

Q: Which keyword makes execution continue into the next case?
T: short
A: fallthrough
E: `fallthrough` is explicit and rare in idiomatic Go.

## Functions
slug: functions
minutes: 6
objectives: Declare functions with parameters and results; Call functions and use their results; Understand that Go passes arguments by value
takeaways: func name(params) result { }; Arguments are copied (passed by value); Functions are first-class values

### Concept
```go norun
func add(a int, b int) int {
	return a + b
}
func area(w, h float64) float64 { return w * h } // shared type
```
Go passes **everything by value** — the function gets a copy. (Slices, maps and pointers copy a small header/reference, so they can still let a callee modify shared data.)

Functions are values: you can assign them to variables and pass them around.

### Example
```go
package main

import "fmt"

func square(n int) int { return n * n }

func main() {
	f := square
	fmt.Println(f(6))
}
```

### Exercise
Write `cube(n int) int` and print `cube(3)`.
```text expect
27
```
```go solution
package main

import "fmt"

// BEGIN
func cube(n int) int {
	return n * n * n
}

// END

func main() {
	fmt.Println(cube(3))
}
```

### Check
Q: How does Go pass function arguments?
T: mcq
- [ ] By reference
- [x] By value (a copy)
- [ ] By name
- [ ] Depends on the type
E: Everything is copied. Pointers copy the address, which is how a function can modify the caller's data.

Q: What does this print?
T: output
```go
func double(n int) { n *= 2 }
func main() {
	x := 4
	double(x)
	fmt.Println(x)
}
```
- [x] 4
- [ ] 8
- [ ] 2
- [ ] 0
E: `double` modifies its own copy of `n`; `x` is unchanged.

## Parameters
slug: parameters
minutes: 6
objectives: Group parameters of the same type; Pass slices and maps to functions; Choose value versus pointer parameters
takeaways: Adjacent parameters can share a type; Slices and maps share underlying data when passed; Use pointers when a function must modify a value

### Concept
```go norun
func repeat(s string, n int) string
func between(min, max, v int) bool
```
When a function receives a **slice** it gets a copy of the slice header, but the elements are shared, so changing `xs[0]` inside is visible outside. Appending is not (the caller's length is unchanged).

Use a **pointer parameter** when you need to modify the caller's variable:

```go norun
func inc(n *int) { *n++ }
```

### Example
```go
package main

import "fmt"

func zero(xs []int) { xs[0] = 0 }

func inc(n *int) { *n++ }

func main() {
	xs := []int{1, 2, 3}
	zero(xs)
	n := 1
	inc(&n)
	fmt.Println(xs, n)
}
```

### Exercise
Write `incAll(xs []int)` that adds 1 to every element in place, then print the slice.
```text expect
[2 3 4]
```
```go solution
package main

import "fmt"

// BEGIN
func incAll(xs []int) {
	for i := range xs {
		xs[i]++
	}
}

// END

func main() {
	xs := []int{1, 2, 3}
	incAll(xs)
	fmt.Println(xs)
}
```

### Check
Q: You pass a slice to a function that sets `xs[0] = 99`. What happens to the caller's slice?
T: mcq
- [ ] Unchanged — the slice was copied
- [x] The first element changes too
- [ ] A compile error
- [ ] The whole slice is copied deeply
E: The slice header is copied but points to the same backing array.

Q: To let a function change an int variable owned by the caller you pass...
T: mcq
- [ ] the int by value
- [x] a pointer to the int
- [ ] the int by reference keyword
- [ ] a const
E: A `*int` lets the callee modify the original.

## Return Values
slug: return-values
minutes: 5
objectives: Return values from functions; Use named results carefully; Understand bare returns
takeaways: A function may return any number of values; Named results document meaning and enable defer tweaks; Avoid bare returns in long functions

### Concept
```go norun
func half(n int) int { return n / 2 }

func split(sum int) (x, y int) { // named results
	x = sum * 4 / 9
	y = sum - x
	return // bare return: returns x, y
}
```
Named results are useful for documentation (`(n int, err error)`) and for `defer` blocks that adjust the result. Bare returns hurt readability in long functions; prefer explicit `return x, y`.

### Example
```go
package main

import "fmt"

func minmax(a, b int) (lo, hi int) {
	if a < b {
		return a, b
	}
	return b, a
}

func main() {
	fmt.Println(minmax(9, 4))
}
```

### Exercise
Complete `abs` so it returns the absolute value of `n`.
```text expect
5 5 0
```
```go solution
package main

import "fmt"

func abs(n int) int {
	// BEGIN
	if n < 0 {
		return -n
	}
	return n
	// END
}

func main() {
	fmt.Println(abs(-5), abs(5), abs(0))
}
```

### Check
Q: Every path through a function that has a result type must...
T: mcq
- [ ] print something
- [x] end in a return (or panic)
- [ ] call another function
- [ ] use named results
E: The compiler reports "missing return" otherwise.

Q: What does this print?
T: output
```go
func f() (n int) {
	n = 3
	return 7
}

func main() {
	fmt.Println(f())
}
```
- [ ] 3
- [x] 7
- [ ] 10
- [ ] 0
E: An explicit `return 7` assigns 7 to the named result before returning.

## Multiple Return Values
slug: multiple-returns
minutes: 6
objectives: Return and receive multiple values; Follow the (value, error) convention; Ignore results with the blank identifier
takeaways: Go functions commonly return (result, error); Use _ to discard a value you don't need; Check the error before using the result

### Concept
Multiple returns are the backbone of Go error handling:

```go norun
n, err := strconv.Atoi("42")
if err != nil {
	return err
}
```
Use the **blank identifier** `_` to discard a value: `_, err := os.Stat(path)`. Never discard an `error` without a reason.

### Example
```go
package main

import (
	"errors"
	"fmt"
)

func divide(a, b int) (int, error) {
	if b == 0 {
		return 0, errors.New("divide by zero")
	}
	return a / b, nil
}

func main() {
	if q, err := divide(9, 3); err == nil {
		fmt.Println(q)
	}
	_, err := divide(1, 0)
	fmt.Println(err)
}
```

### Exercise
Write `divmod(a, b int) (int, int)` returning quotient and remainder, and print both.
```text expect
3 1
```
```go solution
package main

import "fmt"

// BEGIN
func divmod(a, b int) (int, int) {
	return a / b, a % b
}

// END

func main() {
	q, r := divmod(10, 3)
	fmt.Println(q, r)
}
```

### Check
Q: What is the conventional order of results for a function that can fail?
T: mcq
- [ ] (error, value)
- [x] (value, error)
- [ ] (bool, value)
- [ ] error only
E: By convention the `error` is the last result.

Q: What does `_` do in `_, err := f()`?
T: mcq
- [ ] Declares a private variable
- [x] Discards the first result
- [ ] Marks it optional
- [ ] Causes a compile error
E: The blank identifier drops a value without naming it.

## Variadic Functions
slug: variadic
minutes: 6
objectives: Declare variadic parameters with ...; Call with individual values or a slice...; Recognise variadics in the standard library
takeaways: The last parameter can be ...T and arrives as a []T; Pass an existing slice with xs...; fmt.Println and append are variadic

### Concept
```go norun
func sum(nums ...int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}
sum()          // 0
sum(1, 2, 3)   // 6
xs := []int{4, 5}
sum(xs...)     // 9
```
Inside the function `nums` is an ordinary `[]int` (nil when no arguments were passed). Only the **last** parameter can be variadic.

### Example
```go
package main

import "fmt"

func join(sep string, parts ...string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += sep
		}
		out += p
	}
	return out
}

func main() {
	fmt.Println(join("-", "a", "b", "c"))
	words := []string{"x", "y"}
	fmt.Println(join("+", words...))
}
```

### Exercise
Write `largest(nums ...int) int` returning the largest argument (assume at least one).
```text expect
9
```
```go solution
package main

import "fmt"

// BEGIN
func largest(nums ...int) int {
	m := nums[0]
	for _, n := range nums[1:] {
		if n > m {
			m = n
		}
	}
	return m
}

// END

func main() {
	fmt.Println(largest(3, 9, 4))
}
```

### Check
Q: How do you pass an existing slice `xs` to a variadic `f(nums ...int)`?
T: mcq
- [ ] f(xs)
- [x] f(xs...)
- [ ] f(...xs)
- [ ] f([]xs)
E: A trailing `...` unpacks the slice into the variadic parameter.

Q: What is the type of `nums` inside `func f(nums ...int)`?
T: short
A: []int
E: A variadic parameter is received as a slice.

## Anonymous Functions
slug: anonymous-functions
minutes: 7
objectives: Define and call function literals; Capture variables in closures; Use closures for counters and callbacks
takeaways: A function literal has no name and can be called immediately; Closures capture variables, not values; Closures are how Go does callbacks and generators

### Concept
```go norun
f := func(x int) int { return x * 2 }
func() { fmt.Println("now") }() // define and call
```
A **closure** remembers the variables of the scope it was created in:

```go norun
func counter() func() int {
	n := 0
	return func() int { n++; return n }
}
```
Each call to `counter()` produces an independent `n`.

### Example
```go
package main

import "fmt"

func counter() func() int {
	n := 0
	return func() int {
		n++
		return n
	}
}

func main() {
	a, b := counter(), counter()
	fmt.Println(a(), a(), b())
}
```

### Exercise
Write `adder(step int) func(int) int` that returns a function adding `step` to its argument.
```text expect
15 20
```
```go solution
package main

import "fmt"

// BEGIN
func adder(step int) func(int) int {
	return func(n int) int { return n + step }
}

// END

func main() {
	add5 := adder(5)
	add10 := adder(10)
	fmt.Println(add5(10), add10(10))
}
```

### Check
Q: What does this print?
T: output
```go
c := func() func() int {
	n := 0
	return func() int { n++; return n }
}()
c()
fmt.Println(c())
```
- [ ] 1
- [x] 2
- [ ] 0
- [ ] 3
E: The closure keeps its own `n`, so the second call returns 2.

Q: A closure copies the values of the variables it uses at creation time.
T: tf
A: false
E: Closures capture the variables themselves, so later changes are visible to the closure.
