# Core Language Challenges

## fizzbuzz
title: FizzBuzz
difficulty: beginner
module: control-flow
lesson: control-flow/for-loops
skill: functions

### Problem
Write `FizzBuzz(n int) []string` returning the numbers 1 to `n` as strings, except: multiples of 3 become `"Fizz"`, multiples of 5 become `"Buzz"`, and multiples of both become `"FizzBuzz"`. For `n <= 0` return an empty slice.

### Constraints
- `n` can be 0 or negative
- Return strings, not ints

### Starter
```go
package main

func FizzBuzz(n int) []string {
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

func TestFizzBuzz(t *testing.T) {
	got := FizzBuzz(15)
	want := []string{"1", "2", "Fizz", "4", "Buzz", "Fizz", "7", "8", "Fizz", "Buzz", "11", "Fizz", "13", "14", "FizzBuzz"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v", got)
	}
	if len(FizzBuzz(0)) != 0 || len(FizzBuzz(-3)) != 0 {
		t.Error("non-positive n must give an empty result")
	}
}
```

### Hints
- Check divisibility by 15 (or both 3 and 5) first
- `strconv.Itoa` converts an int to a string

### Solution
```go
package main

import "strconv"

func FizzBuzz(n int) []string {
	var out []string
	for i := 1; i <= n; i++ {
		switch {
		case i%15 == 0:
			out = append(out, "FizzBuzz")
		case i%3 == 0:
			out = append(out, "Fizz")
		case i%5 == 0:
			out = append(out, "Buzz")
		default:
			out = append(out, strconv.Itoa(i))
		}
	}
	return out
}

func main() {}
```

### Explanation
A tagless `switch` reads top to bottom, so the most specific case (`%15`) must come first.

## collatz-steps
title: Collatz Steps
difficulty: beginner
module: control-flow
lesson: control-flow/infinite-loops
skill: functions

### Problem
The Collatz process: if `n` is even, halve it; otherwise replace it with `3n+1`. Repeat until `n` is 1. Write `CollatzSteps(n int) int` returning how many steps it takes. `CollatzSteps(1)` is 0. For `n < 1` return -1.

### Constraints
- Use a loop with a clear exit
- `n < 1` returns -1

### Starter
```go
package main

func CollatzSteps(n int) int {
	return 0
}

func main() {}
```

### Tests
```go
package main

import "testing"

func TestCollatz(t *testing.T) {
	tests := map[int]int{1: 0, 2: 1, 3: 7, 6: 8, 27: 111, 0: -1, -5: -1}
	for in, want := range tests {
		if got := CollatzSteps(in); got != want {
			t.Errorf("CollatzSteps(%d) = %d, want %d", in, got, want)
		}
	}
}
```

### Hints
- `for n != 1 { ... steps++ }`
- Handle the invalid input before looping

### Solution
```go
package main

func CollatzSteps(n int) int {
	if n < 1 {
		return -1
	}
	steps := 0
	for n != 1 {
		if n%2 == 0 {
			n /= 2
		} else {
			n = 3*n + 1
		}
		steps++
	}
	return steps
}

func main() {}
```

### Explanation
The loop condition is the exit path. Validating input first avoids an endless loop for `n <= 0`.

## variadic-stats
title: Variadic Statistics
difficulty: beginner
module: control-flow
lesson: control-flow/variadic
skill: functions

### Problem
Write two variadic functions: `Sum(nums ...int) int` and `Average(nums ...int) (float64, bool)`. `Average` returns `false` when called with no arguments.

### Constraints
- `Sum()` with no arguments is 0

### Starter
```go
package main

func Sum(nums ...int) int {
	return 0
}

func Average(nums ...int) (float64, bool) {
	return 0, false
}

func main() {}
```

### Tests
```go
package main

import "testing"

func TestSum(t *testing.T) {
	if Sum() != 0 || Sum(5) != 5 || Sum(1, 2, 3) != 6 {
		t.Error("Sum wrong")
	}
	xs := []int{4, 5, 6}
	if Sum(xs...) != 15 {
		t.Error("Sum with spread slice wrong")
	}
}

func TestAverage(t *testing.T) {
	if avg, ok := Average(1, 2, 3, 4); !ok || avg != 2.5 {
		t.Errorf("got %v %v", avg, ok)
	}
	if _, ok := Average(); ok {
		t.Error("Average() should report !ok")
	}
}
```

### Hints
- Inside the function `nums` is a `[]int`
- Reuse `Sum` in `Average`

### Solution
```go
package main

func Sum(nums ...int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

func Average(nums ...int) (float64, bool) {
	if len(nums) == 0 {
		return 0, false
	}
	return float64(Sum(nums...)) / float64(len(nums)), true
}

func main() {}
```

### Explanation
`nums...` forwards the slice into another variadic call. Returning `(value, ok)` avoids a division by zero.

## make-counter
title: Counter Generator
difficulty: beginner
module: control-flow
lesson: control-flow/anonymous-functions
skill: functions

### Problem
Write `MakeCounter(start, step int) func() int`. Each call to the returned function returns the current value and then advances it by `step`. Separate counters must not share state.

### Constraints
- Use a closure
- No package-level variables

### Starter
```go
package main

func MakeCounter(start, step int) func() int {
	return func() int { return 0 }
}

func main() {}
```

### Tests
```go
package main

import "testing"

func TestCounter(t *testing.T) {
	c := MakeCounter(10, 5)
	for _, want := range []int{10, 15, 20} {
		if got := c(); got != want {
			t.Errorf("got %d, want %d", got, want)
		}
	}
	a, b := MakeCounter(0, 1), MakeCounter(100, -1)
	a()
	a()
	if got := b(); got != 100 {
		t.Errorf("counters share state: b() = %d", got)
	}
	if got := a(); got != 2 {
		t.Errorf("a() = %d, want 2", got)
	}
}
```

### Hints
- Declare a local variable in `MakeCounter` and modify it inside the returned function
- Return the value *before* adding the step

### Solution
```go
package main

func MakeCounter(start, step int) func() int {
	n := start
	return func() int {
		v := n
		n += step
		return v
	}
}

func main() {}
```

### Explanation
Each call to `MakeCounter` creates a new `n`, captured by its own closure, so counters are independent.

## min-max
title: Min and Max
difficulty: beginner
module: control-flow
lesson: control-flow/multiple-returns
skill: functions

### Problem
Write `MinMax(xs []int) (min, max int, err error)` returning the smallest and largest values. For an empty slice return a non-nil error.

### Constraints
- Single pass over the slice

### Starter
```go
package main

func MinMax(xs []int) (min, max int, err error) {
	return
}

func main() {}
```

