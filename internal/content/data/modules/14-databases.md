# Databases
id: databases
number: 14
track: advanced
paths: beginner, pro
skill: databases
requires: rest-api
project: go-postgres-backend
summary: SQL, database/sql, transactions, pools, migrations and the repository pattern.

## SQL Fundamentals
slug: sql-fundamentals
minutes: 9
challenges: sql-query-builder
objectives: Read and write basic SELECT, INSERT, UPDATE and DELETE statements; Explain why parameters (not string concatenation) carry values; Build parameterised queries safely
takeaways: SQL statements read and change relational tables; Values must travel as parameters ($1, $2 or ?) — never concatenated into the SQL text; Identifiers (table/column names) can't be parameters, so allow-list them

### Concept
```sql
CREATE TABLE users (
  id    BIGSERIAL PRIMARY KEY,
  email TEXT NOT NULL UNIQUE,
  name  TEXT NOT NULL,
  age   INT
);
INSERT INTO users (email, name) VALUES ($1, $2) RETURNING id;
SELECT id, name FROM users WHERE age >= $1 ORDER BY name LIMIT 20;
UPDATE users SET name = $1 WHERE id = $2;
DELETE FROM users WHERE id = $1;
```
**Never** build SQL by concatenating user input (`"... WHERE name = '" + name + "'"`): that is **SQL injection**. Send values as *parameters*; the driver transmits them separately from the SQL text so they can never be interpreted as SQL.

Placeholders: PostgreSQL uses `$1, $2…`, SQLite and MySQL use `?`. Table and column names can't be parameters — if they come from input, check them against an allow-list.

### Example
```go
package main

import (
	"fmt"
	"sort"
	"strings"
)

func where(filters map[string]any) (string, []any) {
	cols := make([]string, 0, len(filters))
	for c := range filters {
		cols = append(cols, c)
	}
	sort.Strings(cols)
	var parts []string
	var args []any
	for i, c := range cols {
		parts = append(parts, fmt.Sprintf("%s = $%d", c, i+1))
		args = append(args, filters[c])
	}
	if len(parts) == 0 {
		return "", nil
	}
	return " WHERE " + strings.Join(parts, " AND "), args
}

func main() {
	clause, args := where(map[string]any{"name": "Ada", "age": 36})
	fmt.Println("SELECT * FROM users"+clause, args)
}
```

### Exercise
Write `insertSQL(table string, cols []string) string` returning `INSERT INTO <table> (<cols>) VALUES ($1, $2, …)` with one numbered placeholder per column.
```text expect
INSERT INTO users (email, name) VALUES ($1, $2)
```
```go solution
package main

import (
	"fmt"
	"strings"
)

// BEGIN
func insertSQL(table string, cols []string) string {
	ph := make([]string, len(cols))
	for i := range cols {
		ph[i] = fmt.Sprintf("$%d", i+1)
	}
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(cols, ", "), strings.Join(ph, ", "))
}

// END

func main() {
	fmt.Println(insertSQL("users", []string{"email", "name"}))
}
```

### Check
Q: How should user-supplied values reach a SQL statement?
T: mcq
- [ ] Concatenated into the string with quotes
- [x] As bound parameters ($1, $2 or ?)
- [ ] Through fmt.Sprintf
- [ ] Escaped by hand
E: Parameters are sent separately from the SQL text, which prevents injection.

Q: Can a table name be passed as a query parameter?
T: tf
A: false
E: Parameters stand for values only; identifiers must be validated against an allow-list.

## PostgreSQL
slug: postgresql
status: planned

## SQLite
slug: sqlite
status: planned

## database/sql
slug: database-sql
minutes: 9
objectives: Open a database and run queries with database/sql; Iterate rows correctly (Next, Scan, Err, Close); Handle sql.ErrNoRows and NULLs
takeaways: sql.Open returns a pool (*sql.DB) — create one per database and share it; QueryContext + rows loop: defer rows.Close(), for rows.Next(), check rows.Err(); QueryRowContext(...).Scan(...) returns sql.ErrNoRows when nothing matches

