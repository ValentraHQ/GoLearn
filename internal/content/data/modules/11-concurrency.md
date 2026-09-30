# Concurrency
id: concurrency
number: 11
track: advanced
paths: beginner, pro
skill: concurrency
requires: errors
project: concurrent-job-processing-system
summary: Goroutines, channels, synchronisation, pipelines, context and graceful shutdown.

## What Is Concurrency?
slug: what-is-concurrency
minutes: 6
objectives: Distinguish concurrency from parallelism; Explain why Go makes concurrency a core feature; Know the model: share memory by communicating
takeaways: Concurrency is structuring a program as independently executing tasks; Parallelism is running tasks at the same instant on multiple cores; Go's proverb: don't communicate by sharing memory, share memory by communicating

### Concept
**Concurrency** is about *dealing with* many things at once (structure). **Parallelism** is *doing* many things at once (execution). A concurrent program can run on one core and still be correct; on more cores it may also run in parallel.

Go's building blocks:

- **goroutines** — lightweight tasks scheduled by the runtime (thousands are normal),
- **channels** — typed pipes for passing values between goroutines,
- **`sync`** and **`context`** — locks, waiting, cancellation.

> Don't communicate by sharing memory; share memory by communicating. — Go Proverb

### Example
```go
package main

import (
	"fmt"
	"runtime"
)

func main() {
	fmt.Println("CPUs available:", runtime.NumCPU() > 0)
	fmt.Println("GOMAXPROCS positive:", runtime.GOMAXPROCS(0) > 0)
}
```

### Exercise
Print `true` if `runtime.NumGoroutine()` equals 1 at program start (only the main goroutine exists).
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
	fmt.Println(runtime.NumGoroutine() == 1)
	// END
}
```

### Check
Q: What is the difference between concurrency and parallelism?
T: mcq
- [ ] They are synonyms
- [x] Concurrency is structuring independent tasks; parallelism is executing them simultaneously
- [ ] Parallelism needs goroutines, concurrency does not
- [ ] Concurrency requires multiple machines
E: Concurrent code may run on a single core; parallelism needs several.

Q: According to Go's proverb you should share memory by communicating.
T: tf
A: true
E: Prefer passing ownership of data over channels to guarding shared state with locks — though both tools exist.

## Goroutines
slug: goroutines
minutes: 7
objectives: Start goroutines with the go keyword; Understand that main exiting kills them; Wait for them with a WaitGroup
takeaways: go f() runs f concurrently; The program exits when main returns — even if goroutines are still running; Use sync.WaitGroup (or channels) to wait for completion

### Concept
```go norun
go doWork()             // runs concurrently
go func() { ... }()     // anonymous function
```
Goroutines are cheap (a few KB of stack, grown as needed) and multiplexed onto OS threads by the Go scheduler. The function's return values are discarded.

**When `main` returns the program ends** — other goroutines are killed mid-flight. Never use `time.Sleep` to "wait"; use a `sync.WaitGroup`: `Add` before `go`, `Done` (deferred) inside, `Wait` in the parent.

### Example
```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	results := make([]int, 5)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i] = i * i // each goroutine writes its own slot
		}()
	}
	wg.Wait()
	fmt.Println(results)
}
```

### Exercise
Launch 4 goroutines; goroutine `i` stores `i+10` in `out[i]`. Wait for them and print `out`.
```text expect
[10 11 12 13]
```
```go solution
package main

import (
	"fmt"
	"sync"
)

func main() {
	out := make([]int, 4)
	var wg sync.WaitGroup
	// BEGIN
	for i := range out {
		wg.Add(1)
		go func() {
			defer wg.Done()
			out[i] = i + 10
		}()
	}
	wg.Wait()
	// END
	fmt.Println(out)
}
```

### Check
Q: What happens to running goroutines when `main` returns?
T: mcq
- [ ] They finish first
- [x] They are terminated with the program
- [ ] They become daemons
- [ ] main blocks until they end
E: The process exits as soon as `main` returns.

Q: Why is `time.Sleep` a poor way to wait for goroutines?
T: mcq
- [ ] It's deprecated
- [x] It guesses at timing instead of synchronising on completion
- [ ] It blocks all goroutines
- [ ] It requires a mutex
E: Sleeping is either too short (races) or too long (slow); use a WaitGroup or channel.

## Goroutine Lifecycle
slug: goroutine-lifecycle
minutes: 7
objectives: Describe how a goroutine starts, runs, blocks and ends; Know the ways a goroutine can terminate; Recognise that the runtime cannot kill a goroutine for you
takeaways: A goroutine ends when its function returns or the program exits; You can't kill a goroutine from outside — you signal it (channel/context) to stop; Blocked goroutines still consume memory: every goroutine needs an exit path

### Concept
States: **runnable** → **running** → **waiting** (on a channel, mutex, I/O, sleep) → runnable → … → **finished**.

A goroutine ends when its function returns (or panics — which crashes the program unless recovered in that goroutine). There is **no** `Kill(goroutine)`. Cooperative shutdown means: give every long-running goroutine a way to be told to stop (a `done` channel or a `context`) and to notice it.

A goroutine blocked forever is a **leak**: it holds its stack and everything it references.

### Example
```go
package main

import (
	"fmt"
	"time"
)

func worker(done <-chan struct{}, out chan<- int) {
	for i := 0; ; i++ {
		select {
		case <-done:
			close(out)
			return
		case out <- i:
		}
	}
}

func main() {
	done := make(chan struct{})
	out := make(chan int)
	go worker(done, out)
	fmt.Println(<-out, <-out, <-out)
	close(done)
	time.Sleep(10 * time.Millisecond)
	_, open := <-out
	fmt.Println("worker stopped:", !open)
}
```

### Exercise
Write `spin(done <-chan struct{}) int` that increments a counter in a loop until `done` is closed, then returns a `true`-ish count (`> 0`). Print `true` after closing done from `main`.
```text expect
true
```
```go solution
package main

import (
	"fmt"
	"time"
)

// BEGIN
func spin(done <-chan struct{}) int {
	n := 0
	for {
		select {
		case <-done:
			return n
		default:
			n++
		}
	}
}

// END

func main() {
	done := make(chan struct{})
	res := make(chan int)
	go func() { res <- spin(done) }()
	time.Sleep(5 * time.Millisecond)
	close(done)
	fmt.Println(<-res > 0)
}
```

### Check
Q: How do you stop a goroutine from outside?
T: mcq
- [ ] goroutine.Kill()
- [ ] runtime.Goexit(id)
- [x] Signal it (channel or context) and let it return
- [ ] You can't; leave it running
E: Go has cooperative cancellation only: the goroutine must check a signal and exit.

Q: A goroutine blocked forever on a channel receive is harmless.
T: tf
A: false
E: It leaks memory and whatever it references. Every goroutine needs a guaranteed exit path.

## Channels
slug: channels
minutes: 8
objectives: Create channels with make; Send and receive with <-; Use channels to pass results between goroutines
takeaways: ch := make(chan T) creates a channel; ch <- v sends and v := <-ch receives; A send and a receive synchronise the two goroutines

### Concept
```go norun
ch := make(chan int)
go func() { ch <- 42 }()   // send
v := <-ch                  // receive (blocks until a value arrives)
```
A channel is a typed conduit. Sending and receiving **block** until the other side is ready (for unbuffered channels), so channels synchronise goroutines as well as move data. The zero value of a channel is `nil`; operations on a nil channel block forever.

### Example
```go
package main

