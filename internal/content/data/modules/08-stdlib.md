# Standard Library
id: stdlib
number: 08
track: core
paths: beginner, pro
skill: stdlib
requires: modules
summary: The packages you will use every day: fmt, strings, strconv, sort, time, regexp, context, sync and more.

## fmt
slug: fmt
minutes: 8
challenges: fmt-table
objectives: Choose the right print function and verb; Format widths, precision and alignment; Use Sprintf and Fprintf
takeaways: Print* writes to stdout, Sprint* returns a string, Fprint* writes to an io.Writer; %v %+v %#v %T %q %x cover most debugging needs; Width and precision (%-8s, %6.2f) align tabular output

### Concept
**Explanation.** `fmt` formats and prints values. Three families: `Print`/`Println`/`Printf` (stdout), `Sprint*` (return a string), `Fprint*` (any `io.Writer`, e.g. `os.Stderr`).

**Common verbs:** `%v` value, `%+v` with field names, `%#v` Go syntax, `%T` type, `%d` int, `%s` string, `%q` quoted, `%f`/`%.2f` float, `%x` hex, `%t` bool, `%p` pointer.

**Layout:** `%-8s` left-align in 8 columns, `%6.2f` width 6 with 2 decimals, `%05d` zero-pad.

**Use cases:** logs, CLI tables, building error messages (`fmt.Errorf`), implementing `String() string` on your types.

### Example
```go
package main

import "fmt"

type Point struct{ X, Y int }

func (p Point) String() string { return fmt.Sprintf("(%d,%d)", p.X, p.Y) }

func main() {
	fmt.Printf("|%-6s|%6.2f|%05d|\n", "go", 3.14159, 42)
	fmt.Println(Point{1, 2}, fmt.Sprintf("%q", "hi"))
}
```

### Exercise
Print `Name      Qty` header and one row for `apple` with quantity `3`, using `%-8s` for the name and `%3d` for the quantity, so the output is exactly as expected.
```text expect
Name      Qty
apple       3
```
```go solution
package main

import "fmt"

func main() {
	// BEGIN
	fmt.Printf("%-8s  %3s\n", "Name", "Qty")
	fmt.Printf("%-8s  %3d\n", "apple", 3)
	// END
}
```

### Check
Q: Which verb prints a value's type?
T: short
A: %T
E: `%T` prints the Go type of the operand.

Q: What does this print?
T: output
```go
fmt.Printf("%5.1f|%-4d|", 3.14159, 7)
```
- [ ] 3.14|7   |
- [x]   3.1|7   |
- [ ] 3.1  |7|
- [ ] 3.1|   7|
E: `%5.1f` is width 5 with one decimal (right-aligned) and `%-4d` pads on the right.

## strings
slug: strings
minutes: 8
challenges: title-case
objectives: Search, split, join and replace text; Trim and normalise input; Use strings.Builder and strings.NewReplacer
takeaways: The strings package is the toolbox for text; Cut, Fields, Split, Join, TrimSpace, Contains and Replace cover most tasks; Use strings.Builder to build large strings efficiently

### Concept
**Explanation.** All functions treat strings as UTF-8 and never modify the input (strings are immutable).

**Common use cases:** `Contains/HasPrefix/HasSuffix/Index` searching; `Split/Fields/Cut` parsing; `Join/Repeat/Replace` building; `TrimSpace/Trim/TrimPrefix` cleaning; `ToUpper/ToLower/EqualFold` case handling; `Builder` for concatenation loops.

`strings.Cut(s, "=")` splits on the first separator and reports whether it was found — the cleanest way to parse `key=value`.

### Example
```go
package main

import (
	"fmt"
	"strings"
)

func main() {
	key, val, ok := strings.Cut("port=8080", "=")
	fmt.Println(key, val, ok)
	fmt.Println(strings.Fields("  a  b\tc\n"), strings.EqualFold("Go", "GO"))
	r := strings.NewReplacer("<", "&lt;", ">", "&gt;")
	fmt.Println(r.Replace("<b>hi</b>"))
}
```

### Exercise
Write `parse(line string) (key, value string)` splitting on the first `=` and trimming spaces around both parts.
```text expect
host|db.local
```
```go solution
package main

import (
	"fmt"
	"strings"
)

// BEGIN
func parse(line string) (key, value string) {
	k, v, _ := strings.Cut(line, "=")
	return strings.TrimSpace(k), strings.TrimSpace(v)
}

// END

func main() {
	k, v := parse("  host = db.local ")
	fmt.Println(k + "|" + v)
}
```

### Check
Q: What does `strings.Fields("  a  b ")` return?
T: mcq
- [ ] ["", "", "a", "", "b", ""]
- [x] ["a" "b"]
- [ ] "a b"
- [ ] ["a  b"]
E: `Fields` splits around runs of whitespace and drops empty fields.

Q: What does this print?
T: output
```go
before, after, found := strings.Cut("a=b=c", "=")
fmt.Println(before, after, found)
```
- [ ] a b true
- [x] a b=c true
- [ ] a=b c true
- [ ] a=b=c  false
E: `Cut` splits at the first separator only.

