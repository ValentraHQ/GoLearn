# Standard Library Challenges

## fmt-table
title: Format a Table Row
difficulty: beginner
module: stdlib
lesson: stdlib/fmt
skill: stdlib

### Problem
Write `FormatRow(name string, qty int, price float64) string` that returns one row of a text table: the name left-aligned in 10 columns, the quantity right-aligned in 3 columns, and the price right-aligned in 8 columns with 2 decimals, each column separated by a single space.

### Expected
`FormatRow("apple", 3, 1.5)` returns `apple        3     1.50`.

### Constraints
- Use `fmt.Sprintf` with width flags
- Do not add a trailing newline

### Starter
```go
package main

func FormatRow(name string, qty int, price float64) string {
	return ""
}

func main() {}
```

### Tests
```go
package main

import "testing"

func TestFormatRow(t *testing.T) {
	tests := []struct {
		name  string
		qty   int
		price float64
		want  string
	}{
		{"apple", 3, 1.5, "apple        3     1.50"},
		{"watermelon", 12, 10, "watermelon  12    10.00"},
		{"x", 0, 0.005, "x            0     0.01"},
	}
	for _, tc := range tests {
		if got := FormatRow(tc.name, tc.qty, tc.price); got != tc.want {
			t.Errorf("FormatRow(%q,%d,%v) = %q, want %q", tc.name, tc.qty, tc.price, got, tc.want)
		}
	}
}
```

### Hints
- `%-10s` left-aligns in 10 columns; `%3d` right-aligns integers
- `%8.2f` is width 8 with two decimals

### Solution
```go
package main

import "fmt"

func FormatRow(name string, qty int, price float64) string {
	return fmt.Sprintf("%-10s %3d %8.2f", name, qty, price)
}

func main() {}
```

### Explanation
Width and precision flags in `fmt` verbs do all the alignment work: `-` left-aligns, a number is the minimum width, and `.2` after it sets float precision.

## title-case
title: Title Case
difficulty: beginner
module: stdlib
lesson: stdlib/strings
skill: stdlib

### Problem
Write `TitleCase(s string) string` that uppercases the first letter of every word and lowercases the rest. Words are separated by any amount of whitespace; the result joins them with single spaces. Empty or all-space input returns `""`.

### Constraints
- Use the `strings` package
- Assume ASCII input

### Starter
```go
package main

func TitleCase(s string) string {
	return ""
}

func main() {}
```

### Tests
```go
package main

import "testing"

func TestTitleCase(t *testing.T) {
	tests := map[string]string{
		"hello world":       "Hello World",
		"  gO   lAnG  ":     "Go Lang",
		"":                  "",
		"   ":               "",
		"single":            "Single",
		"multiple  spaces ": "Multiple Spaces",
	}
	for in, want := range tests {
		if got := TitleCase(in); got != want {
			t.Errorf("TitleCase(%q) = %q, want %q", in, got, want)
		}
	}
}
```

### Hints
- `strings.Fields` handles any run of whitespace
- Build each word from `strings.ToUpper(w[:1]) + strings.ToLower(w[1:])`
- `strings.Join` puts the words back together

### Solution
```go
package main

import "strings"

func TitleCase(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		words[i] = strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
	}
	return strings.Join(words, " ")
}

func main() {}
```

### Explanation
`Fields` splits on whitespace runs and drops empties, so the edge cases (empty, extra spaces) fall out naturally.

## sum-numeric-strings
title: Sum Numeric Strings
difficulty: beginner
module: stdlib
lesson: stdlib/strconv
skill: stdlib

### Problem
Write `SumNumbers(items []string) (int, []string)` that parses each item as an integer and returns the sum of the valid ones plus the list of invalid items in their original order.

### Constraints
- Use `strconv.Atoi`
- Surrounding whitespace makes an item invalid

### Starter
```go
package main

func SumNumbers(items []string) (int, []string) {
	return 0, nil
}

func main() {}
```

### Tests
```go
package main

import (
	"reflect"
	"testing"
)

func TestSumNumbers(t *testing.T) {
	sum, bad := SumNumbers([]string{"1", "2", "x", "40", " 5", "-3", ""})
	if sum != 40 {
		t.Errorf("sum = %d, want 40", sum)
	}
	want := []string{"x", " 5", ""}
	if !reflect.DeepEqual(bad, want) {
		t.Errorf("invalid = %#v, want %#v", bad, want)
	}
	sum, bad = SumNumbers(nil)
	if sum != 0 || len(bad) != 0 {
		t.Errorf("empty input: %d %v", sum, bad)
	}
}
```

### Hints
- `Atoi` returns an error for anything that isn't a plain integer
- Append failures to a slice in the loop