import "fmt"

func square(n int, out chan<- int) { out <- n * n }

func main() {
	ch := make(chan int)
	go square(7, ch)
	fmt.Println(<-ch)
}
```

### Exercise
Start a goroutine that sums `nums` and sends the total on a channel; receive and print it.
```text expect
15
```
```go solution
package main

import "fmt"

func main() {
	nums := []int{1, 2, 3, 4, 5}
	ch := make(chan int)
	// BEGIN
	go func() {
		sum := 0
		for _, n := range nums {
			sum += n
		}
		ch <- sum
	}()
	fmt.Println(<-ch)
	// END
}
```

### Check
Q: What happens on `v := <-ch` when no value has been sent yet?
T: mcq
- [ ] v is 0
- [x] The goroutine blocks until a value arrives
- [ ] A panic
- [ ] It returns immediately
E: Receives block until data is available (or the channel is closed).

Q: What is the zero value of a channel?
T: short
A: nil
E: A nil channel blocks forever on send and receive.

## Buffered Channels
slug: buffered-channels
minutes: 6
objectives: Create channels with capacity; Predict when sends block; Use buffering to decouple bursts
takeaways: make(chan T, n) creates a buffer of n values; Sends block only when the buffer is full; receives block only when it's empty; Buffers absorb bursts — they are not a substitute for backpressure design

### Concept
```go norun
ch := make(chan string, 2)
ch <- "a"   // doesn't block
ch <- "b"   // doesn't block
// ch <- "c" would block until someone receives
fmt.Println(len(ch), cap(ch)) // 2 2
```
Buffering lets a producer run ahead of a consumer by up to `cap` items. Use it for known bursts, semaphores and result collection where the number of results is known. An unbounded backlog just hides overload: pick the capacity deliberately.

### Example
```go
package main

import "fmt"

func main() {
	ch := make(chan int, 3)
	for i := 1; i <= 3; i++ {
		ch <- i
	}
	fmt.Println(len(ch), cap(ch))
	fmt.Println(<-ch, <-ch)
	fmt.Println(len(ch))
}
```

### Exercise
Create a buffered channel of capacity 3, send `"a"`, `"b"`, `"c"` without any goroutine, then receive and print them in order on one line.
```text expect
a b c
```
```go solution
package main

import "fmt"

func main() {
	// BEGIN
	ch := make(chan string, 3)
	ch <- "a"
	ch <- "b"
	ch <- "c"
	fmt.Println(<-ch, <-ch, <-ch)
	// END
}
```

### Check
Q: When does a send on a buffered channel block?
T: mcq
- [ ] Always
- [ ] Never
- [x] When the buffer is full
- [ ] When the buffer is empty
E: A buffered send only blocks once `len(ch) == cap(ch)`.

Q: What does this print?
T: output
```go
ch := make(chan int, 4)
ch <- 1
ch <- 2
fmt.Println(len(ch), cap(ch))
```
- [ ] 4 2
- [x] 2 4
- [ ] 2 2
- [ ] 0 4
E: `len` counts queued items (2); `cap` is the buffer size (4).

## Unbuffered Channels
slug: unbuffered-channels
minutes: 6
objectives: Explain rendezvous semantics; Recognise a deadlock; Use unbuffered channels for synchronisation
takeaways: An unbuffered channel hands a value directly from sender to receiver; Both sides must be ready — otherwise they block; If all goroutines are blocked the runtime reports a deadlock

### Concept
With no buffer a send completes only when a receiver takes the value — a **rendezvous**. That makes unbuffered channels excellent for synchronisation ("the worker has started", "the result is ready").

```go norun
ch := make(chan int)
ch <- 1 // DEADLOCK: no other goroutine will ever receive
```
If every goroutine is blocked, the runtime aborts with `fatal error: all goroutines are asleep - deadlock!`. That check only catches the whole program; a *partial* deadlock in a server just hangs quietly.

### Example
```go
package main

import "fmt"

func main() {
	ready := make(chan struct{})
	go func() {
		fmt.Println("worker: starting")
		close(ready)
	}()
	<-ready
	fmt.Println("main: worker signalled")
}
```

### Exercise
Use an unbuffered `done` channel so `main` waits for the goroutine, which prints `working` first; then `main` prints `finished`.
```text expect
working
finished
```
```go solution
package main

import "fmt"

func main() {
	// BEGIN
	done := make(chan struct{})
	go func() {
		fmt.Println("working")
		done <- struct{}{}
	}()
	<-done
	fmt.Println("finished")
	// END
}
```

### Check
Q: When does a send on an unbuffered channel complete?
T: mcq
- [ ] Immediately
- [x] When a receiver takes the value
- [ ] After a timeout
- [ ] When the channel is closed
E: Unbuffered channels synchronise sender and receiver.

Q: What happens if the only goroutine sends on an unbuffered channel that nobody receives from?
T: mcq
- [ ] The value is dropped
- [ ] It succeeds
- [x] The runtime detects a deadlock and the program crashes
- [ ] The send times out
E: `fatal error: all goroutines are asleep - deadlock!`

## Channel Direction
slug: channel-direction
minutes: 5
objectives: Declare send-only and receive-only channel types; Use direction to document and enforce ownership; Convert bidirectional channels implicitly
takeaways: chan<- T can only send, <-chan T can only receive; A bidirectional channel converts to either direction automatically; Direction in signatures makes data flow obvious and prevents misuse

### Concept
```go norun
func producer(out chan<- int)  { out <- 1 }   // send only
func consumer(in <-chan int)   { <-in }       // receive only
```
The compiler rejects receiving from a `chan<-` or closing a `<-chan`. Pass the narrowest type a function needs: it documents intent and prevents bugs. Only the **sender** side should close a channel, which the types encode nicely.

### Example
```go
package main

import "fmt"

func gen(n int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for i := 1; i <= n; i++ {
			out <- i
		}
	}()
	return out
}

func main() {
	for v := range gen(3) {
		fmt.Print(v, " ")
	}
	fmt.Println()
}
```

### Exercise
Write `gen(words ...string) <-chan string` that emits the words and closes the channel. Print them joined by commas.
```text expect
go,is,fun
```
```go solution
package main

import (
	"fmt"
	"strings"
)

// BEGIN
func gen(words ...string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for _, w := range words {
			out <- w
		}
	}()
	return out
}

// END

