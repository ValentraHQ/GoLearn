# Kubernetes Challenges

## reconcile-plan
title: Reconciliation Plan
difficulty: advanced
module: kubernetes
lesson: kubernetes/reconciliation
skill: kubernetes

### Problem
A controller compares **desired** and **actual** Deployments (both `map[string]Spec`, keyed by name) and decides what to do.

```go
type Spec struct { Image string; Replicas int }
type Action struct { Op, Name string } // Op is "create", "update" or "delete"
```

Write `Plan(desired, actual map[string]Spec) []Action`:

- name in desired but not actual → `create`
- name in both with a different `Spec` → `update`
- name in actual but not desired → `delete`
- identical → no action

Return actions ordered: all creates, then all updates, then all deletes; within each group sort by name. Applying the plan and planning again must produce **no** actions (idempotence).

### Constraints
- Deterministic output despite map iteration order
- Do not modify the input maps

### Starter
```go
package main

type Spec struct {
	Image    string
	Replicas int
}

type Action struct {
	Op   string
	Name string
}

func Plan(desired, actual map[string]Spec) []Action {
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

func TestPlan(t *testing.T) {
	desired := map[string]Spec{
		"web":    {"nginx:1.27", 3},
		"api":    {"api:2", 2},
		"worker": {"worker:1", 1},
	}
	actual := map[string]Spec{
		"web":   {"nginx:1.26", 3}, // image differs
		"api":   {"api:2", 2},      // identical
		"old":   {"old:1", 1},      // not desired
		"legacy": {"legacy:9", 4},  // not desired
	}
	got := Plan(desired, actual)
	want := []Action{
		{"create", "worker"},
		{"update", "web"},
		{"delete", "legacy"},
		{"delete", "old"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got  %v\nwant %v", got, want)
	}
}

func apply(actual map[string]Spec, desired map[string]Spec, actions []Action) map[string]Spec {
	out := map[string]Spec{}
	for k, v := range actual {
		out[k] = v
	}
	for _, a := range actions {
		switch a.Op {
		case "create", "update":
			out[a.Name] = desired[a.Name]
		case "delete":
			delete(out, a.Name)
		}
	}
	return out
}

func TestIdempotent(t *testing.T) {
	desired := map[string]Spec{"a": {"x", 1}, "b": {"y", 2}}
	actual := map[string]Spec{"b": {"y", 1}, "c": {"z", 1}}
	first := Plan(desired, actual)
	if len(first) != 3 {
		t.Fatalf("first plan: %v", first)
	}
	converged := apply(actual, desired, first)
	if again := Plan(desired, converged); len(again) != 0 {
		t.Errorf("plan after converging should be empty, got %v", again)
	}
}

func TestEmptyAndUntouchedInputs(t *testing.T) {
	if got := Plan(nil, nil); len(got) != 0 {
		t.Errorf("nil/nil: %v", got)
	}
	desired := map[string]Spec{"a": {"x", 1}}
	actual := map[string]Spec{}
	Plan(desired, actual)
	if len(actual) != 0 || len(desired) != 1 {
		t.Error("inputs were modified")
	}
}
```

### Hints
- Collect names of each category into separate slices and `sort.Strings` each
- `Spec` is comparable, so `desired[n] != actual[n]` works

### Solution
```go
package main

import "sort"

type Spec struct {
	Image    string
	Replicas int
}

type Action struct {
	Op   string
	Name string
}

func Plan(desired, actual map[string]Spec) []Action {
	var creates, updates, deletes []string
	for name, d := range desired {
		a, ok := actual[name]
		switch {
		case !ok:
			creates = append(creates, name)
		case a != d:
			updates = append(updates, name)
		}
	}
	for name := range actual {
		if _, ok := desired[name]; !ok {
			deletes = append(deletes, name)
		}
	}
	sort.Strings(creates)
	sort.Strings(updates)
	sort.Strings(deletes)
	var out []Action
	for _, n := range creates {
		out = append(out, Action{"create", n})
	}
	for _, n := range updates {
		out = append(out, Action{"update", n})
	}
	for _, n := range deletes {
		out = append(out, Action{"delete", n})
	}
	return out
}

func main() {}
```

### Explanation
A controller's core is a pure diff of desired versus actual state. Because the plan depends only on the two states — not on history — it is idempotent and safe to run repeatedly.

## workqueue-dedupe
title: Deduplicating Work Queue
difficulty: expert
module: kubernetes
lesson: kubernetes/controllers
skill: kubernetes

### Problem
Implement the semantics of a Kubernetes controller work queue:

- `NewQueue() *Queue`
- `Add(key string)` — if `key` is already **waiting**, do nothing. If it is currently **being processed**, remember it so it is queued again once `Done` is called. Otherwise enqueue it (FIFO).
- `Get() (key string, shutdown bool)` — blocks until a key is available and marks it as *processing*. After `ShutDown` and once the queue is empty, returns `"", true`.
- `Done(key string)` — marks processing finished; if the key was re-added meanwhile, enqueue it now.
- `Len() int` — number of waiting keys.
- `ShutDown()` — no more keys are accepted; blocked `Get` calls wake up.

A key is never handed to two workers at the same time.

### Constraints
- Safe for concurrent use
- FIFO order for distinct keys
- A key being processed is never given out again until `Done`

### Starter
```go
package main

type Queue struct {
}

func NewQueue() *Queue {
	return &Queue{}
}

func (q *Queue) Add(key string)          {}
func (q *Queue) Get() (string, bool)     { return "", true }
func (q *Queue) Done(key string)         {}
func (q *Queue) Len() int                { return 0 }
func (q *Queue) ShutDown()               {}

func main() {}
```