### Solution
```go
package main

import "strconv"

func SumNumbers(items []string) (int, []string) {
	sum := 0
	var bad []string
	for _, it := range items {
		n, err := strconv.Atoi(it)
		if err != nil {
			bad = append(bad, it)
			continue
		}
		sum += n
	}
	return sum, bad
}

func main() {}
```

### Explanation
Parsing errors are ordinary values — collect them instead of aborting, which is how many real importers behave.

## byte-buffer-lines
title: Join Lines With a Buffer
difficulty: beginner
module: stdlib
lesson: stdlib/bytes
skill: stdlib

### Problem
Write `JoinLines(lines []string) []byte` returning every line followed by a newline (`\n`), built with a `bytes.Buffer`. An empty slice yields an empty (non-nil or nil) result of length 0.

### Constraints
- Use `bytes.Buffer`

### Starter
```go
package main

func JoinLines(lines []string) []byte {
	return nil
}

func main() {}
```

### Tests
```go
package main

import (
	"bytes"
	"testing"
)

func TestJoinLines(t *testing.T) {
	if got := JoinLines([]string{"a", "bc"}); !bytes.Equal(got, []byte("a\nbc\n")) {
		t.Errorf("got %q", got)
	}
	if got := JoinLines(nil); len(got) != 0 {
		t.Errorf("empty input gave %q", got)
	}
	if got := JoinLines([]string{""}); !bytes.Equal(got, []byte("\n")) {
		t.Errorf("blank line gave %q", got)
	}
}
```

### Hints
- The zero value of `bytes.Buffer` is ready to use
- `buf.Bytes()` returns the accumulated data

### Solution
```go
package main

import "bytes"

func JoinLines(lines []string) []byte {
	var buf bytes.Buffer
	for _, l := range lines {
		buf.WriteString(l)
		buf.WriteByte('\n')
	}
	return buf.Bytes()
}

func main() {}
```

### Explanation
`bytes.Buffer` avoids the repeated copying of `+=` and needs no initialisation.

## sort-people
title: Sort People
difficulty: beginner
module: stdlib
lesson: stdlib/sort
skill: stdlib

### Problem
Given `type Person struct { Name string; Age int }`, write `SortPeople(ps []Person)` that sorts the slice in place by age **descending**; people with the same age are ordered by name **ascending**.

### Constraints
- Sort in place
- Use `sort.Slice` or `slices.SortFunc`

### Starter
```go
package main

type Person struct {
	Name string
	Age  int
}

func SortPeople(ps []Person) {
}

func main() {}
```

### Tests
```go
package main

import (
	"reflect"
	"testing"
)

func TestSortPeople(t *testing.T) {
	ps := []Person{{"Bob", 30}, {"Ada", 30}, {"Cy", 41}, {"Di", 25}}
	SortPeople(ps)
	want := []Person{{"Cy", 41}, {"Ada", 30}, {"Bob", 30}, {"Di", 25}}
	if !reflect.DeepEqual(ps, want) {
		t.Errorf("got %v, want %v", ps, want)
	}
	SortPeople(nil)
	one := []Person{{"Solo", 1}}
	SortPeople(one)
}
```

### Hints
- In the less function compare ages first (`>` for descending)
- Only when ages are equal compare names

### Solution
```go
package main

import "sort"

type Person struct {
	Name string
	Age  int
}

func SortPeople(ps []Person) {
	sort.Slice(ps, func(i, j int) bool {
		if ps[i].Age != ps[j].Age {
			return ps[i].Age > ps[j].Age
		}
		return ps[i].Name < ps[j].Name
	})
}

func main() {}
```

### Explanation
A multi-key sort is one `less` function: decide on the first key unless equal, then fall through to the next.

## slices-dedupe-sorted
title: Sorted Unique Values
difficulty: beginner
module: stdlib
lesson: stdlib/slices
skill: stdlib

### Problem
Write `Dedupe(xs []int) []int` returning the sorted, duplicate-free values of `xs`. The input slice must not be modified.

### Constraints
- Do not mutate the argument
- Return an empty slice (length 0) for empty input

### Starter
```go
package main

func Dedupe(xs []int) []int {
	return nil
}

func main() {}
```

### Tests
```go
package main

import (
	"reflect"
	"testing"
)

func TestDedupe(t *testing.T) {
	in := []int{3, 1, 3, 2, 1}
	got := Dedupe(in)
	if !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Errorf("got %v", got)
	}
	if !reflect.DeepEqual(in, []int{3, 1, 3, 2, 1}) {
		t.Errorf("input was modified: %v", in)
	}
	if got := Dedupe(nil); len(got) != 0 {
		t.Errorf("nil input gave %v", got)
	}
	if got := Dedupe([]int{5, 5, 5}); !reflect.DeepEqual(got, []int{5}) {
		t.Errorf("got %v", got)
	}
}
```