## strconv
slug: strconv
minutes: 7
challenges: sum-numeric-strings
objectives: Parse ints, floats and bools with errors; Format numbers as strings; Handle *strconv.NumError
takeaways: Atoi/ParseInt/ParseFloat/ParseBool parse text and return errors; Itoa/FormatInt/FormatFloat convert back; Quote and Unquote handle string literals

### Concept
**Explanation.** `strconv` converts between strings and basic types. Parsing always returns an `error`; bad input never silently becomes 0.

**Use cases:** reading numbers from flags, environment variables, CSV/JSON text, query strings.

`ParseInt(s, base, bitSize)` handles other bases (`16`) and sizes (`64`). Errors are `*strconv.NumError` with `Func`, `Num`, and `Err` (`strconv.ErrSyntax` or `strconv.ErrRange`).

### Example
```go
package main

import (
	"errors"
	"fmt"
	"strconv"
)

func main() {
	n, err := strconv.ParseInt("ff", 16, 64)
	fmt.Println(n, err)
	_, err = strconv.Atoi("12a")
	fmt.Println(err, errors.Is(err, strconv.ErrSyntax))
	fmt.Println(strconv.FormatFloat(3.14159, 'f', 2, 64), strconv.Quote("hi\n"))
}
```

### Exercise
Write `parseBool(s string) string` returning `yes` if `strconv.ParseBool` says true, `no` if false and `invalid` on error.
```text expect
yes no invalid
```
```go solution
package main

import (
	"fmt"
	"strconv"
)

// BEGIN
func parseBool(s string) string {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return "invalid"
	}
	if b {
		return "yes"
	}
	return "no"
}

// END

func main() {
	fmt.Println(parseBool("true"), parseBool("0"), parseBool("maybe"))
}
```

### Check
Q: What does `strconv.Atoi("42abc")` return?
T: mcq
- [ ] 42, nil
- [ ] 0, nil
- [x] 0 and a non-nil error
- [ ] It panics
E: Parsing failures are reported as errors — the result is 0 alongside the error.

Q: Which function parses hexadecimal `"ff"` into an integer?
T: mcq
- [ ] strconv.Atoi("ff")
- [x] strconv.ParseInt("ff", 16, 64)
- [ ] strconv.Hex("ff")
- [ ] strconv.Itoa("ff")
E: `ParseInt` takes the base as its second argument.

## bytes
slug: bytes
minutes: 6
challenges: byte-buffer-lines
objectives: Use bytes.Buffer as a growable buffer; Apply strings-like helpers to []byte; Read and write through the io interfaces
takeaways: bytes mirrors strings for []byte; bytes.Buffer is an io.Reader and io.Writer with a useful zero value; Use it to build output or feed readers in tests

### Concept
**Explanation.** The `bytes` package has the same helpers as `strings` (`Contains`, `Split`, `TrimSpace`, `Equal`) but for `[]byte`, plus `bytes.Buffer` — a variable-size buffer implementing `io.Reader`, `io.Writer`, and `io.ByteWriter`.

**Use cases:** assembling binary or text output, capturing what a function writes (in tests), passing data to APIs that want an `io.Reader` (`bytes.NewReader`).

Compare byte slices with `bytes.Equal`, never `==`.

### Example
```go
package main

import (
	"bytes"
	"fmt"
)

func main() {
	var buf bytes.Buffer // zero value is ready
	buf.WriteString("hello ")
	fmt.Fprintf(&buf, "%d items", 3)
	fmt.Println(buf.String(), buf.Len())
	fmt.Println(bytes.Equal([]byte("a"), []byte("a")))
}
```

### Exercise
Write `render(items []string) string` that writes each item plus a newline into a `bytes.Buffer` and returns the string.
```text expect
a
b
```
```go solution
package main

import (
	"bytes"
	"fmt"
)

// BEGIN
func render(items []string) string {
	var buf bytes.Buffer
	for _, it := range items {
		buf.WriteString(it)
		buf.WriteByte('\n')
	}
	return buf.String()
}

// END

func main() {
	fmt.Print(render([]string{"a", "b"}))
}
```

### Check
Q: How should two `[]byte` values be compared?
T: mcq
- [ ] a == b
- [x] bytes.Equal(a, b)
- [ ] a.Equals(b)
- [ ] strings.Equal
E: Slices can't be compared with `==` (except to nil); use `bytes.Equal`.

Q: `bytes.Buffer` needs an explicit constructor before use.
T: tf
A: false
E: Its zero value is an empty buffer ready to use.

## sort
slug: sort
minutes: 7
challenges: sort-people
objectives: Sort slices of basic types; Sort structs with sort.Slice; Keep sorts stable when order of equals matters
takeaways: sort.Ints/Strings/Float64s sort common slices; sort.Slice takes a less function for anything else; sort.SliceStable preserves the order of equal elements

### Concept
**Explanation.** `sort.Slice(xs, func(i, j int) bool {...})` sorts any slice given a *less* function. Use `sort.SliceStable` to keep equal elements in their original order, and `sort.Search` for binary search on a sorted slice.

**Use cases:** ordering records for display, deterministic map iteration, leaderboards.

