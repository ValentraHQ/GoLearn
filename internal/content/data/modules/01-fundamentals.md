# Go Fundamentals
id: fundamentals
number: 01
track: core
paths: beginner, pro
skill: fundamentals
requires: setup
project: hello-go-cli
summary: Packages, variables, constants, types and basic input/output — the vocabulary of every Go program.

## Hello World
slug: hello-world
minutes: 5
objectives: Read every line of a Go program; Use fmt.Println and fmt.Printf; Explain what package main and func main do
takeaways: A program is package main plus func main; fmt.Println adds spaces and a newline; fmt.Printf uses verbs like %s and %d and needs an explicit \n

### Concept
A Go file starts with a **package** clause, then **imports**, then declarations. Execution begins at `main`.

- `fmt.Println(a, b)` prints operands separated by spaces, then a newline.
- `fmt.Printf("%s is %d\n", name, n)` formats with **verbs**: `%s` string, `%d` integer, `%v` any value, `%q` quoted string.

### Example
```go
package main

import "fmt"

func main() {
	name := "Gopher"
	fmt.Println("Hello,", name)
	fmt.Printf("%s has %d letters\n", name, len(name))
}
```

### Exercise
Print `Hello, Go!` with `Println`, then `7 wonders` using `Printf` with the number 7 and `%d`.
```text expect
Hello, Go!
7 wonders
```
```go solution
package main

import "fmt"

func main() {
	// BEGIN
	fmt.Println("Hello,", "Go!")
	fmt.Printf("%d wonders\n", 7)
	// END
}
```

### Check
Q: What does this print?
T: output
```go
fmt.Println("a", "b", 3)
```
- [x] a b 3
- [ ] ab3
- [ ] a, b, 3
- [ ] a b3
E: `Println` puts a space between every operand and ends with a newline.

Q: `Printf` adds a newline automatically.
T: tf
A: false
E: Only `Println` adds a newline. With `Printf` you write `\n` yourself.

## Packages
slug: packages
minutes: 5
objectives: Explain what a package is; Distinguish package main from library packages; Know that a directory holds one package
takeaways: Every Go file belongs to exactly one package; package main builds an executable; All files in a directory share one package name

### Concept
A **package** is a directory of `.go` files that share a `package name` line. Packages are Go's unit of code organisation and reuse.

- `package main` → the compiler produces an **executable**.
- Any other name (`package greet`) → a **library** other packages import.
- Names starting with a capital letter are visible outside the package (you will use this in Module 07).

### Example
Both files below are in the same directory, so they are the same package and can call each other's functions.
```go norun
// main.go
package main

func main() {
	sayHello() // defined in greet.go, same package
}

// greet.go
package main

import "fmt"

func sayHello() { fmt.Println("hello") }
```

### Exercise
The helper `greeting` lives in the same package as `main`. Call it and print its result.
```text expect
Hello from the package
```
```go solution
package main

import "fmt"

func greeting() string { return "Hello from the package" }

func main() {
	// BEGIN
	fmt.Println(greeting())
	// END
}
```

### Check
Q: Which package name makes `go build` produce an executable?
T: short
A: main
E: Only `package main` (with a `main` function) builds a runnable program.

Q: Two `.go` files in the same directory may declare different package names.
T: tf
A: false
E: All files in one directory must belong to the same package (test files may add a `_test` suffix).

## Imports
slug: imports
minutes: 5
objectives: Import one or many packages; Use an alias and the blank identifier; Understand why unused imports are errors
takeaways: Group imports in one parenthesised block; goimports and gofmt keep them sorted; Unused imports are compile errors

### Concept
Import paths name packages: `"fmt"`, `"strings"`, `"net/http"`. The last path element is the name you use in code.

```go norun
import (
	"fmt"
	str "strings" // alias
	_ "embed"     // side-effect only
)
```

Go refuses to compile with an **unused import**. That keeps dependencies honest. Your editor removes them for you on save.

### Example
```go
package main

import (
	"fmt"
	"strings"
)

func main() {
	fmt.Println(strings.ToUpper("go"))
	fmt.Println(strings.Repeat("=", 5))
}
```

