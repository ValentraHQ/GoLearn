package runner

// Trusted grading for challenges.
//
// `go test -json` cannot be trusted on its own: the learner's code runs in the
// same process as the tests and can print "--- PASS: TestX" and call
// os.Exit(0). Instead, the sandbox entrypoint (a separate, trusted process)
//
//  1. rewrites the platform-owned test file so every top-level test records
//     its own outcome when it finishes,
//  2. adds a TestMain that, after m.Run() returns, writes a verdict file
//     authenticated with a per-run random key compiled into that test binary,
//  3. reads the verdict back and verifies the MAC.
//
// The server accepts a challenge only from a verified verdict; test events
// parsed from stdout are used purely to show failure messages.
//
// Limits (also in docs/sandbox.md): learner code runs in the same process as
// the recorder, so a learner who reverse-engineers their own compiled test
// binary within the run can still extract the per-run key and forge a verdict.
// Printing lines, exiting early, skipping tests, or writing a guessed verdict
// no longer works. Closing the remaining gap needs out-of-process (black-box)
// tests.

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"
)

// Verdict is the authenticated outcome of a graded `go test` run.
type Verdict struct {
	// Valid is true only when the verdict file's MAC verified.
	Valid bool `json:"valid"`
	// ExitCode is what m.Run() returned (0 means every selected test passed).
	ExitCode int `json:"exitCode"`
	// Tests maps top-level test names that ran to whether they passed.
	Tests map[string]bool `json:"tests,omitempty"`
}

// Passed reports whether the named test ran and passed under a valid verdict.
func (v *Verdict) Passed(name string) bool {
	return v != nil && v.Valid && v.Tests[name]
}

// AllPassed reports a valid, clean verdict in which every expected test ran
// and passed.
func (v *Verdict) AllPassed(expected []string) bool {
	if v == nil || !v.Valid || v.ExitCode != 0 || len(expected) == 0 {
		return false
	}
	for _, n := range expected {
		if !v.Tests[n] {
			return false
		}
	}
	return true
}

// Grader holds the per-run secrets shared between the generated test files and
// the entrypoint.
type Grader struct {
	Key         string // hex, compiled into the test binary
	VerdictPath string // where TestMain writes the verdict
	suffix      string // randomises the recorder's symbol name
}

// NewGrader creates fresh random secrets under dir.
func NewGrader(dir string) (*Grader, error) {
	b := make([]byte, 40)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return &Grader{
		Key:         hex.EncodeToString(b[:32]),
		VerdictPath: dir + "/.v" + hex.EncodeToString(b[32:]),
		suffix:      hex.EncodeToString(b[32:36]),
	}, nil
}

func isTestFunc(fd *ast.FuncDecl) bool {
	n := fd.Name.Name
	if fd.Recv != nil || !strings.HasPrefix(n, "Test") {
		return false
	}
	if len(n) > 4 && n[4] >= 'a' && n[4] <= 'z' { // Testfoo is not a test
		return false
	}
	ps := fd.Type.Params
	if ps == nil || len(ps.List) != 1 {
		return false
	}
	star, ok := ps.List[0].Type.(*ast.StarExpr)
	if !ok {
		return false
	}
	sel, ok := star.X.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "T" {
		return false
	}
	id, ok := sel.X.(*ast.Ident)
	return ok && id.Name == "testing"
}

// Instrument rewrites a platform-owned test file and returns it together with
// the generated grader file. It refuses files that define their own TestMain,
// which would bypass the recorder.
func (g *Grader) Instrument(testSrc string) (instrumented, graderFile string, err error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "main_test.go", testSrc, parser.ParseComments)
	if err != nil {
		return "", "", fmt.Errorf("parse tests: %w", err)
	}
	if f.Name.Name != "main" {
		return "", "", errors.New("tests must be in package main")
	}
	rec := "gradeRecord_" + g.suffix
	count := 0
	for _, d := range f.Decls {
		fd, ok := d.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if fd.Name.Name == "TestMain" && fd.Recv == nil {
			return "", "", errors.New("tests must not define TestMain")
		}
		if !isTestFunc(fd) || fd.Body == nil {
			continue
		}
		field := fd.Type.Params.List[0]
		if len(field.Names) == 0 || field.Names[0].Name == "_" {
			field.Names = []*ast.Ident{ast.NewIdent("graderT")}
		}
		tname := field.Names[0].Name
		stmt, err := parseStmt(fmt.Sprintf("%s.Cleanup(func() { %s(%s.Name(), !%s.Failed() && !%s.Skipped()) })", tname, rec, tname, tname, tname))
		if err != nil {
			return "", "", err
		}
		fd.Body.List = append([]ast.Stmt{stmt}, fd.Body.List...)
		count++
	}
	if count == 0 {
		return "", "", errors.New("no tests found")
	}
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, f); err != nil {
		return "", "", err
	}
	return buf.String(), g.graderSource(rec), nil
}

