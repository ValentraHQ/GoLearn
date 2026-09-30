# Memory & Performance
id: performance
number: 21
track: advanced-go
paths: pro
skill: performance
requires: concurrency
summary: Stack, heap, escape analysis, garbage collection, profiling with pprof and optimisation.

## Stack
slug: stack
minutes: 7
objectives: Explain goroutine stacks and stack frames; Understand that stacks grow dynamically; Recognise why stack allocation is cheap
takeaways: Each goroutine has its own small stack (starting around 2 KB) that grows and shrinks automatically; Local variables that don't outlive the function live on the stack — allocation is nearly free and needs no GC; Deep recursion can still overflow the maximum stack size (1 GB on 64-bit)

### Concept
When a function is called, its locals live in a **stack frame**; when it returns the frame is discarded — no garbage collector involved. Go stacks start tiny and are **copied to a bigger block** when needed, which is why you can run hundreds of thousands of goroutines.

Cost model:

| Where | Allocate | Free |
| --- | --- | --- |
| stack | bump a pointer | automatic on return |
| heap | runtime allocator | garbage collector |

You don't choose: the compiler decides using **escape analysis** (next lessons). Your job is to write code that lets values stay on the stack when possible — e.g. small structs passed by value, no needless pointers.

### Example
```go
package main

import "fmt"

func sum(n int) int {
	if n == 0 {
		return 0
	}
	return n + sum(n-1) // each call pushes a frame; Go stacks grow as needed
}

func main() {
	fmt.Println(sum(100000))
}
```

### Exercise
Compute the sum of `1..n` recursively for `n = 50000` (works because goroutine stacks grow dynamically) and print it.
```text expect
1250025000
```
```go solution
package main

import "fmt"

// BEGIN
func sum(n int) int {
	if n == 0 {
		return 0
	}
	return n + sum(n-1)
}

// END

func main() {
	fmt.Println(sum(50000))
}
```

### Check
Q: How big is a goroutine's initial stack?
T: mcq
- [ ] 8 MB like an OS thread
- [x] A few kilobytes, growing on demand
- [ ] 1 GB
- [ ] Fixed at 64 KB
E: Small growable stacks are why goroutines are so cheap.

Q: Does freeing stack memory involve the garbage collector?
T: tf
A: false
E: Stack frames are popped when functions return; only heap objects are garbage collected.

## Heap
slug: heap
minutes: 7
objectives: Explain what lives on the heap; Describe the cost of heap allocation; Identify common causes of heap allocations
takeaways: The heap holds values that outlive their function or whose size is unknown at compile time; Heap allocation is slower and adds GC work; Pointers returned from functions, closures capturing variables, interfaces holding non-pointer values, and growing slices/maps commonly allocate

### Concept
Heap memory is shared by all goroutines and reclaimed by the GC. Typical reasons a value goes to the heap:

- Its address is returned or stored somewhere that outlives the function (`return &x`).
- It's captured by a closure that escapes.
- It's stored in an interface and doesn't fit the interface word (boxing), e.g. `fmt.Println(x)`.
- It's a slice/map/channel whose size is dynamic (`make([]int, n)` with non-constant `n`).
- It's too large for the stack.

Heap allocations aren't bad — they're a cost to **measure**. The tool: `go test -bench . -benchmem` shows `allocs/op`; `go build -gcflags=-m` shows escape decisions.

### Example
```go
package main

import (
	"fmt"
	"testing"
)

type Point struct{ X, Y int }

func byValue() Point    { return Point{1, 2} }
func byPointer() *Point { return &Point{1, 2} }

func main() {
	a := testing.AllocsPerRun(1000, func() { _ = byValue() })
	b := testing.AllocsPerRun(1000, func() { _ = byPointer() })
	fmt.Println("value allocs:", a, "pointer allocs >= value:", b >= a)
}
```