### Concept
```go norun
import (
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" driver
)

db, err := sql.Open("pgx", dsn)            // doesn't connect yet
err = db.PingContext(ctx)                  // verifies connectivity

rows, err := db.QueryContext(ctx, `SELECT id, name FROM users WHERE age >= $1`, 18)
if err != nil { return err }
defer rows.Close()
for rows.Next() {
	var u User
	if err := rows.Scan(&u.ID, &u.Name); err != nil { return err }
	users = append(users, u)
}
return rows.Err() // errors that ended iteration early
```
`db.QueryRowContext(...).Scan(&x)` returns `sql.ErrNoRows` when no row matches — map it to your own `ErrNotFound`. Use `sql.NullString`/`*string` for nullable columns. Always use the `…Context` variants so cancellation and timeouts work.

`sql.DB` is a **connection pool**, safe for concurrent use; open it once at startup.

### Example
```go
package main

import (
	"fmt"
	"strings"
)

// rowsLike is the subset of *sql.Rows we need — a real *sql.Rows satisfies it.
type rowsLike interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
	Close() error
}

type fake struct {
	data []string
	i    int
}

func (f *fake) Next() bool { f.i++; return f.i <= len(f.data) }
func (f *fake) Scan(dest ...any) error {
	*(dest[0].(*string)) = f.data[f.i-1]
	return nil
}
func (f *fake) Err() error   { return nil }
func (f *fake) Close() error { return nil }

func names(r rowsLike) ([]string, error) {
	defer r.Close()
	var out []string
	for r.Next() {
		var n string
		if err := r.Scan(&n); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, r.Err()
}

func main() {
	got, err := names(&fake{data: []string{"ada", "grace"}})
	fmt.Println(strings.Join(got, ","), err)
}
```

### Exercise
Write `sum(r rowsLike) (int, error)` that scans an `int` from each row, adds them up, closes the rows and returns `rows.Err()`.
```text expect
60 <nil>
```
```go solution
package main

import "fmt"

type rowsLike interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
	Close() error
}

type fake struct {
	data []int
	i    int
}

func (f *fake) Next() bool { f.i++; return f.i <= len(f.data) }
func (f *fake) Scan(dest ...any) error {
	*(dest[0].(*int)) = f.data[f.i-1]
	return nil
}
func (f *fake) Err() error   { return nil }
func (f *fake) Close() error { return nil }

// BEGIN
func sum(r rowsLike) (int, error) {
	defer r.Close()
	total := 0
	for r.Next() {
		var n int
		if err := r.Scan(&n); err != nil {
			return 0, err
		}
		total += n
	}
	return total, r.Err()
}

// END

func main() {
	fmt.Println(sum(&fake{data: []int{10, 20, 30}}))
}
```

### Check
Q: What must you check after the `for rows.Next()` loop ends?
T: mcq
- [ ] Nothing
- [x] rows.Err()
- [ ] db.Ping()
- [ ] rows.Next() again
E: `Next` returns false on both end-of-data and errors; `Err` tells you which.

Q: What error does `QueryRow(...).Scan` return when no row matches?
T: short
A: sql.ErrNoRows
E: Translate it to your own domain error (e.g. `ErrNotFound`) at the repository boundary.

## Connections
slug: connections
status: planned

## Queries
slug: queries
status: planned

## Prepared Statements
slug: prepared-statements
status: planned

## Transactions
slug: transactions
minutes: 9
challenges: with-tx
objectives: Group statements atomically with BEGIN/COMMIT/ROLLBACK; Write a helper that commits on success and rolls back on error or panic; Avoid holding transactions open across slow work
takeaways: A transaction makes several statements succeed or fail together; Always defer a rollback (a no-op after commit) and commit explicitly; Keep transactions short — they hold locks and a pooled connection

### Concept
```go norun
tx, err := db.BeginTx(ctx, nil)
if err != nil { return err }
defer tx.Rollback() // safe after Commit: returns sql.ErrTxDone, ignore it

if _, err := tx.ExecContext(ctx, `UPDATE accounts SET balance = balance - $1 WHERE id = $2`, amt, from); err != nil {
	return err
}
if _, err := tx.ExecContext(ctx, `UPDATE accounts SET balance = balance + $1 WHERE id = $2`, amt, to); err != nil {
	return err
}
return tx.Commit()
```
Wrap this in a helper so every caller gets correct commit/rollback (including on panic). Do only database work inside the transaction — no HTTP calls or sleeps. Concurrent transactions can conflict; understand your isolation level and retry on serialization failures (`40001`).