### Exercise
Use the `strings` package: print `GOLEARN` (uppercase) and then `go-go-go` using `strings.Repeat("go-", 3)` trimmed with `strings.TrimSuffix`.
```text expect
GOLEARN
go-go-go
```
```go solution
package main

import (
	"fmt"
	"strings"
)

func main() {
	// BEGIN
	fmt.Println(strings.ToUpper("golearn"))
	fmt.Println(strings.TrimSuffix(strings.Repeat("go-", 3), "-"))
	// END
}
```

### Check
Q: What happens if you import a package and never use it?
T: mcq
- [ ] A warning is printed
- [x] The program does not compile
- [ ] The import is ignored silently
- [ ] The package is loaded lazily
E: Unused imports are compile errors in Go.

Q: Which line correctly imports `fmt` and `os`?
T: mcq
- [ ] import fmt, os
- [ ] import "fmt", "os"
- [x] import ("fmt"; "os")
- [ ] using fmt, os
E: A parenthesised import block accepts one path per line (or separated by semicolons).

## Variables
slug: variables
minutes: 6
objectives: Declare variables with var; Assign and reassign; Declare several variables at once
takeaways: var name type = value declares a variable; The type can be inferred from the value; Unused local variables are compile errors

### Concept
`var` declares a variable with a type, a value, or both.

```go norun
var age int          // zero value: 0
var name = "Ada"     // type inferred: string
var x, y int = 1, 2  // several at once
```

Variables are **statically typed**: once `age` is an `int`, it can only hold ints. And, like imports, a **local variable that is never used** is a compile error.

### Example
```go
package main

import "fmt"

var greeting = "Hello" // package-level

func main() {
	var name string = "Ada"
	var year int
	year = 1815
	fmt.Println(greeting, name, year)
}
```

### Exercise
Declare `var language = "Go"` and `var version int = 1` then print them as `Go 1`.
```text expect
Go 1
```
```go solution
package main

import "fmt"

func main() {
	// BEGIN
	var language = "Go"
	var version int = 1
	fmt.Println(language, version)
	// END
}
```

### Check
Q: What is the type of `x` after `var x = 3.5`?
T: mcq
- [ ] int
- [ ] float32
- [x] float64
- [ ] number
E: Untyped floating-point constants default to `float64`.

Q: Assigning a string to an `int` variable compiles if the string looks like a number.
T: tf
A: false
E: Go never converts types implicitly. Use `strconv.Atoi` to parse.

## Short Variable Declaration
slug: short-var-decl
minutes: 5
objectives: Use := inside functions; Know when := redeclares versus declares; Avoid accidental shadowing
takeaways: := declares and initialises, inside functions only; At least one new variable must appear on the left of :=; Inner-scope := creates a new variable that shadows the outer one

### Concept
Inside a function, `name := value` declares a variable and infers its type.

```go norun
count := 10          // int
pi := 3.14           // float64
ok, msg := true, "y" // two variables
```

**Shadowing:** in a nested block, `:=` creates a *new* variable that hides the outer one. It is legal, and a common source of bugs.

### Example
```go
package main

import "fmt"

func main() {
	x := 1
	if true {
		x := 2 // new variable, shadows outer x
		fmt.Println("inner", x)
	}
	fmt.Println("outer", x)
}
```

### Exercise
Use `:=` to create `a := 4` and `b := 6`, then print their sum, `10`.
```text expect
10
```
```go solution
package main

import "fmt"

func main() {
	// BEGIN
	a := 4
	b := 6
	fmt.Println(a + b)
	// END
}
```

### Check
Q: What does this print?
T: output
```go
x := 5
if x > 0 {
	x := 10
	_ = x
}
fmt.Println(x)
```
- [ ] 10
- [x] 5
- [ ] 15
- [ ] 0
E: The inner `x := 10` is a different variable; the outer `x` is unchanged.

Q: Where can `:=` be used?
T: mcq
- [ ] At package level
- [x] Only inside functions
- [ ] Only in main
- [ ] Anywhere a var is allowed
E: Package-level declarations must use `var`.

## Constants
slug: constants
minutes: 6
objectives: Declare constants with const; Use iota for enumerations; Explain untyped constants
takeaways: const values are fixed at compile time; iota generates incrementing values inside a const block; Untyped constants adapt to the context they are used in

### Concept
`const` declares values fixed at compile time. Only numbers, strings and booleans can be constants.

```go norun
const Pi = 3.14159
const (
	Small = iota // 0
	Medium       // 1
	Large        // 2
)
```