### Tests
```go
package main

import (
	"sync"
	"testing"
	"time"
)

func TestDedupeAndOrder(t *testing.T) {
	q := NewQueue()
	for _, k := range []string{"a", "b", "a", "c", "b"} {
		q.Add(k)
	}
	if q.Len() != 3 {
		t.Fatalf("Len = %d, want 3", q.Len())
	}
	var got []string
	for i := 0; i < 3; i++ {
		k, sd := q.Get()
		if sd {
			t.Fatal("unexpected shutdown")
		}
		got = append(got, k)
		q.Done(k)
	}
	if got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Errorf("order = %v", got)
	}
}

func TestReAddWhileProcessing(t *testing.T) {
	q := NewQueue()
	q.Add("x")
	k, _ := q.Get()
	q.Add("x") // arrives while x is being processed
	if q.Len() != 0 {
		t.Errorf("x must not be queued while processing, Len = %d", q.Len())
	}
	q.Done(k)
	if q.Len() != 1 {
		t.Errorf("x should be queued again after Done, Len = %d", q.Len())
	}
	k2, _ := q.Get()
	if k2 != "x" {
		t.Errorf("got %q", k2)
	}
	q.Done(k2)
	if q.Len() != 0 {
		t.Errorf("Len = %d", q.Len())
	}
}

func TestGetBlocksUntilAdd(t *testing.T) {
	q := NewQueue()
	got := make(chan string, 1)
	go func() {
		k, _ := q.Get()
		got <- k
	}()
	select {
	case <-got:
		t.Fatal("Get returned before anything was added")
	case <-time.After(30 * time.Millisecond):
	}
	q.Add("late")
	select {
	case k := <-got:
		if k != "late" {
			t.Errorf("got %q", k)
		}
	case <-time.After(time.Second):
		t.Fatal("Get did not wake up after Add")
	}
}

func TestShutdown(t *testing.T) {
	q := NewQueue()
	q.Add("a")
	done := make(chan bool, 1)
	q.ShutDown()
	q.Add("ignored")
	k, sd := q.Get() // queued item is still delivered
	if sd || k != "a" {
		t.Fatalf("got %q %v, want the queued item first", k, sd)
	}
	q.Done(k)
	go func() {
		_, sd := q.Get()
		done <- sd
	}()
	select {
	case sd := <-done:
		if !sd {
			t.Error("drained queue after ShutDown should report shutdown")
		}
	case <-time.After(time.Second):
		t.Fatal("Get blocked after ShutDown")
	}
}

func TestShutdownWakesBlockedGet(t *testing.T) {
	q := NewQueue()
	done := make(chan bool, 1)
	go func() {
		_, sd := q.Get()
		done <- sd
	}()
	time.Sleep(20 * time.Millisecond)
	q.ShutDown()
	select {
	case sd := <-done:
		if !sd {
			t.Error("want shutdown=true")
		}
	case <-time.After(time.Second):
		t.Fatal("blocked Get was not woken by ShutDown")
	}
}

func TestNeverProcessedConcurrently(t *testing.T) {
	q := NewQueue()
	var mu sync.Mutex
	active := map[string]int{}
	var wg sync.WaitGroup
	for w := 0; w < 4; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				k, sd := q.Get()
				if sd {
					return
				}
				mu.Lock()
				active[k]++
				if active[k] > 1 {
					t.Errorf("key %q handed to two workers at once", k)
				}
				mu.Unlock()
				time.Sleep(time.Millisecond)
				mu.Lock()
				active[k]--
				mu.Unlock()
				q.Done(k)
			}
		}()
	}
	for i := 0; i < 200; i++ {
		q.Add([]string{"a", "b", "c"}[i%3])
	}
	time.Sleep(100 * time.Millisecond)
	q.ShutDown()
	wg.Wait()
}
```

### Hints
- Keep three pieces of state: a FIFO slice of waiting keys, a `dirty` set (keys that need processing) and a `processing` set
- `Add`: skip if shut down or already dirty; mark dirty; if not processing, append to the queue and `Signal`
- `Get`: wait on a `sync.Cond` while the queue is empty and not shut down; pop, add to `processing`, remove from `dirty`
- `Done`: remove from `processing`; if the key is dirty again, re-queue it

### Solution
```go
package main

import "sync"

type Queue struct {
	mu         sync.Mutex
	cond       *sync.Cond
	queue      []string
	dirty      map[string]bool
	processing map[string]bool
	shutdown   bool
}

func NewQueue() *Queue {
	q := &Queue{dirty: map[string]bool{}, processing: map[string]bool{}}
	q.cond = sync.NewCond(&q.mu)
	return q
}

func (q *Queue) Add(key string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.shutdown || q.dirty[key] {
		return
	}
	q.dirty[key] = true
	if q.processing[key] {
		return // will be re-queued by Done
	}
	q.queue = append(q.queue, key)
	q.cond.Signal()
}

func (q *Queue) Get() (string, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for len(q.queue) == 0 && !q.shutdown {
		q.cond.Wait()
	}
	if len(q.queue) == 0 {
		return "", true
	}
	key := q.queue[0]
	q.queue = q.queue[1:]
	q.processing[key] = true
	delete(q.dirty, key)
	return key, false
}

func (q *Queue) Done(key string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	delete(q.processing, key)
	if q.dirty[key] {
		q.queue = append(q.queue, key)
		q.cond.Signal()
	}
}

func (q *Queue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.queue)
}

func (q *Queue) ShutDown() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.shutdown = true
	q.cond.Broadcast()
}

func main() {}
```

### Explanation
This is the essential algorithm behind `client-go`'s `workqueue`. The `dirty` set deduplicates waiting keys; the `processing` set guarantees a key is never worked on by two workers at once while still remembering updates that arrive mid-flight, so no change is lost.