### Tests
```go
package main

import "testing"

func TestMinMax(t *testing.T) {
	mn, mx, err := MinMax([]int{3, -1, 9, 4})
	if err != nil || mn != -1 || mx != 9 {
		t.Errorf("got %d %d %v", mn, mx, err)
	}
	mn, mx, err = MinMax([]int{7})
	if err != nil || mn != 7 || mx != 7 {
		t.Errorf("single: %d %d %v", mn, mx, err)
	}
	if _, _, err = MinMax(nil); err == nil {
		t.Error("expected an error for empty input")
	}
}
```

### Hints
- Start `min` and `max` from the first element
- `errors.New("empty slice")` for the error

### Solution
```go
package main

import "errors"

func MinMax(xs []int) (min, max int, err error) {
	if len(xs) == 0 {
		return 0, 0, errors.New("empty slice")
	}
	min, max = xs[0], xs[0]
	for _, x := range xs[1:] {
		if x < min {
			min = x
		}
		if x > max {
			max = x
		}
	}
	return min, max, nil
}

func main() {}
```

### Explanation
Named results document the return values; seeding from `xs[0]` avoids inventing sentinel values.

## primes-up-to
title: Primes Up To N
difficulty: beginner
module: control-flow
lesson: control-flow/break
skill: functions

### Problem
Write `PrimesUpTo(n int) []int` returning all primes ≤ `n` in ascending order. Use `break` to stop testing a candidate as soon as a divisor is found (or use `continue` on the outer loop).

### Constraints
- `n < 2` returns an empty slice

### Starter
```go
package main

func PrimesUpTo(n int) []int {
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

func TestPrimes(t *testing.T) {
	if got := PrimesUpTo(30); !reflect.DeepEqual(got, []int{2, 3, 5, 7, 11, 13, 17, 19, 23, 29}) {
		t.Errorf("got %v", got)
	}
	if got := PrimesUpTo(2); !reflect.DeepEqual(got, []int{2}) {
		t.Errorf("got %v", got)
	}
	if len(PrimesUpTo(1)) != 0 || len(PrimesUpTo(-4)) != 0 {
		t.Error("expected none")
	}
}
```

### Hints
- Test divisors `d` from 2 while `d*d <= candidate`
- A found divisor means "not prime" — stop searching

### Solution
```go
package main

func PrimesUpTo(n int) []int {
	var primes []int
	for c := 2; c <= n; c++ {
		isPrime := true
		for d := 2; d*d <= c; d++ {
			if c%d == 0 {
				isPrime = false
				break
			}
		}
		if isPrime {
			primes = append(primes, c)
		}
	}
	return primes
}

func main() {}
```

### Explanation
Checking divisors only up to √c and breaking at the first hit keeps the inner loop short.

## largest-number
title: Largest Number
difficulty: beginner
module: data-structures
lesson: data-structures/slices
skill: data-structures

### Problem
Create a function that returns the largest number in a slice of integers: `Largest(xs []int) (int, error)`. An empty slice returns an error.

### Constraints
- Negative numbers are allowed
- Do not modify the slice

### Starter
```go
package main

func Largest(xs []int) (int, error) {
	return 0, nil
}

func main() {}
```

### Tests
```go
package main

import "testing"

func TestLargest(t *testing.T) {
	tests := []struct {
		in   []int
		want int
	}{
		{[]int{3, 9, 4}, 9},
		{[]int{-5, -2, -9}, -2},
		{[]int{7}, 7},
		{[]int{1, 1, 1}, 1},
	}
	for _, tc := range tests {
		got, err := Largest(tc.in)
		if err != nil || got != tc.want {
			t.Errorf("Largest(%v) = %d, %v; want %d", tc.in, got, err, tc.want)
		}
	}
	if _, err := Largest(nil); err == nil {
		t.Error("expected error for empty slice")
	}
}
```

### Hints
- Start with the first element, not 0 (all values might be negative)
- Return an error before indexing an empty slice

### Solution
```go
package main

import "errors"

func Largest(xs []int) (int, error) {
	if len(xs) == 0 {
		return 0, errors.New("empty slice")
	}
	largest := xs[0]
	for _, x := range xs[1:] {
		if x > largest {
			largest = x
		}
	}
	return largest, nil
}

func main() {}
```

### Explanation
Seeding with `xs[0]` (rather than 0) makes negative-only input work, and the empty check prevents an index-out-of-range panic.

## rotate-slice
title: Rotate a Slice
difficulty: intermediate
module: data-structures
lesson: data-structures/copy
skill: data-structures

### Problem
Write `Rotate(xs []int, k int) []int` returning a **new** slice rotated to the right by `k` positions: `Rotate([1 2 3 4 5], 2)` is `[4 5 1 2 3]`. `k` may be larger than the length or negative (negative rotates left). The input must not change.

### Constraints
- Return a new slice; leave `xs` untouched
- Empty input returns an empty slice

### Starter
```go
package main

func Rotate(xs []int, k int) []int {
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

func TestRotate(t *testing.T) {
	in := []int{1, 2, 3, 4, 5}
	tests := []struct {
		k    int
		want []int
	}{
		{0, []int{1, 2, 3, 4, 5}},
		{2, []int{4, 5, 1, 2, 3}},
		{5, []int{1, 2, 3, 4, 5}},
		{7, []int{4, 5, 1, 2, 3}},
		{-1, []int{2, 3, 4, 5, 1}},
		{-6, []int{2, 3, 4, 5, 1}},
	}
	for _, tc := range tests {
		if got := Rotate(in, tc.k); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("Rotate(k=%d) = %v, want %v", tc.k, got, tc.want)
		}
	}
	if !reflect.DeepEqual(in, []int{1, 2, 3, 4, 5}) {
		t.Error("input modified")
	}
	if got := Rotate(nil, 3); len(got) != 0 {
		t.Errorf("nil: %v", got)
	}
}
```

### Hints
- Normalise `k` with `((k % n) + n) % n`
- `copy(out, xs[n-k:])` then `copy(out[k:], xs[:n-k])`

### Solution
```go
package main

func Rotate(xs []int, k int) []int {
	n := len(xs)
	out := make([]int, n)
	if n == 0 {
		return out
	}
	k = ((k % n) + n) % n
	copy(out, xs[n-k:])
	copy(out[k:], xs[:n-k])
	return out
}

func main() {}
```

### Explanation
Two `copy` calls place the tail first and the head after it. Normalising `k` handles large and negative rotations uniformly.

## first-unique-char
title: First Unique Character
difficulty: intermediate
module: data-structures
lesson: data-structures/maps
skill: data-structures

### Problem
Write `FirstUnique(s string) int` returning the **byte index** of the first character (rune) that appears exactly once in `s`, or `-1` if none does.

