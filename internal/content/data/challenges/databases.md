# Database Challenges

## sql-query-builder
title: Safe SELECT Builder
difficulty: intermediate
module: databases
lesson: databases/sql-fundamentals
skill: databases

### Problem
Write `Select(table string, cols []string, filters []Filter, limit int) (string, []any, error)` that builds a parameterised PostgreSQL query.

`type Filter struct { Column, Op string; Value any }`

- Result format: `SELECT c1, c2 FROM t WHERE a = $1 AND b >= $2 LIMIT 10` (no `WHERE` without filters; no `LIMIT` when `limit <= 0`)
- Values go to the returned args slice in placeholder order — never into the SQL text
- Table and column names must match `^[a-z_][a-z0-9_]*$`; otherwise return an error
- Allowed operators: `=`, `<>`, `<`, `<=`, `>`, `>=`, `LIKE` (anything else is an error)
- `cols` must not be empty

### Constraints
- Never interpolate values into the SQL string
- Validate every identifier and operator

### Starter
```go
package main

type Filter struct {
	Column string
	Op     string
	Value  any
}

func Select(table string, cols []string, filters []Filter, limit int) (string, []any, error) {
	return "", nil, nil
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

func TestSelect(t *testing.T) {
	q, args, err := Select("users", []string{"id", "name"}, []Filter{{"age", ">=", 18}, {"name", "LIKE", "A%"}}, 10)
	if err != nil {
		t.Fatal(err)
	}
	if want := "SELECT id, name FROM users WHERE age >= $1 AND name LIKE $2 LIMIT 10"; q != want {
		t.Errorf("query = %q\nwant    %q", q, want)
	}
	if !reflect.DeepEqual(args, []any{18, "A%"}) {
		t.Errorf("args = %v", args)
	}
}

func TestSelectNoFilters(t *testing.T) {
	q, args, err := Select("orders", []string{"id"}, nil, 0)
	if err != nil || q != "SELECT id FROM orders" || len(args) != 0 {
		t.Errorf("%q %v %v", q, args, err)
	}
}

func TestInjectionAttemptsRejected(t *testing.T) {
	cases := []struct {
		name  string
		table string
		cols  []string
		fs    []Filter
	}{
		{"table", "users; DROP TABLE users", []string{"id"}, nil},
		{"column", "users", []string{"id, password"}, nil},
		{"filter column", "users", []string{"id"}, []Filter{{"name = 'x' OR 1=1 --", "=", 1}}},
		{"operator", "users", []string{"id"}, []Filter{{"name", "= 1 OR", 1}}},
		{"no columns", "users", nil, nil},
		{"uppercase table", "Users", []string{"id"}, nil},
	}
	for _, c := range cases {
		if q, _, err := Select(c.table, c.cols, c.fs, 0); err == nil {
			t.Errorf("%s: expected an error, got %q", c.name, q)
		}
	}
}

func TestValueNeverInSQL(t *testing.T) {
	q, args, _ := Select("users", []string{"id"}, []Filter{{"name", "=", "Robert'); DROP TABLE users;--"}}, 0)
	if q != "SELECT id FROM users WHERE name = $1" || len(args) != 1 {
		t.Errorf("%q %v", q, args)
	}
}
```

### Hints
- Compile `^[a-z_][a-z0-9_]*$` once with `regexp.MustCompile`
- Use `fmt.Sprintf("%s %s $%d", col, op, i+1)` for each condition
- An allow-list `map[string]bool` handles operators

### Solution
```go
package main

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type Filter struct {
	Column string
	Op     string
	Value  any
}

var identRe = regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)

var allowedOps = map[string]bool{"=": true, "<>": true, "<": true, "<=": true, ">": true, ">=": true, "LIKE": true}

func Select(table string, cols []string, filters []Filter, limit int) (string, []any, error) {
	if !identRe.MatchString(table) {
		return "", nil, fmt.Errorf("invalid table name %q", table)
	}
	if len(cols) == 0 {
		return "", nil, errors.New("no columns selected")
	}
	for _, c := range cols {
		if !identRe.MatchString(c) {
			return "", nil, fmt.Errorf("invalid column name %q", c)
		}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "SELECT %s FROM %s", strings.Join(cols, ", "), table)
	var args []any
	for i, f := range filters {
		if !identRe.MatchString(f.Column) {
			return "", nil, fmt.Errorf("invalid filter column %q", f.Column)
		}
		if !allowedOps[f.Op] {
			return "", nil, fmt.Errorf("operator %q not allowed", f.Op)
		}
		if i == 0 {
			b.WriteString(" WHERE ")
		} else {
			b.WriteString(" AND ")
		}
		fmt.Fprintf(&b, "%s %s $%d", f.Column, f.Op, i+1)
		args = append(args, f.Value)
	}
	if limit > 0 {
		fmt.Fprintf(&b, " LIMIT %d", limit)
	}
	return b.String(), args, nil
}

func main() {}
```

### Explanation
Values become `$n` parameters; the only text that reaches the SQL string is validated identifiers and allow-listed operators, so there is no injection surface.

