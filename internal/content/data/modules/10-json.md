# JSON
id: json
number: 10
track: core
paths: beginner, pro
skill: stdlib
requires: stdlib
project: json-data-processing-tool
summary: Encode, decode, stream and validate JSON with struct tags.

## JSON Fundamentals
slug: json-fundamentals
minutes: 6
objectives: Recognise JSON's six value types; Map JSON to Go types; Know JSON's limits (no comments, numbers are doubles)
takeaways: JSON has object, array, string, number, boolean and null; Objects map to structs or map[string]any, arrays to slices; Numbers decoded into any become float64

### Concept
| JSON | Go (typed target) | Go (into `any`) |
| --- | --- | --- |
| object | struct or `map[string]T` | `map[string]any` |
| array | `[]T` | `[]any` |
| string | `string` | `string` |
| number | `int`, `float64`, … | `float64` |
| true/false | `bool` | `bool` |
| null | pointer/slice/map = nil | `nil` |

JSON is UTF-8 text with no comments and no trailing commas. Prefer decoding into **typed structs** — you get validation and autocomplete; use `any` only for truly dynamic shapes.

### Example
```go
package main

import (
	"encoding/json"
	"fmt"
)

func main() {
	var v any
	_ = json.Unmarshal([]byte(`{"name":"Ada","age":36,"tags":["a","b"],"vip":true,"x":null}`), &v)
	m := v.(map[string]any)
	fmt.Printf("%T %T %T %T %v\n", m["name"], m["age"], m["tags"], m["vip"], m["x"])
}
```

### Exercise
Decode `{"a":1,"b":2}` into `map[string]int` and print the sum of the values.
```text expect
3
```
```go solution
package main

import (
	"encoding/json"
	"fmt"
)

func main() {
	var m map[string]int
	// BEGIN
	_ = json.Unmarshal([]byte(`{"a":1,"b":2}`), &m)
	sum := 0
	for _, v := range m {
		sum += v
	}
	fmt.Println(sum)
	// END
}
```

### Check
Q: What Go type does a JSON number become when decoding into `any`?
T: short
A: float64
E: The default for numbers in an `any` target is `float64`.

Q: JSON allows `//` comments.
T: tf
A: false
E: Standard JSON has no comments (some tools tolerate them, `encoding/json` does not).

## Struct Tags
slug: struct-tags
minutes: 7
objectives: Rename fields with json tags; Use omitempty, omitzero and "-"; Embed structs and use string options
takeaways: `json:"name"` sets the key; omitempty drops empty values; `json:"-"` hides a field; Only exported fields participate

### Concept
```go norun
type User struct {
	ID       int       `json:"id"`
	Name     string    `json:"name"`
	Email    string    `json:"email,omitempty"`  // dropped when ""
	Password string    `json:"-"`                // never encoded
	Age      int       `json:"age,string"`       // "36" as a JSON string
	Created  time.Time `json:"created,omitzero"` // Go 1.24: drop zero time
}
```
`omitempty` omits false, 0, nil pointers/slices/maps and "" — but **not** zero structs; `omitzero` (Go 1.24+) handles that. Embedded structs are flattened. Without a tag the key is the Go field name (`Name`); decoding matches keys case-insensitively.

### Example
```go
package main

import (
	"encoding/json"
	"fmt"
)

type User struct {
	Name     string `json:"name"`
	Email    string `json:"email,omitempty"`
	Password string `json:"-"`
}

func main() {
	b, _ := json.Marshal(User{Name: "Ada", Password: "secret"})
	fmt.Println(string(b))
}
```

### Exercise
Define `Product` with fields `SKU` (`sku`), `Price` (`price`) and `Note` (`note`, omitted when empty) and print the JSON for `Product{"A-1", 9.99, ""}`.
```text expect
{"sku":"A-1","price":9.99}
```
```go solution
package main

import (
	"encoding/json"
	"fmt"
)

// BEGIN
type Product struct {
	SKU   string  `json:"sku"`
	Price float64 `json:"price"`
	Note  string  `json:"note,omitempty"`
}

// END

func main() {
	b, _ := json.Marshal(Product{"A-1", 9.99, ""})
	fmt.Println(string(b))
}
```

