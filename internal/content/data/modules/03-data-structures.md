# Data Structures
id: data-structures
number: 03
track: core
paths: beginner, pro
skill: data-structures
requires: control-flow
project: contact-manager-cli
summary: Arrays, slices, maps and the truth about strings, bytes and runes.

## Arrays
slug: arrays
minutes: 5
objectives: Declare fixed-size arrays; Explain that arrays are values; Know when arrays (rather than slices) are appropriate
takeaways: An array's length is part of its type; Arrays are copied on assignment and when passed to functions; Slices are used far more often than arrays

### Concept
An array has a **fixed length** that is part of its type: `[3]int` and `[4]int` are different types.

```go norun
var a [3]int              // [0 0 0]
b := [3]string{"x", "y", "z"}
c := [...]int{1, 2, 3, 4} // length inferred: [4]int
```
Arrays are **values**: `d := a` copies all elements, and passing an array to a function copies it. Use arrays for small fixed-size data (a `[16]byte` UUID, a `[4]int` vector) and slices for everything else.

### Example
```go
package main

import "fmt"

func main() {
	a := [3]int{1, 2, 3}
	b := a // copy
	b[0] = 99
	fmt.Println(a, b, len(a))
}
```

### Exercise
Create the array `[5]int{1, 2, 3, 4, 5}` and print the sum of its elements.
```text expect
15
```
```go solution
package main

import "fmt"

func main() {
	// BEGIN
	nums := [5]int{1, 2, 3, 4, 5}
	sum := 0
	for _, n := range nums {
		sum += n
	}
	fmt.Println(sum)
	// END
}
```

### Check
Q: What does this print?
T: output
```go
a := [2]int{1, 2}
b := a
b[0] = 9
fmt.Println(a[0], b[0])
```
- [ ] 9 9
- [x] 1 9
- [ ] 9 1
- [ ] 1 1
E: Arrays are copied on assignment, so changing `b` does not affect `a`.

Q: `[3]int` and `[4]int` are the same type.
T: tf
A: false
E: The length is part of an array's type.

## Slices
slug: slices
minutes: 8
objectives: Create slices with literals and make; Slice existing slices with s[i:j]; Understand a slice as a view onto an array
takeaways: A slice is a view (pointer, length, capacity) onto an array; s[i:j] shares memory with s; Prefer slices to arrays in APIs

### Concept
A **slice** describes a window onto an underlying array: a pointer, a length and a capacity.

```go norun
s := []int{10, 20, 30, 40}
t := s[1:3]        // [20 30] — shares memory with s
u := make([]int, 3)    // len 3, cap 3, zeros
v := make([]int, 0, 10) // len 0, cap 10
```
Because `t` shares memory with `s`, `t[0] = 99` changes `s[1]`. `s[i:j]` includes `i` and excludes `j`; either may be omitted (`s[:2]`, `s[2:]`).

### Example
```go
package main

import "fmt"

func main() {
	s := []int{10, 20, 30, 40}
	t := s[1:3]
	t[0] = 99
	fmt.Println(s, t)
}
```

### Exercise
Given `s := []int{1, 2, 3, 4, 5}`, print the middle three elements `s[1:4]`.
```text expect
[2 3 4]
```
```go solution
package main

import "fmt"

func main() {
	s := []int{1, 2, 3, 4, 5}
	// BEGIN
	fmt.Println(s[1:4])
	// END
}
```

### Check
Q: What does this print?
T: output
```go
s := []int{1, 2, 3}
t := s[:2]
t[0] = 9
fmt.Println(s)
```
- [ ] [1 2 3]
- [x] [9 2 3]
- [ ] [9 2]
- [ ] [1 2]
E: `t` and `s` share the same backing array, so the change is visible through `s`.

Q: Which expression gives the last two elements of slice `s` of length 5?
T: mcq
- [ ] s[-2:]
- [ ] s[2:2]
- [x] s[3:]
- [ ] s[:2]
E: Go has no negative indexes; `s[3:]` starts at index 3 and runs to the end.

## Slice Length
slug: slice-length
minutes: 4
objectives: Use len to count elements; Handle empty and nil slices; Iterate safely without out-of-range panics
takeaways: len(s) is the number of elements; Indexing beyond len panics at run time; A nil slice and an empty slice both have length 0

### Concept
`len(s)` returns the number of elements you can index. Valid indexes are `0` to `len(s)-1`; anything else panics with `index out of range`.