`iota` restarts at 0 in every `const` block and increments per line. Constants without an explicit type are **untyped** and take the type of the context, so `Pi` works with both `float32` and `float64`.

### Example
```go
package main

import "fmt"

type Weekday int

const (
	Sunday Weekday = iota
	Monday
	Tuesday
)

func main() {
	fmt.Println(Sunday, Monday, Tuesday)
}
```

### Exercise
Define a `const` block with `Low = iota`, `Mid`, `High` and print `High`, which is `2`.
```text expect
2
```
```go solution
package main

import "fmt"

// BEGIN
const (
	Low = iota
	Mid
	High
)

// END

func main() {
	fmt.Println(High)
}
```

### Check
Q: What does `iota` produce for the third name in a const block?
T: short
A: 2
E: `iota` starts at 0 and increases by one for each constant specification.

Q: You can declare a slice as a constant.
T: tf
A: false
E: Only basic types (numbers, strings, booleans) can be constants.

## Data Types
slug: data-types
minutes: 7
objectives: Name Go's basic types; Choose between int sizes and float64; Recognise rune and byte
takeaways: Basic types: bool, string, int/uint variants, float32/64, byte, rune; int is the default integer type; string is immutable UTF-8 text

### Concept
| Kind | Types |
| --- | --- |
| Boolean | `bool` |
| Text | `string`, `byte` (uint8), `rune` (int32) |
| Integers | `int`, `int8`…`int64`, `uint`, `uint8`…`uint64` |
| Floats | `float32`, `float64` |
| Complex | `complex64`, `complex128` |

Use plain `int` and `float64` unless you have a reason (file formats, memory layout, APIs) to pick a sized type. `%T` prints a value's type.

### Example
```go
package main

import "fmt"

func main() {
	var (
		ok   bool    = true
		n    int     = 42
		f    float64 = 2.5
		s    string  = "go"
		r    rune    = 'G'
	)
	fmt.Printf("%T %T %T %T %T\n", ok, n, f, s, r)
}
```

### Exercise
Print the type of the untyped literal `3.0` using `%T`, then the type of `'x'`. Use one `Printf` for each, ending with a newline.
```text expect
float64
int32
```
```go solution
package main

import "fmt"

func main() {
	// BEGIN
	fmt.Printf("%T\n", 3.0)
	fmt.Printf("%T\n", 'x')
	// END
}
```

### Check
Q: What is a `rune`?
T: mcq
- [ ] A 16-bit integer
- [ ] An alias for string
- [x] An alias for int32 representing a Unicode code point
- [ ] A single byte
E: `rune` is `int32` and holds one Unicode code point.

Q: What does this print?
T: output
```go
fmt.Printf("%T", 'a')
```
- [ ] byte
- [ ] rune
- [x] int32
- [ ] char
E: `%T` reports the underlying type name, `int32`, because `rune` is only an alias.

## Zero Values
slug: zero-values
minutes: 5
objectives: State the zero value of each basic type; Rely on zero values instead of initialising everything; Avoid nil surprises
takeaways: Every variable has a usable zero value; Numbers 0, bool false, string empty, pointers/slices/maps nil; Design types so their zero value is useful

### Concept
Go never leaves memory uninitialised. A variable declared without a value gets its type's **zero value**.

| Type | Zero value |
| --- | --- |
| numbers | `0` |
| `bool` | `false` |
| `string` | `""` |
| pointer, slice, map, channel, func, interface | `nil` |

A nil slice is safe to `append` to and `len` returns 0. A nil map can be read but not written. Idiomatic Go designs types so the zero value is ready to use (`sync.Mutex`, `bytes.Buffer`).

### Example
```go
package main

import "fmt"

func main() {
	var n int
	var s string
	var b bool
	var p *int
	var xs []int
	fmt.Printf("%d %q %t %v %v %d\n", n, s, b, p, xs, len(xs))
}
```

### Exercise
Declare `var count int` and `var label string` without values, then print `count`, then `label` in quotes using `%q`, on one line: `0 ""`.
```text expect
0 ""
```
```go solution
package main

import "fmt"

func main() {
	// BEGIN
	var count int
	var label string
	fmt.Printf("%d %q\n", count, label)
	// END
}
```

### Check
Q: What is the zero value of a `string`?
T: mcq
- [ ] nil
- [x] "" (the empty string)
- [ ] "0"
- [ ] undefined
E: Strings are never nil in Go; the zero value is the empty string.