### Check
Q: What does the tag `json:"-"` do?
T: mcq
- [ ] Renames the field to "-"
- [x] Excludes the field from encoding and decoding
- [ ] Makes it optional
- [ ] Encodes it as null
E: A tag of just `-` hides the field entirely.

Q: Which fields does the JSON encoder see?
T: mcq
- [ ] All fields
- [x] Exported fields only
- [ ] Tagged fields only
- [ ] Pointer fields only
E: Unexported fields are skipped regardless of tags.

## Marshal
slug: marshal
minutes: 6
objectives: Convert Go values to JSON with json.Marshal; Pretty-print with MarshalIndent; Implement custom encoding with MarshalJSON
takeaways: json.Marshal returns []byte and an error; MarshalIndent adds indentation for humans; Types can customise output by implementing json.Marshaler

### Concept
```go norun
b, err := json.Marshal(v)
pretty, _ := json.MarshalIndent(v, "", "  ")
```
Marshal fails for unsupported values (channels, funcs, cyclic data). **Map keys are sorted** so output is deterministic. `[]byte` becomes base64; `time.Time` uses RFC 3339; `nil` slices encode as `null` while empty slices encode as `[]` — matter for API clients!

Implement `MarshalJSON() ([]byte, error)` to control a type's encoding.

### Example
```go
package main

import (
	"encoding/json"
	"fmt"
)

type Temp float64

func (t Temp) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%.1f°C"`, float64(t))), nil
}

func main() {
	b, _ := json.Marshal(map[string]any{"b": Temp(21.55), "a": []int(nil), "c": []int{}})
	fmt.Println(string(b))
}
```

### Exercise
Marshal `[]string{"go", "gopher"}` with `json.MarshalIndent` using a two-space indent and print it.
```text expect
[
  "go",
  "gopher"
]
```
```go solution
package main

import (
	"encoding/json"
	"fmt"
)

func main() {
	// BEGIN
	b, _ := json.MarshalIndent([]string{"go", "gopher"}, "", "  ")
	fmt.Println(string(b))
	// END
}
```

### Check
Q: What does a nil slice encode to?
T: mcq
- [ ] []
- [x] null
- [ ] ""
- [ ] 0
E: nil slices become `null`; use an empty (non-nil) slice to get `[]`.

Q: What does this print?
T: output
```go
b, _ := json.Marshal(map[string]int{"b": 2, "a": 1})
fmt.Println(string(b))
```
- [ ] {"b":2,"a":1}
- [x] {"a":1,"b":2}
- [ ] {a:1,b:2}
- [ ] [1,2]
E: `encoding/json` sorts map keys, making output deterministic.

## Unmarshal
slug: unmarshal
minutes: 7
objectives: Decode JSON into structs; Handle type errors and missing fields; Distinguish absent from zero with pointers
takeaways: json.Unmarshal needs a pointer target; Unknown keys are ignored by default; Use pointer fields to tell "missing" from "zero"

### Concept
```go norun
var u User
if err := json.Unmarshal(data, &u); err != nil {
	return fmt.Errorf("decode user: %w", err)
}
```
Errors include `*json.SyntaxError` (with `Offset`) and `*json.UnmarshalTypeError` (field and value). Missing keys leave fields at their zero value, so `{"active": false}` and `{}` look identical for a `bool` — declare `Active *bool` when the distinction matters. Unknown keys are silently ignored unless you use `Decoder.DisallowUnknownFields()`.

### Example
```go
package main

import (
	"encoding/json"
	"errors"
	"fmt"
)

type Settings struct {
	Theme string `json:"theme"`
	Beta  *bool  `json:"beta"`
}

func main() {
	var s Settings
	_ = json.Unmarshal([]byte(`{"theme":"dark"}`), &s)
	fmt.Println(s.Theme, s.Beta == nil)
	err := json.Unmarshal([]byte(`{"theme": 5}`), &s)
	var te *json.UnmarshalTypeError
	fmt.Println(errors.As(err, &te), te.Field)
}
```

### Exercise
Decode `{"name":"Ada","langs":["go","c"]}` into a struct and print `Ada knows 2 languages`.
```text expect
Ada knows 2 languages
```
```go solution
package main