### Constraints
- Must handle multi-byte UTF-8 characters
- Return the byte offset, as `range` reports it

### Starter
```go
package main

func FirstUnique(s string) int {
	return 0
}

func main() {}
```

### Tests
```go
package main

import "testing"

func TestFirstUnique(t *testing.T) {
	tests := map[string]int{
		"leetcode":  0,
		"loveleet":  1,
		"aabb":      -1,
		"":          -1,
		"aabbc":     4,
		"ééa":       4,
		"éaé":       2,
		"日本日":       3,
	}
	for in, want := range tests {
		if got := FirstUnique(in); got != want {
			t.Errorf("FirstUnique(%q) = %d, want %d", in, got, want)
		}
	}
}
```

### Hints
- First pass: count runes in a `map[rune]int`
- Second pass: `range` again and return the first index with count 1

### Solution
```go
package main

func FirstUnique(s string) int {
	counts := map[rune]int{}
	for _, r := range s {
		counts[r]++
	}
	for i, r := range s {
		if counts[r] == 1 {
			return i
		}
	}
	return -1
}

func main() {}
```

### Explanation
Ranging over a string yields byte offsets and runes, so the same loop handles ASCII and multi-byte text correctly.

## group-anagrams
title: Group Anagrams
difficulty: intermediate
module: data-structures
lesson: data-structures/map-operations
skill: data-structures

### Problem
Write `GroupAnagrams(words []string) [][]string` that groups words which are anagrams of each other. Within each group keep the input order. Order the groups by the position of their first word in the input.

### Constraints
- Words contain lowercase ASCII letters only
- Do not modify the input

### Starter
```go
package main

func GroupAnagrams(words []string) [][]string {
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

func TestGroupAnagrams(t *testing.T) {
	got := GroupAnagrams([]string{"eat", "tea", "tan", "ate", "nat", "bat"})
	want := [][]string{{"eat", "tea", "ate"}, {"tan", "nat"}, {"bat"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
	if got := GroupAnagrams(nil); len(got) != 0 {
		t.Errorf("nil input: %v", got)
	}
	got = GroupAnagrams([]string{"a", "a"})
	if !reflect.DeepEqual(got, [][]string{{"a", "a"}}) {
		t.Errorf("duplicates: %v", got)
	}
}
```

### Hints
- The sorted letters of a word are a canonical key
- Track first-seen order of keys in a separate slice

### Solution
```go
package main

import "sort"

func GroupAnagrams(words []string) [][]string {
	index := map[string]int{}
	var groups [][]string
	for _, w := range words {
		b := []byte(w)
		sort.Slice(b, func(i, j int) bool { return b[i] < b[j] })
		key := string(b)
		i, ok := index[key]
		if !ok {
			i = len(groups)
			index[key] = i
			groups = append(groups, nil)
		}
		groups[i] = append(groups[i], w)
	}
	return groups
}

func main() {}
```

### Explanation
Maps give O(1) lookup by canonical key, and remembering each key's group index preserves deterministic output despite random map iteration.

## is-palindrome
title: Palindrome Check
difficulty: beginner
module: data-structures
lesson: data-structures/runes
skill: data-structures

### Problem
Write `IsPalindrome(s string) bool` that reports whether `s` reads the same forwards and backwards, considering only letters and digits and ignoring case. The empty string is a palindrome.

### Constraints
- Must work with non-ASCII letters
- Use runes, not bytes

### Starter
```go
package main

func IsPalindrome(s string) bool {
	return false
}

func main() {}
```

### Tests
```go
package main

import "testing"

func TestIsPalindrome(t *testing.T) {
	yes := []string{"", "a", "racecar", "A man, a plan, a canal: Panama", "Was it a car or a cat I saw?", "ÉtÉ", "12321"}
	no := []string{"ab", "hello", "Go lang", "12345"}
	for _, s := range yes {
		if !IsPalindrome(s) {
			t.Errorf("%q should be a palindrome", s)
		}
	}
	for _, s := range no {
		if IsPalindrome(s) {
			t.Errorf("%q should not be a palindrome", s)
		}
	}
}
```

### Hints
- Build a `[]rune` of `unicode.ToLower(r)` for letters and digits only
- Compare from both ends toward the middle

### Solution
```go
package main

import "unicode"

func IsPalindrome(s string) bool {
	var rs []rune
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			rs = append(rs, unicode.ToLower(r))
		}
	}
	for i, j := 0, len(rs)-1; i < j; i, j = i+1, j-1 {
		if rs[i] != rs[j] {
			return false
		}
	}
	return true
}

func main() {}
```

### Explanation
Filtering into a `[]rune` first makes the two-pointer comparison trivial and keeps multi-byte characters intact.

## run-length-encode
title: Run-Length Encoding
difficulty: intermediate
module: data-structures
lesson: data-structures/strings
skill: data-structures

### Problem
Write `RLE(s string) string` that compresses runs of the same character: `"aaabcc"` becomes `"a3b1c2"`. Each run is the character followed by its decimal count (including 1). Work on runes so multi-byte characters are not split.

### Constraints
- Empty input returns empty output
- Use `strings.Builder`

### Starter
```go
package main

func RLE(s string) string {
	return ""
}

func main() {}
```

### Tests
```go
package main

import "testing"

func TestRLE(t *testing.T) {
	tests := map[string]string{
		"":          "",
		"a":         "a1",
		"aaabcc":    "a3b1c2",
		"abc":       "a1b1c1",
		"zzzzzzzzzzzz": "z12",
		"ééa":       "é2a1",
		"日日日本":      "日3本1",
	}
	for in, want := range tests {
		if got := RLE(in); got != want {
			t.Errorf("RLE(%q) = %q, want %q", in, got, want)
		}
	}
}
```

### Hints
- Convert to `[]rune` and walk with an index
- `strconv.Itoa(count)` for the count

### Solution
```go
package main

import (
	"strconv"
	"strings"
)

func RLE(s string) string {
	rs := []rune(s)
	var b strings.Builder
	for i := 0; i < len(rs); {
		j := i
		for j < len(rs) && rs[j] == rs[i] {
			j++
		}
		b.WriteRune(rs[i])
		b.WriteString(strconv.Itoa(j - i))
		i = j
	}
	return b.String()
}

func main() {}
```

### Explanation
The inner loop finds the end of a run, so each character is visited once. A `Builder` avoids quadratic string concatenation.

## stack
title: Implement a Stack
difficulty: beginner
module: structs
lesson: structs/pointer-receivers
skill: structs

### Problem
Implement `type Stack struct` of ints with pointer-receiver methods `Push(int)`, `Pop() (int, bool)`, `Peek() (int, bool)` and `Len() int`. `Pop` and `Peek` return `false` when empty. The zero value must be a usable empty stack.