```go norun
var nilSlice []int   // nil, len 0
empty := []int{}     // non-nil, len 0
```
Both behave the same with `len`, `range` and `append`. Check `len(s) == 0` rather than `s == nil` unless you must distinguish them (for example when encoding JSON).

### Example
```go
package main

import "fmt"

func first(xs []int) (int, bool) {
	if len(xs) == 0 {
		return 0, false
	}
	return xs[0], true
}

func main() {
	fmt.Println(first(nil))
	fmt.Println(first([]int{7, 8}))
}
```

### Exercise
Write `last(xs []int) (int, bool)` returning the last element and `false` when the slice is empty.
```text expect
0 false
9 true
```
```go solution
package main

import "fmt"

// BEGIN
func last(xs []int) (int, bool) {
	if len(xs) == 0 {
		return 0, false
	}
	return xs[len(xs)-1], true
}

// END

func main() {
	fmt.Println(last(nil))
	fmt.Println(last([]int{4, 9}))
}
```

### Check
Q: What happens when you access `xs[len(xs)]`?
T: mcq
- [ ] It returns the zero value
- [x] The program panics with index out of range
- [ ] It wraps around
- [ ] It returns nil
E: Slice indexes are bounds-checked at run time.

Q: Is `len` of a nil slice 0?
T: tf
A: true
E: A nil slice has length 0 and is safe to range over and append to.

## Slice Capacity
slug: slice-capacity
minutes: 6
objectives: Read a slice's capacity with cap; Preallocate with make([]T, 0, n); Predict when append reallocates
takeaways: cap(s) is how many elements fit before reallocation; Preallocating avoids repeated copying; Slicing from the start keeps the remaining capacity

### Concept
The **capacity** is the size of the underlying array from the slice's start. `append` writes in place while `len < cap`, otherwise it allocates a bigger array and copies.

```go norun
s := make([]int, 0, 4)   // len 0, cap 4
s = append(s, 1, 2)      // len 2, cap 4 — no allocation
t := s[:1]               // len 1, cap 4 — remaining capacity is kept
```
If you know how many elements you will add, `make([]T, 0, n)` avoids repeated growth.

### Example
```go
package main

import "fmt"

func main() {
	s := make([]int, 2, 5)
	fmt.Println(len(s), cap(s))
	t := s[1:]
	fmt.Println(len(t), cap(t))
}
```

### Exercise
Create a slice with length 0 and capacity 3 using `make`, then print its `len` and `cap`.
```text expect
0 3
```
```go solution
package main

import "fmt"

func main() {
	// BEGIN
	s := make([]string, 0, 3)
	fmt.Println(len(s), cap(s))
	// END
}
```

### Check
Q: What does this print?
T: output
```go
s := make([]int, 2, 6)
t := s[1:3]
fmt.Println(len(t), cap(t))
```
- [ ] 2 6
- [x] 2 5
- [ ] 2 4
- [ ] 1 5
E: `t` starts at index 1 of an array of capacity 6, so its capacity is 5; its length is 3-1 = 2.

Q: Why preallocate with `make([]T, 0, n)`?
T: mcq
- [ ] To make the slice immutable
- [x] To avoid repeated reallocation while appending
- [ ] To save the zero-value initialisation
- [ ] To prevent index panics
E: Extra capacity means `append` doesn't need to grow and copy the array.

## Append
slug: append
minutes: 7
objectives: Grow slices with append; Always assign the result of append; Append one slice to another with ...
takeaways: append returns a new slice header — assign it; append may or may not reallocate; append(a, b...) concatenates

### Concept
```go norun
s = append(s, 4)          // one element
s = append(s, 5, 6)       // several
s = append(s, other...)   // another slice
```
Always assign the result (`s = append(s, x)`): the returned slice may point to a new array. When capacity is exceeded Go allocates a larger array and copies, so old and new slices stop sharing memory.

### Example
```go
package main

import "fmt"

func main() {
	var s []int
	for i := 1; i <= 5; i++ {
		s = append(s, i*i)
	}
	fmt.Println(s, len(s))
	s = append(s, []int{100, 200}...)
	fmt.Println(s)
}
```

### Exercise
Write `evens(n int) []int` returning the even numbers from 0 up to and including `n` using `append`.
```text expect
[0 2 4 6]
```
```go solution
package main

import "fmt"

// BEGIN
func evens(n int) []int {
	var out []int
	for i := 0; i <= n; i += 2 {
		out = append(out, i)
	}
	return out
}

// END

func main() {
	fmt.Println(evens(7))
}
```