import (
	"encoding/json"
	"fmt"
)

// BEGIN
type Dev struct {
	Name  string   `json:"name"`
	Langs []string `json:"langs"`
}

// END

func main() {
	var d Dev
	_ = json.Unmarshal([]byte(`{"name":"Ada","langs":["go","c"]}`), &d)
	fmt.Printf("%s knows %d languages\n", d.Name, len(d.Langs))
}
```

### Check
Q: Why use `*bool` instead of `bool` for an optional flag?
T: mcq
- [ ] Pointers are faster
- [x] nil can mean "not provided", distinct from false
- [ ] JSON requires pointers
- [ ] To avoid copying
E: A plain `bool` can't tell a missing key from an explicit false.

Q: `json.Unmarshal(data, v)` with `v` a struct value (not a pointer) fills the struct.
T: tf
A: false
E: You must pass a pointer; otherwise it returns an error and changes nothing.

## Encoder
slug: encoder
minutes: 6
objectives: Stream JSON to an io.Writer; Configure indentation and HTML escaping; Write JSON lines
takeaways: json.NewEncoder(w).Encode(v) writes one value plus a newline; SetIndent and SetEscapeHTML tune output; Encoders are ideal for HTTP responses and JSON-lines logs

### Concept
```go norun
enc := json.NewEncoder(w)
enc.SetIndent("", "  ")
enc.SetEscapeHTML(false) // keep <, > and & readable
if err := enc.Encode(v); err != nil { ... }
```
`Encode` appends a newline, so calling it repeatedly produces **JSON Lines** (one object per line) — perfect for log files and streaming APIs. In HTTP handlers, encode straight into the `http.ResponseWriter` after setting `Content-Type: application/json`.

### Example
```go
package main

import (
	"encoding/json"
	"os"
)

func main() {
	enc := json.NewEncoder(os.Stdout)
	for i := 1; i <= 2; i++ {
		_ = enc.Encode(map[string]int{"n": i})
	}
	enc.SetEscapeHTML(false)
	_ = enc.Encode("a<b>&c")
}
```

### Exercise
Write `writeLines(w io.Writer, items []string) error` that encodes each item as a JSON object `{"item":"..."}` on its own line.
```text expect
{"item":"a"}
{"item":"b"}
```
```go solution
package main

import (
	"encoding/json"
	"io"
	"os"
)

// BEGIN
func writeLines(w io.Writer, items []string) error {
	enc := json.NewEncoder(w)
	for _, it := range items {
		if err := enc.Encode(map[string]string{"item": it}); err != nil {
			return err
		}
	}
	return nil
}

// END

func main() {
	_ = writeLines(os.Stdout, []string{"a", "b"})
}
```

### Check
Q: What does `Encoder.Encode` append after each value?
T: mcq
- [ ] A comma
- [x] A newline
- [ ] Nothing
- [ ] A null byte
E: The trailing newline makes repeated calls produce JSON Lines.

Q: Which method keeps `<` and `&` from being escaped as < and &?
T: short
A: SetEscapeHTML
E: `SetEscapeHTML(false)` disables HTML-safe escaping.

## Decoder
slug: decoder
minutes: 8
objectives: Stream-decode values from an io.Reader; Reject unknown fields and trailing data; Process large arrays token by token
takeaways: json.NewDecoder(r).Decode reads one value at a time; DisallowUnknownFields makes typos in payloads an error; Decoder.Token lets you stream a huge array without loading it

### Concept
```go norun
dec := json.NewDecoder(r)
dec.DisallowUnknownFields()
var cfg Config
if err := dec.Decode(&cfg); err != nil { ... }
```
A decoder over a stream of concatenated values (JSON Lines) decodes them one by one until `io.EOF`. To stream a big top-level array:

```go norun
dec.Token()              // consume '['
for dec.More() {
	var item Item
	dec.Decode(&item)    // one element at a time
}
dec.Token()              // consume ']'
```
For HTTP bodies wrap the reader with `http.MaxBytesReader` first.

### Example
```go
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