func main() {
	var got []string
	for w := range gen("go", "is", "fun") {
		got = append(got, w)
	}
	fmt.Println(strings.Join(got, ","))
}
```

### Check
Q: What does the type `<-chan int` allow?
T: mcq
- [ ] Only sending
- [x] Only receiving
- [ ] Sending and receiving
- [ ] Closing only
E: The arrow's position shows the direction relative to `chan`: `<-chan` is receive-only.

Q: Can a bidirectional `chan int` be passed to a function expecting `chan<- int`?
T: tf
A: true
E: Bidirectional channels convert implicitly to the restricted directions.

## Closing Channels
slug: closing-channels
minutes: 6
objectives: Close a channel to signal "no more values"; Detect closure with the comma-ok receive; Avoid closing twice or sending on a closed channel
takeaways: close(ch) means no more values will be sent; Receiving from a closed channel yields buffered values then zero values with ok == false; Only the sender closes; closing twice or sending after close panics

### Concept
```go norun
close(ch)
v, ok := <-ch // ok == false once closed and drained
```
Rules:

- Only the **sender** (or the goroutine that owns the channel) should close it.
- Sending on a closed channel **panics**; closing a closed or nil channel panics.
- You **don't have to** close channels — only when receivers need to know "the end". GC reclaims unused channels.

A closed channel also works as a broadcast: every receiver wakes at once (`<-done`).

### Example
```go
package main

import "fmt"

func main() {
	ch := make(chan int, 2)
	ch <- 1
	ch <- 2
	close(ch)
	for i := 0; i < 3; i++ {
		v, ok := <-ch
		fmt.Println(v, ok)
	}
}
```

### Exercise
Write `drain(ch <-chan int) int` that reads with the comma-ok form until the channel is closed and returns the sum.
```text expect
6
```
```go solution
package main

import "fmt"

// BEGIN
func drain(ch <-chan int) int {
	sum := 0
	for {
		v, ok := <-ch
		if !ok {
			return sum
		}
		sum += v
	}
}

// END

func main() {
	ch := make(chan int, 3)
	ch <- 1
	ch <- 2
	ch <- 3
	close(ch)
	fmt.Println(drain(ch))
}
```

### Check
Q: What does receiving from a closed, empty channel return?
T: mcq
- [ ] It blocks
- [ ] It panics
- [x] The zero value and ok == false
- [ ] The last value again
E: Closed channels never block receivers.

Q: What happens if you send on a closed channel?
T: mcq
- [ ] The value is discarded
- [ ] It blocks
- [x] A run-time panic
- [ ] It returns an error
E: Sending on a closed channel panics — which is why only the sending side should close.

## Range Over Channels
slug: range-over-channels
minutes: 5
objectives: Consume a channel with range; Understand that range stops on close; Combine with a producer goroutine
takeaways: for v := range ch receives until ch is closed and drained; Forgetting to close makes the loop block forever; The producer closes, the consumer ranges

### Concept
```go norun
for v := range ch {
	fmt.Println(v)
}
```
`range` on a channel receives values until the channel is **closed and empty**. If nobody closes it, the loop (and your goroutine) blocks forever. The canonical shape: producer goroutine `defer close(out)`, consumer `for range`.

### Example
```go
package main

import "fmt"

func main() {
	ch := make(chan int)
	go func() {
		defer close(ch)
		for i := 1; i <= 4; i++ {
			ch <- i * i
		}
	}()
	for v := range ch {
		fmt.Print(v, " ")
	}
	fmt.Println()
}
```

### Exercise
Have a goroutine send the letters `a`, `b`, `c` and close the channel; range over it in `main` and print the letters concatenated (`abc`).
```text expect
abc
```
```go solution
package main

import "fmt"

func main() {
	ch := make(chan string)
	// BEGIN
	go func() {
		defer close(ch)
		for _, s := range []string{"a", "b", "c"} {
			ch <- s
		}
	}()
	out := ""
	for s := range ch {
		out += s
	}
	fmt.Println(out)
	// END
}
```

### Check
Q: When does `for v := range ch` stop?
T: mcq
- [ ] After one value
- [ ] When the channel is empty
- [x] When the channel is closed and drained
- [ ] After a timeout
E: Range keeps receiving until close; an unclosed channel blocks it forever.

Q: Who should close the channel in a producer–consumer pair?
T: mcq
- [ ] The consumer
- [x] The producer
- [ ] Either, at random
- [ ] Nobody must
E: The sender knows when there's no more data; closing from the receiving side risks a send-on-closed panic.

## Select
slug: select
minutes: 8
objectives: Wait on multiple channel operations with select; Add a default case and timeouts; Recognise that a ready case is chosen at random
takeaways: select blocks until one of its cases can proceed; With default it never blocks; If several cases are ready one is chosen at random; time.After in a select implements a timeout

### Concept
```go norun
select {
case v := <-ch1:
	// got a value
case ch2 <- x:
	// sent
case <-time.After(time.Second):
	// timeout
default:
	// nothing ready (makes select non-blocking)
}
```
`select` is Go's multiplexer. Common uses: timeouts, cancellation (`case <-ctx.Done()`), non-blocking try-send/receive, merging channels. A `select {}` with no cases blocks forever. A nil channel case is never ready — handy for disabling a case dynamically.

### Example
```go
package main

import (
	"fmt"
	"time"
)

func main() {
	ch := make(chan string)
	go func() {
		time.Sleep(50 * time.Millisecond)
		ch <- "result"
	}()
	select {
	case v := <-ch:
		fmt.Println(v)
	case <-time.After(10 * time.Millisecond):
		fmt.Println("timeout")
	}
}
```

### Exercise
Write `tryRecv(ch <-chan int) (int, bool)` that receives a value if one is immediately available and otherwise returns `0, false` without blocking.
```text expect
5 true
0 false
```
```go solution
package main

import "fmt"

// BEGIN
func tryRecv(ch <-chan int) (int, bool) {
	select {
	case v := <-ch:
		return v, true
	default:
		return 0, false
	}
}

// END

func main() {
	ch := make(chan int, 1)
	ch <- 5
	fmt.Println(tryRecv(ch))
	fmt.Println(tryRecv(ch))
}
```

### Check
Q: What does a `default` case do in a select?
T: mcq
- [ ] Runs after all others
- [x] Runs immediately if no other case is ready, making the select non-blocking
- [ ] Is required
- [ ] Runs on timeout
E: With `default`, `select` never blocks.

Q: If two cases in a select are ready at the same time, which runs?
T: mcq
- [ ] The first one written
- [ ] The last one written
- [x] One chosen at random
- [ ] Both, in order
E: The pseudo-random choice prevents starvation of later cases.

## WaitGroup
slug: waitgroup
minutes: 6
objectives: Coordinate goroutines with sync.WaitGroup; Follow the Add/Done/Wait discipline; Use WaitGroup.Go (Go 1.25+) or the classic pattern
takeaways: Add before starting the goroutine, defer Done inside, Wait in the parent; A WaitGroup must not be copied; Negative counters panic

### Concept
```go norun
var wg sync.WaitGroup
for _, url := range urls {
	wg.Add(1)
	go func() {
		defer wg.Done()
		fetch(url)
	}()
}
wg.Wait()
```
Rules: call `Add(1)` **before** `go` (never inside the goroutine), `defer Done()`, and pass the WaitGroup by pointer. Go 1.25 adds `wg.Go(func(){...})` which does the Add/Done for you. A WaitGroup waits — it does not collect results or errors; use channels or `errgroup` for that.

### Example
```go
package main