### Hints
- `slices.Clone` before sorting protects the caller's data
- `slices.Compact` removes adjacent duplicates

### Solution
```go
package main

import "slices"

func Dedupe(xs []int) []int {
	out := slices.Clone(xs)
	slices.Sort(out)
	return slices.Compact(out)
}

func main() {}
```

### Explanation
Clone → sort → compact is the idiomatic three-step recipe, and it keeps the function free of side effects.

## invert-map
title: Group Keys by Value
difficulty: intermediate
module: stdlib
lesson: stdlib/maps
skill: stdlib

### Problem
Write `Invert(m map[string]string) map[string][]string` that maps each **value** to the **sorted** list of keys that had it.

### Constraints
- Key lists must be sorted ascending
- The input map must not be modified

### Starter
```go
package main

func Invert(m map[string]string) map[string][]string {
	return nil
}

func main() {}
```

### Tests
```go
package main

import (
	"reflect"
	"testing"
)

func TestInvert(t *testing.T) {
	in := map[string]string{"b": "x", "a": "x", "c": "y"}
	got := Invert(in)
	want := map[string][]string{"x": {"a", "b"}, "y": {"c"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
	if len(in) != 3 {
		t.Error("input modified")
	}
	if got := Invert(nil); len(got) != 0 {
		t.Errorf("nil input gave %v", got)
	}
}
```

### Hints
- Iterate the map, appending each key under its value
- Sort every list at the end (map order is random)

### Solution
```go
package main

import "sort"

func Invert(m map[string]string) map[string][]string {
	out := make(map[string][]string)
	for k, v := range m {
		out[v] = append(out[v], k)
	}
	for _, keys := range out {
		sort.Strings(keys)
	}
	return out
}

func main() {}
```

### Explanation
Map iteration order is random, so anything you build from it must be sorted before it is observable.

## parse-duration
title: Total Duration
difficulty: intermediate
module: stdlib
lesson: stdlib/time
skill: stdlib

### Problem
Write `Total(parts []string) (time.Duration, error)` that parses each string with `time.ParseDuration` and returns their sum. On the first invalid entry return a zero duration and an error whose message contains the offending string, wrapping the original error with `%w`.

### Constraints
- Use `time.ParseDuration`
- Wrap the cause with `%w`

### Starter
```go
package main

import "time"

func Total(parts []string) (time.Duration, error) {
	return 0, nil
}

func main() {}
```

### Tests
```go
package main

import (
	"strings"
	"testing"
	"time"
)

func TestTotal(t *testing.T) {
	got, err := Total([]string{"1h", "30m", "45s", "500ms"})
	if err != nil || got != time.Hour+30*time.Minute+45*time.Second+500*time.Millisecond {
		t.Fatalf("got %v, %v", got, err)
	}
	got, err = Total(nil)
	if err != nil || got != 0 {
		t.Fatalf("empty: %v %v", got, err)
	}
	got, err = Total([]string{"1h", "banana", "2h"})
	if err == nil || got != 0 {
		t.Fatalf("want error, got %v %v", got, err)
	}
	if !strings.Contains(err.Error(), "banana") {
		t.Errorf("error should mention the bad input: %v", err)
	}
	if !strings.Contains(err.Error(), "invalid duration") {
		t.Errorf("error should wrap the parse error: %v", err)
	}
}
```

### Hints
- `Duration` values can be added with `+`
- Format the error as `fmt.Errorf("parse %q: %w", s, err)`

### Solution
```go
package main

import (
	"fmt"
	"time"
)

func Total(parts []string) (time.Duration, error) {
	var sum time.Duration
	for _, p := range parts {
		d, err := time.ParseDuration(p)
		if err != nil {
			return 0, fmt.Errorf("parse %q: %w", p, err)
		}
		sum += d
	}
	return sum, nil
}

func main() {}
```

### Explanation
`Duration` is just an `int64` of nanoseconds, so summing is plain addition; the interesting part is returning a useful, wrapped error.

## basic-stats
title: Basic Statistics
difficulty: intermediate
module: stdlib
lesson: stdlib/math
skill: stdlib

### Problem
Write `Stats(xs []float64) (min, max, mean, median float64)` for a **non-empty** slice. The median of an even-length slice is the average of the two middle values. Do not modify the input.

### Constraints
- The input is never empty
- Do not mutate `xs`

### Starter
```go
package main

func Stats(xs []float64) (min, max, mean, median float64) {
	return
}

func main() {}
```