func main() {
	dec := json.NewDecoder(strings.NewReader(`{"n":1} {"n":2}
{"n":3}`))
	for {
		var v struct{ N int }
		if err := dec.Decode(&v); err == io.EOF {
			break
		} else if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Print(v.N, " ")
	}
	fmt.Println()
}
```

### Exercise
Write `sumStream(r io.Reader) (int, error)` that decodes JSON-lines objects `{"n":<int>}` until EOF and returns the total.
```text expect
6 <nil>
```
```go solution
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// BEGIN
func sumStream(r io.Reader) (int, error) {
	dec := json.NewDecoder(r)
	total := 0
	for {
		var v struct {
			N int `json:"n"`
		}
		err := dec.Decode(&v)
		if err == io.EOF {
			return total, nil
		}
		if err != nil {
			return total, err
		}
		total += v.N
	}
}

// END

func main() {
	fmt.Println(sumStream(strings.NewReader("{\"n\":1}\n{\"n\":2}\n{\"n\":3}\n")))
}
```

### Check
Q: Which method makes decoding fail on keys that don't exist in the target struct?
T: short
A: DisallowUnknownFields
E: It catches typos such as `"prot"` instead of `"port"`.

Q: How do you detect the end of a stream of JSON values?
T: mcq
- [ ] Decode returns nil forever
- [x] Decode returns io.EOF
- [ ] It panics
- [ ] You must count values first
E: A clean end of input is reported as `io.EOF`.

## Nested JSON
slug: nested-json
minutes: 7
objectives: Model nested objects and arrays with structs; Use json.RawMessage for deferred decoding; Decode polymorphic payloads
takeaways: Nested objects are nested structs (or maps); json.RawMessage delays decoding of part of a document; A discriminator field plus RawMessage decodes polymorphic data safely

### Concept
```go norun
type Order struct {
	ID    int `json:"id"`
	Buyer struct {
		Name string `json:"name"`
	} `json:"buyer"`
	Items []Item `json:"items"`
}
```
When part of the document's shape depends on another field, keep it as `json.RawMessage` and decode it once you know the type:

```go norun
type Event struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}
```

### Example
```go
package main

import (
	"encoding/json"
	"fmt"
)

type Event struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func main() {
	in := `[{"type":"click","data":{"x":1,"y":2}},{"type":"key","data":{"key":"a"}}]`
	var evs []Event
	_ = json.Unmarshal([]byte(in), &evs)
	for _, e := range evs {
		switch e.Type {
		case "click":
			var d struct{ X, Y int }
			_ = json.Unmarshal(e.Data, &d)
			fmt.Println("click", d.X, d.Y)
		case "key":
			var d struct{ Key string }
			_ = json.Unmarshal(e.Data, &d)
			fmt.Println("key", d.Key)
		}
	}
}
```

### Exercise
Decode `{"user":{"name":"Ada","address":{"city":"London"}}}` into nested structs and print the city.
```text expect
London
```
```go solution
package main

import (
	"encoding/json"
	"fmt"
)

// BEGIN
type Payload struct {
	User struct {
		Name    string `json:"name"`
		Address struct {
			City string `json:"city"`
		} `json:"address"`
	} `json:"user"`
}

// END

func main() {
	var p Payload
	_ = json.Unmarshal([]byte(`{"user":{"name":"Ada","address":{"city":"London"}}}`), &p)
	fmt.Println(p.User.Address.City)
}
```

### Check
Q: What is `json.RawMessage` used for?
T: mcq
- [ ] Encoding binary data
- [x] Keeping part of a document undecoded until its type is known
- [ ] Speeding up marshalling of maps
- [ ] Validating UTF-8
E: `RawMessage` stores the raw JSON bytes so you can decode them later, e.g. after reading a discriminator.

Q: Anonymous nested structs are always better than named types.
T: tf
A: false
E: Named types are reusable and easier to test; anonymous structs suit one-off payloads.

## JSON Validation
slug: json-validation
minutes: 7
objectives: Check syntax with json.Valid; Validate business rules after decoding; Report errors with locations
takeaways: json.Valid checks syntax only; Decoding validates types but not business rules — validate the struct afterwards; SyntaxError.Offset pinpoints where parsing failed

### Concept
Three layers of validation:

1. **Syntax** — `json.Valid(data)` or the error from decoding (`*json.SyntaxError` has `Offset`).
2. **Shape** — types and unknown fields (`UnmarshalTypeError`, `DisallowUnknownFields`).
3. **Rules** — ranges, required fields, formats. Write `func (c Config) Validate() error`, returning all problems (`errors.Join`) so users fix them in one go.

Required fields aren't a JSON concept: use pointers (`*string`) or check zero values. Limit input size before decoding untrusted JSON.

### Example
```go
package main