import (
	"fmt"
	"sort"
	"sync"
)

func main() {
	var (
		wg sync.WaitGroup
		mu sync.Mutex
		got []int
	)
	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			got = append(got, i*i)
			mu.Unlock()
		}()
	}
	wg.Wait()
	sort.Ints(got)
	fmt.Println(got)
}
```

### Exercise
Run 3 goroutines that each add their id (1, 2, 3) to a shared `total` guarded by a mutex; wait and print `6`.
```text expect
6
```
```go solution
package main

import (
	"fmt"
	"sync"
)

func main() {
	var (
		wg    sync.WaitGroup
		mu    sync.Mutex
		total int
	)
	// BEGIN
	for id := 1; id <= 3; id++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			total += id
			mu.Unlock()
		}()
	}
	wg.Wait()
	// END
	fmt.Println(total)
}
```

### Check
Q: Where should `wg.Add(1)` be called?
T: mcq
- [ ] Inside the goroutine
- [x] Before the `go` statement, in the parent
- [ ] After wg.Wait()
- [ ] Anywhere
E: Otherwise `Wait` might run before the `Add`, returning too early.

Q: A `sync.WaitGroup` returns the results of the goroutines it waits for.
T: tf
A: false
E: It only waits. Collect results via channels or shared state.

## Mutex
slug: mutex
minutes: 7
objectives: Protect shared data with sync.Mutex; Keep critical sections small; Avoid copying and forgetting to unlock
takeaways: Lock/Unlock guard a critical section; defer Unlock immediately after Lock; Put the mutex next to the data it protects and never copy it

### Concept
```go norun
type SafeCounter struct {
	mu sync.Mutex
	n  map[string]int
}

func (c *SafeCounter) Inc(k string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.n[k]++
}
```
Without the lock, concurrent map writes crash the program and counter increments are lost (a **race condition**). Guidelines: hold locks briefly, never call unknown code while holding one, always `defer Unlock()`, embed the mutex in the struct that owns the data, and never copy a struct containing a mutex (use pointer receivers).

### Example
```go
package main

import (
	"fmt"
	"sync"
)

type Counter struct {
	mu sync.Mutex
	n  int
}

func (c *Counter) Inc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.n++
}

func main() {
	var c Counter
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); c.Inc() }()
	}
	wg.Wait()
	fmt.Println(c.n)
}
```

### Exercise
Make `Bank.Deposit` and `Bank.Balance` safe for concurrent use with a mutex. 100 goroutines deposit 10 each; print `1000`.
```text expect
1000
```
```go solution
package main

import (
	"fmt"
	"sync"
)

type Bank struct {
	// BEGIN
	mu      sync.Mutex
	balance int
	// END
}

func (b *Bank) Deposit(n int) {
	// BEGIN
	b.mu.Lock()
	defer b.mu.Unlock()
	b.balance += n
	// END
}

func (b *Bank) Balance() int {
	// BEGIN
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.balance
	// END
}

func main() {
	var b Bank
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); b.Deposit(10) }()
	}
	wg.Wait()
	fmt.Println(b.Balance())
}
```

### Check
Q: Why `defer mu.Unlock()` right after `Lock()`?
T: mcq
- [ ] It is faster
- [x] It guarantees unlock on every return path, including panics
- [ ] Unlock must be deferred by the compiler
- [ ] It makes the lock reentrant
E: Forgetting to unlock on an early return deadlocks the next caller.

Q: Are Go mutexes reentrant (can the same goroutine lock twice)?
T: tf
A: false
E: Locking a mutex you already hold deadlocks. Restructure code so locked helpers don't re-lock.

## RWMutex
slug: rwmutex
minutes: 6
objectives: Allow many concurrent readers with sync.RWMutex; Choose between Mutex and RWMutex; Avoid upgrading a read lock to a write lock
takeaways: RLock allows many readers, Lock allows one writer and no readers; RWMutex pays off only for read-heavy, non-trivial critical sections; You can't upgrade an RLock to Lock — release first

### Concept
```go norun
type Cache struct {
	mu   sync.RWMutex
	data map[string]string
}

func (c *Cache) Get(k string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.data[k]
	return v, ok
}
```
`RWMutex` has more overhead than `Mutex`; use it when reads greatly outnumber writes *and* the critical section does real work. Trying to take `Lock` while holding `RLock` deadlocks. Measure before assuming it's faster.

### Example
```go
package main

import (
	"fmt"
	"sync"
)

type Cache struct {
	mu   sync.RWMutex
	data map[string]int
}

func (c *Cache) Get(k string) (int, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.data[k]
	return v, ok
}

func (c *Cache) Set(k string, v int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.data == nil {
		c.data = map[string]int{}
	}
	c.data[k] = v
}

func main() {
	var c Cache
	c.Set("a", 1)
	fmt.Println(c.Get("a"))
	fmt.Println(c.Get("b"))
}
```

### Exercise
Implement `Len()` on `Cache` using a read lock. After setting two keys print `2`.
```text expect
2
```
```go solution
package main

import (
	"fmt"
	"sync"
)

type Cache struct {
	mu   sync.RWMutex
	data map[string]int
}

func (c *Cache) Set(k string, v int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.data == nil {
		c.data = map[string]int{}
	}
	c.data[k] = v
}

// BEGIN
func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.data)
}

// END

func main() {
	var c Cache
	c.Set("a", 1)
	c.Set("b", 2)
	fmt.Println(c.Len())
}
```

### Check
Q: How many goroutines may hold an `RLock` at the same time?
T: mcq
- [ ] One
- [x] Any number, as long as no writer holds Lock
- [ ] Exactly two
- [ ] None while others wait
E: Readers share the lock; a writer needs exclusive access.

Q: It is safe to call `Lock()` while holding `RLock()` on the same RWMutex.
T: tf
A: false
E: That deadlocks; release the read lock before taking the write lock.

## Atomic Operations
slug: atomic-operations
minutes: 6
objectives: Use sync/atomic for simple shared counters and flags; Prefer atomic.Int64 and friends over raw functions; Know when a mutex is the better choice
takeaways: atomic.Int64, Bool and Value offer lock-free operations on single values; Atomics protect one variable, not a multi-step invariant; Use a mutex when several fields must change together

### Concept
```go norun
var hits atomic.Int64
hits.Add(1)
n := hits.Load()