func parseStmt(src string) (ast.Stmt, error) {
	e, err := parser.ParseExpr(src)
	if err != nil {
		return nil, err
	}
	return &ast.ExprStmt{X: e}, nil
}

func (g *Grader) graderSource(rec string) string {
	return `package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
)

var (
	gradeMu  sync.Mutex
	gradeRes = map[string]bool{}
)

// gradeWrite stores the authenticated verdict; the caller holds gradeMu. While
// the run is incomplete (a test is still running, or the binary may crash) the
// code is -1, which can never count as a pass.
func gradeWrite(code int) {
	names := make([]string, 0, len(gradeRes))
	for n := range gradeRes {
		names = append(names, n)
	}
	sort.Strings(names)
	var b strings.Builder
	b.WriteString("code=" + strconv.Itoa(code) + "\n")
	for _, n := range names {
		r := "fail"
		if gradeRes[n] {
			r = "pass"
		}
		b.WriteString(r + " " + n + "\n")
	}
	mac := hmac.New(sha256.New, []byte(` + strconv.Quote(g.Key) + `))
	mac.Write([]byte(b.String()))
	b.WriteString("mac=" + hex.EncodeToString(mac.Sum(nil)) + "\n")
	_ = os.WriteFile(` + strconv.Quote(g.VerdictPath) + `, []byte(b.String()), 0o600)
}

func ` + rec + `(name string, ok bool) {
	gradeMu.Lock()
	defer gradeMu.Unlock()
	if prev, seen := gradeRes[name]; seen {
		ok = ok && prev
	}
	gradeRes[name] = ok
	gradeWrite(-1)
}

func TestMain(m *testing.M) {
	code := m.Run()
	gradeMu.Lock()
	gradeWrite(code)
	gradeMu.Unlock()
	os.Exit(code)
}
`
}

// ReadVerdict loads and authenticates the verdict file. A missing, malformed
// or wrongly signed file yields Valid=false rather than an error: it simply
// means the run produced no trustworthy result.
func (g *Grader) ReadVerdict() *Verdict {
	raw, err := os.ReadFile(g.VerdictPath)
	if err != nil {
		return &Verdict{}
	}
	text := string(raw)
	i := strings.LastIndex(text, "mac=")
	if i < 0 || i == 0 || text[i-1] != '\n' {
		return &Verdict{}
	}
	body, macHex := text[:i], strings.TrimSpace(text[i+4:])
	got, err := hex.DecodeString(macHex)
	if err != nil {
		return &Verdict{}
	}
	mac := hmac.New(sha256.New, []byte(g.Key))
	mac.Write([]byte(body))
	if !hmac.Equal(got, mac.Sum(nil)) {
		return &Verdict{}
	}
	v := &Verdict{Valid: true, Tests: map[string]bool{}}
	lines := strings.Split(strings.TrimSuffix(body, "\n"), "\n")
	if len(lines) == 0 || !strings.HasPrefix(lines[0], "code=") {
		return &Verdict{}
	}
	code, err := strconv.Atoi(strings.TrimPrefix(lines[0], "code="))
	if err != nil {
		return &Verdict{}
	}
	v.ExitCode = code
	for _, l := range lines[1:] {
		status, name, ok := strings.Cut(l, " ")
		if !ok || (status != "pass" && status != "fail") {
			return &Verdict{}
		}
		v.Tests[name] = status == "pass"
	}
	return v
}