Since Go 1.21 the `slices` package offers generic `slices.Sort` and `slices.SortFunc` which are usually nicer (next lesson).

### Example
```go
package main

import (
	"fmt"
	"sort"
)

type Person struct {
	Name string
	Age  int
}

func main() {
	people := []Person{{"Bob", 30}, {"Ada", 30}, {"Cy", 25}}
	sort.SliceStable(people, func(i, j int) bool { return people[i].Age < people[j].Age })
	fmt.Println(people)
}
```

### Exercise
Sort `words` by length, shortest first, breaking ties alphabetically. Print the result.
```text expect
[go is fun code]
```
```go solution
package main

import (
	"fmt"
	"sort"
)

func main() {
	words := []string{"code", "go", "fun", "is"}
	// BEGIN
	sort.Slice(words, func(i, j int) bool {
		if len(words[i]) != len(words[j]) {
			return len(words[i]) < len(words[j])
		}
		return words[i] < words[j]
	})
	// END
	fmt.Println(words)
}
```

### Check
Q: Which sort keeps equal elements in their original relative order?
T: mcq
- [ ] sort.Slice
- [x] sort.SliceStable
- [ ] sort.Sort
- [ ] sort.Ints
E: Only stable sorts guarantee that equal elements aren't reordered.

Q: What does this print?
T: output
```go
xs := []int{3, 1, 2}
sort.Ints(xs)
fmt.Println(xs)
```
- [ ] [3 1 2]
- [x] [1 2 3]
- [ ] [3 2 1]
- [ ] [1 3 2]
E: `sort.Ints` sorts in place in ascending order.

## slices
slug: slices
minutes: 7
challenges: slices-dedupe-sorted
objectives: Use generic helpers such as Sort, Contains, Index and Reverse; Search sorted data with BinarySearch; Clone and compare slices
takeaways: slices provides type-safe generic helpers (Go 1.21+); slices.Sort, SortFunc, Contains, Index, Max, Min, Reverse, Clone, Equal, BinarySearch; Prefer it to hand-written loops

### Concept
**Explanation.** The `slices` package works with any slice type using generics.

**Common operations:** `slices.Sort(xs)`, `slices.SortFunc(xs, cmp)`, `slices.Contains(xs, x)`, `slices.Index`, `slices.Max/Min`, `slices.Reverse`, `slices.Clone`, `slices.Equal`, `slices.BinarySearch`, `slices.Compact` (remove adjacent duplicates), `slices.Insert/Delete`.

`cmp.Compare(a, b)` and `strings.Compare` return -1, 0, 1 for `SortFunc`.

### Example
```go
package main

import (
	"cmp"
	"fmt"
	"slices"
)

func main() {
	xs := []int{5, 2, 8, 2, 9}
	slices.Sort(xs)
	fmt.Println(xs, slices.Contains(xs, 8), slices.Max(xs))
	xs = slices.Compact(xs)
	fmt.Println(xs)
	type P struct {
		Name string
		Age  int
	}
	ps := []P{{"b", 2}, {"a", 2}, {"c", 1}}
	slices.SortFunc(ps, func(x, y P) int { return cmp.Or(cmp.Compare(x.Age, y.Age), cmp.Compare(x.Name, y.Name)) })
	fmt.Println(ps)
}
```

### Exercise
Return the sorted, de-duplicated version of `xs` using `slices.Sort` and `slices.Compact` on a clone.
```text expect
[1 2 3 5]
```
```go solution
package main

import (
	"fmt"
	"slices"
)

// BEGIN
func uniqueSorted(xs []int) []int {
	out := slices.Clone(xs)
	slices.Sort(out)
	return slices.Compact(out)
}

// END

func main() {
	fmt.Println(uniqueSorted([]int{3, 1, 5, 3, 2, 1}))
}
```

### Check
Q: What does `slices.Compact` do?
T: mcq
- [ ] Sorts the slice
- [x] Removes consecutive duplicate elements
- [ ] Shrinks capacity
- [ ] Copies the slice
E: `Compact` removes runs of equal adjacent elements — sort first to remove all duplicates.

Q: Which function tells you whether a slice contains a value?
T: short
A: slices.Contains
E: `slices.Contains(xs, x)` reports membership in a slice.

## maps
slug: maps
minutes: 6
challenges: invert-map
objectives: Copy and compare maps; Extract keys and values as sorted slices; Delete entries by predicate
takeaways: The maps package (Go 1.21+) provides Clone, Copy, Equal, DeleteFunc, Keys and Values; maps.Keys returns an iterator — use slices.Sorted to order it; Ranging over a map is unordered

### Concept
**Explanation.** `maps` has generic helpers: `maps.Clone`, `maps.Copy(dst, src)`, `maps.Equal`, `maps.DeleteFunc(m, func(k, v) bool)`. `maps.Keys(m)` and `maps.Values(m)` return **iterators** (Go 1.23+); combine with `slices.Sorted` / `slices.Collect`.

**Use cases:** deterministic output (`slices.Sorted(maps.Keys(m))`), defensive copies before returning internal state, filtering.