var ready atomic.Bool
ready.Store(true)
```
Typed atomics (Go 1.19+) are hard to misuse. They are ideal for counters, gauges and flags. But an atomic can't make "check balance, then withdraw" safe — that spans two operations; use a mutex. `atomic.Value` / `atomic.Pointer[T]` swap whole immutable snapshots (configuration reloads).

### Example
```go
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func main() {
	var hits atomic.Int64
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); hits.Add(1) }()
	}
	wg.Wait()
	fmt.Println(hits.Load())
}
```

### Exercise
Count how many of 100 goroutines see an even `id` using an `atomic.Int32`; print the count (`50`).
```text expect
50
```
```go solution
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func main() {
	var evens atomic.Int32
	var wg sync.WaitGroup
	// BEGIN
	for id := 0; id < 100; id++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if id%2 == 0 {
				evens.Add(1)
			}
		}()
	}
	wg.Wait()
	// END
	fmt.Println(evens.Load())
}
```

### Check
Q: What can atomics NOT do safely by themselves?
T: mcq
- [ ] Increment a counter
- [ ] Set a boolean flag
- [x] Protect an invariant spanning multiple variables or steps
- [ ] Publish a pointer
E: Each atomic operation is individually safe; multi-step logic needs a mutex.

Q: Which type would you use for a shared request counter?
T: mcq
- [ ] int with no protection
- [x] atomic.Int64
- [ ] chan int
- [ ] sync.WaitGroup
E: `atomic.Int64` provides lock-free, race-free increments.

## Worker Pools
slug: worker-pools
minutes: 9
challenges: worker-pool-results
objectives: Bound concurrency with a fixed set of workers; Feed jobs and collect results through channels; Close channels in the right order
takeaways: A worker pool runs N goroutines reading from one jobs channel; Close jobs when done submitting; close results after all workers finish (WaitGroup); Bounded workers protect downstream systems from overload

### Concept
```go norun
jobs := make(chan Job)
results := make(chan Result)
for w := 0; w < N; w++ {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for j := range jobs {
			results <- process(j)
		}
	}()
}
go func() { wg.Wait(); close(results) }() // closer
for _, j := range all { jobs <- j }
close(jobs)
```
Starting a goroutine per task is fine for a few hundred, but for thousands of HTTP calls or DB queries you want a **limit**. The closer goroutine closes `results` only after every worker exits, so the consumer's `range results` terminates.

### Example
```go
package main

import (
	"fmt"
	"sort"
	"sync"
)

func main() {
	jobs := make(chan int)
	results := make(chan int)
	var wg sync.WaitGroup
	for w := 0; w < 3; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for n := range jobs {
				results <- n * n
			}
		}()
	}
	go func() { wg.Wait(); close(results) }()
	go func() {
		for i := 1; i <= 6; i++ {
			jobs <- i
		}
		close(jobs)
	}()
	var out []int
	for r := range results {
		out = append(out, r)
	}
	sort.Ints(out)
	fmt.Println(out)
}
```

### Exercise
Implement `double(nums []int, workers int) []int` using a worker pool that returns the doubled values **in the original order** (send index+value pairs).
```text expect
[2 4 6 8 10]
```
```go solution
package main

import (
	"fmt"
	"sync"
)

// BEGIN
func double(nums []int, workers int) []int {
	type job struct{ i, n int }
	jobs := make(chan job)
	out := make([]int, len(nums))
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				out[j.i] = j.n * 2 // each index is written by exactly one worker
			}
		}()
	}
	for i, n := range nums {
		jobs <- job{i, n}
	}
	close(jobs)
	wg.Wait()
	return out
}

// END

func main() {
	fmt.Println(double([]int{1, 2, 3, 4, 5}, 3))
}
```

### Check
Q: Who should close the `results` channel in a worker pool?
T: mcq
- [ ] Each worker when it finishes
- [x] A single goroutine after wg.Wait() confirms all workers exited
- [ ] The consumer
- [ ] Nobody
E: Closing from any worker would panic other workers still sending.

Q: Why use a fixed number of workers rather than one goroutine per job?
T: mcq
- [ ] Goroutines are expensive
- [x] To bound concurrent load on CPU, memory or downstream services
- [ ] The scheduler requires it
- [ ] It preserves order
E: Goroutines are cheap, but unbounded concurrency can overwhelm a database or API.

## Fan-Out
slug: fan-out
minutes: 6
objectives: Distribute one stream across several goroutines; Recognise that fan-out reads from a shared channel; Combine with fan-in for parallel stages
takeaways: Fan-out = several goroutines receiving from the same channel; Work is distributed automatically to whichever worker is free; Results arrive in nondeterministic order

### Concept
Fan-out is exactly what worker pools do: many goroutines read from **one** input channel, so each item is processed by exactly one of them. It parallelises a slow stage.

```go norun
in := gen(1, 2, 3, 4, 5, 6)
c1 := square(in)
c2 := square(in) // both consume from 'in'
```
Because scheduling isn't deterministic, output order varies — sort or index results when order matters.

### Example
```go
package main

import (
	"fmt"
	"sort"
	"sync"
)

func gen(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			out <- n
		}
	}()
	return out
}

func square(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			out <- n * n
		}
	}()
	return out
}

func main() {
	in := gen(1, 2, 3, 4, 5, 6)
	outs := []<-chan int{square(in), square(in), square(in)}
	var (
		wg  sync.WaitGroup
		mu  sync.Mutex
		all []int
	)
	for _, c := range outs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for v := range c {
				mu.Lock()
				all = append(all, v)
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	sort.Ints(all)
	fmt.Println(all)
}
```

### Exercise
Given the `gen` and `square` stages, start **two** `square` goroutines on the same input and print the sorted, combined results using the provided merge loop.
```text expect
[1 4 9 16]
```
```go solution
package main

import (
	"fmt"
	"sort"
)

func gen(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			out <- n
		}
	}()
	return out
}

func square(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			out <- n * n
		}
	}()
	return out
}

func main() {
	in := gen(1, 2, 3, 4)
	// BEGIN
	c1, c2 := square(in), square(in)
	// END
	var all []int
	for c1 != nil || c2 != nil {
		select {
		case v, ok := <-c1:
			if !ok {
				c1 = nil
				continue
			}
			all = append(all, v)
		case v, ok := <-c2:
			if !ok {
				c2 = nil
				continue
			}
			all = append(all, v)
		}
	}
	sort.Ints(all)
	fmt.Println(all)
}
```

### Check
Q: In fan-out, how many workers process a given item?
T: mcq
- [ ] All of them
- [x] Exactly one
- [ ] Two
- [ ] It depends on buffering
E: An item received from a channel is delivered to a single receiver.

Q: Is output order guaranteed to match input order after a fan-out?
T: tf
A: false
E: Workers finish at different times; index or sort results to restore order.

## Fan-In
slug: fan-in
minutes: 7
challenges: merge-channels
objectives: Merge multiple channels into one; Close the merged channel exactly once; Use a WaitGroup to know when all inputs are drained
takeaways: Fan-in multiplexes many input channels onto one output channel; One goroutine per input forwards values; a WaitGroup closes the output when all inputs end

### Concept
```go norun
func merge(cs ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup
	wg.Add(len(cs))
	for _, c := range cs {
		go func() {
			defer wg.Done()
			for v := range c {
				out <- v
			}
		}()
	}
	go func() { wg.Wait(); close(out) }()
	return out
}
```
The closer goroutine is essential: `out` must be closed exactly once, after **all** forwarders exit. (A `select` over a fixed number of channels works too but doesn't scale to N inputs.)

### Example
```go
package main