### Check
Q: What is wrong with this line?
T: debug
```go
append(nums, 42)
```
- [ ] append needs a third argument
- [x] The result is discarded; it must be assigned: nums = append(nums, 42)
- [ ] append only works on arrays
- [ ] Nothing, it modifies nums in place
E: `append` returns the updated slice; ignoring it (the compiler even complains) loses the change.

Q: What does this print?
T: output
```go
a := []int{1, 2}
b := append(a, 3)
fmt.Println(len(a), len(b))
```
- [ ] 3 3
- [x] 2 3
- [ ] 2 2
- [ ] 3 2
E: `a` keeps its own length of 2; `b` is a new slice header of length 3.

## Copy
slug: copy
minutes: 5
objectives: Copy elements with copy; Avoid aliasing bugs by cloning slices; Use slices.Clone
takeaways: copy(dst, src) copies min(len(dst), len(src)) elements; Assigning a slice never copies the data; slices.Clone makes an independent copy

### Concept
`dst := src` copies only the header. To get an independent copy:

```go norun
dst := make([]int, len(src))
n := copy(dst, src) // number of elements copied
// or, Go 1.21+:
dst = slices.Clone(src)
```
`copy` copies `min(len(dst), len(src))` elements and handles overlapping slices correctly. Use it before mutating a slice you received when the caller must not see the change.

### Example
```go
package main

import "fmt"

func main() {
	src := []int{1, 2, 3}
	dst := make([]int, len(src))
	copy(dst, src)
	dst[0] = 99
	fmt.Println(src, dst)
}
```

### Exercise
Write `clone(xs []int) []int` returning an independent copy using `make` and `copy`.
```text expect
[1 2 3] [9 2 3]
```
```go solution
package main

import "fmt"

// BEGIN
func clone(xs []int) []int {
	out := make([]int, len(xs))
	copy(out, xs)
	return out
}

// END

func main() {
	a := []int{1, 2, 3}
	b := clone(a)
	b[0] = 9
	fmt.Println(a, b)
}
```

### Check
Q: What does this print?
T: output
```go
dst := make([]int, 2)
n := copy(dst, []int{7, 8, 9})
fmt.Println(n, dst)
```
- [ ] 3 [7 8 9]
- [x] 2 [7 8]
- [ ] 2 [8 9]
- [ ] 0 [0 0]
E: `copy` copies only as many elements as fit in `dst`.

Q: `b := a` for slices creates an independent copy of the elements.
T: tf
A: false
E: Only the slice header is copied; both slices refer to the same backing array.

## Maps
slug: maps
minutes: 7
objectives: Create maps with literals and make; Read, write and delete keys; Use the comma-ok idiom
takeaways: A map is an unordered hash table: map[K]V; Reading a missing key yields the zero value; Use v, ok := m[k] to distinguish missing from zero

### Concept
```go norun
ages := map[string]int{"ada": 36, "linus": 28}
m := make(map[string]bool)
m["go"] = true
delete(m, "go")
v, ok := ages["grace"] // 0, false
```
Keys must be **comparable** (strings, numbers, bools, arrays, structs of those; not slices, maps or funcs). Reading a missing key returns the zero value, so use the comma-ok form when it matters. A **nil map** can be read but panics on write — always `make` it.

### Example
```go
package main

import "fmt"

func main() {
	stock := map[string]int{"apple": 3}
	stock["pear"] = 0
	if n, ok := stock["pear"]; ok {
		fmt.Println("pear:", n)
	}
	_, ok := stock["kiwi"]
	fmt.Println(ok, len(stock))
}
```

### Exercise
Count how many times each word appears in `words` using a `map[string]int`, then print the count for `"go"`.
```text expect
3
```
```go solution
package main

import "fmt"

func main() {
	words := []string{"go", "is", "go", "fun", "go"}
	// BEGIN
	counts := make(map[string]int)
	for _, w := range words {
		counts[w]++
	}
	fmt.Println(counts["go"])
	// END
}
```

### Check
Q: What happens when you write to a nil map?
T: mcq
- [ ] It is created automatically
- [ ] Nothing
- [x] The program panics
- [ ] Compile error
E: Assignment to an entry in a nil map panics; create maps with `make` or a literal.

Q: What does this print?
T: output
```go
m := map[string]int{"a": 1}
v, ok := m["b"]
fmt.Println(v, ok)
```
- [ ] 1 true
- [ ] <nil> false
- [x] 0 false
- [ ] 0 true
E: A missing key yields the zero value and `ok == false`.

## Map Operations
slug: map-operations
minutes: 7
objectives: Iterate a map and understand random order; Sort keys for deterministic output; Use maps as sets and for grouping
takeaways: Map iteration order is randomised; Collect and sort keys when you need order; map[T]struct{} is an idiomatic set