### Example
```go
package main

import (
	"fmt"
	"maps"
	"slices"
)

func main() {
	m := map[string]int{"b": 2, "a": 1, "c": 3}
	fmt.Println(slices.Sorted(maps.Keys(m)))
	c := maps.Clone(m)
	maps.DeleteFunc(c, func(k string, v int) bool { return v < 2 })
	fmt.Println(len(m), len(c))
}
```

### Exercise
Write `invert(m map[string]int) map[int]string` and print the inverted map (`fmt` prints maps sorted by key).
```text expect
map[1:a 2:b]
```
```go solution
package main

import "fmt"

// BEGIN
func invert(m map[string]int) map[int]string {
	out := make(map[int]string, len(m))
	for k, v := range m {
		out[v] = k
	}
	return out
}

// END

func main() {
	fmt.Println(invert(map[string]int{"a": 1, "b": 2}))
}
```

### Check
Q: What does `fmt.Println` do with the key order of a map?
T: mcq
- [ ] Random order
- [x] It prints keys in sorted order
- [ ] Insertion order
- [ ] Reverse order
E: `fmt` sorts map keys when printing so output is deterministic.

Q: `maps.Keys` returns a sorted slice.
T: tf
A: false
E: It returns an iterator in unspecified order; use `slices.Sorted(maps.Keys(m))` to sort.

## time
slug: time
minutes: 8
challenges: parse-duration
objectives: Work with time.Time and time.Duration; Format and parse with reference layouts; Use timers, tickers and deadlines
takeaways: time.Duration is an int64 of nanoseconds — multiply by time.Second etc.; Layouts use the reference time Mon Jan 2 15:04:05 MST 2006; Store and compare times in UTC and use time.Since for elapsed time

### Concept
**Explanation.** `time.Now()`, `time.Date(...)`, `t.Add(d)`, `t.Sub(u)`, `t.Before(u)`, `time.Since(t)`. Durations: `90*time.Second`, `d.Minutes()`.

**Formatting** uses a reference time instead of `%Y-%m-%d`: `t.Format("2006-01-02 15:04")`. Parse with `time.Parse(layout, s)`; `time.ParseDuration("1h30m")`.

**Timers:** `time.After(d)`, `time.NewTimer`, `time.NewTicker` — always `Stop()` timers and tickers you no longer need.

**Tip:** pass a clock/`now func() time.Time` into code you need to test.

### Example
```go
package main

import (
	"fmt"
	"time"
)

func main() {
	t := time.Date(2024, time.March, 9, 14, 30, 0, 0, time.UTC)
	fmt.Println(t.Format("2006-01-02 15:04"), t.Weekday())
	fmt.Println(t.Add(36 * time.Hour).Format(time.RFC3339))
	d, _ := time.ParseDuration("1h30m")
	fmt.Println(d, d.Minutes())
}
```

### Exercise
Write `daysBetween(a, b string) int` parsing both dates as `2006-01-02` and returning the number of whole days between them.
```text expect
30
```
```go solution
package main

import (
	"fmt"
	"time"
)

// BEGIN
func daysBetween(a, b string) int {
	ta, _ := time.Parse("2006-01-02", a)
	tb, _ := time.Parse("2006-01-02", b)
	return int(tb.Sub(ta).Hours() / 24)
}

// END

func main() {
	fmt.Println(daysBetween("2024-01-01", "2024-01-31"))
}
```

### Check
Q: What is the layout for a `YYYY-MM-DD` date in Go?
T: short
A: 2006-01-02
E: Go layouts use the fixed reference time: Jan 2, 2006 at 15:04:05.

Q: What does this print?
T: output
```go
d := 90 * time.Minute
fmt.Println(d)
```
- [ ] 90m
- [x] 1h30m0s
- [ ] 5400
- [ ] 1.5h
E: `Duration.String()` normalises to hours, minutes and seconds.

## math
slug: math
minutes: 6
challenges: basic-stats
objectives: Use common math functions and constants; Recognise float precision limits and NaN/Inf; Use math/rand/v2 for non-security randomness
takeaways: math offers Sqrt, Pow, Floor, Ceil, Round, Abs, Max, Min, Inf and constants like Pi; Floating-point arithmetic is inexact — compare with a tolerance; Use crypto/rand for secrets, math/rand/v2 for simulations

### Concept
**Explanation.** `math.Sqrt`, `math.Pow`, `math.Floor/Ceil/Round/Trunc`, `math.Abs`, `math.MaxInt64`, `math.Pi`. Integer division truncates; convert to float for fractional results.

Floats are binary approximations: `0.1 + 0.2 != 0.3`. Compare with `math.Abs(a-b) < 1e-9`. Division by float zero gives `+Inf`; `math.NaN()` is not equal to itself — test with `math.IsNaN`.

`math/rand/v2` (Go 1.22+) is for games and sampling; use `crypto/rand` for tokens and keys.

### Example
```go
package main

import (
	"fmt"
	"math"
)

func main() {
	fmt.Println(math.Sqrt(2), math.Round(2.5), math.Floor(-1.5))
	fmt.Println(0.1+0.2 == 0.3, math.Abs(0.1+0.2-0.3) < 1e-9)
	fmt.Println(math.Inf(1), math.IsNaN(math.NaN()))
}
```