import (
	"fmt"
	"sort"
	"sync"
)

func gen(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			out <- n
		}
	}()
	return out
}

func merge(cs ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup
	wg.Add(len(cs))
	for _, c := range cs {
		go func() {
			defer wg.Done()
			for v := range c {
				out <- v
			}
		}()
	}
	go func() { wg.Wait(); close(out) }()
	return out
}

func main() {
	var all []int
	for v := range merge(gen(1, 2), gen(3, 4), gen(5)) {
		all = append(all, v)
	}
	sort.Ints(all)
	fmt.Println(all)
}
```

### Exercise
Write `merge(cs ...<-chan int) <-chan int`. Print the sum of merging `gen(1,2,3)` and `gen(10,20)`.
```text expect
36
```
```go solution
package main

import (
	"fmt"
	"sync"
)

func gen(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			out <- n
		}
	}()
	return out
}

// BEGIN
func merge(cs ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup
	wg.Add(len(cs))
	for _, c := range cs {
		go func() {
			defer wg.Done()
			for v := range c {
				out <- v
			}
		}()
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

// END

func main() {
	sum := 0
	for v := range merge(gen(1, 2, 3), gen(10, 20)) {
		sum += v
	}
	fmt.Println(sum)
}
```

### Check
Q: When should the merged output channel be closed?
T: mcq
- [ ] When the first input closes
- [x] After every input has been fully drained
- [ ] Immediately after starting the forwarders
- [ ] Never
E: Closing early would cause a panic when remaining forwarders send.

Q: Fan-in preserves the relative order of items across different input channels.
T: tf
A: false
E: Items from different inputs interleave nondeterministically; only order within one input is preserved.

## Pipelines
slug: pipelines
minutes: 8
challenges: pipeline-stages
objectives: Chain stages connected by channels; Make each stage own and close its output channel; Combine stages of different speeds
takeaways: A pipeline is a series of stages linked by channels: generate → transform → consume; Each stage receives on one channel and sends on another, closing its output when its input closes; Backpressure is automatic with unbuffered channels

### Concept
```go norun
nums := gen(1, 2, 3)
sq := square(nums)
out := double(sq)
for v := range out { ... }
```
Each stage is a function `func(in <-chan T) <-chan U` that starts a goroutine, ranges over `in`, sends results to a fresh `out`, and `defer close(out)`. Stages run concurrently, and a slow stage naturally slows the ones before it (backpressure). To be leak-free stages must also stop when downstream stops — that's what **cancellation** (`context`) adds in the next lessons.

### Example
```go
package main

import "fmt"

func gen(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			out <- n
		}
	}()
	return out
}

func stage(in <-chan int, f func(int) int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for v := range in {
			out <- f(v)
		}
	}()
	return out
}

func main() {
	out := stage(stage(gen(1, 2, 3, 4), func(n int) int { return n * n }), func(n int) int { return n + 1 })
	for v := range out {
		fmt.Print(v, " ")
	}
	fmt.Println()
}
```

### Exercise
Add a `filterEven` stage that forwards only even numbers, and run `gen(1..6) → filterEven → square`. Print the results one per line.
```text expect
4
16
36
```
```go solution
package main

import "fmt"

func gen(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			out <- n
		}
	}()
	return out
}

func square(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for v := range in {
			out <- v * v
		}
	}()
	return out
}

// BEGIN
func filterEven(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for v := range in {
			if v%2 == 0 {
				out <- v
			}
		}
	}()
	return out
}

// END

func main() {
	for v := range square(filterEven(gen(1, 2, 3, 4, 5, 6))) {
		fmt.Println(v)
	}
}
```

### Check
Q: What must each pipeline stage do when its input channel is closed and drained?
T: mcq
- [ ] Panic
- [x] Close its own output channel
- [ ] Keep running
- [ ] Restart
E: Closing propagates completion down the pipeline so the final consumer's range ends.

Q: What provides natural backpressure in a pipeline of unbuffered channels?
T: mcq
- [ ] Mutexes
- [x] A send blocks until the next stage is ready to receive
- [ ] The scheduler
- [ ] Panics
E: A slow stage makes upstream sends block, throttling producers automatically.

## Context
slug: context
minutes: 9
objectives: Pass a context through call chains; Derive child contexts and know cancellation propagates downward; Store only request-scoped values, sparingly
takeaways: context.Context carries cancellation signals, deadlines and request-scoped values across API boundaries; Cancelling a parent cancels all its children; ctx is the first parameter of functions that block or do I/O

### Concept
```go norun
ctx, cancel := context.WithCancel(parent)
defer cancel()
```
A context forms a tree: cancelling a node cancels all descendants. Blocking code must respect it:

```go norun
select {
case <-ctx.Done():
	return ctx.Err()
case res := <-work:
	return use(res)
}
```
`context.WithValue` carries request-scoped data (request ID, auth principal) — use **typed unexported keys**, never for optional parameters. Do not store contexts in structs; pass them explicitly. `context.Cause` / `WithCancelCause` (Go 1.20+) explains *why* something was cancelled.

### Example
```go
package main

import (
	"context"
	"fmt"
)

type ctxKey struct{}

func main() {
	parent, cancel := context.WithCancel(context.Background())
	child := context.WithValue(parent, ctxKey{}, "req-42")
	fmt.Println(child.Value(ctxKey{}), child.Err())
	cancel()
	<-child.Done() // cancellation propagates to children
	fmt.Println(child.Err())
}
```

### Exercise
Write `worker(ctx context.Context, out chan<- int)` that sends 0, 1, 2, … until `ctx` is cancelled, then returns. In `main`, read three values, cancel, and print `stopped`.
```text expect
0 1 2 stopped
```
```go solution
package main

import (
	"context"
	"fmt"
)

// BEGIN
func worker(ctx context.Context, out chan<- int) {
	for i := 0; ; i++ {
		select {
		case <-ctx.Done():
			return
		case out <- i:
		}
	}
}