## with-tx
title: Transaction Runner
difficulty: advanced
module: databases
lesson: databases/transactions
skill: databases

### Problem
Implement `WithTx(begin func() (Tx, error), fn func(Tx) error) error` with this contract:

- If `begin` fails, return its error (do not call `fn`).
- If `fn` returns an error → roll back and return **fn's** error (ignore the rollback error).
- If `fn` panics → roll back, then re-panic with the same value.
- Otherwise commit and return the commit error (a failed commit must not be followed by a rollback call).

`Tx` is `interface{ Commit() error; Rollback() error }`.

### Constraints
- Exactly one of Commit/Rollback is called per transaction
- Preserve the original panic value

### Starter
```go
package main

type Tx interface {
	Commit() error
	Rollback() error
}

func WithTx(begin func() (Tx, error), fn func(Tx) error) error {
	return nil
}

func main() {}
```

### Tests
```go
package main

import (
	"errors"
	"reflect"
	"testing"
)

type fakeTx struct {
	log         *[]string
	commitErr   error
	rollbackErr error
}

func (t *fakeTx) Commit() error   { *t.log = append(*t.log, "commit"); return t.commitErr }
func (t *fakeTx) Rollback() error { *t.log = append(*t.log, "rollback"); return t.rollbackErr }

func beginWith(tx *fakeTx) func() (Tx, error) { return func() (Tx, error) { return tx, nil } }

func TestCommit(t *testing.T) {
	var log []string
	err := WithTx(beginWith(&fakeTx{log: &log}), func(Tx) error { log = append(log, "work"); return nil })
	if err != nil || !reflect.DeepEqual(log, []string{"work", "commit"}) {
		t.Errorf("err=%v log=%v", err, log)
	}
}

func TestRollbackOnError(t *testing.T) {
	var log []string
	boom := errors.New("boom")
	tx := &fakeTx{log: &log, rollbackErr: errors.New("rollback failed")}
	err := WithTx(beginWith(tx), func(Tx) error { return boom })
	if !errors.Is(err, boom) || !reflect.DeepEqual(log, []string{"rollback"}) {
		t.Errorf("err=%v log=%v", err, log)
	}
}

func TestRollbackOnPanic(t *testing.T) {
	var log []string
	defer func() {
		if r := recover(); r != "oops" {
			t.Errorf("recovered %v, want the original panic value", r)
		}
		if !reflect.DeepEqual(log, []string{"rollback"}) {
			t.Errorf("log=%v", log)
		}
	}()
	_ = WithTx(beginWith(&fakeTx{log: &log}), func(Tx) error { panic("oops") })
	t.Error("expected a panic")
}

func TestCommitErrorIsReturnedWithoutRollback(t *testing.T) {
	var log []string
	ce := errors.New("commit failed")
	err := WithTx(beginWith(&fakeTx{log: &log, commitErr: ce}), func(Tx) error { return nil })
	if !errors.Is(err, ce) || !reflect.DeepEqual(log, []string{"commit"}) {
		t.Errorf("err=%v log=%v", err, log)
	}
}

func TestBeginError(t *testing.T) {
	be := errors.New("no connection")
	called := false
	err := WithTx(func() (Tx, error) { return nil, be }, func(Tx) error { called = true; return nil })
	if !errors.Is(err, be) || called {
		t.Errorf("err=%v fnCalled=%v", err, called)
	}
}
```

### Hints
- A deferred `recover()` can roll back and then `panic(r)` again
- Track whether the transaction finished (committed or rolled back) to avoid double calls

### Solution
```go
package main

type Tx interface {
	Commit() error
	Rollback() error
}

func WithTx(begin func() (Tx, error), fn func(Tx) error) (err error) {
	tx, err := begin()
	if err != nil {
		return err
	}
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func main() {}
```

### Explanation
Every path performs exactly one of commit or rollback. The deferred recover only runs the rollback for panics, then re-panics so callers still see the failure.

## memory-repository
title: In-Memory Repository
difficulty: intermediate
module: databases
lesson: databases/repository-pattern
skill: databases

### Problem
Implement `Repo`, an in-memory user repository with a **usable zero value**:

- `Create(email, name string) (User, error)` — assigns ids 1, 2, 3…; emails are unique (case-insensitive) → `ErrDuplicate`; empty email → `ErrInvalid`
- `Get(id int) (User, error)` → `ErrNotFound` if missing
- `Update(id int, name string) error` → `ErrNotFound` if missing
- `Delete(id int) error` → `ErrNotFound` if missing
- `List() []User` sorted by id

`User` has `ID int`, `Email string`, `Name string`. The errors and `User` are provided. Must be safe for concurrent use.

### Constraints
- Emails are compared case-insensitively but stored as given
- Deleting frees the email for reuse
- Race-free