### Exercise
Write `mean(xs []float64) float64` and print the mean of `2, 4, 9` rounded to 2 decimals with `%.2f`.
```text expect
5.00
```
```go solution
package main

import "fmt"

// BEGIN
func mean(xs []float64) float64 {
	sum := 0.0
	for _, x := range xs {
		sum += x
	}
	return sum / float64(len(xs))
}

// END

func main() {
	fmt.Printf("%.2f\n", mean([]float64{2, 4, 9}))
}
```

### Check
Q: What does `math.Round(2.5)` return?
T: mcq
- [ ] 2
- [x] 3
- [ ] 2.5
- [ ] 4
E: `math.Round` rounds half away from zero.

Q: Which package should generate authentication tokens?
T: mcq
- [ ] math/rand
- [ ] math/rand/v2
- [x] crypto/rand
- [ ] time
E: Only `crypto/rand` is cryptographically secure.

## regexp
slug: regexp
minutes: 8
challenges: extract-emails
objectives: Compile and match regular expressions; Extract submatches and replace with templates; Understand Go's RE2 guarantees
takeaways: Compile once with MustCompile at package level; FindString/FindAllString/FindStringSubmatch extract matches; RE2 has no backreferences but guarantees linear-time matching

### Concept
**Explanation.** `re := regexp.MustCompile(`...`)` (use raw strings). `re.MatchString(s)`, `re.FindString(s)`, `re.FindAllString(s, -1)`, `re.FindStringSubmatch(s)`, `re.ReplaceAllString(s, "$1")`.

Named groups: `(?P<year>\d{4})` with `re.SubexpNames()`.

Go uses **RE2**: no lookaheads or backreferences, but matching time is linear in input size, so untrusted input can't cause catastrophic backtracking. Compile once and reuse — compiling is expensive.

### Example
```go
package main

import (
	"fmt"
	"regexp"
)

var dateRe = regexp.MustCompile(`(\d{4})-(\d{2})-(\d{2})`)

func main() {
	s := "released 2024-03-09, patched 2024-04-01"
	fmt.Println(dateRe.FindString(s))
	fmt.Println(dateRe.FindAllString(s, -1))
	fmt.Println(dateRe.FindStringSubmatch(s)[1:])
	fmt.Println(dateRe.ReplaceAllString(s, "$3/$2/$1"))
}
```

### Exercise
Write `hashtags(s string) []string` returning all `#word` tags (including the `#`) using a regexp.
```text expect
[#go #learn]
```
```go solution
package main

import (
	"fmt"
	"regexp"
)

// BEGIN
var tagRe = regexp.MustCompile(`#\w+`)

func hashtags(s string) []string {
	return tagRe.FindAllString(s, -1)
}

// END

func main() {
	fmt.Println(hashtags("I love #go and want to #learn more"))
}
```

### Check
Q: Which regexp feature does Go's engine (RE2) NOT support?
T: mcq
- [ ] Character classes
- [ ] Non-greedy quantifiers
- [x] Backreferences
- [ ] Named groups
E: RE2 omits backreferences and lookaround in exchange for linear-time matching.

Q: Where should `regexp.MustCompile` normally be called?
T: mcq
- [ ] Inside the function that uses it, on every call
- [x] Once, at package level
- [ ] In init only
- [ ] It doesn't matter
E: Compilation is costly; compile once and reuse the `*Regexp`.

## encoding/json
slug: encoding-json
minutes: 8
challenges: json-roundtrip
objectives: Marshal Go values to JSON; Unmarshal JSON into structs and maps; Control field names with struct tags
takeaways: json.Marshal / json.Unmarshal convert between Go values and JSON bytes; Only exported fields are encoded — use tags for names and omitempty; Unmarshal into map[string]any only when the shape is unknown

### Concept
**Explanation.** `json.Marshal(v)` → `[]byte`; `json.Unmarshal(data, &v)` (pass a **pointer**). Tags: `` `json:"user_name,omitempty"` `` rename or omit fields; `` `json:"-"` `` hides a field.

**Use cases:** HTTP APIs, config files, storing documents. For streams use `json.NewEncoder(w)` / `json.NewDecoder(r)` (Module 10 covers this in depth).

Numbers decode into `float64` when the target is `any`.

### Example
```go
package main

import (
	"encoding/json"
	"fmt"
)

type User struct {
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
	Admin bool   `json:"-"`
}

func main() {
	b, _ := json.Marshal(User{Name: "Ada", Admin: true})
	fmt.Println(string(b))
	var u User
	err := json.Unmarshal([]byte(`{"name":"Bob","email":"b@x.io"}`), &u)
	fmt.Println(u, err)
}
```

### Exercise
Marshal the struct `Item{Name: "pen", Price: 1.5}` with tags `name` and `price` and print the JSON.
```text expect
{"name":"pen","price":1.5}
```
```go solution
package main

import (
	"encoding/json"
	"fmt"
)

// BEGIN
type Item struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

// END