### Tests
```go
package main

import (
	"math"
	"testing"
)

func near(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestStats(t *testing.T) {
	in := []float64{7, 1, 3, 9, 5}
	mn, mx, mean, med := Stats(in)
	if !near(mn, 1) || !near(mx, 9) || !near(mean, 5) || !near(med, 5) {
		t.Errorf("odd: %v %v %v %v", mn, mx, mean, med)
	}
	if in[0] != 7 || in[1] != 1 {
		t.Error("input was modified")
	}
	mn, mx, mean, med = Stats([]float64{4, 1, 2, 3})
	if !near(mn, 1) || !near(mx, 4) || !near(mean, 2.5) || !near(med, 2.5) {
		t.Errorf("even: %v %v %v %v", mn, mx, mean, med)
	}
	mn, mx, mean, med = Stats([]float64{-2})
	if !near(mn, -2) || !near(mx, -2) || !near(mean, -2) || !near(med, -2) {
		t.Errorf("single: %v %v %v %v", mn, mx, mean, med)
	}
}
```

### Hints
- Copy and sort to find the median; min and max are then the ends
- Even length: average `s[n/2-1]` and `s[n/2]`

### Solution
```go
package main

import "sort"

func Stats(xs []float64) (min, max, mean, median float64) {
	s := make([]float64, len(xs))
	copy(s, xs)
	sort.Float64s(s)
	sum := 0.0
	for _, x := range s {
		sum += x
	}
	n := len(s)
	min, max, mean = s[0], s[n-1], sum/float64(n)
	if n%2 == 1 {
		median = s[n/2]
	} else {
		median = (s[n/2-1] + s[n/2]) / 2
	}
	return
}

func main() {}
```

### Explanation
Sorting a copy gives min, max and median in one pass and keeps the function side-effect free.

## extract-emails
title: Extract Email Addresses
difficulty: intermediate
module: stdlib
lesson: stdlib/regexp
skill: stdlib

### Problem
Write `Emails(s string) []string` that returns every email address found in `s`, in order of appearance, without duplicates. Use the pattern `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`. Trailing sentence punctuation is not part of an address.

### Constraints
- Compile the regexp once at package level
- Preserve first-seen order

### Starter
```go
package main

func Emails(s string) []string {
	return nil
}

func main() {}
```

### Tests
```go
package main

import (
	"reflect"
	"testing"
)

func TestEmails(t *testing.T) {
	in := "Contact ada@example.com, bob.smith+news@mail.example.org. Again: ada@example.com!"
	want := []string{"ada@example.com", "bob.smith+news@mail.example.org"}
	if got := Emails(in); !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
	if got := Emails("no addresses here @ all"); len(got) != 0 {
		t.Errorf("expected none, got %v", got)
	}
}
```

### Hints
- `FindAllString(s, -1)` returns all matches
- Track seen addresses in a map to remove duplicates

### Solution
```go
package main

import "regexp"

var emailRe = regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)

func Emails(s string) []string {
	seen := map[string]bool{}
	var out []string
	for _, m := range emailRe.FindAllString(s, -1) {
		if !seen[m] {
			seen[m] = true
			out = append(out, m)
		}
	}
	return out
}

func main() {}
```

### Explanation
The greedy domain part backtracks so a trailing `.` is left out of the match. A `seen` map keeps first-seen order while removing duplicates.

## json-roundtrip
title: JSON Config Round Trip
difficulty: intermediate
module: stdlib
lesson: stdlib/encoding-json
skill: stdlib

### Problem
Define `type Config struct { Name string; Port int; Tags []string; Debug bool }` with JSON keys `name`, `port`, `tags`, `debug`. `tags` and `debug` are omitted when empty/false.

Write `Encode(c Config) (string, error)` returning compact JSON, and `Decode(s string) (Config, error)` that rejects unknown fields and ports outside 1–65535.

### Constraints
- Use struct tags with `omitempty`
- Decode must fail on unknown fields
- Validate `Port` after decoding

### Starter
```go
package main

type Config struct {
	Name string
	Port int
	Tags []string
	Debug bool
}

func Encode(c Config) (string, error) {
	return "", nil
}

func Decode(s string) (Config, error) {
	return Config{}, nil
}

func main() {}
```

