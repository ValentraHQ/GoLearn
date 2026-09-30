package runner

import (
	"encoding/json"
	"strings"
)

type testEvent struct {
	Action string `json:"Action"`
	Test   string `json:"Test"`
	Output string `json:"Output"`
}

// ParseTestJSON turns `go test -json` output into per-test results.
// Subtests are folded into their parent so results map to top-level test names.
func ParseTestJSON(out string) []TestResult {
	var order []string
	byName := map[string]*TestResult{}
	var failedSub = map[string]bool{}
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "{") {
			continue
		}
		var ev testEvent
		if json.Unmarshal([]byte(line), &ev) != nil || ev.Test == "" {
			continue
		}
		top := ev.Test
		if i := strings.Index(ev.Test, "/"); i >= 0 {
			top = ev.Test[:i]
			if ev.Action == "fail" {
				failedSub[top] = true
			}
			// keep subtest output on the parent
			if ev.Action == "output" {
				appendOutput(byName, &order, top, ev.Output)
			}
			continue
		}
		tr := byName[top]
		if tr == nil {
			tr = &TestResult{Name: top}
			byName[top] = tr
			order = append(order, top)
		}
		switch ev.Action {
		case "output":
			appendOutput(byName, &order, top, ev.Output)
		case "pass":
			tr.Passed = true
		case "fail", "skip":
			tr.Passed = false
		}
	}
	res := make([]TestResult, 0, len(order))
	for _, n := range order {
		tr := byName[n]
		if failedSub[n] {
			tr.Passed = false
		}
		tr.Output = trimTestOutput(tr.Output)
		res = append(res, *tr)
	}
	return res
}

func appendOutput(m map[string]*TestResult, order *[]string, name, s string) {
	tr := m[name]
	if tr == nil {
		tr = &TestResult{Name: name}
		m[name] = tr
		*order = append(*order, name)
	}
	if len(tr.Output) < 4096 {
		tr.Output += s
	}
}

// trimTestOutput drops the framework's "=== RUN" / "--- PASS" chatter.
func trimTestOutput(s string) string {
	var keep []string
	for _, l := range strings.Split(s, "\n") {
		t := strings.TrimSpace(l)
		if strings.HasPrefix(t, "=== RUN") || strings.HasPrefix(t, "=== PAUSE") || strings.HasPrefix(t, "=== CONT") ||
			strings.HasPrefix(t, "--- PASS") || strings.HasPrefix(t, "--- FAIL") || strings.HasPrefix(t, "--- SKIP") || t == "" {
			continue
		}
		keep = append(keep, strings.TrimRight(l, " \t"))
	}
	return strings.Join(keep, "\n")
}