### Example
```go
package main

import "fmt"

type Tx interface {
	Commit() error
	Rollback() error
}

type logTx struct{ log *[]string }

func (t logTx) Commit() error   { *t.log = append(*t.log, "commit"); return nil }
func (t logTx) Rollback() error { *t.log = append(*t.log, "rollback"); return nil }

func withTx(begin func() (Tx, error), fn func(Tx) error) error {
	tx, err := begin()
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func main() {
	var log []string
	begin := func() (Tx, error) { return logTx{&log}, nil }
	_ = withTx(begin, func(Tx) error { return nil })
	_ = withTx(begin, func(Tx) error { return fmt.Errorf("boom") })
	fmt.Println(log)
}
```

### Exercise
Extend `withTx` so a **panic** inside `fn` also rolls back (then re-panics). Print the log after recovering.
```text expect
[rollback] recovered: oops
```
```go solution
package main

import "fmt"

type Tx interface {
	Commit() error
	Rollback() error
}

type logTx struct{ log *[]string }

func (t logTx) Commit() error   { *t.log = append(*t.log, "commit"); return nil }
func (t logTx) Rollback() error { *t.log = append(*t.log, "rollback"); return nil }

// BEGIN
func withTx(begin func() (Tx, error), fn func(Tx) error) (err error) {
	tx, err := begin()
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

// END

func main() {
	var log []string
	begin := func() (Tx, error) { return logTx{&log}, nil }
	defer func() {
		r := recover()
		fmt.Println(log, "recovered:", r)
	}()
	_ = withTx(begin, func(Tx) error { panic("oops") })
}
```

### Check
Q: Why `defer tx.Rollback()` right after `BeginTx`?
T: mcq
- [ ] To speed up commits
- [x] It guarantees rollback on every early-return path; after Commit it's a harmless no-op
- [ ] Rollback is required after Commit
- [ ] It opens a new connection
E: Forgetting to roll back leaks the connection and holds locks.

Q: It is fine to make an HTTP call to another service while a database transaction is open.
T: tf
A: false
E: Slow external calls keep locks and a pool connection busy; do them outside the transaction.

## Connection Pools
slug: connection-pools
status: planned

## Indexes
slug: indexes
status: planned

## Migrations
slug: migrations
minutes: 8
objectives: Version the schema with ordered migration files; Track applied versions in the database; Write safe, reversible changes
takeaways: Migrations are ordered, immutable scripts that evolve the schema; A schema_migrations table records what has been applied; Prefer additive, backward-compatible changes (expand → migrate → contract)

### Concept
Never change production schemas by hand. Keep numbered files under version control:

```
migrations/
  0001_create_users.up.sql     0001_create_users.down.sql
  0002_add_users_age.up.sql    0002_add_users_age.down.sql
```
A runner (`golang-migrate`, `goose`, `atlas` or your own) creates a `schema_migrations(version)` table, applies pending files **in order inside a transaction where possible** and records each version.

**Rules:** never edit an applied migration; make deploys compatible with the previous app version (add a nullable column, deploy code, backfill, then add constraints — *expand/contract*); take care with long locks on big tables (`CREATE INDEX CONCURRENTLY`).

### Example
```go
package main

import (
	"fmt"
	"sort"
)

type Migration struct {
	Version int
	Name    string
}

func pending(applied map[int]bool, all []Migration) []Migration {
	sort.Slice(all, func(i, j int) bool { return all[i].Version < all[j].Version })
	var out []Migration
	for _, m := range all {
		if !applied[m.Version] {
			out = append(out, m)
		}
	}
	return out
}

func main() {
	all := []Migration{{3, "add_index"}, {1, "create_users"}, {2, "add_age"}}
	for _, m := range pending(map[int]bool{1: true}, all) {
		fmt.Println(m.Version, m.Name)
	}
}
```

### Exercise
Write `latest(applied map[int]bool) int` returning the highest applied version (0 when none).
```text expect
3 0
```
```go solution
package main

import "fmt"

// BEGIN
func latest(applied map[int]bool) int {
	max := 0
	for v, ok := range applied {
		if ok && v > max {
			max = v
		}
	}
	return max
}

// END

func main() {
	fmt.Println(latest(map[int]bool{1: true, 3: true, 2: true}), latest(nil))
}
```