func main() {
	b, _ := json.Marshal(Item{Name: "pen", Price: 1.5})
	fmt.Println(string(b))
}
```

### Check
Q: Which fields does `encoding/json` encode?
T: mcq
- [ ] All fields
- [x] Only exported fields
- [ ] Only tagged fields
- [ ] Only string fields
E: Unexported fields are invisible to reflection-based encoders.

Q: `json.Unmarshal(data, u)` (passing a struct value, not a pointer) works.
T: tf
A: false
E: You must pass a pointer, otherwise Unmarshal returns an `InvalidUnmarshalError`.

## encoding/csv
slug: encoding-csv
minutes: 6
challenges: csv-totals
objectives: Read CSV records with csv.Reader; Write CSV with csv.Writer; Handle quoting and headers
takeaways: csv.NewReader(r).ReadAll() returns [][]string; csv.NewWriter(w) must be Flushed; The package handles quoting and embedded commas for you

### Concept
**Explanation.** `r := csv.NewReader(src)` where `src` is any `io.Reader`. `r.Read()` returns one record; `r.ReadAll()` loads them all. Set `r.Comma = ';'` for other delimiters and `r.FieldsPerRecord = -1` to allow ragged rows.

`w := csv.NewWriter(dst)`; `w.Write(record)`; **`w.Flush()`** and check `w.Error()`.

**Use cases:** spreadsheet exports, log ingestion, data-import tools. Skip the header row explicitly, and parse numeric columns with `strconv`.

### Example
```go
package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
)

func main() {
	in := "name,qty\n\"Smith, J\",3\nbob,5\n"
	recs, err := csv.NewReader(strings.NewReader(in)).ReadAll()
	fmt.Println(recs, err)
	w := csv.NewWriter(os.Stdout)
	_ = w.Write([]string{"a,b", "c"})
	w.Flush()
}
```

### Exercise
Write `total(csvText string) int` summing the second column (`qty`) of all data rows (skip the header).
```text expect
8
```
```go solution
package main

import (
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
)

// BEGIN
func total(csvText string) int {
	recs, err := csv.NewReader(strings.NewReader(csvText)).ReadAll()
	if err != nil {
		return 0
	}
	sum := 0
	for _, r := range recs[1:] {
		n, _ := strconv.Atoi(r[1])
		sum += n
	}
	return sum
}

// END

func main() {
	fmt.Println(total("name,qty\napple,3\npear,5\n"))
}
```

### Check
Q: What must you call after writing records with `csv.Writer`?
T: short
A: Flush
E: `Writer` is buffered; call `Flush()` (and check `Error()`).

Q: The csv package handles fields containing commas correctly.
T: tf
A: true
E: It quotes and unquotes fields following RFC 4180.

## os
slug: os
minutes: 7
challenges: read-env-config
objectives: Read environment variables and arguments; Create, write and read files; Exit with a status code
takeaways: os.Getenv/LookupEnv read the environment, os.Args the arguments; os.ReadFile/WriteFile handle whole-file I/O; os.Exit skips deferred calls — return from main instead when you can

### Concept
**Explanation.** `os.Args`, `os.Getenv("HOME")`, `os.LookupEnv` (distinguishes unset from empty), `os.ReadFile`, `os.WriteFile(path, data, 0o644)`, `os.MkdirAll`, `os.Remove`, `os.Stat`, `os.Exit(code)`.

Always check errors and compare with `errors.Is(err, fs.ErrNotExist)`. Use `os.CreateTemp` for scratch files. `os.Exit` terminates immediately — **deferred functions don't run**, so keep it in `main` only.

### Example
```go
package main

import (
	"fmt"
	"os"
)

func main() {
	os.Setenv("APP_MODE", "dev")
	mode, ok := os.LookupEnv("APP_MODE")
	fmt.Println(mode, ok)
	_, ok = os.LookupEnv("SURELY_UNSET_VAR")
	fmt.Println(ok)
}
```

### Exercise
Write `getenv(key, def string) string` returning `def` when the variable is unset or empty.
```text expect
8080 debug
```
```go solution
package main

import (
	"fmt"
	"os"
)

// BEGIN
func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// END

func main() {
	os.Setenv("LEVEL", "debug")
	fmt.Println(getenv("PORT_UNSET", "8080"), getenv("LEVEL", "info"))
}
```

### Check
Q: What happens to deferred functions when `os.Exit(1)` is called?
T: mcq
- [ ] They run first
- [x] They are skipped
- [ ] They run in reverse order
- [ ] Only the first runs
E: `os.Exit` terminates the process immediately without running defers.

Q: Which function tells you whether an environment variable is set at all (even to an empty string)?
T: short
A: os.LookupEnv
E: `LookupEnv` returns `(value, ok)` so unset and empty can be distinguished.

## io
slug: io
minutes: 7
challenges: count-bytes
objectives: Use io.Reader and io.Writer together; Copy, limit and tee streams; Read fully with io.ReadAll
takeaways: io.Copy moves data between a Reader and a Writer without loading it all; io.LimitReader, io.TeeReader and io.MultiWriter compose streams; io.ReadAll loads everything — guard it with a limit for untrusted input

### Concept
**Explanation.** `io.Reader`/`io.Writer` are the universal streaming interfaces. Helpers: `io.Copy(dst, src)`, `io.ReadAll(r)`, `io.LimitReader(r, n)`, `io.TeeReader(r, w)` (copy while reading), `io.MultiWriter(a, b)`, `io.Pipe()`, `io.EOF` (end of stream — not a failure).

**Use cases:** HTTP bodies, files, compression, hashing while copying (`io.Copy(hasher, file)`).

Never `io.ReadAll` an untrusted body without `io.LimitReader` — it can exhaust memory.

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
	r := io.LimitReader(strings.NewReader("abcdefghij"), 4)
	n, _ := io.Copy(os.Stdout, r)
	fmt.Println("\ncopied", n)
	both := io.MultiWriter(os.Stdout, io.Discard)
	fmt.Fprintln(both, "written twice, seen once")
}
```

