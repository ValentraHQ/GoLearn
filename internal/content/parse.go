package content

import (
	"fmt"
	"regexp"
	"strings"
)

// The authoring format is Markdown with three heading levels:
//
//	# Module title            (module file header, followed by key: value lines)
//	## Item title             (lesson / challenge / project / glossary term, followed by key: value lines)
//	### Section               (free-form body)
//
// Headings inside fenced code blocks are ignored.

type section struct {
	name string
	body string
}

type item struct {
	title string
	meta  map[string]string
	subs  []section
}

func (it *item) sub(name string) (string, bool) {
	for _, s := range it.subs {
		if strings.EqualFold(s.name, name) {
			return s.body, true
		}
	}
	return "", false
}

type document struct {
	head  item
	items []*item
}

var metaRe = regexp.MustCompile(`^([a-z][a-z_-]*):\s?(.*)$`)

func parseDocument(src string) (*document, error) {
	doc := &document{}
	var cur *item
	var sec *section
	var buf []string
	inFence := false
	metaOpen := false
	flush := func() {
		if sec != nil && cur != nil {
			sec.body = strings.Trim(strings.Join(buf, "\n"), "\n")
			cur.subs = append(cur.subs, *sec)
		}
		sec, buf = nil, nil
	}
	for i, line := range strings.Split(strings.ReplaceAll(src, "\r\n", "\n"), "\n") {
		if strings.HasPrefix(line, "```") {
			inFence = !inFence
		}
		if !inFence {
			switch {
			case strings.HasPrefix(line, "# "):
				flush()
				doc.head = item{title: strings.TrimSpace(line[2:]), meta: map[string]string{}}
				cur = &doc.head
				metaOpen = true
				continue
			case strings.HasPrefix(line, "## "):
				flush()
				it := &item{title: strings.TrimSpace(line[3:]), meta: map[string]string{}}
				doc.items = append(doc.items, it)
				cur = it
				metaOpen = true
				continue
			case strings.HasPrefix(line, "### "):
				if cur == nil {
					return nil, fmt.Errorf("line %d: section before any item", i+1)
				}
				flush()
				sec = &section{name: strings.TrimSpace(line[4:])}
				metaOpen = false
				continue
			}
			if metaOpen && cur != nil {
				if m := metaRe.FindStringSubmatch(line); m != nil {
					cur.meta[m[1]] = strings.TrimSpace(m[2])
					continue
				}
				if strings.TrimSpace(line) == "" {
					if len(cur.meta) > 0 {
						metaOpen = false
					}
					continue
				}
				return nil, fmt.Errorf("line %d: unexpected text after heading %q (put prose under a ### section)", i+1, cur.title)
			}
		}
		if sec != nil {
			buf = append(buf, line)
		}
	}
	flush()
	if inFence {
		return nil, fmt.Errorf("unterminated code fence")
	}
	return doc, nil
}

type fence struct {
	info string
	code string
}

// splitFences separates a body into prose and fenced code blocks.
func splitFences(body string) (prose string, fences []fence) {
	var p, c []string
	in := false
	var info string
	for _, line := range strings.Split(body, "\n") {
		if strings.HasPrefix(line, "```") {
			if !in {
				in = true
				info = strings.TrimSpace(line[3:])
				c = nil
			} else {
				in = false
				fences = append(fences, fence{info: info, code: strings.Join(c, "\n")})
			}
			continue
		}
		if in {
			c = append(c, line)
		} else {
			p = append(p, line)
		}
	}
	return strings.Trim(strings.Join(p, "\n"), "\n"), fences
}

func splitList(s string, sep string) []string {
	out := []string{}
	for _, p := range strings.Split(s, sep) {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// bullets returns "- x" list items (multi-line items are joined).
func bullets(body string) []string {
	out := []string{}
	for _, line := range strings.Split(body, "\n") {
		t := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(t, "- "):
			out = append(out, strings.TrimSpace(t[2:]))
		case t != "" && len(out) > 0 && strings.HasPrefix(line, "  "):
			out[len(out)-1] += " " + t
		}
	}
	return out
}

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	return strings.Trim(slugRe.ReplaceAllString(strings.ToLower(s), "-"), "-")
}

const (
	beginMarker = "// BEGIN"
	endMarker   = "// END"
)

// splitSolution returns the runnable solution (markers removed) and a starter
// where each BEGIN/END region is replaced by a TODO comment.
func splitSolution(code string) (solution, starter string, marked bool) {
	var sol, st []string
	skipping := false
	for _, line := range strings.Split(code, "\n") {
		t := strings.TrimSpace(line)
		switch {
		case t == beginMarker:
			marked = true
			skipping = true
			indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
			st = append(st, indent+"// TODO: write your code here")
		case t == endMarker:
			skipping = false
		default:
			sol = append(sol, line)
			if !skipping {
				st = append(st, line)
			}
		}
	}
	return strings.Join(sol, "\n") + "\n", strings.Join(st, "\n") + "\n", marked
}