### Constraints
- Zero value must work
- LIFO order

### Starter
```go
package main

type Stack struct {
}

func (s *Stack) Push(v int) {
}

func (s *Stack) Pop() (int, bool) {
	return 0, false
}

func (s *Stack) Peek() (int, bool) {
	return 0, false
}

func (s *Stack) Len() int {
	return 0
}

func main() {}
```

### Tests
```go
package main

import "testing"

func TestStack(t *testing.T) {
	var s Stack
	if _, ok := s.Pop(); ok {
		t.Error("Pop on empty should fail")
	}
	if _, ok := s.Peek(); ok {
		t.Error("Peek on empty should fail")
	}
	s.Push(1)
	s.Push(2)
	s.Push(3)
	if s.Len() != 3 {
		t.Errorf("Len = %d", s.Len())
	}
	if v, ok := s.Peek(); !ok || v != 3 || s.Len() != 3 {
		t.Errorf("Peek = %d %v (len %d)", v, ok, s.Len())
	}
	for _, want := range []int{3, 2, 1} {
		if v, ok := s.Pop(); !ok || v != want {
			t.Errorf("Pop = %d %v, want %d", v, ok, want)
		}
	}
	if s.Len() != 0 {
		t.Error("stack should be empty")
	}
}
```

### Hints
- A struct with a `[]int` field is enough
- Pop: take the last element and shrink the slice

### Solution
```go
package main

type Stack struct {
	items []int
}

func (s *Stack) Push(v int) { s.items = append(s.items, v) }

func (s *Stack) Pop() (int, bool) {
	if len(s.items) == 0 {
		return 0, false
	}
	v := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return v, true
}

func (s *Stack) Peek() (int, bool) {
	if len(s.items) == 0 {
		return 0, false
	}
	return s.items[len(s.items)-1], true
}

func (s *Stack) Len() int { return len(s.items) }

func main() {}
```

### Explanation
Pointer receivers are required because `Push` and `Pop` change the slice header. A nil slice is a valid empty stack, so no constructor is needed.

## bank-account
title: Bank Account
difficulty: intermediate
module: structs
lesson: structs/methods
skill: structs

### Problem
Implement an `Account` type in cents with `Deposit(cents int) error`, `Withdraw(cents int) error` and `Balance() int`. Both mutators reject non-positive amounts with `ErrInvalidAmount`; `Withdraw` also fails with `ErrInsufficientFunds` if the balance would go negative. Failed operations must not change the balance. `ErrInvalidAmount` and `ErrInsufficientFunds` are declared for you.

### Constraints
- Failed operations leave the balance unchanged
- Return the sentinel errors so callers can use `errors.Is`

### Starter
```go
package main

import "errors"

var (
	ErrInvalidAmount     = errors.New("invalid amount")
	ErrInsufficientFunds = errors.New("insufficient funds")
)

type Account struct {
}

func (a *Account) Deposit(cents int) error {
	return nil
}

func (a *Account) Withdraw(cents int) error {
	return nil
}

func (a *Account) Balance() int {
	return 0
}

func main() {}
```

### Tests
```go
package main

import (
	"errors"
	"testing"
)

func TestAccount(t *testing.T) {
	var a Account
	if err := a.Deposit(1000); err != nil || a.Balance() != 1000 {
		t.Fatalf("deposit: %v %d", err, a.Balance())
	}
	if err := a.Withdraw(300); err != nil || a.Balance() != 700 {
		t.Fatalf("withdraw: %v %d", err, a.Balance())
	}
	if err := a.Withdraw(701); !errors.Is(err, ErrInsufficientFunds) || a.Balance() != 700 {
		t.Errorf("overdraft: %v %d", err, a.Balance())
	}
	for _, amt := range []int{0, -5} {
		if err := a.Deposit(amt); !errors.Is(err, ErrInvalidAmount) {
			t.Errorf("Deposit(%d): %v", amt, err)
		}
		if err := a.Withdraw(amt); !errors.Is(err, ErrInvalidAmount) {
			t.Errorf("Withdraw(%d): %v", amt, err)
		}
	}
	if a.Balance() != 700 {
		t.Errorf("balance changed by failed operations: %d", a.Balance())
	}
	if err := a.Withdraw(700); err != nil || a.Balance() != 0 {
		t.Errorf("withdraw all: %v %d", err, a.Balance())
	}
}
```

### Hints
- Validate before mutating
- Check `cents <= 0` first, then the balance

### Solution
```go
package main

import "errors"

var (
	ErrInvalidAmount     = errors.New("invalid amount")
	ErrInsufficientFunds = errors.New("insufficient funds")
)

type Account struct {
	balance int
}

func (a *Account) Deposit(cents int) error {
	if cents <= 0 {
		return ErrInvalidAmount
	}
	a.balance += cents
	return nil
}

func (a *Account) Withdraw(cents int) error {
	if cents <= 0 {
		return ErrInvalidAmount
	}
	if cents > a.balance {
		return ErrInsufficientFunds
	}
	a.balance -= cents
	return nil
}

func (a *Account) Balance() int { return a.balance }

func main() {}
```

### Explanation
Validating before mutating gives all-or-nothing operations. Money is stored in integer cents to avoid floating-point drift.

## reverse-linked-list
title: Reverse a Linked List
difficulty: intermediate
module: structs
lesson: structs/dereferencing
skill: pointers

### Problem
Given `type Node struct { Val int; Next *Node }`, write `Reverse(head *Node) *Node` that reverses a singly linked list **in place** and returns the new head. `Reverse(nil)` is `nil`.

### Constraints
- Reverse in place — do not allocate new nodes
- Handle nil and single-node lists

### Starter
```go
package main

type Node struct {
	Val  int
	Next *Node
}

func Reverse(head *Node) *Node {
	return head
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

func build(vals ...int) *Node {
	var head *Node
	for i := len(vals) - 1; i >= 0; i-- {
		head = &Node{vals[i], head}
	}
	return head
}

func toSlice(n *Node) []int {
	var out []int
	for ; n != nil; n = n.Next {
		out = append(out, n.Val)
	}
	return out
}

func TestReverse(t *testing.T) {
	if Reverse(nil) != nil {
		t.Error("Reverse(nil) should be nil")
	}
	one := build(1)
	if got := Reverse(one); got != one || got.Next != nil {
		t.Error("single node")
	}
	head := build(1, 2, 3, 4)
	first := head
	got := Reverse(head)
	if !reflect.DeepEqual(toSlice(got), []int{4, 3, 2, 1}) {
		t.Errorf("got %v", toSlice(got))
	}
	if first.Next != nil {
		t.Error("old head should now be the tail (must reuse nodes)")
	}
}
```

