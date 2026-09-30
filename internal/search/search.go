// Package search provides in-memory full-text search over the curriculum.
package search

import (
	"sort"
	"strings"
	"unicode"

	"github.com/valentrahq/golearn/internal/content"
)

type Kind string

const (
	KindLesson    Kind = "lesson"
	KindModule    Kind = "module"
	KindChallenge Kind = "challenge"
	KindProject   Kind = "project"
	KindExample   Kind = "example"
	KindGlossary  Kind = "glossary"
)

type Result struct {
	Kind     Kind   `json:"kind"`
	ID       string `json:"id"`
	Title    string `json:"title"`
	Group    string `json:"group"` // module title for lessons, otherwise a section name
	Subtitle string `json:"subtitle,omitempty"`
	Snippet  string `json:"snippet,omitempty"`
	Path     string `json:"path"`
	Score    int    `json:"-"`
}

type doc struct {
	res    Result
	title  string // lowercased
	strong string // lowercased high-value text
	body   string // lowercased low-value text
}

type Index struct{ docs []doc }

func Build(c *content.Catalog) *Index {
	ix := &Index{}
	add := func(r Result, strong, body string) {
		ix.docs = append(ix.docs, doc{res: r, title: strings.ToLower(r.Title), strong: strings.ToLower(strong), body: strings.ToLower(body)})
	}
	for _, m := range c.Modules {
		add(Result{Kind: KindModule, ID: m.ID, Title: m.Number + " " + m.Title, Group: "Modules", Subtitle: m.Summary, Path: "/learn/" + m.ID},
			m.Title, m.Summary)
		for _, l := range m.Lessons {
			if l.Status != content.StatusPublished {
				continue
			}
			path := "/learn/" + m.ID + "/" + l.Slug
			add(Result{Kind: KindLesson, ID: l.ID, Title: l.Title, Group: m.Title, Path: path},
				l.Title+" "+strings.Join(l.Objectives, " "), l.Concept+" "+strings.Join(l.Takeaways, " "))
			for _, ex := range l.Examples {
				if ex.Lang == "" {
					continue
				}
				add(Result{Kind: KindExample, ID: l.ID, Title: l.Title, Group: "Code examples", Subtitle: m.Title, Path: path, Snippet: firstLines(ex.Code, 3)},
					"", ex.Code)
			}
		}
	}
	for _, ch := range c.Challenges {
		add(Result{Kind: KindChallenge, ID: ch.ID, Title: ch.Title, Group: "Challenges", Subtitle: ch.Difficulty, Path: "/challenge/" + ch.ID},
			ch.Title, ch.Problem)
	}
	for _, p := range c.Projects {
		add(Result{Kind: KindProject, ID: p.ID, Title: p.Title, Group: "Projects", Subtitle: p.Level, Path: "/projects/" + p.ID},
			p.Title+" "+p.Summary, p.Description)
	}
	for _, t := range c.Glossary {
		add(Result{Kind: KindGlossary, ID: t.ID, Title: t.Term, Group: "Glossary", Subtitle: t.Simple, Path: "/glossary#" + t.ID},
			t.Term, t.Simple+" "+t.Technical)
	}
	return ix
}

func firstLines(s string, n int) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	if len(lines) > n {
		lines = lines[:n]
	}
	return strings.Join(lines, "\n")
}

func tokenize(q string) []string {
	return strings.FieldsFunc(strings.ToLower(q), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '.' && r != '_' && r != '/' && r != ':'
	})
}

// Query returns up to limit results; every query token must appear somewhere in the document.
func (ix *Index) Query(q string, limit int) []Result {
	toks := tokenize(q)
	if len(toks) == 0 {
		return nil
	}
	var out []Result
	for _, d := range ix.docs {
		score := 0
		ok := true
		for _, t := range toks {
			s := 0
			switch {
			case d.title == t:
				s = 100
			case strings.HasPrefix(d.title, t) || strings.Contains(d.title, " "+t):
				s = 50
			case strings.Contains(d.title, t):
				s = 30
			case strings.Contains(d.strong, t):
				s = 12
			case strings.Contains(d.body, t):
				s = 3
			}
			if s == 0 {
				ok = false
				break
			}
			score += s
		}
		if !ok {
			continue
		}
		r := d.res
		r.Score = score
		out = append(out, r)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Score != out[j].Score {
			return out[i].Score > out[j].Score
		}
		return kindRank(out[i].Kind) < kindRank(out[j].Kind)
	})
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out
}

func kindRank(k Kind) int {
	switch k {
	case KindModule:
		return 0
	case KindLesson:
		return 1
	case KindChallenge:
		return 2
	case KindProject:
		return 3
	case KindGlossary:
		return 4
	}
	return 5
}