import (
	"encoding/json"
	"errors"
	"fmt"
)

func main() {
	fmt.Println(json.Valid([]byte(`{"a":1}`)), json.Valid([]byte(`{"a":}`)))
	var v map[string]any
	err := json.Unmarshal([]byte(`{"a":1,}`), &v)
	var se *json.SyntaxError
	if errors.As(err, &se) {
		fmt.Println("syntax error at byte", se.Offset)
	}
}
```

### Exercise
Write `validate(name string, port int) error` returning `errors.Join` of every problem: `name is required` (empty) and `port out of range` (not 1–65535). Return nil when valid.
```text expect
name is required
port out of range
<nil>
```
```go solution
package main

import (
	"errors"
	"fmt"
)

// BEGIN
func validate(name string, port int) error {
	var errs []error
	if name == "" {
		errs = append(errs, errors.New("name is required"))
	}
	if port < 1 || port > 65535 {
		errs = append(errs, errors.New("port out of range"))
	}
	return errors.Join(errs...)
}

// END

func main() {
	fmt.Println(validate("", 0))
	fmt.Println(validate("api", 80))
}
```

### Check
Q: What does `json.Valid` check?
T: mcq
- [ ] That the data matches a struct
- [x] That the bytes are syntactically valid JSON
- [ ] Business rules
- [ ] Field types
E: It validates syntax only.

Q: How do you make a field "required" when decoding JSON in Go?
T: mcq
- [ ] A `json:"name,required"` tag
- [x] Use a pointer type or validate after decoding
- [ ] Encoding/json does it automatically
- [ ] Use `omitempty`
E: There is no `required` option; check for nil or zero values yourself.

## Working with APIs
slug: working-with-apis
minutes: 8
objectives: Call a JSON API with http.Client; Decode the response safely; Handle non-2xx statuses and timeouts
takeaways: Always set a timeout on http.Client; Check the status code before decoding; Close resp.Body and limit how much you read

### Concept
```go norun
client := &http.Client{Timeout: 5 * time.Second}
resp, err := client.Get(url)
if err != nil { return err }
defer resp.Body.Close()
if resp.StatusCode != http.StatusOK {
	return fmt.Errorf("unexpected status %s", resp.Status)
}
var out Response
err = json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&out)
```
The default `http.Client` has **no timeout** — never use it in production. Non-2xx responses are not Go errors; check `StatusCode`. The lesson exercise uses a local `httptest` server so it needs no network.

### Example
```go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
)

func main() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"greeting": "hello"})
	}))
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	var out map[string]string
	json.NewDecoder(resp.Body).Decode(&out)
	fmt.Println(resp.StatusCode, out["greeting"])
}
```

### Exercise
Write `fetchName(url string) (string, error)` that GETs `url`, returns an error containing the status if it isn't 200, and otherwise decodes `{"name": "..."}` and returns the name.
```text expect
Ada <nil>
 unexpected status: 404 Not Found
```
```go solution
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"
)

// BEGIN
func fetchName(url string) (string, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %s", resp.Status)
	}
	var out struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return out.Name, nil
}

// END

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"name":"Ada"}`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	fmt.Println(fetchName(srv.URL + "/ok"))
	fmt.Println(fetchName(srv.URL + "/missing"))
}
```

### Check
Q: Does `client.Get` return an error for a 404 response?
T: mcq
- [ ] Yes
- [x] No — a 404 is a normal response; check resp.StatusCode
- [ ] Only if the body is empty
- [ ] Only with a timeout
E: Errors mean the request couldn't complete; HTTP error statuses are part of a successful exchange.

Q: The zero-value `http.Client` has a request timeout.
T: tf
A: false
E: There is no timeout by default; always set `Timeout` (or use contexts).