### Hints
- Keep `prev` and `cur` pointers
- Save `cur.Next` before you overwrite it

### Solution
```go
package main

type Node struct {
	Val  int
	Next *Node
}

func Reverse(head *Node) *Node {
	var prev *Node
	for cur := head; cur != nil; {
		next := cur.Next
		cur.Next = prev
		prev, cur = cur, next
	}
	return prev
}

func main() {}
```

### Explanation
Each step points the current node backwards; saving `next` first prevents losing the rest of the list.

## employee-payroll
title: Employee Payroll
difficulty: intermediate
module: structs
lesson: structs/struct-composition
skill: structs

### Problem
`Employee` has a `Name`, a `Salary` and a `Cost()` method. `Manager` **embeds** `Employee` and also has a slice of `Reports` (employees).

Implement `Employee.Summary()` returning `"<Name> (<Salary>)"` — `Manager` must get it for free through embedding — and `Manager.Cost()` returning the manager's own salary plus the `Cost()` of every report.

### Constraints
- `Manager` must embed `Employee`
- Use the promoted `Summary` — do not redefine it on `Manager`

### Starter
```go
package main

type Employee struct {
	Name   string
	Salary int
}

func (e Employee) Cost() int { return e.Salary }

func (e Employee) Summary() string {
	return ""
}

type Manager struct {
	Employee
	Reports []Employee
}

func (m Manager) Cost() int {
	return 0
}

func main() {}
```

### Tests
```go
package main

import "testing"

func TestPayroll(t *testing.T) {
	m := Manager{
		Employee: Employee{"Grace", 200},
		Reports:  []Employee{{"Ada", 100}, {"Bob", 90}},
	}
	if got := m.Cost(); got != 390 {
		t.Errorf("Manager.Cost() = %d, want 390", got)
	}
	if got := m.Summary(); got != "Grace (200)" {
		t.Errorf("Summary() = %q", got)
	}
	if got := m.Name; got != "Grace" {
		t.Errorf("promoted field Name = %q", got)
	}
	solo := Manager{Employee: Employee{"Solo", 50}}
	if solo.Cost() != 50 {
		t.Errorf("no reports: %d", solo.Cost())
	}
	var e Employee = Employee{"Ada", 100}
	if e.Cost() != 100 || e.Summary() != "Ada (100)" {
		t.Error("Employee methods broken")
	}
}
```

### Hints
- `type Manager struct { Employee; Reports []Employee }`
- Inside `Manager.Cost` call `m.Employee.Cost()` to reach the embedded method
- `fmt.Sprintf("%s (%d)", e.Name, e.Salary)`

### Solution
```go
package main

import "fmt"

type Employee struct {
	Name   string
	Salary int
}

func (e Employee) Cost() int { return e.Salary }

func (e Employee) Summary() string {
	return fmt.Sprintf("%s (%d)", e.Name, e.Salary)
}

type Manager struct {
	Employee
	Reports []Employee
}

func (m Manager) Cost() int {
	total := m.Employee.Cost()
	for _, r := range m.Reports {
		total += r.Cost()
	}
	return total
}

func main() {}
```

### Explanation
`Manager` gets `Name`, `Salary` and `Summary` by embedding, and overrides `Cost` while still reaching the embedded version via `m.Employee.Cost()`.

## shapes-total-area
title: Shapes and Total Area
difficulty: beginner
module: interfaces
lesson: interfaces/what-are-interfaces
skill: interfaces

### Problem
`Shape` is an interface with `Area() float64` and `Perimeter() float64`. Fill in the `Area` and `Perimeter` methods of `Rect{W, H float64}` and `Circle{R float64}` (they already satisfy `Shape` but return 0), and write `TotalArea(shapes []Shape) float64`. Use `math.Pi` for circles.

### Constraints
- Use value receivers
- `TotalArea(nil)` is 0

### Starter
```go
package main

type Shape interface {
	Area() float64
	Perimeter() float64
}

type Rect struct{ W, H float64 }

func (r Rect) Area() float64      { return 0 }
func (r Rect) Perimeter() float64 { return 0 }

type Circle struct{ R float64 }

func (c Circle) Area() float64      { return 0 }
func (c Circle) Perimeter() float64 { return 0 }

func TotalArea(shapes []Shape) float64 {
	return 0
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

func TestShapes(t *testing.T) {
	var r Shape = Rect{3, 4}
	var c Shape = Circle{2}
	if !near(r.Area(), 12) || !near(r.Perimeter(), 14) {
		t.Errorf("rect: %v %v", r.Area(), r.Perimeter())
	}
	if !near(c.Area(), math.Pi*4) || !near(c.Perimeter(), 4*math.Pi) {
		t.Errorf("circle: %v %v", c.Area(), c.Perimeter())
	}
	if got := TotalArea([]Shape{r, c}); !near(got, 12+4*math.Pi) {
		t.Errorf("total = %v", got)
	}
	if TotalArea(nil) != 0 {
		t.Error("TotalArea(nil) should be 0")
	}
}
```

### Hints
- Circle perimeter is `2*math.Pi*R`
- `TotalArea` just loops and calls `Area()`

### Solution
```go
package main

import "math"

type Shape interface {
	Area() float64
	Perimeter() float64
}

type Rect struct{ W, H float64 }

func (r Rect) Area() float64      { return r.W * r.H }
func (r Rect) Perimeter() float64 { return 2 * (r.W + r.H) }

type Circle struct{ R float64 }

func (c Circle) Area() float64      { return math.Pi * c.R * c.R }
func (c Circle) Perimeter() float64 { return 2 * math.Pi * c.R }

func TotalArea(shapes []Shape) float64 {
	total := 0.0
	for _, s := range shapes {
		total += s.Area()
	}
	return total
}

func main() {}
```

### Explanation
`TotalArea` works with any future shape without change — that's the point of coding against an interface.

## describe-any
title: Describe Any Value
difficulty: intermediate
module: interfaces
lesson: interfaces/type-switches
skill: interfaces

### Problem
Write `Describe(v any) string` using a type switch:
- `int` → `"int:<n>"`
- `string` → `"string:<len>"` where `<len>` is the number of runes
- `bool` → `"bool:yes"` or `"bool:no"`
- `[]int` → `"ints:<len>"`
- `nil` → `"nil"`
- anything else → `"other:<type>"` (using `%T`)

### Constraints
- Use a type switch
- Count runes, not bytes, for strings

### Starter
```go
package main

func Describe(v any) string {
	return ""
}

func main() {}
```