### Exercise
Write `readAtMost(r io.Reader, n int64) (string, error)` using `io.ReadAll` over an `io.LimitReader`.
```text expect
hello
```
```go solution
package main

import (
	"fmt"
	"io"
	"strings"
)

// BEGIN
func readAtMost(r io.Reader, n int64) (string, error) {
	b, err := io.ReadAll(io.LimitReader(r, n))
	return string(b), err
}

// END

func main() {
	s, _ := readAtMost(strings.NewReader("hello world"), 5)
	fmt.Println(s)
}
```

### Check
Q: What does `io.EOF` signal?
T: mcq
- [ ] A read error
- [x] A normal end of the stream
- [ ] A closed file
- [ ] A timeout
E: `EOF` is returned when no more input is available; it is not a failure.

Q: Why wrap untrusted input in `io.LimitReader` before `io.ReadAll`?
T: mcq
- [ ] To speed it up
- [x] To prevent unbounded memory use
- [ ] To decode UTF-8
- [ ] To close it
E: `ReadAll` reads until EOF; a limit bounds how much an attacker can make you allocate.

## bufio
slug: bufio
minutes: 7
challenges: word-count
objectives: Read lines and words with bufio.Scanner; Buffer writes with bufio.Writer; Handle long lines
takeaways: bufio.Scanner splits input into lines (default) or words; bufio.Writer batches small writes — remember to Flush; Scanner has a 64 KB line limit unless you call Buffer

### Concept
**Explanation.** `sc := bufio.NewScanner(r)`; `sc.Scan()` advances, `sc.Text()` returns the token, `sc.Err()` reports failure. `sc.Split(bufio.ScanWords)` switches to words. For lines longer than 64 KB call `sc.Buffer(make([]byte, 0, 64*1024), 1<<20)`.

`bufio.NewReader(r)` provides `ReadString('\n')` and `ReadRune`. `bufio.NewWriter(w)` buffers output — **`Flush()`** at the end.

**Use cases:** processing log files, parsing stdin, writing many small lines efficiently.

### Example
```go
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	sc := bufio.NewScanner(strings.NewReader("the quick brown fox"))
	sc.Split(bufio.ScanWords)
	n := 0
	for sc.Scan() {
		n++
	}
	fmt.Println(n)
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	fmt.Fprintln(w, "buffered")
}
```

### Exercise
Write `lines(s string) int` counting lines with a `bufio.Scanner`.
```text expect
3
```
```go solution
package main

import (
	"bufio"
	"fmt"
	"strings"
)

// BEGIN
func lines(s string) int {
	sc := bufio.NewScanner(strings.NewReader(s))
	n := 0
	for sc.Scan() {
		n++
	}
	return n
}

// END

func main() {
	fmt.Println(lines("a\nb\nc\n"))
}
```

### Check
Q: What must you do at the end when using a `bufio.Writer`?
T: short
A: Flush
E: Buffered data is only written to the underlying writer on `Flush` (or when the buffer fills).

Q: What is the default maximum token size for `bufio.Scanner`?
T: mcq
- [ ] 1 KB
- [x] 64 KB
- [ ] 1 MB
- [ ] Unlimited
E: Longer lines make `Scan` fail with `bufio.ErrTooLong` unless you enlarge the buffer.

## filepath
slug: filepath
minutes: 6
challenges: clean-paths
objectives: Join, split and clean file paths portably; Walk directory trees; Match names with globs
takeaways: filepath.Join builds OS-correct paths and cleans them; filepath.Base/Dir/Ext split paths; filepath.WalkDir traverses a tree; use path for URLs and slash-separated paths