### Exercise
Use `testing.AllocsPerRun` to measure `make([]int, 0, 10)` in a function that keeps the result in a package-level `sink` variable (forcing it to the heap). Print `true` if it allocates at least once per run.
```text expect
true
```
```go solution
package main

import (
	"fmt"
	"testing"
)

var sink []int

func main() {
	// BEGIN
	allocs := testing.AllocsPerRun(100, func() {
		sink = make([]int, 0, 10)
	})
	fmt.Println(allocs >= 1)
	// END
}
```

### Check
Q: Which tool shows the number of allocations per operation in a benchmark?
T: mcq
- [ ] go vet
- [x] go test -bench . -benchmem
- [ ] gofmt
- [ ] go mod why
E: `-benchmem` adds `B/op` and `allocs/op` columns.

Q: Every heap allocation is a performance bug.
T: tf
A: false
E: Heap allocations are normal; measure to find the ones that matter.

## Pointers and Performance
slug: pointers-and-performance
minutes: 7
objectives: Weigh copying versus pointer indirection; Avoid pointers to tiny values; Understand how pointers affect GC scanning
takeaways: Pass small structs by value — copying a few words is cheap and keeps data on the stack; Pointers can force heap allocation and add GC scan work; Use pointers for mutation, large structs, or when identity/nil matters — not by reflex

### Concept
"Use pointers for speed" is a myth for small types. Passing `Point{X,Y}` (16 bytes) by value is as cheap as passing a pointer and avoids indirection, allocation and GC work. A slice of pointers (`[]*Item`) scatters items across the heap (poor cache locality, more GC pointers to trace) while `[]Item` keeps them contiguous.

Use pointers when:
- the method must **mutate** the receiver,
- the struct is **large** (hundreds of bytes),
- you need **shared identity** or `nil` to mean "absent",
- the type contains a `sync.Mutex`.

Measure before optimising.

### Example
```go
package main

import (
	"fmt"
	"testing"
)

type Item struct{ A, B, C, D int }

func sumValues(xs []Item) (s int) {
	for _, x := range xs {
		s += x.A
	}
	return
}

func sumPointers(xs []*Item) (s int) {
	for _, x := range xs {
		s += x.A
	}
	return
}

func main() {
	vals := make([]Item, 1000)
	ptrs := make([]*Item, 1000)
	for i := range ptrs {
		ptrs[i] = &Item{}
	}
	a := testing.AllocsPerRun(100, func() { sumValues(vals) })
	b := testing.AllocsPerRun(100, func() { sumPointers(ptrs) })
	fmt.Println(a, b)
}
```

### Exercise
Write `alloc(n int) []Item` returning a slice of `n` values, and `allocPtr(n int) []*Item` returning `n` pointers to new items. Print whether the pointer version performs more allocations per run for `n = 100`.
```text expect
true
```
```go solution
package main

import (
	"fmt"
	"testing"
)

type Item struct{ A, B int }

var sink any

// BEGIN
func alloc(n int) []Item {
	return make([]Item, n)
}

func allocPtr(n int) []*Item {
	out := make([]*Item, n)
	for i := range out {
		out[i] = &Item{}
	}
	return out
}

// END

func main() {
	a := testing.AllocsPerRun(50, func() { sink = alloc(100) })
	b := testing.AllocsPerRun(50, func() { sink = allocPtr(100) })
	fmt.Println(b > a)
}
```

### Check
Q: Which is usually faster for small element types?
T: mcq
- [ ] []*Item
- [x] []Item
- [ ] map[int]*Item
- [ ] They are identical
E: Contiguous values have better cache locality and fewer allocations.

Q: Which is a good reason to use a pointer receiver?
T: mcq
- [ ] It makes the code look advanced
- [x] The method mutates the receiver
- [ ] Pointers are always faster
- [ ] It avoids all allocation
E: Mutation (or very large structs) justify pointers; speed alone usually doesn't.

## Allocations
slug: allocations
minutes: 8
objectives: Find allocations with benchmarks; Reduce allocations by preallocating and reusing buffers; Use sync.Pool for short-lived temporary objects
takeaways: Preallocate slices and maps when the size is known; Reuse buffers (bytes.Buffer, strings.Builder, sync.Pool) in hot paths; Measure allocs/op before and after — optimise what the profile shows