Q: What does this print?
T: output
```go
var xs []int
xs = append(xs, 1)
fmt.Println(len(xs), xs == nil)
```
- [ ] 0 true
- [x] 1 false
- [ ] 1 true
- [ ] 0 false
E: Appending to a nil slice allocates a backing array, so the result is non-nil with length 1.

## Type Conversion
slug: type-conversion
minutes: 6
objectives: Convert between numeric types explicitly; Parse strings with strconv; Avoid the string(int) trap
takeaways: Go has no implicit conversions — write T(x); Use strconv.Itoa/Atoi to convert between ints and strings; Converting float to int truncates toward zero

### Concept
Every conversion is explicit: `float64(n)`, `int(f)`, `[]byte(s)`.

- Float to int **truncates**: `int(3.9)` is `3`.
- Number ↔ text uses `strconv`: `strconv.Itoa(42)` → `"42"`, `strconv.Atoi("42")` → `42, nil`.
- `string(65)` yields `"A"` (a code point), **not** `"65"`. `go vet` warns about this.

### Example
```go
package main

import (
	"fmt"
	"strconv"
)

func main() {
	n := 7
	f := float64(n) / 2
	fmt.Println(f, int(f))
	s := strconv.Itoa(n)
	back, err := strconv.Atoi(s + "0")
	fmt.Println(s, back, err)
}
```

### Exercise
Convert the int `total := 7` and `count := 2` to `float64`, divide, and print `3.5`.
```text expect
3.5
```
```go solution
package main

import "fmt"

func main() {
	total, count := 7, 2
	// BEGIN
	fmt.Println(float64(total) / float64(count))
	// END
}
```

### Check
Q: What does this print?
T: output
```go
f := 9.99
fmt.Println(int(f))
```
- [ ] 10
- [x] 9
- [ ] 9.99
- [ ] 9.0
E: Converting a float to an integer truncates the fractional part.

Q: Which call converts the string "42" into the int 42?
T: mcq
- [ ] int("42")
- [ ] strconv.Itoa("42")
- [x] strconv.Atoi("42")
- [ ] string.ToInt("42")
E: `Atoi` (ASCII to integer) returns the number and an error.

## Basic Input and Output
slug: basic-io
minutes: 8
objectives: Print with the fmt family; Read a line from standard input with bufio; Write to stderr
takeaways: fmt.Print/Println/Printf write to stdout; bufio.Scanner reads lines from os.Stdin; Diagnostics belong on os.Stderr

### Concept
Output: `fmt.Print`, `fmt.Println`, `fmt.Printf`, and `fmt.Fprintln(os.Stderr, …)` for errors.

Input: wrap `os.Stdin` in a `bufio.Scanner`.

```go norun
scanner := bufio.NewScanner(os.Stdin)
for scanner.Scan() {
	line := scanner.Text()
	fmt.Println("you typed:", line)
}
```

`fmt.Sscanf` and `fmt.Sscan` parse values out of a string, useful for quick parsing.

### Example
```go
package main

import (
	"bufio"
	"fmt"
	"strings"
)

func main() {
	input := "alpha\nbeta\n"
	sc := bufio.NewScanner(strings.NewReader(input))
	for sc.Scan() {
		fmt.Println("read:", sc.Text())
	}
}
```

### Exercise
Scan the lines in `input` and print each one uppercased. Use `strings.ToUpper`.
```text expect
ONE
TWO
```
```go solution
package main

import (
	"bufio"
	"fmt"
	"strings"
)

func main() {
	input := "one\ntwo\n"
	sc := bufio.NewScanner(strings.NewReader(input))
	// BEGIN
	for sc.Scan() {
		fmt.Println(strings.ToUpper(sc.Text()))
	}
	// END
}
```

### Check
Q: Which package reads lines from standard input efficiently?
T: mcq
- [ ] io/ioutil only
- [x] bufio (Scanner or Reader) with os.Stdin
- [ ] fmt.Println
- [ ] strings
E: `bufio.Scanner` wraps any `io.Reader`, including `os.Stdin`.

Q: Error messages should be written to `os.Stderr`, not `os.Stdout`.
T: tf
A: true
E: Keeping errors on stderr lets users pipe or redirect normal output without mixing in diagnostics.