### Tests
```go
package main

import "testing"

func TestDescribe(t *testing.T) {
	tests := []struct {
		in   any
		want string
	}{
		{42, "int:42"},
		{"héllo", "string:5"},
		{true, "bool:yes"},
		{false, "bool:no"},
		{[]int{1, 2, 3}, "ints:3"},
		{nil, "nil"},
		{3.5, "other:float64"},
		{map[string]int{}, "other:map[string]int"},
	}
	for _, tc := range tests {
		if got := Describe(tc.in); got != tc.want {
			t.Errorf("Describe(%v) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
```

### Hints
- `switch x := v.(type)`
- `utf8.RuneCountInString(x)` counts runes

### Solution
```go
package main

import (
	"fmt"
	"unicode/utf8"
)

func Describe(v any) string {
	switch x := v.(type) {
	case nil:
		return "nil"
	case int:
		return fmt.Sprintf("int:%d", x)
	case string:
		return fmt.Sprintf("string:%d", utf8.RuneCountInString(x))
	case bool:
		if x {
			return "bool:yes"
		}
		return "bool:no"
	case []int:
		return fmt.Sprintf("ints:%d", len(x))
	default:
		return fmt.Sprintf("other:%T", v)
	}
}

func main() {}
```

### Explanation
Inside each case `x` has that case's concrete type, so no separate assertion is needed.

## notifier-fake
title: Notify With a Fake
difficulty: intermediate
module: interfaces
lesson: interfaces/interfaces-for-testing
skill: interfaces

### Problem
Write `NotifyAll(s Sender, users []string, msg string) (sent int, err error)` where `Sender` is `interface{ Send(to, msg string) error }`. It sends `msg` to every user. If a send fails, keep going with the remaining users, and at the end return how many succeeded and an error describing the failures (`nil` if none). The error message must list the failed recipients separated by commas.

### Constraints
- Do not stop at the first failure
- Failed recipients appear in the error text, in input order

### Starter
```go
package main

type Sender interface {
	Send(to, msg string) error
}

func NotifyAll(s Sender, users []string, msg string) (sent int, err error) {
	return 0, nil
}

func main() {}
```

### Tests
```go
package main

import (
	"errors"
	"strings"
	"testing"
)

type fake struct {
	calls []string
	fail  map[string]bool
}

func (f *fake) Send(to, msg string) error {
	f.calls = append(f.calls, to+":"+msg)
	if f.fail[to] {
		return errors.New("boom")
	}
	return nil
}

func TestNotifyAll(t *testing.T) {
	f := &fake{}
	n, err := NotifyAll(f, []string{"a", "b"}, "hi")
	if n != 2 || err != nil || strings.Join(f.calls, ",") != "a:hi,b:hi" {
		t.Errorf("all ok: %d %v %v", n, err, f.calls)
	}
}

func TestNotifyAllPartialFailure(t *testing.T) {
	f := &fake{fail: map[string]bool{"b": true, "d": true}}
	n, err := NotifyAll(f, []string{"a", "b", "c", "d"}, "x")
	if n != 2 {
		t.Errorf("sent = %d, want 2", n)
	}
	if err == nil || !strings.Contains(err.Error(), "b, d") {
		t.Errorf("error should list b, d: %v", err)
	}
	if len(f.calls) != 4 {
		t.Errorf("should attempt every user, made %d calls", len(f.calls))
	}
}

func TestNotifyAllNobody(t *testing.T) {
	n, err := NotifyAll(&fake{}, nil, "x")
	if n != 0 || err != nil {
		t.Errorf("%d %v", n, err)
	}
}
```

### Hints
- Collect failed recipients in a `[]string`
- `strings.Join(failed, ", ")` builds the message

### Solution
```go
package main

import (
	"fmt"
	"strings"
)

type Sender interface {
	Send(to, msg string) error
}

func NotifyAll(s Sender, users []string, msg string) (sent int, err error) {
	var failed []string
	for _, u := range users {
		if e := s.Send(u, msg); e != nil {
			failed = append(failed, u)
			continue
		}
		sent++
	}
	if len(failed) > 0 {
		err = fmt.Errorf("failed to notify: %s", strings.Join(failed, ", "))
	}
	return sent, err
}

func main() {}
```

### Explanation
Because `NotifyAll` depends only on the small `Sender` interface, the test can pass a hand-written fake that records calls and injects failures.

## parse-age
title: Parse an Age
difficulty: beginner
module: errors
lesson: errors/error-wrapping
skill: errors

### Problem
Write `ParseAge(s string) (int, error)`. Valid ages are integers 0–150. If `s` is not a number, return an error that wraps both `ErrInvalidAge` and the underlying parse error (use `%w` twice, Go 1.20+). If it is a number outside the range, return an error wrapping `ErrInvalidAge` whose message contains the number. `ErrInvalidAge` is provided.

### Constraints
- `errors.Is(err, ErrInvalidAge)` must be true for every failure
- Wrap with `%w`

### Starter
```go
package main

import "errors"

var ErrInvalidAge = errors.New("invalid age")

func ParseAge(s string) (int, error) {
	return 0, nil
}

func main() {}
```

### Tests
```go
package main

import (
	"errors"
	"strconv"
	"strings"
	"testing"
)

func TestParseAge(t *testing.T) {
	for in, want := range map[string]int{"0": 0, "42": 42, "150": 150} {
		if got, err := ParseAge(in); err != nil || got != want {
			t.Errorf("ParseAge(%q) = %d, %v", in, got, err)
		}
	}
}

func TestParseAgeErrors(t *testing.T) {
	_, err := ParseAge("abc")
	if !errors.Is(err, ErrInvalidAge) || !errors.Is(err, strconv.ErrSyntax) {
		t.Errorf("abc: %v", err)
	}
	for _, in := range []string{"-1", "151"} {
		_, err := ParseAge(in)
		if !errors.Is(err, ErrInvalidAge) || !strings.Contains(err.Error(), in) {
			t.Errorf("%s: %v", in, err)
		}
	}
}
```

### Hints
- `fmt.Errorf("%w: %w", ErrInvalidAge, err)` wraps two errors
- `strconv.Atoi` returns `*strconv.NumError` wrapping `strconv.ErrSyntax`

### Solution
```go
package main

import (
	"errors"
	"fmt"
	"strconv"
)

var ErrInvalidAge = errors.New("invalid age")

func ParseAge(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrInvalidAge, err)
	}
	if n < 0 || n > 150 {
		return 0, fmt.Errorf("%w: %d is out of range", ErrInvalidAge, n)
	}
	return n, nil
}

func main() {}
```

### Explanation
Since Go 1.20 `fmt.Errorf` accepts multiple `%w` verbs, so callers can match on your sentinel *and* still reach the low-level cause.

## validation-error
title: Structured Validation Error
difficulty: intermediate
module: errors
lesson: errors/custom-errors
skill: errors