### Concept
`for k, v := range m` visits entries in an **unspecified, deliberately randomised** order. To print deterministically, collect the keys and sort them:

```go norun
keys := make([]string, 0, len(m))
for k := range m {
	keys = append(keys, k)
}
sort.Strings(keys)
```
Since Go 1.21: `slices.Sorted(maps.Keys(m))`. A set is a map with empty-struct values: `seen := map[string]struct{}{}`. Deleting during iteration is safe.

### Example
```go
package main

import (
	"fmt"
	"sort"
)

func main() {
	m := map[string]int{"b": 2, "a": 1, "c": 3}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Println(k, m[k])
	}
}
```

### Exercise
Return the distinct values of `xs` in first-seen order using a set (`map[int]struct{}`).
```text expect
[3 1 2]
```
```go solution
package main

import "fmt"

// BEGIN
func unique(xs []int) []int {
	seen := make(map[int]struct{})
	var out []int
	for _, x := range xs {
		if _, ok := seen[x]; ok {
			continue
		}
		seen[x] = struct{}{}
		out = append(out, x)
	}
	return out
}

// END

func main() {
	fmt.Println(unique([]int{3, 1, 3, 2, 1}))
}
```

### Check
Q: Is the iteration order of a Go map guaranteed?
T: tf
A: false
E: It is intentionally randomised; sort the keys if you need a stable order.

Q: Which type is the idiomatic way to model a set of strings?
T: mcq
- [ ] []string
- [ ] map[string]bool only
- [x] map[string]struct{}
- [ ] set[string]
E: `struct{}` occupies no memory, so it makes an efficient set (`map[string]bool` also works and is often clearer).

## Strings
slug: strings
minutes: 7
objectives: Treat strings as immutable byte sequences; Use the strings package for common tasks; Build strings efficiently
takeaways: Strings are immutable; s[i] is a byte, not a character; Use strings.Builder to concatenate in loops

### Concept
A Go `string` is an **immutable** sequence of bytes (usually UTF-8 text). `len(s)` counts **bytes**. The `strings` package covers most needs: `Contains`, `Split`, `Join`, `Fields`, `TrimSpace`, `HasPrefix`, `Replace`, `ToUpper`.

Repeated `+=` in a loop copies every time; use `strings.Builder`:

```go norun
var b strings.Builder
for i := 0; i < 3; i++ {
	fmt.Fprintf(&b, "%d,", i)
}
s := b.String()
```

### Example
```go
package main

import (
	"fmt"
	"strings"
)

func main() {
	s := "  go is fun  "
	words := strings.Fields(s)
	fmt.Println(len(words), strings.Join(words, "-"))
	fmt.Println(strings.Contains(s, "fun"), strings.ToUpper(strings.TrimSpace(s)))
}
```

### Exercise
Write `initials(name string) string` that returns the first letter of each space-separated word, uppercased. `"ada byron lovelace"` → `"ABL"`.
```text expect
ABL
```
```go solution
package main

import (
	"fmt"
	"strings"
)

// BEGIN
func initials(name string) string {
	var b strings.Builder
	for _, w := range strings.Fields(name) {
		b.WriteString(strings.ToUpper(w[:1]))
	}
	return b.String()
}

// END

func main() {
	fmt.Println(initials("ada byron lovelace"))
}
```

### Check
Q: What does `len("héllo")` return?
T: mcq
- [ ] 5
- [x] 6
- [ ] 4
- [ ] 7
E: `é` takes 2 bytes in UTF-8, and `len` counts bytes.

Q: What does this print?
T: output
```go
fmt.Println(strings.Split("a,b,c", ","))
```
- [ ] a b c
- [x] [a b c]
- [ ] [a,b,c]
- [ ] ["a" "b" "c"]
E: `Split` returns a `[]string`, which `Println` prints in brackets separated by spaces.

## Bytes
slug: bytes
minutes: 6
objectives: Convert between string and []byte; Mutate text through a byte slice; Use the bytes package
takeaways: []byte(s) copies the string into a mutable slice; string(b) copies back; The bytes package mirrors strings for []byte

### Concept
Strings are immutable; a `[]byte` is mutable. Converting **copies** the data:

```go norun
b := []byte("gopher")
b[0] = 'G'
s := string(b) // "Gopher"
```
The `bytes` package has the same helpers as `strings` (`bytes.Contains`, `bytes.Split`, `bytes.ToUpper`) and `bytes.Buffer` for accumulating data. Bytes are the right unit for binary data and I/O; runes are for text characters.