### Starter
```go
package main

import "errors"

var (
	ErrNotFound  = errors.New("not found")
	ErrDuplicate = errors.New("duplicate email")
	ErrInvalid   = errors.New("invalid input")
)

type User struct {
	ID    int
	Email string
	Name  string
}

type Repo struct {
}

func (r *Repo) Create(email, name string) (User, error) {
	return User{}, nil
}

func (r *Repo) Get(id int) (User, error) {
	return User{}, nil
}

func (r *Repo) Update(id int, name string) error {
	return nil
}

func (r *Repo) Delete(id int) error {
	return nil
}

func (r *Repo) List() []User {
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
	"reflect"
	"sync"
	"testing"
)

func TestCreateGetList(t *testing.T) {
	var r Repo
	a, err := r.Create("ada@x.io", "Ada")
	if err != nil || a.ID != 1 {
		t.Fatalf("%+v %v", a, err)
	}
	b, _ := r.Create("bob@x.io", "Bob")
	if b.ID != 2 {
		t.Errorf("second id = %d", b.ID)
	}
	got, err := r.Get(1)
	if err != nil || got != a {
		t.Errorf("Get: %+v %v", got, err)
	}
	if !reflect.DeepEqual(r.List(), []User{a, b}) {
		t.Errorf("List: %v", r.List())
	}
	var empty Repo
	if l := empty.List(); len(l) != 0 {
		t.Errorf("empty list: %v", l)
	}
}

func TestErrors(t *testing.T) {
	var r Repo
	r.Create("ada@x.io", "Ada")
	if _, err := r.Create("ADA@x.io", "Other"); !errors.Is(err, ErrDuplicate) {
		t.Errorf("case-insensitive duplicate: %v", err)
	}
	if _, err := r.Create("", "Nobody"); !errors.Is(err, ErrInvalid) {
		t.Errorf("empty email: %v", err)
	}
	if _, err := r.Get(99); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get: %v", err)
	}
	if err := r.Update(99, "x"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Update: %v", err)
	}
	if err := r.Delete(99); !errors.Is(err, ErrNotFound) {
		t.Errorf("Delete: %v", err)
	}
}

func TestUpdateDeleteAndEmailReuse(t *testing.T) {
	var r Repo
	u, _ := r.Create("ada@x.io", "Ada")
	if err := r.Update(u.ID, "Ada L."); err != nil {
		t.Fatal(err)
	}
	if got, _ := r.Get(u.ID); got.Name != "Ada L." || got.Email != "ada@x.io" {
		t.Errorf("after update: %+v", got)
	}
	if err := r.Delete(u.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Get(u.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("deleted user still found: %v", err)
	}
	again, err := r.Create("ada@x.io", "Ada again")
	if err != nil || again.ID != 2 {
		t.Errorf("email should be reusable and ids never reused: %+v %v", again, err)
	}
}

func TestConcurrentCreates(t *testing.T) {
	var r Repo
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := r.Create(fmt.Sprintf("u%d@x.io", i), "n"); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	list := r.List()
	if len(list) != 50 {
		t.Fatalf("got %d users", len(list))
	}
	for i, u := range list {
		if u.ID != i+1 {
			t.Fatalf("ids not sequential/sorted: %v", list)
		}
	}
}
```

### Hints
- Keep a map by id, a map from lowercase email to id, a counter and a mutex
- Lazily allocate the maps so the zero value works

### Solution
```go
package main

import (
	"errors"
	"sort"
	"strings"
	"sync"
)

var (
	ErrNotFound  = errors.New("not found")
	ErrDuplicate = errors.New("duplicate email")
	ErrInvalid   = errors.New("invalid input")
)

type User struct {
	ID    int
	Email string
	Name  string
}

type Repo struct {
	mu     sync.Mutex
	next   int
	byID   map[int]User
	byMail map[string]int
}

func (r *Repo) init() {
	if r.byID == nil {
		r.byID = map[int]User{}
		r.byMail = map[string]int{}
	}
}

func (r *Repo) Create(email, name string) (User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.init()
	if email == "" {
		return User{}, ErrInvalid
	}
	key := strings.ToLower(email)
	if _, taken := r.byMail[key]; taken {
		return User{}, ErrDuplicate
	}
	r.next++
	u := User{ID: r.next, Email: email, Name: name}
	r.byID[u.ID] = u
	r.byMail[key] = u.ID
	return u, nil
}

func (r *Repo) Get(id int) (User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.byID[id]
	if !ok {
		return User{}, ErrNotFound
	}
	return u, nil
}

func (r *Repo) Update(id int, name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.byID[id]
	if !ok {
		return ErrNotFound
	}
	u.Name = name
	r.byID[id] = u
	return nil
}

func (r *Repo) Delete(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.byID[id]
	if !ok {
		return ErrNotFound
	}
	delete(r.byID, id)
	delete(r.byMail, strings.ToLower(u.Email))
	return nil
}

func (r *Repo) List() []User {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]User, 0, len(r.byID))
	for _, u := range r.byID {
		out = append(out, u)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func main() {}
```

### Explanation
Two maps give O(1) lookup by id and by normalised email; a single mutex protects both plus the counter. Returning domain errors keeps callers independent of the storage engine.