### Concept
Common allocation reducers:

1. **Preallocate:** `make([]T, 0, n)`, `make(map[K]V, n)`.
2. **Avoid repeated string concatenation:** `strings.Builder`.
3. **Reuse buffers:** `sync.Pool` for temporary `*bytes.Buffer`s.
4. **Avoid `fmt` in hot paths:** `strconv.AppendInt` writes into an existing byte slice.
5. **Avoid unnecessary conversions:** `[]byte(s)` / `string(b)` copy.

```go norun
var bufPool = sync.Pool{New: func() any { return new(bytes.Buffer) }}

buf := bufPool.Get().(*bytes.Buffer)
buf.Reset()
defer bufPool.Put(buf)
```
Optimise **after** profiling; keep the clear version unless the numbers justify the complexity.

### Example
```go
package main

import (
	"fmt"
	"strconv"
	"testing"
)

func viaSprintf(n int) string   { return fmt.Sprintf("%d", n) }
func viaItoa(n int) string      { return strconv.Itoa(n) }

func main() {
	a := testing.AllocsPerRun(100, func() { _ = viaSprintf(123456) })
	b := testing.AllocsPerRun(100, func() { _ = viaItoa(123456) })
	fmt.Println(b <= a)
}
```

### Exercise
Write `build(n int) []int` that returns `0..n-1` using a **preallocated** slice, and print whether it needs at most 1 allocation per run for `n = 1000`.
```text expect
true
```
```go solution
package main

import (
	"fmt"
	"testing"
)

var sink []int

// BEGIN
func build(n int) []int {
	out := make([]int, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, i)
	}
	return out
}

// END

func main() {
	allocs := testing.AllocsPerRun(100, func() { sink = build(1000) })
	fmt.Println(allocs <= 1)
}
```

### Check
Q: What does preallocating with `make([]T, 0, n)` avoid?
T: mcq
- [ ] Bounds checks
- [x] Repeated grow-and-copy allocations while appending
- [ ] Garbage collection entirely
- [ ] Nil slices
E: One allocation of the final size replaces many growths.

Q: What is `sync.Pool` for?
T: mcq
- [ ] Sharing state between goroutines permanently
- [x] Reusing temporary objects to reduce allocation pressure
- [ ] Limiting goroutines
- [ ] Caching database rows
E: The pool may drop items at any GC, so it holds only recreatable scratch objects.

## Escape Analysis
slug: escape-analysis
minutes: 8
objectives: Explain how the compiler decides stack versus heap; Read -gcflags=-m output; Write code that avoids unnecessary escapes
takeaways: A value escapes to the heap when its address may outlive the stack frame; go build -gcflags=-m prints escape decisions; Returning values instead of pointers and avoiding interfaces in hot paths keeps data on the stack

### Concept
```bash
go build -gcflags='-m' ./...          # show escape decisions
go build -gcflags='-m -m' ./...       # with reasons
```
Typical output:

```
./main.go:9:9: &Point{...} escapes to heap
./main.go:14:13: x does not escape
```
A value **escapes** when the compiler can't prove its address stays inside the function: returned pointers, values stored in package variables or in heap objects, captured by escaping closures, passed to interface methods or functions taking `any` (e.g. `fmt.Println`), sent on channels.

Inlining can rescue values: if a small function is inlined, the pointer may no longer escape. Treat escape analysis as a *tool for diagnosing* allocation-heavy hot paths, not something to micro-manage everywhere.

### Example
```go
package main

import (
	"fmt"
	"testing"
)

type Point struct{ X, Y int }

var global *Point

func stays() int {
	p := Point{1, 2} // does not escape
	return p.X + p.Y
}

func escapes() {
	p := Point{1, 2}
	global = &p // escapes: stored in a package variable
}

func main() {
	fmt.Println(testing.AllocsPerRun(100, func() { stays() }), testing.AllocsPerRun(100, escapes))
}
```