// END

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	out := make(chan int)
	stopped := make(chan struct{})
	go func() {
		worker(ctx, out)
		close(stopped)
	}()
	fmt.Print(<-out, " ", <-out, " ", <-out, " ")
	cancel()
	<-stopped
	fmt.Println("stopped")
}
```

### Check
Q: If a parent context is cancelled, what happens to its children?
T: mcq
- [ ] Nothing
- [x] They are cancelled too
- [ ] They keep running until their own deadline
- [ ] They panic
E: Cancellation propagates down the tree.

Q: Should optional function parameters be passed through `context.WithValue`?
T: tf
A: false
E: Values are for request-scoped data crossing API boundaries, not a way to smuggle arguments.

## Cancellation
slug: cancellation
minutes: 7
challenges: cancellable-worker
objectives: Stop goroutines cleanly with context cancellation; Always call the cancel function; Clean up resources after cancellation
takeaways: Cancellation is cooperative: goroutines must check ctx.Done(); defer cancel() to avoid leaking timers and child contexts; After ctx.Done() fires, release resources and return promptly

### Concept
Cancel work when the result is no longer needed: the client disconnected, a sibling failed, the user pressed Ctrl+C.

```go norun
ctx, cancel := context.WithCancel(ctx)
defer cancel()          // always
go worker(ctx)
```
A goroutine must *observe* cancellation. In loops check `ctx.Err()` or `select` on `ctx.Done()`; for blocking calls use context-aware APIs (`http.NewRequestWithContext`, `db.QueryContext`). Return `ctx.Err()` so callers can tell cancellation from failure with `errors.Is(err, context.Canceled)`.

### Example
```go
package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func poll(ctx context.Context) error {
	t := time.NewTicker(5 * time.Millisecond)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
		}
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()
	err := poll(ctx)
	fmt.Println(errors.Is(err, context.Canceled))
}
```

### Exercise
Write `countUntilCancel(ctx context.Context) int` that counts loop iterations (checking `ctx.Err()` each time) and returns the count when cancelled. Print `true` if the count is positive.
```text expect
true
```
```go solution
package main

import (
	"context"
	"fmt"
	"time"
)

// BEGIN
func countUntilCancel(ctx context.Context) int {
	n := 0
	for ctx.Err() == nil {
		n++
	}
	return n
}

// END

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	fmt.Println(countUntilCancel(ctx) > 0)
}
```

### Check
Q: Why should you `defer cancel()` even when the context will time out anyway?
T: mcq
- [ ] It's required to start the timer
- [x] It releases the timer and child-context resources immediately
- [ ] It logs the cancellation
- [ ] It's optional and useless
E: Not calling cancel leaks resources until the parent context ends.

Q: How do callers tell a cancelled operation from a failed one?
T: mcq
- [ ] By the error message text
- [x] errors.Is(err, context.Canceled)
- [ ] By checking the return code
- [ ] They can't
E: `context.Canceled` (and `DeadlineExceeded`) are sentinel errors you can test with `errors.Is`.

## Timeouts
slug: timeouts
minutes: 6
objectives: Bound operations with context.WithTimeout; Use time.After in select for simple timeouts; Understand that timeouts must be enforced by the running code
takeaways: context.WithTimeout cancels automatically after a duration; A timeout only works if the operation checks the context; Set timeouts on every network call — the defaults are infinite

### Concept
```go norun
ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
defer cancel()
req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
resp, err := client.Do(req) // returns context.DeadlineExceeded if too slow
```
Timeouts turn "hang forever" into "fail fast". Choose them per dependency and propagate: a request that has 5 s in total shouldn't give each of five calls 5 s. `http.Client{Timeout: ...}` and `http.Server{ReadTimeout, WriteTimeout, IdleTimeout}` exist for the whole-request level.

### Example
```go
package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func slowOp(ctx context.Context) error {
	select {
	case <-time.After(100 * time.Millisecond):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	err := slowOp(ctx)
	fmt.Println(errors.Is(err, context.DeadlineExceeded))
}
```

### Exercise
Write `withTimeout(d time.Duration, work time.Duration) error` that runs a pretend operation lasting `work` under a context timeout of `d`, returning `nil` if it finishes and `context.DeadlineExceeded` otherwise.
```text expect
<nil>
context deadline exceeded
```
```go solution
package main

import (
	"context"
	"fmt"
	"time"
)

// BEGIN
func withTimeout(d, work time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), d)
	defer cancel()
	select {
	case <-time.After(work):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// END

func main() {
	fmt.Println(withTimeout(200*time.Millisecond, 5*time.Millisecond))
	fmt.Println(withTimeout(5*time.Millisecond, 200*time.Millisecond))
}
```

### Check
Q: Does `context.WithTimeout` forcibly stop running code?
T: mcq
- [ ] Yes, it kills the goroutine
- [x] No — it signals; the code must watch ctx.Done()
- [ ] Only for network calls
- [ ] Only in main
E: Cancellation is cooperative.

Q: What is the default timeout of `http.Get`?
T: mcq
- [ ] 30 seconds
- [ ] 5 seconds
- [x] None — it can wait forever
- [ ] 1 minute
E: `http.DefaultClient` has no timeout; always configure one.

## Deadlines
slug: deadlines
minutes: 6
objectives: Use context.WithDeadline for absolute cut-off times; Read a context's deadline with ctx.Deadline(); Budget remaining time across sub-calls
takeaways: WithDeadline sets an absolute time; WithTimeout is WithDeadline(now+d); ctx.Deadline() lets callees adapt their work to the time left; Child contexts can shorten but never extend a parent's deadline

### Concept
```go norun
deadline := time.Now().Add(3 * time.Second)
ctx, cancel := context.WithDeadline(ctx, deadline)
defer cancel()

if dl, ok := ctx.Deadline(); ok {
	remaining := time.Until(dl)
}
```
A deadline is an instant; it stays fixed as the request travels through services, unlike a timeout which restarts at each hop. A callee can inspect the remaining time (skip an optional slow step when there is not enough). A derived context cannot extend its parent's deadline — the earlier one wins.

### Example
```go
package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	parent, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	child, cancel2 := context.WithTimeout(parent, time.Hour) // can't extend
	defer cancel2()
	pd, _ := parent.Deadline()
	cd, _ := child.Deadline()
	fmt.Println(pd.Equal(cd))
}
```

### Exercise
Write `hasTimeLeft(ctx context.Context, need time.Duration) bool` returning true if the context has no deadline or at least `need` remains.
```text expect
true false true
```
```go solution
package main

import (
	"context"
	"fmt"
	"time"
)

// BEGIN
func hasTimeLeft(ctx context.Context, need time.Duration) bool {
	dl, ok := ctx.Deadline()
	if !ok {
		return true
	}
	return time.Until(dl) >= need
}

// END

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	short, cancel2 := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel2()
	fmt.Println(hasTimeLeft(ctx, time.Second), hasTimeLeft(short, time.Second), hasTimeLeft(context.Background(), time.Hour))
}
```

### Check
Q: If a parent context expires in 1 s and you derive a child with a 1-hour timeout, when does the child expire?
T: mcq
- [ ] After 1 hour
- [x] After 1 second — the earlier deadline wins
- [ ] Never
- [ ] After 1 hour and 1 second
E: A child can only shorten the parent's deadline.

Q: `ctx.Deadline()` returns a second value telling you...
T: mcq
- [ ] The error
- [x] Whether a deadline is set at all
- [ ] The remaining seconds
- [ ] The parent context
E: `(deadline, ok)`: `ok` is false for contexts without a deadline.

## Graceful Shutdown
slug: graceful-shutdown
minutes: 9
challenges: graceful-stop
objectives: Catch SIGINT/SIGTERM to begin shutdown; Stop accepting work and drain in-flight jobs; Enforce a shutdown deadline
takeaways: Shutdown order: stop intake → drain in-flight work → close resources → exit; Use signal.NotifyContext plus http.Server.Shutdown with a deadline; If draining exceeds the deadline, exit anyway — a hung shutdown is worse than an abrupt one

### Concept
```go norun
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()