### Problem
Define `type ValidationError struct { Field, Reason string }` with an `Error()` method returning `"<Field>: <Reason>"` (pointer receiver). Write `Validate(name string, age int) error` returning a `*ValidationError` for the first problem: empty name (`Field: "name"`, `Reason: "required"`), or age outside 0–150 (`Field: "age"`, `Reason: "out of range"`). Return a true `nil` when valid.

### Constraints
- Return `nil`, not a typed nil pointer, when valid
- Checks run in order: name, then age

### Starter
```go
package main

type ValidationError struct {
	Field  string
	Reason string
}

func (e *ValidationError) Error() string {
	return ""
}

func Validate(name string, age int) error {
	return nil
}

func main() {}
```

### Tests
```go
package main

import (
	"errors"
	"fmt"
	"testing"
)

func TestValidate(t *testing.T) {
	if err := Validate("Ada", 36); err != nil {
		t.Fatalf("valid input returned %v", err)
	}
	var ve *ValidationError
	err := Validate("", 30)
	if !errors.As(err, &ve) || ve.Field != "name" || ve.Reason != "required" {
		t.Errorf("empty name: %#v", err)
	}
	if err.Error() != "name: required" {
		t.Errorf("message = %q", err.Error())
	}
	err = Validate("Bob", 200)
	if !errors.As(err, &ve) || ve.Field != "age" || err.Error() != "age: out of range" {
		t.Errorf("age: %#v", err)
	}
	wrapped := fmt.Errorf("signup: %w", Validate("", 1))
	if !errors.As(wrapped, &ve) || ve.Field != "name" {
		t.Error("errors.As must see through wrapping")
	}
}
```

### Hints
- `fmt.Sprintf("%s: %s", e.Field, e.Reason)`
- Return `&ValidationError{...}` for failures and plain `nil` otherwise

### Solution
```go
package main

import "fmt"

type ValidationError struct {
	Field  string
	Reason string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Reason)
}

func Validate(name string, age int) error {
	if name == "" {
		return &ValidationError{"name", "required"}
	}
	if age < 0 || age > 150 {
		return &ValidationError{"age", "out of range"}
	}
	return nil
}

func main() {}
```

### Explanation
A typed error carries machine-readable fields that `errors.As` can extract, while returning a literal `nil` avoids the typed-nil-in-interface trap.

## recover-to-error
title: Panic to Error
difficulty: intermediate
module: errors
lesson: errors/recover
skill: errors

### Problem
Write `Safe(f func()) (err error)` that runs `f`. If `f` panics, `Safe` returns an error whose message starts with `panic: ` followed by the panic value; if the panic value is itself an `error`, it must remain reachable through `errors.Is`. If `f` doesn't panic, return `nil`.

### Constraints
- Use `defer` and `recover`
- Runtime panics (nil map write, index out of range) must be caught too

### Starter
```go
package main

func Safe(f func()) (err error) {
	f()
	return nil
}

func main() {}
```

### Tests
```go
package main

import (
	"errors"
	"strings"
	"testing"
)

func TestSafe(t *testing.T) {
	if err := Safe(func() {}); err != nil {
		t.Errorf("no panic: %v", err)
	}
	err := Safe(func() { panic("boom") })
	if err == nil || err.Error() != "panic: boom" {
		t.Errorf("string panic: %v", err)
	}
	sentinel := errors.New("sentinel")
	err = Safe(func() { panic(sentinel) })
	if !errors.Is(err, sentinel) || !strings.HasPrefix(err.Error(), "panic: ") {
		t.Errorf("error panic: %v", err)
	}
	err = Safe(func() {
		var m map[string]int
		m["a"] = 1
	})
	if err == nil || !strings.Contains(err.Error(), "nil map") {
		t.Errorf("runtime panic: %v", err)
	}
	err = Safe(func() {
		xs := []int{1}
		i := 5
		_ = xs[i]
	})
	if err == nil || !strings.Contains(err.Error(), "index out of range") {
		t.Errorf("index panic: %v", err)
	}
}
```

### Hints
- `defer func() { if r := recover(); r != nil { ... } }()` assigns to the named result
- Use a type switch on `r`: `error` → `fmt.Errorf("panic: %w", e)`, otherwise `fmt.Errorf("panic: %v", r)`

### Solution
```go
package main

import "fmt"

func Safe(f func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = fmt.Errorf("panic: %w", e)
			} else {
				err = fmt.Errorf("panic: %v", r)
			}
		}
	}()
	f()
	return nil
}

func main() {}
```

### Explanation
Assigning to the named result inside the deferred closure is the only way to change what a function returns after a panic. Runtime errors implement `error`, so wrapping with `%w` preserves them.

## retry
title: Retry With Context
difficulty: intermediate
module: errors
lesson: errors/error-best-practices
skill: errors

### Problem
Write `Retry(attempts int, f func() error) error`. It calls `f` up to `attempts` times, stopping at the first success (return `nil`). If every attempt fails, return an error `after N attempts: <last error>` wrapping the last error with `%w`. If `attempts < 1`, return an error without calling `f`.

### Constraints
- No sleeping is needed for this challenge
- Wrap the last error with `%w`

### Starter
```go
package main

func Retry(attempts int, f func() error) error {
	return nil
}

func main() {}
```

### Tests
```go
package main

import (
	"errors"
	"testing"
)

func TestRetry(t *testing.T) {
	calls := 0
	err := Retry(3, func() error {
		calls++
		if calls < 3 {
			return errors.New("flaky")
		}
		return nil
	})
	if err != nil || calls != 3 {
		t.Errorf("flaky then ok: %v after %d calls", err, calls)
	}

	calls = 0
	if err := Retry(5, func() error { calls++; return nil }); err != nil || calls != 1 {
		t.Errorf("first success: %v after %d", err, calls)
	}

	boom := errors.New("boom")
	calls = 0
	err = Retry(3, func() error { calls++; return boom })
	if !errors.Is(err, boom) || calls != 3 {
		t.Errorf("always fails: %v after %d", err, calls)
	}
	if err == nil || err.Error() != "after 3 attempts: boom" {
		t.Errorf("message = %v", err)
	}

	calls = 0
	if err := Retry(0, func() error { calls++; return nil }); err == nil || calls != 0 {
		t.Errorf("zero attempts: %v after %d", err, calls)
	}
}
```

### Hints
- Keep the last error in a variable
- `fmt.Errorf("after %d attempts: %w", attempts, last)`

### Solution
```go
package main

import (
	"errors"
	"fmt"
)

func Retry(attempts int, f func() error) error {
	if attempts < 1 {
		return errors.New("attempts must be at least 1")
	}
	var last error
	for i := 0; i < attempts; i++ {
		if last = f(); last == nil {
			return nil
		}
	}
	return fmt.Errorf("after %d attempts: %w", attempts, last)
}

func main() {}
```