### Concept
**Explanation.** `filepath.Join("a", "b", "../c")` → `a/c` (uses `\` on Windows). `Base`, `Dir`, `Ext`, `Clean`, `Abs`, `Rel`, `Match`, `Glob`, and `WalkDir(root, fn)`.

Use `path` (not `filepath`) for URL paths; they always use `/`. **Security:** when joining user input to a base directory, check the result is still inside the base to avoid `../` traversal.

### Example
```go
package main

import (
	"fmt"
	"path/filepath"
)

func main() {
	p := filepath.Join("data", "logs", "..", "app.log")
	fmt.Println(p, filepath.Base(p), filepath.Ext(p), filepath.Dir(p))
	ok, _ := filepath.Match("*.go", "main.go")
	fmt.Println(ok)
}
```

### Exercise
Write `changeExt(name, ext string) string` replacing the file extension (`report.txt`, `.md` → `report.md`).
```text expect
report.md
```
```go solution
package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

// BEGIN
func changeExt(name, ext string) string {
	return strings.TrimSuffix(name, filepath.Ext(name)) + ext
}

// END

func main() {
	fmt.Println(changeExt("report.txt", ".md"))
}
```

### Check
Q: Which package should you use for URL paths?
T: mcq
- [ ] path/filepath
- [x] path
- [ ] os
- [ ] net
E: URL paths always use forward slashes; `path` is OS-independent.

Q: What does `filepath.Join("a", "../b")` return?
T: mcq
- [ ] a/../b
- [x] b
- [ ] a/b
- [ ] ../b
E: `Join` calls `Clean`, which resolves `..` elements lexically.

## context
slug: context
minutes: 9
challenges: context-timeout-call
objectives: Create cancellable contexts and timeouts; Pass ctx as the first parameter; Check ctx.Done() and ctx.Err()
takeaways: context carries cancellation, deadlines and request-scoped values; ctx is the first parameter, named ctx, never stored in structs; Always call the cancel function returned by WithCancel/WithTimeout

### Concept
**Explanation.** `context.Background()` is the root. Derive with `WithCancel`, `WithTimeout`, `WithDeadline`; each returns a `cancel` function you must call (usually `defer cancel()`).

Long-running or blocking work selects on `ctx.Done()`; `ctx.Err()` is `context.Canceled` or `context.DeadlineExceeded`.

**Use cases:** HTTP handlers, database queries, RPCs, worker shutdown. Standard-library APIs accept a context (`http.NewRequestWithContext`, `db.QueryContext`).

Module 11 goes deeper; here you learn the API.

### Example
```go
package main

import (
	"context"
	"fmt"
	"time"
)

func slow(ctx context.Context) error {
	select {
	case <-time.After(200 * time.Millisecond):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	fmt.Println(slow(ctx))
}
```

### Exercise
Write `waitOrCancel(ctx context.Context, d time.Duration) string` returning `done` after `d` or `cancelled` if the context ends first.
```text expect
done cancelled
```
```go solution
package main

import (
	"context"
	"fmt"
	"time"
)

// BEGIN
func waitOrCancel(ctx context.Context, d time.Duration) string {
	select {
	case <-time.After(d):
		return "done"
	case <-ctx.Done():
		return "cancelled"
	}
}

// END

func main() {
	a := waitOrCancel(context.Background(), 5*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	b := waitOrCancel(ctx, time.Second)
	fmt.Println(a, b)
}
```

### Check
Q: Where does `ctx` conventionally appear in a function's parameters?
T: mcq
- [ ] Last
- [x] First
- [ ] Anywhere
- [ ] As a struct field
E: `func Do(ctx context.Context, ...)` is the universal convention.

Q: Why must you call the `cancel` function returned by `WithTimeout`?
T: mcq
- [ ] To start the timer
- [x] To release resources associated with the context
- [ ] To log the error
- [ ] It is optional
E: Not calling it leaks the timer and child-context bookkeeping until the parent ends.

## sync
slug: sync
minutes: 8
challenges: safe-counter
objectives: Protect shared state with sync.Mutex; Wait for goroutines with sync.WaitGroup; Run initialisation once with sync.Once
takeaways: Mutex guards shared data; lock, defer unlock; WaitGroup waits for a set of goroutines; Once runs a function exactly once, safely

### Concept
**Explanation.** `sync.Mutex` (`Lock`/`Unlock`), `sync.RWMutex` (many readers or one writer), `sync.WaitGroup` (`Add`, `Done`, `Wait`), `sync.Once`, `sync.Pool`, `sync.Map`.

**Use cases:** counters and caches accessed from several goroutines, waiting for workers to finish, lazy singleton initialisation.

The rules: **never copy** a mutex/WaitGroup after first use; hold locks briefly; `defer mu.Unlock()`; call `wg.Add` **before** starting the goroutine. Detect mistakes with `go test -race`.

### Example
```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	var (
		mu sync.Mutex
		wg sync.WaitGroup
		n  int
	)
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

### Exercise
Increment the shared counter 50 times from 50 goroutines safely and print `50`.
```text expect
50
```
```go solution
package main

import (
	"fmt"
	"sync"
)

func main() {
	var (
		mu sync.Mutex
		wg sync.WaitGroup
		n  int
	)
	// BEGIN
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			defer mu.Unlock()
			n++
		}()
	}
	wg.Wait()
	// END
	fmt.Println(n)
}
```

### Check
Q: When must `wg.Add(1)` be called relative to starting the goroutine?
T: mcq
- [ ] Inside the goroutine
- [x] Before starting the goroutine
- [ ] After wg.Wait()
- [ ] It doesn't matter
E: Calling `Add` inside the goroutine races with `Wait`, which may return early.

Q: It is safe to copy a `sync.Mutex` after it has been used.
T: tf
A: false
E: Copying creates a second, independent lock state; pass mutexes (and structs containing them) by pointer. `go vet` catches this.