go srv.ListenAndServe()
<-ctx.Done()                                         // signal received

shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
defer cancel()
_ = srv.Shutdown(shutdownCtx)                        // stops listener, waits for handlers
```
Kubernetes sends SIGTERM and after `terminationGracePeriodSeconds` sends SIGKILL, so your drain deadline must be shorter than that. Workers: close the jobs channel, wait for them with a `WaitGroup` (bounded by the deadline), then close databases.

### Example
```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	jobs := make(chan int, 10)
	var wg sync.WaitGroup
	for w := 0; w < 2; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range jobs { // keeps draining until closed
			}
		}()
	}
	for i := 0; i < 5; i++ {
		jobs <- i
	}
	close(jobs) // stop intake
	wg.Wait()   // drain
	fmt.Println("clean shutdown")
}
```

### Exercise
Write `drain(jobs chan int, workers int, process func(int)) ` that starts workers, closes intake (`jobs`) is done by the caller — your function must wait for all workers to finish. Process 5 jobs and print `processed 5`.
```text expect
processed 5
```
```go solution
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// BEGIN
func drain(jobs <-chan int, workers int, process func(int)) {
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				process(j)
			}
		}()
	}
	wg.Wait()
}

// END

func main() {
	jobs := make(chan int)
	var n atomic.Int32
	go func() {
		for i := 0; i < 5; i++ {
			jobs <- i
		}
		close(jobs)
	}()
	drain(jobs, 3, func(int) { n.Add(1) })
	fmt.Println("processed", n.Load())
}
```

### Check
Q: Which signal do orchestrators like Kubernetes send to request a graceful stop?
T: short
A: SIGTERM
E: SIGTERM asks politely; SIGKILL follows after the grace period.

Q: What should a graceful shutdown do first?
T: mcq
- [ ] Close the database
- [x] Stop accepting new work
- [ ] Exit immediately
- [ ] Discard in-flight jobs
E: Stop intake first, then drain what is already running, then release resources.

## Goroutine Leaks
slug: goroutine-leaks
minutes: 8
challenges: fix-goroutine-leak
objectives: Recognise the common causes of leaked goroutines; Detect leaks with runtime.NumGoroutine and tests; Fix them with buffered results, contexts and closed channels
takeaways: A leak is a goroutine that can never finish — usually blocked on a channel; Typical causes: unbuffered result nobody reads, missing close, no cancellation path; Every goroutine you start needs a known way to end

### Concept
Classic leak:

```go norun
func first(urls []string) string {
	ch := make(chan string) // unbuffered
	for _, u := range urls {
		go func() { ch <- fetch(u) }() // losers block forever!
	}
	return <-ch
}
```
Only one goroutine's send is received; the rest block forever holding memory. Fixes: **buffer** the channel to `len(urls)`, or pass a `ctx` and `select` on `ctx.Done()` around the send.

Detect with `runtime.NumGoroutine()` (before/after in tests), `go.uber.org/goleak`, or the `pprof` goroutine profile in production (Module 21).

### Example
```go
package main

import (
	"fmt"
	"runtime"
	"time"
)

func leaky() {
	ch := make(chan int)
	go func() { ch <- 1 }() // nobody receives: leaked
}

func main() {
	before := runtime.NumGoroutine()
	for i := 0; i < 10; i++ {
		leaky()
	}
	time.Sleep(20 * time.Millisecond)
	fmt.Println("leaked:", runtime.NumGoroutine()-before)
}
```

### Exercise
Fix `first`: make the result channel buffered so the slower goroutines can send and exit. `main` prints the winner and `leaked: 0` after a short wait.
```text expect
fast
leaked: 0
```
```go solution
package main

import (
	"fmt"
	"runtime"
	"time"
)

func fetch(name string, d time.Duration) string {
	time.Sleep(d)
	return name
}

func first() string {
	// BEGIN
	ch := make(chan string, 3) // buffered: no sender ever blocks
	// END
	go func() { ch <- fetch("slow", 60*time.Millisecond) }()
	go func() { ch <- fetch("fast", 1*time.Millisecond) }()
	go func() { ch <- fetch("medium", 30*time.Millisecond) }()
	return <-ch
}

func main() {
	before := runtime.NumGoroutine()
	fmt.Println(first())
	time.Sleep(150 * time.Millisecond)
	fmt.Println("leaked:", runtime.NumGoroutine()-before)
}
```

### Check
Q: What most commonly causes a goroutine leak?
T: mcq
- [ ] Too many CPUs
- [x] Blocking forever on a channel operation nobody will complete
- [ ] Using WaitGroup
- [ ] Calling close twice
E: A send or receive with no counterpart keeps the goroutine (and its memory) alive.

Q: Buffering a result channel to the number of senders guarantees none of them blocks on send.
T: tf
A: true
E: If capacity ≥ number of sends, every send completes immediately, so the goroutines can exit even if nobody reads.

## Race Conditions
slug: race-conditions
minutes: 8
challenges: compute-once-cache
objectives: Define a data race; Detect races with go test -race and go run -race; Fix races with locks, atomics or by confining data to one goroutine
takeaways: A data race is unsynchronised concurrent access where at least one access is a write; Races cause lost updates and corrupted maps — and the outcome is undefined; Run tests with -race in CI; fix with mutexes, atomics or channels

### Concept
```go norun
counter := 0
for i := 0; i < 1000; i++ {
	go func() { counter++ }() // DATA RACE: read-modify-write is not atomic
}
```
`counter++` is three steps (load, add, store); interleaved goroutines overwrite each other. The result is unpredictable, and concurrent map writes crash the program (`fatal error: concurrent map writes`).

Find them: **`go test -race ./...`** and **`go run -race`** instrument memory accesses and report the two conflicting goroutines with stack traces. It only finds races that *happen* during the run, so run it in CI with good tests.

Fix by (1) not sharing (confine data to one goroutine, pass copies), (2) a mutex, (3) atomics, or (4) channels that transfer ownership.

### Example
```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	var (
		wg sync.WaitGroup
		mu sync.Mutex
		n  int
	)
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			n++ // safe: protected by mu
			mu.Unlock()
		}()
	}
	wg.Wait()
	fmt.Println(n)
}
```

### Exercise
Fix the racy `total` computation using `atomic.AddInt64` on an `int64` (no mutex). It should always print `1000`.
```text expect
1000
```
```go solution
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func main() {
	var wg sync.WaitGroup
	var total int64
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// BEGIN
			atomic.AddInt64(&total, 1)
			// END
		}()
	}
	wg.Wait()
	fmt.Println(atomic.LoadInt64(&total))
}
```

### Check
Q: Which command reports data races in tests?
T: short
A: go test -race
E: The `-race` flag builds with the race detector and reports conflicting accesses.

Q: `counter++` from many goroutines without synchronisation is safe because it is a single instruction.
T: tf
A: false
E: It compiles to load, add and store; interleaving causes lost updates — a data race.