### Example
```go
package main

import (
	"bytes"
	"fmt"
)

func main() {
	b := []byte("hello")
	b[0] = 'H'
	fmt.Println(string(b), b[0])
	fmt.Println(bytes.Contains(b, []byte("ell")))
}
```

### Exercise
Reverse the ASCII string `"golang"` by swapping bytes in a `[]byte` in place and print the result.
```text expect
gnalog
```
```go solution
package main

import "fmt"

func main() {
	b := []byte("golang")
	// BEGIN
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
	// END
	fmt.Println(string(b))
}
```

### Check
Q: What is `[]byte("A")[0]`?
T: mcq
- [ ] "A"
- [x] 65
- [ ] 'a'
- [ ] 1
E: A byte is a `uint8`; the character `A` has the value 65.

Q: `[]byte(s)` shares memory with the string `s`.
T: tf
A: false
E: The conversion copies, because strings are immutable and byte slices are not.

## Runes
slug: runes
minutes: 7
objectives: Explain what a rune is; Iterate a string by rune with range; Count characters instead of bytes
takeaways: A rune is a Unicode code point (int32); range over a string decodes runes; Use utf8.RuneCountInString or []rune(s) to count characters

### Concept
A **rune** is one Unicode code point. `for i, r := range s` walks the string rune by rune, yielding the byte offset `i` and the rune `r`.

```go norun
s := "héllo"
len(s)                      // 6 bytes
utf8.RuneCountInString(s)   // 5 runes
r := []rune(s)              // convert to modify by character
```
Indexing `s[1]` gives a **byte**, which may be half of a character. Use runes whenever a "character" matters (reversing, truncating, counting).

### Example
```go
package main

import "fmt"

func main() {
	for i, r := range "héy" {
		fmt.Println(i, string(r))
	}
}
```

### Exercise
Reverse a string by runes so that non-ASCII characters survive: `"héllo"` → `"olléh"`.
```text expect
olléh
```
```go solution
package main

import "fmt"

// BEGIN
func reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

// END

func main() {
	fmt.Println(reverse("héllo"))
}
```

### Check
Q: What does this print?
T: output
```go
for i, r := range "aé" {
	fmt.Print(i, ":", string(r), ";")
}
```
- [ ] 0:a;2:é;
- [x] 0:a;1:é;
- [ ] 0:a;1:e;
- [ ] a;é;
E: `range` yields the byte offset of each rune. `a` starts at 0 and `é` at 1.

Q: Which converts a string to a slice you can modify by character?
T: short
A: []rune(s)
E: `[]rune(s)` decodes the string into code points that can be edited independently.

## UTF-8
slug: utf8
minutes: 6
objectives: Describe how UTF-8 encodes code points; Validate and decode UTF-8 with unicode/utf8; Avoid slicing strings mid-character
takeaways: UTF-8 uses 1–4 bytes per code point; ASCII is a subset of UTF-8; Slice strings at rune boundaries, not arbitrary byte offsets

### Concept
UTF-8 encodes each Unicode code point in **1 to 4 bytes**: ASCII takes 1, most European letters 2, many Asian characters 3, emoji 4. Go source files and strings are UTF-8 by convention.

The `unicode/utf8` package helps: `utf8.RuneLen(r)`, `utf8.ValidString(s)`, `utf8.DecodeRuneInString(s)`.

```go norun
s := "日本"
s[:1] // invalid: cuts the first character in half
```
Truncate text by runes, not bytes.

### Example
```go
package main

import (
	"fmt"
	"unicode/utf8"
)

func main() {
	for _, s := range []string{"a", "é", "日", "🙂"} {
		fmt.Println(s, len(s), utf8.RuneCountInString(s))
	}
}
```

### Exercise
Write `truncate(s string, n int) string` that keeps at most `n` **runes**.
```text expect
héll
日本
```
```go solution
package main

import "fmt"

// BEGIN
func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}

// END

func main() {
	fmt.Println(truncate("héllo", 4))
	fmt.Println(truncate("日本語", 2))
}
```

### Check
Q: How many bytes does the emoji 🙂 take in UTF-8?
T: short
A: 4
E: Code points above U+FFFF need four bytes.

Q: What does this print?
T: output
```go
fmt.Println(len("日本"), utf8.RuneCountInString("日本"))
```
- [ ] 2 2
- [ ] 6 6
- [x] 6 2
- [ ] 2 6
E: Each of the two characters takes 3 bytes, so 6 bytes but 2 runes.