### Tests
```go
package main

import (
	"reflect"
	"testing"
)

func TestEncode(t *testing.T) {
	got, err := Encode(Config{Name: "api", Port: 8080})
	if err != nil || got != `{"name":"api","port":8080}` {
		t.Errorf("got %q, %v", got, err)
	}
	got, _ = Encode(Config{Name: "x", Port: 1, Tags: []string{"a", "b"}, Debug: true})
	if got != `{"name":"x","port":1,"tags":["a","b"],"debug":true}` {
		t.Errorf("got %q", got)
	}
}

func TestDecode(t *testing.T) {
	c, err := Decode(`{"name":"api","port":9000,"tags":["x"],"debug":true}`)
	want := Config{Name: "api", Port: 9000, Tags: []string{"x"}, Debug: true}
	if err != nil || !reflect.DeepEqual(c, want) {
		t.Errorf("got %+v, %v", c, err)
	}
	for _, bad := range []string{
		`{"name":"a","port":0}`,
		`{"name":"a","port":70000}`,
		`{"name":"a","port":80,"extra":1}`,
		`{"name":`,
	} {
		if _, err := Decode(bad); err == nil {
			t.Errorf("Decode(%s) should fail", bad)
		}
	}
}
```

### Hints
- Tags look like `` `json:"tags,omitempty"` ``
- `dec := json.NewDecoder(strings.NewReader(s)); dec.DisallowUnknownFields()`

### Solution
```go
package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Config struct {
	Name  string   `json:"name"`
	Port  int      `json:"port"`
	Tags  []string `json:"tags,omitempty"`
	Debug bool     `json:"debug,omitempty"`
}

func Encode(c Config) (string, error) {
	b, err := json.Marshal(c)
	return string(b), err
}

func Decode(s string) (Config, error) {
	var c Config
	dec := json.NewDecoder(strings.NewReader(s))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&c); err != nil {
		return Config{}, err
	}
	if c.Port < 1 || c.Port > 65535 {
		return Config{}, fmt.Errorf("port %d out of range", c.Port)
	}
	return c, nil
}

func main() {}
```

### Explanation
Struct tags drive encoding, and a `json.Decoder` with `DisallowUnknownFields` gives strict decoding. Validation of business rules still belongs after decoding.

## csv-totals
title: CSV Category Totals
difficulty: intermediate
module: stdlib
lesson: stdlib/encoding-csv
skill: stdlib

### Problem
Write `Totals(csvText string) (map[string]int, error)`. The first row is a header (`category,amount`). Sum `amount` per `category`. If an amount is not an integer, return an error mentioning the 1-based line number of the bad row (the header is line 1).

### Constraints
- Use `encoding/csv`
- Header row is skipped

### Starter
```go
package main

func Totals(csvText string) (map[string]int, error) {
	return nil, nil
}

func main() {}
```

### Tests
```go
package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestTotals(t *testing.T) {
	in := "category,amount\nfood,10\nrent,500\nfood,5\n\"misc, other\",7\n"
	got, err := Totals(in)
	want := map[string]int{"food": 15, "rent": 500, "misc, other": 7}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, %v", got, err)
	}
	_, err = Totals("category,amount\nfood,10\nrent,abc\n")
	if err == nil || !strings.Contains(err.Error(), "line 3") {
		t.Errorf("want error mentioning line 3, got %v", err)
	}
	got, err = Totals("category,amount\n")
	if err != nil || len(got) != 0 {
		t.Errorf("header only: %v %v", got, err)
	}
}
```

### Hints
- `ReadAll` returns records including the header; loop with an index
- Line number of `recs[i]` is `i+1`

### Solution
```go
package main

import (
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
)

func Totals(csvText string) (map[string]int, error) {
	recs, err := csv.NewReader(strings.NewReader(csvText)).ReadAll()
	if err != nil {
		return nil, err
	}
	out := map[string]int{}
	for i, r := range recs {
		if i == 0 {
			continue
		}
		n, err := strconv.Atoi(r[1])
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid amount %q", i+1, r[1])
		}
		out[r[0]] += n
	}
	return out, nil
}

func main() {}
```

### Explanation
`csv.Reader` deals with quoting so `"misc, other"` stays one field. Reporting the line number turns a generic parse error into something a user can act on.

## read-env-config
title: Configuration From the Environment
difficulty: intermediate
module: stdlib
lesson: stdlib/os
skill: stdlib

### Problem
Write `Load(getenv func(string) string) (Config, error)` for
`type Config struct { Host string; Port int; Debug bool }`.

- `HOST` defaults to `localhost`
- `PORT` defaults to `8080`, must be an integer 1–65535
- `DEBUG` defaults to `false`, parsed with `strconv.ParseBool`

Invalid values return an error naming the variable. `getenv` is injected so the code is testable.

### Constraints
- Do not call `os.Getenv` directly
- Empty values count as unset

### Starter
```go
package main

type Config struct {
	Host  string
	Port  int
	Debug bool
}

func Load(getenv func(string) string) (Config, error) {
	return Config{}, nil
}

func main() {}
```

### Tests
```go
package main

import (
	"strings"
	"testing"
)

func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestLoadDefaults(t *testing.T) {
	c, err := Load(env(nil))
	if err != nil || c != (Config{"localhost", 8080, false}) {
		t.Fatalf("got %+v, %v", c, err)
	}
}

func TestLoadValues(t *testing.T) {
	c, err := Load(env(map[string]string{"HOST": "db", "PORT": "5432", "DEBUG": "true"}))
	if err != nil || c != (Config{"db", 5432, true}) {
		t.Fatalf("got %+v, %v", c, err)
	}
}

func TestLoadErrors(t *testing.T) {
	cases := map[string]map[string]string{
		"PORT":  {"PORT": "abc"},
		"PORT ": {"PORT": "70000"},
		"DEBUG": {"DEBUG": "maybe"},
	}
	for name, m := range cases {
		_, err := Load(env(m))
		if err == nil || !strings.Contains(err.Error(), strings.TrimSpace(name)) {
			t.Errorf("%v: want error naming %s, got %v", m, name, err)
		}
	}
}
```

### Hints
- Write a small helper `get(key, def)` that treats `""` as unset
- Wrap parse errors: `fmt.Errorf("PORT: %w", err)`

### Solution
```go
package main

import (
	"fmt"
	"strconv"
)

type Config struct {
	Host  string
	Port  int
	Debug bool
}

func Load(getenv func(string) string) (Config, error) {
	get := func(k, def string) string {
		if v := getenv(k); v != "" {
			return v
		}
		return def
	}
	c := Config{Host: get("HOST", "localhost")}
	port, err := strconv.Atoi(get("PORT", "8080"))
	if err != nil {
		return Config{}, fmt.Errorf("PORT: %w", err)
	}
	if port < 1 || port > 65535 {
		return Config{}, fmt.Errorf("PORT: %d out of range", port)
	}
	c.Port = port
	dbg, err := strconv.ParseBool(get("DEBUG", "false"))
	if err != nil {
		return Config{}, fmt.Errorf("DEBUG: %w", err)
	}
	c.Debug = dbg
	return c, nil
}

func main() {}
```

### Explanation
Injecting `getenv` turns global process state into a plain function argument, so tests need no `os.Setenv` juggling.

## count-bytes
title: Count Bytes in a Stream
difficulty: intermediate
module: stdlib
lesson: stdlib/io
skill: stdlib

### Problem
Write `CountBytes(r io.Reader) (int64, error)` that reads `r` to the end, without holding all of it in memory, and returns the number of bytes read. If the reader fails part-way, return the count read so far **and** the error.

### Constraints
- Do not use `io.ReadAll`
- `io.EOF` is not an error

### Starter
```go
package main

import "io"

func CountBytes(r io.Reader) (int64, error) {
	return 0, nil
}

func main() {}
```

### Tests
```go
package main

import (
	"errors"
	"io"
	"strings"
	"testing"
	"testing/iotest"
)

func TestCountBytes(t *testing.T) {
	n, err := CountBytes(strings.NewReader(strings.Repeat("x", 100000)))
	if err != nil || n != 100000 {
		t.Errorf("got %d, %v", n, err)
	}
	n, err = CountBytes(strings.NewReader(""))
	if err != nil || n != 0 {
		t.Errorf("empty: %d, %v", n, err)
	}
}

func TestCountBytesError(t *testing.T) {
	boom := errors.New("boom")
	r := io.MultiReader(strings.NewReader("abc"), iotest.ErrReader(boom))
	n, err := CountBytes(r)
	if n != 3 || !errors.Is(err, boom) {
		t.Errorf("got %d, %v; want 3 and boom", n, err)
	}
}
```

### Hints
- `io.Copy(io.Discard, r)` returns the number of bytes copied and any error
- `io.Copy` already treats `io.EOF` as success

### Solution
```go
package main

import "io"

func CountBytes(r io.Reader) (int64, error) {
	return io.Copy(io.Discard, r)
}

func main() {}
```

### Explanation
`io.Copy` streams in chunks, returns the byte count even on failure, and hides `io.EOF` — one line does everything the challenge asks.

## word-count
title: Word Frequency
difficulty: intermediate
module: stdlib
lesson: stdlib/bufio
skill: stdlib

### Problem
Write `WordCount(r io.Reader) map[string]int` that counts how many times each word appears. Words are separated by whitespace, compared case-insensitively, and stripped of the punctuation characters `. , ! ? ; : " '` at both ends. Empty tokens are ignored.

### Constraints
- Use `bufio.Scanner` with `bufio.ScanWords`
- Do not read the whole input into memory

### Starter
```go
package main

import "io"

func WordCount(r io.Reader) map[string]int {
	return nil
}

func main() {}
```

### Tests
```go
package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestWordCount(t *testing.T) {
	in := "Go is fun. go is FAST!\n\"Go\", said Ann; 'is' it?"
	got := WordCount(strings.NewReader(in))
	want := map[string]int{"go": 3, "is": 3, "fun": 1, "fast": 1, "said": 1, "ann": 1, "it": 1}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v\nwant %v", got, want)
	}
	if got := WordCount(strings.NewReader(" ... !!! ")); len(got) != 0 {
		t.Errorf("punctuation-only input gave %v", got)
	}
}
```

### Hints
- `sc.Split(bufio.ScanWords)`
- `strings.Trim(w, ".,!?;:\"'")` then `strings.ToLower`

### Solution
```go
package main

import (
	"bufio"
	"io"
	"strings"
)

func WordCount(r io.Reader) map[string]int {
	counts := map[string]int{}
	sc := bufio.NewScanner(r)
	sc.Split(bufio.ScanWords)
	for sc.Scan() {
		w := strings.ToLower(strings.Trim(sc.Text(), ".,!?;:\"'"))
		if w != "" {
			counts[w]++
		}
	}
	return counts
}

func main() {}
```

### Explanation
`ScanWords` streams tokens, so memory use is independent of input size. Normalising each token before counting keeps the map clean.

## clean-paths
title: Safe Path Join
difficulty: advanced
module: stdlib
lesson: stdlib/filepath
skill: stdlib

### Problem
Write `SafeJoin(base, userPath string) (string, error)` that joins a user-supplied relative path onto `base` and returns the cleaned result — but returns an error if the result would escape `base` (path traversal such as `../../etc/passwd`). `base` itself is an acceptable result.

### Constraints
- Use `path/filepath`
- Absolute user paths are treated as relative to base
- `/srv/data-evil` is **not** inside `/srv/data`

### Starter
```go
package main

func SafeJoin(base, userPath string) (string, error) {
	return "", nil
}

func main() {}
```

### Tests
```go
package main

import "testing"

func TestSafeJoin(t *testing.T) {
	ok := map[string]string{
		"a/b.txt":      "/srv/data/a/b.txt",
		"./a//b.txt":   "/srv/data/a/b.txt",
		"a/../b.txt":   "/srv/data/b.txt",
		"/abs/file":    "/srv/data/abs/file",
		"":             "/srv/data",
		".":            "/srv/data",
	}
	for in, want := range ok {
		got, err := SafeJoin("/srv/data", in)
		if err != nil || got != want {
			t.Errorf("SafeJoin(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	for _, bad := range []string{"../etc/passwd", "a/../../etc", "../data-evil/x", "a/../.."} {
		if got, err := SafeJoin("/srv/data", bad); err == nil {
			t.Errorf("SafeJoin(%q) = %q, want error", bad, got)
		}
	}
}
```

### Hints
- `filepath.Clean(userPath)` resolves `.` and `..` lexically
- After cleaning, a relative path that is `..` or starts with `../` tries to escape
- Strip a leading `/` so absolute-looking input becomes relative to base

### Solution
```go
package main

import (
	"errors"
	"path/filepath"
	"strings"
)

func SafeJoin(base, userPath string) (string, error) {
	base = filepath.Clean(base)
	sep := string(filepath.Separator)
	rel := strings.TrimLeft(filepath.Clean(userPath), sep)
	if rel == ".." || strings.HasPrefix(rel, ".."+sep) {
		return "", errors.New("path escapes base directory")
	}
	joined := filepath.Join(base, rel)
	// Defence in depth: the result must still be inside base.
	if joined != base && !strings.HasPrefix(joined, base+sep) {
		return "", errors.New("path escapes base directory")
	}
	return joined, nil
}

func main() {}
```

### Explanation
Cleaning first collapses `.` and `a/..` pairs, so anything that still begins with `..` genuinely climbs out of the base and is rejected. The final prefix check (including the separator, so `/srv/data-evil` does not match `/srv/data`) is a second line of defence.

## context-timeout-call
title: Call With Timeout
difficulty: advanced
module: stdlib
lesson: stdlib/context
skill: stdlib

### Problem
Write `CallWithTimeout(d time.Duration, f func(ctx context.Context) (string, error)) (string, error)`.
It runs `f` with a context that expires after `d`. It returns `f`'s result if `f` finishes in time; otherwise it returns `"", context.DeadlineExceeded` **without waiting for `f` to return**. The goroutine running `f` must not leak when `f` returns after the timeout.

### Constraints
- Use `context.WithTimeout` and `defer cancel()`
- Run `f` in a goroutine and use a buffered result channel
- Return promptly on timeout

### Starter
```go
package main

import (
	"context"
	"time"
)

func CallWithTimeout(d time.Duration, f func(ctx context.Context) (string, error)) (string, error) {
	return "", nil
}

func main() {}
```

### Tests
```go
package main

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCallSucceeds(t *testing.T) {
	got, err := CallWithTimeout(time.Second, func(ctx context.Context) (string, error) { return "ok", nil })
	if err != nil || got != "ok" {
		t.Errorf("got %q, %v", got, err)
	}
	boom := errors.New("boom")
	_, err = CallWithTimeout(time.Second, func(ctx context.Context) (string, error) { return "", boom })
	if !errors.Is(err, boom) {
		t.Errorf("want boom, got %v", err)
	}
}

func TestCallTimesOut(t *testing.T) {
	release := make(chan struct{})
	defer close(release)
	start := time.Now()
	got, err := CallWithTimeout(30*time.Millisecond, func(ctx context.Context) (string, error) {
		<-release // ignores ctx: only returns when the test ends
		return "late", nil
	})
	if !errors.Is(err, context.DeadlineExceeded) || got != "" {
		t.Errorf("got %q, %v", got, err)
	}
	if time.Since(start) > time.Second {
		t.Error("did not return promptly on timeout")
	}
}

func TestContextPassedToFunc(t *testing.T) {
	_, err := CallWithTimeout(20*time.Millisecond, func(ctx context.Context) (string, error) {
		<-ctx.Done()
		return "", ctx.Err()
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("got %v", err)
	}
}
```

### Hints
- The result channel needs capacity 1 so the goroutine can always send and exit
- `select` on `ctx.Done()` and the result channel

### Solution
```go
package main

import (
	"context"
	"time"
)

func CallWithTimeout(d time.Duration, f func(ctx context.Context) (string, error)) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), d)
	defer cancel()
	type result struct {
		s   string
		err error
	}
	ch := make(chan result, 1) // buffered: the goroutine never blocks, so it can't leak
	go func() {
		s, err := f(ctx)
		ch <- result{s, err}
	}()
	select {
	case r := <-ch:
		return r.s, r.err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func main() {}
```

### Explanation
The buffered channel is the key: if we time out and stop receiving, the goroutine can still deposit its result and exit. An unbuffered channel would leave it blocked forever — a classic goroutine leak.

## safe-counter
title: Concurrency-Safe Counter
difficulty: intermediate
module: stdlib
lesson: stdlib/sync
skill: stdlib

### Problem
Implement a `Counter` type that can be used from many goroutines: `Inc(key string)`, `Get(key string) int`, and `Snapshot() map[string]int` (a copy that callers may modify freely). The zero value must be ready to use.

### Constraints
- Zero value of `Counter` must work (no constructor)
- Must pass the race detector
- `Snapshot` returns a copy

### Starter
```go
package main

type Counter struct {
}

func (c *Counter) Inc(key string) {
}

func (c *Counter) Get(key string) int {
	return 0
}

func (c *Counter) Snapshot() map[string]int {
	return nil
}

func main() {}
```

### Tests
```go
package main

import (
	"sync"
	"testing"
)

func TestCounterConcurrent(t *testing.T) {
	var c Counter
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				c.Inc("a")
				if j%2 == 0 {
					c.Inc("b")
				}
				_ = c.Get("a")
			}
		}()
	}
	wg.Wait()
	if c.Get("a") != 10000 || c.Get("b") != 5000 || c.Get("missing") != 0 {
		t.Errorf("a=%d b=%d", c.Get("a"), c.Get("b"))
	}
}

func TestSnapshotIsCopy(t *testing.T) {
	var c Counter
	c.Inc("x")
	s := c.Snapshot()
	s["x"] = 99
	s["y"] = 1
	if c.Get("x") != 1 || c.Get("y") != 0 {
		t.Error("Snapshot must return a copy")
	}
	var empty Counter
	if s := empty.Snapshot(); s == nil || len(s) != 0 {
		t.Errorf("empty snapshot should be a non-nil empty map, got %v", s)
	}
}
```

### Hints
- Embed a `sync.Mutex` (or `RWMutex`) in the struct
- Lazily allocate the map in `Inc`

### Solution
```go
package main

import "sync"

type Counter struct {
	mu sync.Mutex
	m  map[string]int
}

func (c *Counter) Inc(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.m == nil {
		c.m = make(map[string]int)
	}
	c.m[key]++
}

func (c *Counter) Get(key string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.m[key]
}

func (c *Counter) Snapshot() map[string]int {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make(map[string]int, len(c.m))
	for k, v := range c.m {
		out[k] = v
	}
	return out
}

func main() {}
```

### Explanation
A mutex plus a lazily created map gives a useful zero value. Snapshot copies under the lock so callers can't touch shared state.