### Explanation
Wrapping the final error adds context (how many tries) without hiding the cause — callers can still `errors.Is` it.

## semver-compare
title: Compare Semantic Versions
difficulty: intermediate
module: modules
lesson: modules/version-management
skill: packages

### Problem
Write `Compare(a, b string) (int, error)` comparing two semantic versions of the form `vMAJOR.MINOR.PATCH` (the `v` prefix is required; MINOR and PATCH may be omitted and default to 0, so `v1.2` equals `v1.2.0`). Return -1, 0 or 1. Return an error for malformed versions.

### Constraints
- Numeric comparison, not string comparison (`v1.10.0 > v1.9.0`)
- Reject missing `v`, empty parts, non-numeric parts and more than three parts

### Starter
```go
package main

func Compare(a, b string) (int, error) {
	return 0, nil
}

func main() {}
```

### Tests
```go
package main

import "testing"

func TestCompare(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"v1.2.3", "v1.2.3", 0},
		{"v1.2.3", "v1.2.4", -1},
		{"v1.10.0", "v1.9.0", 1},
		{"v2.0.0", "v1.99.99", 1},
		{"v1.2", "v1.2.0", 0},
		{"v1", "v1.0.1", -1},
		{"v0.0.1", "v0.0.2", -1},
	}
	for _, tc := range tests {
		got, err := Compare(tc.a, tc.b)
		if err != nil || got != tc.want {
			t.Errorf("Compare(%q,%q) = %d, %v; want %d", tc.a, tc.b, got, err, tc.want)
		}
	}
	for _, bad := range []string{"1.2.3", "v", "v1..2", "v1.x.3", "v1.2.3.4", "", "v-1.0.0"} {
		if _, err := Compare(bad, "v1.0.0"); err == nil {
			t.Errorf("Compare(%q, ...) should fail", bad)
		}
		if _, err := Compare("v1.0.0", bad); err == nil {
			t.Errorf("Compare(..., %q) should fail", bad)
		}
	}
}
```

### Hints
- Write a `parse(string) ([3]int, error)` helper
- `strings.Split(strings.TrimPrefix(v, "v"), ".")`
- Reject negative numbers (`strconv.Atoi("-1")` succeeds!)

### Solution
```go
package main

import (
	"fmt"
	"strconv"
	"strings"
)

func parse(v string) ([3]int, error) {
	var out [3]int
	if !strings.HasPrefix(v, "v") {
		return out, fmt.Errorf("version %q must start with v", v)
	}
	parts := strings.Split(v[1:], ".")
	if len(parts) > 3 {
		return out, fmt.Errorf("version %q has too many parts", v)
	}
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 || strings.HasPrefix(p, "-") || strings.HasPrefix(p, "+") {
			return out, fmt.Errorf("version %q has an invalid part %q", v, p)
		}
		out[i] = n
	}
	return out, nil
}

func Compare(a, b string) (int, error) {
	pa, err := parse(a)
	if err != nil {
		return 0, err
	}
	pb, err := parse(b)
	if err != nil {
		return 0, err
	}
	for i := range pa {
		switch {
		case pa[i] < pb[i]:
			return -1, nil
		case pa[i] > pb[i]:
			return 1, nil
		}
	}
	return 0, nil
}

func main() {}
```

### Explanation
Parsing into a fixed `[3]int` makes omitted parts default to zero, and arrays compare lexicographically part by part. Comparing numbers (not strings) is what makes `v1.10.0` greater than `v1.9.0`.

## functional-options
title: Functional Options
difficulty: intermediate
module: modules
lesson: modules/package-design
skill: packages

### Problem
Configure a `Server` through functional options. Make `NewServer(addr string, opts ...Option) *Server` apply the defaults `Timeout = 30s`, `MaxConns = 100`, `TLS = false`, and implement the options `WithTimeout(time.Duration)`, `WithMaxConns(int)` and `WithTLS()` (currently no-ops). Later options override earlier ones. `WithMaxConns` ignores values below 1 (keeps the previous value).

### Constraints
- `Server` fields `Addr`, `Timeout`, `MaxConns`, `TLS` are exported
- No options → defaults

### Starter
```go
package main

import "time"

type Server struct {
	Addr     string
	Timeout  time.Duration
	MaxConns int
	TLS      bool
}

type Option func(*Server)

func WithTimeout(d time.Duration) Option { return func(s *Server) {} }
func WithMaxConns(n int) Option          { return func(s *Server) {} }
func WithTLS() Option                    { return func(s *Server) {} }

func NewServer(addr string, opts ...Option) *Server {
	return &Server{Addr: addr}
}

func main() {}
```

### Tests
```go
package main

import (
	"testing"
	"time"
)

func TestDefaults(t *testing.T) {
	s := NewServer(":80")
	if s.Addr != ":80" || s.Timeout != 30*time.Second || s.MaxConns != 100 || s.TLS {
		t.Errorf("defaults wrong: %+v", s)
	}
}

func TestOptions(t *testing.T) {
	s := NewServer(":443", WithTimeout(5*time.Second), WithMaxConns(10), WithTLS())
	if s.Timeout != 5*time.Second || s.MaxConns != 10 || !s.TLS {
		t.Errorf("options not applied: %+v", s)
	}
	s = NewServer(":1", WithMaxConns(5), WithMaxConns(7))
	if s.MaxConns != 7 {
		t.Errorf("later option should win: %d", s.MaxConns)
	}
	s = NewServer(":1", WithMaxConns(5), WithMaxConns(0), WithMaxConns(-3))
	if s.MaxConns != 5 {
		t.Errorf("invalid values must be ignored: %d", s.MaxConns)
	}
}
```

### Hints
- An `Option` is a function that modifies a `*Server`
- Set defaults first, then loop over the options

### Solution
```go
package main

import "time"

type Server struct {
	Addr     string
	Timeout  time.Duration
	MaxConns int
	TLS      bool
}

type Option func(*Server)

func WithTimeout(d time.Duration) Option { return func(s *Server) { s.Timeout = d } }

func WithMaxConns(n int) Option {
	return func(s *Server) {
		if n >= 1 {
			s.MaxConns = n
		}
	}
}

func WithTLS() Option { return func(s *Server) { s.TLS = true } }

func NewServer(addr string, opts ...Option) *Server {
	s := &Server{Addr: addr, Timeout: 30 * time.Second, MaxConns: 100}
	for _, o := range opts {
		o(s)
	}
	return s
}

func main() {}
```

### Explanation
Options are plain functions applied over defaults, so new ones can be added without changing the constructor's signature — the property that makes the pattern popular in library APIs.