### Exercise
Make `escapes()` allocate **zero** times by not storing the pointer: rewrite `use` to return the sum computed from a local `Point` value. The program prints the allocations per run.
```text expect
0
```
```go solution
package main

import (
	"fmt"
	"testing"
)

type Point struct{ X, Y int }

// BEGIN
func use() int {
	p := Point{3, 4}
	return p.X + p.Y
}

// END

func main() {
	fmt.Println(testing.AllocsPerRun(100, func() { use() }))
}
```

### Check
Q: Which flag prints the compiler's escape analysis decisions?
T: mcq
- [ ] -race
- [x] -gcflags=-m
- [ ] -ldflags=-s
- [ ] -cover
E: `go build -gcflags=-m` reports which values escape to the heap.

Q: Passing a value to `fmt.Println` can cause it to escape.
T: tf
A: true
E: The argument is converted to `any`, which may box (and heap-allocate) the value.

## Garbage Collection
slug: garbage-collection
minutes: 8
objectives: Describe Go's concurrent, non-generational, tri-colour mark-and-sweep GC; Tune it with GOGC and GOMEMLIMIT; Read runtime.MemStats
takeaways: Go's GC runs concurrently with your program with short stop-the-world pauses; GOGC (default 100) sets how much the heap can grow before the next cycle; GOMEMLIMIT sets a soft memory limit — essential in containers

### Concept
The collector traces reachable objects from roots (stacks, globals) and frees the rest, mostly **concurrently** with your code. Cost scales with the amount of **live heap with pointers**, not with garbage.

Knobs:

- **`GOGC=100`** — next GC when the heap doubles relative to live data after the last GC. Higher = fewer GCs, more memory.
- **`GOMEMLIMIT=1GiB`** — soft cap; the GC works harder near the limit (set it a bit below the container memory limit).
- `debug.SetGCPercent`, `debug.SetMemoryLimit` change them at runtime.

Observe: `runtime.ReadMemStats(&m)` (`HeapAlloc`, `NumGC`, `PauseTotalNs`) or `GODEBUG=gctrace=1`. Reduce GC work by allocating less and by avoiding pointer-heavy data structures.

### Example
```go
package main

import (
	"fmt"
	"runtime"
)

func main() {
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	for i := 0; i < 3; i++ {
		_ = make([]byte, 10<<20)
		runtime.GC()
	}
	runtime.ReadMemStats(&after)
	fmt.Println("GCs ran:", after.NumGC-before.NumGC >= 3)
}
```

### Exercise
Use `runtime.GC()` and `runtime.ReadMemStats` to verify that `NumGC` increases after forcing a collection. Print `true`.
```text expect
true
```
```go solution
package main

import (
	"fmt"
	"runtime"
)

func main() {
	// BEGIN
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	runtime.GC()
	runtime.ReadMemStats(&after)
	fmt.Println(after.NumGC > before.NumGC)
	// END
}
```

### Check
Q: Which setting is most useful for keeping a Go service under a container's memory limit?
T: mcq
- [ ] GOMAXPROCS
- [x] GOMEMLIMIT
- [ ] GOFLAGS
- [ ] GOOS
E: `GOMEMLIMIT` makes the GC more aggressive as the heap approaches a soft limit.

Q: The cost of a GC cycle is proportional to the amount of garbage.
T: tf
A: false
E: Marking cost scales with live, pointer-containing heap; sweeping garbage is comparatively cheap.

## Memory Optimization
slug: memory-optimization
status: planned

## CPU Profiling
slug: cpu-profiling
status: planned

## Memory Profiling
slug: memory-profiling
status: planned

## Goroutine Profiling
slug: goroutine-profiling
status: planned

## pprof
slug: pprof
minutes: 9
objectives: Collect CPU and heap profiles; Read a profile with go tool pprof; Expose pprof safely in a running service
takeaways: pprof samples where your program spends CPU time and holds memory; Add import _ "net/http/pprof" on an internal-only port to profile a live service; Investigate with top, list and web (flame graph) in go tool pprof