### Check
Q: May you edit a migration file that has already been applied in production?
T: mcq
- [ ] Yes, if it is a small change
- [x] No — add a new migration instead
- [ ] Only the down file
- [ ] Only on Fridays
E: Applied migrations are history; changing them makes environments diverge.

Q: What does the expand/contract pattern do?
T: mcq
- [ ] Doubles database size
- [x] Makes schema changes backward compatible across deploys
- [ ] Compresses tables
- [ ] Reorders migrations
E: Add new structures first, migrate data and code, then remove the old ones.

## Repository Pattern
slug: repository-pattern
minutes: 8
challenges: memory-repository
objectives: Hide persistence behind an interface owned by the domain; Return domain errors (not driver errors); Provide an in-memory implementation for tests
takeaways: A repository exposes domain operations (Create, Get, List) and hides SQL; Translate driver errors (sql.ErrNoRows) into domain errors (ErrNotFound); An in-memory fake makes services testable without a database

### Concept
```go norun
type UserRepository interface {
	Create(ctx context.Context, u *User) error
	Get(ctx context.Context, id int64) (*User, error)
	List(ctx context.Context) ([]User, error)
	Delete(ctx context.Context, id int64) error
}
```
Implementations: `postgresRepo` (uses `*sql.DB`), `memoryRepo` (a map + mutex). The **service** depends only on the interface. The repository returns `ErrNotFound`, never `sql.ErrNoRows`, so upper layers don't know which database you use. Run the same conformance tests against every implementation.

### Example
```go
package main

import (
	"errors"
	"fmt"
	"sort"
	"sync"
)

var ErrNotFound = errors.New("not found")

type User struct {
	ID   int
	Name string
}

type MemRepo struct {
	mu    sync.Mutex
	next  int
	users map[int]User
}

func (r *MemRepo) Create(name string) User {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.users == nil {
		r.users = map[int]User{}
	}
	r.next++
	u := User{r.next, name}
	r.users[u.ID] = u
	return u
}

func (r *MemRepo) Get(id int) (User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if u, ok := r.users[id]; ok {
		return u, nil
	}
	return User{}, ErrNotFound
}

func (r *MemRepo) List() []User {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]User, 0, len(r.users))
	for _, u := range r.users {
		out = append(out, u)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func main() {
	var r MemRepo
	r.Create("ada")
	r.Create("grace")
	_, err := r.Get(9)
	fmt.Println(r.List(), errors.Is(err, ErrNotFound))
}
```

### Exercise
Add `Delete(id int) error` to `MemRepo`, returning `ErrNotFound` for unknown ids. Delete id 1, then again; print both results.
```text expect
<nil> not found
```
```go solution
package main

import (
	"errors"
	"fmt"
	"sync"
)

var ErrNotFound = errors.New("not found")

type User struct {
	ID   int
	Name string
}

type MemRepo struct {
	mu    sync.Mutex
	next  int
	users map[int]User
}

func (r *MemRepo) Create(name string) User {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.users == nil {
		r.users = map[int]User{}
	}
	r.next++
	u := User{r.next, name}
	r.users[u.ID] = u
	return u
}

// BEGIN
func (r *MemRepo) Delete(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.users[id]; !ok {
		return ErrNotFound
	}
	delete(r.users, id)
	return nil
}

// END

func main() {
	var r MemRepo
	r.Create("ada")
	fmt.Println(r.Delete(1), r.Delete(1))
}
```

### Check
Q: What should a repository return when a row doesn't exist?
T: mcq
- [ ] sql.ErrNoRows
- [x] A domain error such as ErrNotFound
- [ ] nil and a zero value
- [ ] A panic
E: Keep driver details out of upper layers.

Q: Why write an in-memory repository?
T: mcq
- [ ] It's required for production
- [x] Service tests run fast without a database
- [ ] It replaces migrations
- [ ] It is more durable
E: Fakes make unit tests fast and deterministic.

## sqlc
slug: sqlc
status: planned

## ORM Concepts
slug: orm-concepts
status: planned