### Concept
```bash
# from a benchmark
go test -bench=. -cpuprofile=cpu.out -memprofile=mem.out
go tool pprof -http=:8081 cpu.out       # interactive UI with flame graph

# from a live service
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
go tool pprof http://localhost:6060/debug/pprof/heap
go tool pprof http://localhost:6060/debug/pprof/goroutine
```
```go norun
import _ "net/http/pprof" // registers /debug/pprof/* on http.DefaultServeMux

go http.ListenAndServe("127.0.0.1:6060", nil) // internal only — never expose publicly
```
Inside pprof: `top` (heaviest functions), `list FuncName` (line-level cost), `web` (call graph). The **goroutine profile** finds leaks; the **heap** profile shows who allocates (`-sample_index=alloc_space`); the **block/mutex** profiles reveal contention (enable with `runtime.SetBlockProfileRate`).

`runtime/pprof` also writes profiles programmatically, as below.

### Example
```go
package main

import (
	"bytes"
	"fmt"
	"runtime/pprof"
)

func main() {
	var buf bytes.Buffer
	if err := pprof.Lookup("goroutine").WriteTo(&buf, 1); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(buf.Len() > 0)
}
```

### Exercise
Use `pprof.StartCPUProfile` to write a CPU profile into a `bytes.Buffer` while calling a small busy function, stop the profile, and print `true` if data was written.
```text expect
true
```
```go solution
package main

import (
	"bytes"
	"fmt"
	"runtime/pprof"
)

func busy() int {
	total := 0
	for i := 0; i < 50_000_000; i++ {
		total += i % 7
	}
	return total
}

func main() {
	var buf bytes.Buffer
	// BEGIN
	if err := pprof.StartCPUProfile(&buf); err != nil {
		fmt.Println(err)
		return
	}
	busy()
	pprof.StopCPUProfile()
	fmt.Println(buf.Len() > 0)
	// END
}
```

### Check
Q: Which pprof profile helps find goroutine leaks?
T: mcq
- [ ] CPU
- [ ] Heap
- [x] Goroutine
- [ ] Allocs only
E: The goroutine profile lists every goroutine and where it is blocked.

Q: The pprof HTTP endpoints should be exposed on the public internet.
T: tf
A: false
E: They reveal internals and allow expensive operations; bind to localhost or an internal network.

## Benchmarking
slug: benchmarking
minutes: 7
objectives: Benchmark competing implementations fairly; Compare runs with benchstat; Avoid micro-benchmark traps
takeaways: Benchmark realistic inputs and sizes, several times, and compare with benchstat; Prevent dead-code elimination by using the result; Change one thing at a time; keep the fastest readable version

### Concept
```bash
go test -bench=Parse -benchmem -count=10 > old.txt
# … change the code …
go test -bench=Parse -benchmem -count=10 > new.txt
benchstat old.txt new.txt
```
`benchstat` (golang.org/x/perf/cmd/benchstat) reports the statistical significance of differences (`~` means no significant change). Traps: benchmarking with tiny inputs, letting the compiler remove unused results (assign to a package-level `sink` or use `b.Loop()`), noisy machines (close other apps; pin CPU frequency), and measuring setup instead of the operation (`b.ResetTimer`).

### Example
```go
package main

import (
	"fmt"
	"strings"
	"testing"
)

var sink string

func join1(parts []string) string {
	s := ""
	for _, p := range parts {
		s += p
	}
	return s
}

func join2(parts []string) string { return strings.Join(parts, "") }

func main() {
	parts := strings.Split(strings.Repeat("x,", 100), ",")
	a := testing.Benchmark(func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sink = join1(parts)
		}
	})
	c := testing.Benchmark(func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sink = join2(parts)
		}
	})
	fmt.Println("Join uses fewer allocations:", c.AllocsPerOp() < a.AllocsPerOp())
}
```

### Exercise
Benchmark `strings.Builder` against `+=` for building a 500-character string. Print whether the builder version allocates less per operation.
```text expect
true
```
```go solution
package main

import (
	"fmt"
	"strings"
	"testing"
)

var sink string

func concat(n int) string {
	s := ""
	for i := 0; i < n; i++ {
		s += "x"
	}
	return s
}

func builder(n int) string {
	var b strings.Builder
	b.Grow(n)
	for i := 0; i < n; i++ {
		b.WriteByte('x')
	}
	return b.String()
}

func main() {
	// BEGIN
	slow := testing.Benchmark(func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sink = concat(500)
		}
	})
	fast := testing.Benchmark(func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sink = builder(500)
		}
	})
	fmt.Println(fast.AllocsPerOp() < slow.AllocsPerOp())
	// END
}
```

### Check
Q: Which tool statistically compares two sets of benchmark results?
T: short
A: benchstat
E: It tells you whether an apparent speed-up is real or noise.

Q: Why assign benchmark results to a package-level sink?
T: mcq
- [ ] To print them
- [x] So the compiler can't optimise the computation away
- [ ] To share them between goroutines
- [ ] It's required syntax
E: Otherwise unused results may be eliminated and you'd measure nothing.

## Performance Optimization
slug: performance-optimization
minutes: 8
objectives: Follow a measure → hypothesise → change → re-measure loop; Prioritise algorithmic and allocation wins over micro-tweaks; Know when to stop
takeaways: Profile first: optimise the top of the profile, not what you guess; Algorithmic improvements (O(n²) → O(n)) and fewer allocations usually beat micro-optimisations; Keep code readable and re-benchmark every change

### Concept
A disciplined loop:

1. **Define a target** (p99 latency < 50 ms; 5k req/s at 2 CPUs).
2. **Measure** with a realistic benchmark/load test; capture a CPU + heap profile.
3. **Find the hot spot** (`top`, `list`, flame graph).
4. **Change one thing.**
5. **Re-measure** (benchstat). Keep it only if it helps.
6. **Stop** when the target is met.

Typical wins, in order of impact: better algorithm/data structure, avoiding work (caching, batching), reducing allocations, reducing lock contention, parallelising, then micro-tweaks. Also mind I/O: database round-trips (N+1 queries) dwarf CPU costs in most services. "Premature optimisation is the root of all evil" — Knuth; *measured* optimisation is engineering.

### Example
```go
package main

import "fmt"

func hasDupSlow(xs []int) bool {
	for i := range xs {
		for j := i + 1; j < len(xs); j++ {
			if xs[i] == xs[j] {
				return true
			}
		}
	}
	return false
}

func hasDupFast(xs []int) bool {
	seen := make(map[int]struct{}, len(xs))
	for _, x := range xs {
		if _, ok := seen[x]; ok {
			return true
		}
		seen[x] = struct{}{}
	}
	return false
}

func main() {
	xs := []int{1, 2, 3, 4, 2}
	fmt.Println(hasDupSlow(xs), hasDupFast(xs))
}
```

### Exercise
`intersectSlow` is O(n·m). Write `intersect(a, b []int) []int` in O(n+m) using a set, returning the common values in the order they appear in `a`.
```text expect
[3 5]
```
```go solution
package main

import "fmt"

// BEGIN
func intersect(a, b []int) []int {
	inB := make(map[int]struct{}, len(b))
	for _, x := range b {
		inB[x] = struct{}{}
	}
	var out []int
	for _, x := range a {
		if _, ok := inB[x]; ok {
			out = append(out, x)
		}
	}
	return out
}

// END

func main() {
	fmt.Println(intersect([]int{1, 3, 4, 5}, []int{5, 3, 9}))
}
```

### Check
Q: What should you do before optimising?
T: mcq
- [ ] Rewrite hot paths in assembly
- [x] Measure and profile to find the real bottleneck
- [ ] Add goroutines everywhere
- [ ] Remove all allocations
E: Guessing wastes effort; profiles show where time goes.

Q: Which usually gives the biggest improvement?
T: mcq
- [ ] Renaming variables
- [x] A better algorithm or avoiding unnecessary work
- [ ] Removing comments
- [ ] Using shorter package names
E: Big-O improvements and eliminating work beat micro-tweaks.
