package content

import (
	"embed"
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

//go:embed data
var dataFS embed.FS

// Catalog is the immutable, in-memory curriculum.
type Catalog struct {
	Modules    []*Module
	Challenges []*Challenge
	Projects   []*Project
	Glossary   []*GlossaryTerm
	Skills     []Skill

	moduleByID    map[string]*Module
	lessonByID    map[string]*Lesson
	challengeByID map[string]*Challenge
	projectByID   map[string]*Project
	termByID      map[string]*GlossaryTerm
	skillByID     map[string]Skill
	published     []*Lesson // published lessons in curriculum order
}

func (c *Catalog) Module(id string) *Module       { return c.moduleByID[id] }
func (c *Catalog) Lesson(id string) *Lesson       { return c.lessonByID[id] }
func (c *Catalog) Challenge(id string) *Challenge { return c.challengeByID[id] }
func (c *Catalog) Project(id string) *Project     { return c.projectByID[id] }
func (c *Catalog) Term(id string) *GlossaryTerm   { return c.termByID[id] }
func (c *Catalog) Skill(id string) (Skill, bool)  { s, ok := c.skillByID[id]; return s, ok }

// PublishedLessons returns published lessons in curriculum order.
func (c *Catalog) PublishedLessons() []*Lesson { return c.published }

// NextLesson returns the published lesson following id, or nil.
func (c *Catalog) NextLesson(id string) *Lesson {
	l := c.lessonByID[id]
	if l == nil || l.Status != StatusPublished {
		return nil
	}
	if l.Order+1 < len(c.published) {
		return c.published[l.Order+1]
	}
	return nil
}

func (c *Catalog) PrevLesson(id string) *Lesson {
	l := c.lessonByID[id]
	if l == nil || l.Status != StatusPublished || l.Order == 0 {
		return nil
	}
	return c.published[l.Order-1]
}

var defaultSkills = []Skill{
	{ID: "fundamentals", Name: "Go Fundamentals"},
	{ID: "functions", Name: "Functions"},
	{ID: "data-structures", Name: "Data Structures"},
	{ID: "structs", Name: "Structs"},
	{ID: "pointers", Name: "Pointers"},
	{ID: "interfaces", Name: "Interfaces"},
	{ID: "errors", Name: "Error Handling"},
	{ID: "packages", Name: "Packages"},
	{ID: "stdlib", Name: "Standard Library"},
	{ID: "concurrency", Name: "Concurrency"},
	{ID: "networking", Name: "Networking"},
	{ID: "http", Name: "HTTP"},
	{ID: "rest", Name: "REST"},
	{ID: "databases", Name: "Databases"},
	{ID: "testing", Name: "Testing"},
	{ID: "cli", Name: "CLI"},
	{ID: "docker", Name: "Docker"},
	{ID: "kubernetes", Name: "Kubernetes"},
	{ID: "observability", Name: "Observability"},
	{ID: "generics", Name: "Generics"},
	{ID: "performance", Name: "Performance"},
	{ID: "architecture", Name: "Architecture"},
	{ID: "security", Name: "Security"},
	{ID: "microservices", Name: "Microservices"},
}

// Load parses the embedded curriculum and validates cross-references.
func Load() (*Catalog, error) { return LoadFS(dataFS, "data") }

func LoadFS(fsys fs.FS, root string) (*Catalog, error) {
	c := &Catalog{
		moduleByID: map[string]*Module{}, lessonByID: map[string]*Lesson{},
		challengeByID: map[string]*Challenge{}, projectByID: map[string]*Project{},
		termByID: map[string]*GlossaryTerm{}, skillByID: map[string]Skill{},
	}
	for i, s := range defaultSkills {
		s.Order = i
		c.Skills = append(c.Skills, s)
		c.skillByID[s.ID] = s
	}

	files, err := fs.Glob(fsys, root+"/modules/*.md")
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	for _, f := range files {
		b, err := fs.ReadFile(fsys, f)
		if err != nil {
			return nil, err
		}
		m, err := parseModule(f, string(b))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", f, err)
		}
		if c.moduleByID[m.ID] != nil {
			return nil, fmt.Errorf("%s: duplicate module id %q", f, m.ID)
		}
		c.Modules = append(c.Modules, m)
		c.moduleByID[m.ID] = m
	}
	sort.SliceStable(c.Modules, func(i, j int) bool { return c.Modules[i].Number < c.Modules[j].Number })
	for _, m := range c.Modules {
		for _, l := range m.Lessons {
			if c.lessonByID[l.ID] != nil {
				return nil, fmt.Errorf("duplicate lesson id %q", l.ID)
			}
			c.lessonByID[l.ID] = l
			if l.Status == StatusPublished {
				l.Order = len(c.published)
				c.published = append(c.published, l)
			}
		}
	}

	files, _ = fs.Glob(fsys, root+"/challenges/*.md")
	sort.Strings(files)
	for _, f := range files {
		b, _ := fs.ReadFile(fsys, f)
		chs, err := parseChallenges(string(b))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", f, err)
		}
		for _, ch := range chs {
			if c.challengeByID[ch.ID] != nil {
				return nil, fmt.Errorf("duplicate challenge id %q", ch.ID)
			}
			c.challengeByID[ch.ID] = ch
			c.Challenges = append(c.Challenges, ch)
		}
	}
	if b, err := fs.ReadFile(fsys, root+"/projects.md"); err == nil {
		ps, err := parseProjects(string(b))
		if err != nil {
			return nil, fmt.Errorf("projects.md: %w", err)
		}
		for _, p := range ps {
			c.projectByID[p.ID] = p
			c.Projects = append(c.Projects, p)
		}
	}
	if b, err := fs.ReadFile(fsys, root+"/glossary.md"); err == nil {
		ts, err := parseGlossary(string(b))
		if err != nil {
			return nil, fmt.Errorf("glossary.md: %w", err)
		}
		for _, t := range ts {
			c.termByID[t.ID] = t
			c.Glossary = append(c.Glossary, t)
		}
		sort.Slice(c.Glossary, func(i, j int) bool {
			return strings.ToLower(c.Glossary[i].Term) < strings.ToLower(c.Glossary[j].Term)
		})
	}
	return c, c.validate()
}

func (c *Catalog) validate() error {
	var errs []string
	add := func(f string, a ...any) { errs = append(errs, fmt.Sprintf(f, a...)) }
	for _, m := range c.Modules {
		if _, ok := c.skillByID[m.Skill]; !ok {
			add("module %s: unknown skill %q", m.ID, m.Skill)
		}
		for _, r := range m.Requires {
			if c.moduleByID[r] == nil {
				add("module %s: unknown required module %q", m.ID, r)
			}
		}
		if m.Project != "" && c.projectByID[m.Project] == nil {
			add("module %s: unknown project %q", m.ID, m.Project)
		}
		for _, l := range m.Lessons {
			if _, ok := c.skillByID[l.Skill]; !ok {
				add("lesson %s: unknown skill %q", l.ID, l.Skill)
			}
			for _, ch := range l.Challenges {
				if c.challengeByID[ch] == nil {
					add("lesson %s: unknown challenge %q", l.ID, ch)
				}
			}
			if l.Status == StatusPublished {
				if l.Concept == "" || len(l.Objectives) == 0 || len(l.Takeaways) == 0 || len(l.Examples) == 0 {
					add("lesson %s: published lessons need objectives, concept, an example and takeaways", l.ID)
				}
				if len(l.Quiz) == 0 {
					add("lesson %s: published lessons need a knowledge check", l.ID)
				}
			}
		}
	}
	for _, ch := range c.Challenges {
		if c.moduleByID[ch.ModuleID] == nil {
			add("challenge %s: unknown module %q", ch.ID, ch.ModuleID)
		}
		if ch.Lesson != "" && c.lessonByID[ch.Lesson] == nil {
			add("challenge %s: unknown lesson %q", ch.ID, ch.Lesson)
		}
		if _, ok := c.skillByID[ch.Skill]; !ok {
			add("challenge %s: unknown skill %q", ch.ID, ch.Skill)
		}
	}
	for _, p := range c.Projects {
		if p.ModuleID != "" && c.moduleByID[p.ModuleID] == nil {
			add("project %s: unknown module %q", p.ID, p.ModuleID)
		}
		for _, s := range p.Skills {
			if _, ok := c.skillByID[s]; !ok {
				add("project %s: unknown skill %q", p.ID, s)
			}
		}
	}
	for _, t := range c.Glossary {
		for _, r := range t.Related {
			if c.lessonByID[r] == nil {
				add("glossary %s: unknown lesson %q", t.ID, r)
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("content validation failed:\n  %s", strings.Join(errs, "\n  "))
	}
	return nil
}

func parseModule(file, src string) (*Module, error) {
	doc, err := parseDocument(src)
	if err != nil {
		return nil, err
	}
	h := doc.head
	m := &Module{
		ID: h.meta["id"], Number: h.meta["number"], Title: h.title, Summary: h.meta["summary"],
		Track: h.meta["track"], Paths: splitList(h.meta["paths"], ","), Skill: h.meta["skill"],
		Requires: splitList(h.meta["requires"], ","), Project: h.meta["project"],
	}
	if m.ID == "" || m.Number == "" || m.Skill == "" {
		return nil, fmt.Errorf("module header needs id, number and skill")
	}
	if len(m.Paths) == 0 {
		m.Paths = []string{"beginner", "pro"}
	}
	for _, it := range doc.items {
		l, err := parseLesson(m, it)
		if err != nil {
			return nil, fmt.Errorf("lesson %q: %w", it.title, err)
		}
		m.Lessons = append(m.Lessons, l)
	}
	return m, nil
}

func parseLesson(m *Module, it *item) (*Lesson, error) {
	l := &Lesson{
		Title: it.title, Slug: it.meta["slug"], ModuleID: m.ID, Skill: it.meta["skill"],
		Status: Status(it.meta["status"]), Minutes: 5,
		Objectives: splitList(it.meta["objectives"], ";"),
		Takeaways:  splitList(it.meta["takeaways"], ";"),
		Challenges: splitList(it.meta["challenges"], ","),
	}
	if l.Slug == "" {
		l.Slug = slugify(it.title)
	}
	l.ID = m.ID + "/" + l.Slug
	if l.Skill == "" {
		l.Skill = m.Skill
	}
	if l.Status == "" {
		l.Status = StatusPublished
	}
	if v := it.meta["minutes"]; v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("bad minutes %q", v)
		}
		l.Minutes = n
	}
	if l.Status == StatusPlanned {
		return l, nil
	}
	if b, ok := it.sub("Concept"); ok {
		l.Concept = b
	}
	if b, ok := it.sub("Example"); ok {
		prose, fs := splitFences(b)
		l.Examples = examplesFrom(prose, fs)
	}
	if b, ok := it.sub("Exercise"); ok {
		ex, err := parseExercise(b)
		if err != nil {
			return nil, err
		}
		l.Exercise = ex
	}
	if b, ok := it.sub("Check"); ok {
		qs, err := parseQuestions(b, l.ID)
		if err != nil {
			return nil, err
		}
		l.Quiz = qs
	}
	return l, nil
}

func examplesFrom(prose string, fs []fence) []Example {
	var out []Example
	for i, f := range fs {
		parts := strings.Fields(f.info)
		if len(parts) == 0 {
			parts = []string{"text"}
		}
		ex := Example{Lang: parts[0], Code: strings.TrimRight(f.code, "\n") + "\n"}
		ex.Runnable = ex.Lang == "go" && !contains(parts[1:], "norun")
		if i == 0 {
			ex.Caption = prose
		}
		out = append(out, ex)
	}
	return out
}

func contains(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}

func parseExercise(body string) (*Exercise, error) {
	prose, fs := splitFences(body)
	ex := &Exercise{Prompt: prose}
	var starterOverride string
	for _, f := range fs {
		switch f.info {
		case "text expect", "expect":
			ex.Expected = strings.TrimRight(f.code, "\n") + "\n"
		case "go solution":
			sol, st, marked := splitSolution(f.code)
			ex.Solution = sol
			if marked {
				ex.Starter = st
			}
		case "go starter":
			starterOverride = strings.TrimRight(f.code, "\n") + "\n"
		}
	}
	if starterOverride != "" {
		ex.Starter = starterOverride
	}
	if ex.Solution == "" || ex.Expected == "" || ex.Starter == "" {
		return nil, fmt.Errorf("exercise needs a `go solution` (with // BEGIN / // END markers or a `go starter`) and a `text expect` block")
	}
	return ex, nil
}

var optRe = regexp.MustCompile(`^- \[( |x)\] (.*)$`)

func parseQuestions(body, lessonID string) ([]Question, error) {
	var qs []Question
	var cur *Question
	var codeLines []string
	inFence := false
	finish := func() error {
		if cur == nil {
			return nil
		}
		if len(codeLines) > 0 {
			cur.Code = strings.Join(codeLines, "\n") + "\n"
		}
		codeLines = nil
		switch cur.Type {
		case "tf":
			if len(cur.Options) == 0 {
				cur.Options = []string{"True", "False"}
			}
		case "short":
			if len(cur.Accept) == 0 {
				return fmt.Errorf("question %q: short answer needs A:", cur.Prompt)
			}
		}
		if cur.Type != "short" && len(cur.Correct) == 0 {
			return fmt.Errorf("question %q: no correct option marked", cur.Prompt)
		}
		if cur.Type != "short" && len(cur.Options) < 2 {
			return fmt.Errorf("question %q: needs at least two options", cur.Prompt)
		}
		if cur.Type != "multi" && cur.Type != "short" && len(cur.Correct) != 1 {
			return fmt.Errorf("question %q: exactly one correct option expected", cur.Prompt)
		}
		if cur.Explanation == "" {
			return fmt.Errorf("question %q: missing explanation", cur.Prompt)
		}
		cur.ID = fmt.Sprintf("q%d", len(qs)+1)
		qs = append(qs, *cur)
		cur = nil
		return nil
	}
	for _, line := range strings.Split(body, "\n") {
		if strings.HasPrefix(line, "```") {
			inFence = !inFence
			continue
		}
		if inFence {
			codeLines = append(codeLines, line)
			continue
		}
		t := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(t, "Q:"):
			if err := finish(); err != nil {
				return nil, err
			}
			cur = &Question{Type: "mcq", Prompt: strings.TrimSpace(t[2:])}
		case cur == nil || t == "":
		case strings.HasPrefix(t, "T:"):
			cur.Type = strings.TrimSpace(t[2:])
		case strings.HasPrefix(t, "E:"):
			cur.Explanation = strings.TrimSpace(t[2:])
		case strings.HasPrefix(t, "A:"):
			v := strings.TrimSpace(t[2:])
			if cur.Type == "tf" {
				cur.Options = []string{"True", "False"}
				if strings.EqualFold(v, "true") {
					cur.Correct = []int{0}
				} else {
					cur.Correct = []int{1}
				}
			} else {
				cur.Accept = splitList(v, "|")
			}
		default:
			if m := optRe.FindStringSubmatch(t); m != nil {
				if m[1] == "x" {
					cur.Correct = append(cur.Correct, len(cur.Options))
				}
				cur.Options = append(cur.Options, m[2])
			} else if cur.Explanation != "" {
				cur.Explanation += " " + t
			} else {
				return nil, fmt.Errorf("unexpected line in check: %q", t)
			}
		}
	}
	if err := finish(); err != nil {
		return nil, err
	}
	return qs, nil
}

var testFuncRe = regexp.MustCompile(`(?m)^func (Test\w+)\(t \*testing\.T\)`)

func parseChallenges(src string) ([]*Challenge, error) {
	doc, err := parseDocument(src)
	if err != nil {
		return nil, err
	}
	var out []*Challenge
	for _, it := range doc.items {
		ch := &Challenge{
			ID: slugify(it.title), Title: it.meta["title"], Difficulty: it.meta["difficulty"],
			ModuleID: it.meta["module"], Lesson: it.meta["lesson"], Skill: it.meta["skill"],
		}
		if v := it.meta["id"]; v != "" {
			ch.ID = v
		}
		if ch.Title == "" {
			ch.Title = it.title
		}
		switch ch.Difficulty {
		case "beginner", "intermediate", "advanced", "expert":
		default:
			return nil, fmt.Errorf("challenge %s: bad difficulty %q", ch.ID, ch.Difficulty)
		}
		if b, ok := it.sub("Problem"); ok {
			ch.Problem = b
		}
		if b, ok := it.sub("Input"); ok {
			ch.Input = b
		}
		if b, ok := it.sub("Expected"); ok {
			ch.Expected = b
		}
		if b, ok := it.sub("Constraints"); ok {
			ch.Constraints = bullets(b)
		}
		if b, ok := it.sub("Hints"); ok {
			ch.Hints = bullets(b)
		}
		if b, ok := it.sub("Explanation"); ok {
			ch.Explanation = b
		}
		for name, dst := range map[string]*string{"Starter": &ch.Starter, "Tests": &ch.Tests, "Solution": &ch.Solution} {
			b, ok := it.sub(name)
			if !ok {
				return nil, fmt.Errorf("challenge %s: missing ### %s", ch.ID, name)
			}
			_, fs := splitFences(b)
			if len(fs) != 1 {
				return nil, fmt.Errorf("challenge %s: ### %s needs exactly one code block", ch.ID, name)
			}
			*dst = strings.TrimRight(fs[0].code, "\n") + "\n"
		}
		for _, m := range testFuncRe.FindAllStringSubmatch(ch.Tests, -1) {
			ch.TestNames = append(ch.TestNames, m[1])
		}
		if len(ch.TestNames) == 0 || ch.Problem == "" {
			return nil, fmt.Errorf("challenge %s: needs a problem and at least one Test function", ch.ID)
		}
		out = append(out, ch)
	}
	return out, nil
}

func parseProjects(src string) ([]*Project, error) {
	doc, err := parseDocument(src)
	if err != nil {
		return nil, err
	}
	var out []*Project
	for _, it := range doc.items {
		p := &Project{
			ID: slugify(it.title), Title: it.title, Level: it.meta["level"], ModuleID: it.meta["module"],
			Skills: splitList(it.meta["skills"], ","),
		}
		if v := it.meta["id"]; v != "" {
			p.ID = v
		}
		p.Hours, _ = strconv.Atoi(it.meta["hours"])
		switch p.Level {
		case "beginner", "intermediate", "advanced", "capstone":
		default:
			return nil, fmt.Errorf("project %s: bad level %q", p.ID, p.Level)
		}
		if b, ok := it.sub("Summary"); ok {
			p.Summary = b
		}
		if b, ok := it.sub("Description"); ok {
			p.Description = b
		}
		if b, ok := it.sub("Requirements"); ok {
			p.Requirements = bullets(b)
		}
		if b, ok := it.sub("Architecture"); ok {
			_, fs := splitFences(b)
			if len(fs) > 0 {
				p.Architecture = fs[0].code
			}
		}
		if b, ok := it.sub("Stretch"); ok {
			p.Stretch = bullets(b)
		}
		if b, ok := it.sub("Tasks"); ok {
			for i, t := range bullets(b) {
				title, notes, _ := strings.Cut(t, " :: ")
				p.Tasks = append(p.Tasks, ProjectTask{ID: fmt.Sprintf("t%d", i+1), Title: strings.TrimSpace(title), Notes: strings.TrimSpace(notes)})
			}
		}
		if len(p.Tasks) == 0 || p.Summary == "" {
			return nil, fmt.Errorf("project %s: needs a summary and tasks", p.ID)
		}
		out = append(out, p)
	}
	return out, nil
}

func parseGlossary(src string) ([]*GlossaryTerm, error) {
	doc, err := parseDocument(src)
	if err != nil {
		return nil, err
	}
	var out []*GlossaryTerm
	for _, it := range doc.items {
		t := &GlossaryTerm{ID: slugify(it.title), Term: it.title, Related: splitList(it.meta["related"], ",")}
		t.Simple, _ = it.sub("Simple")
		t.Technical, _ = it.sub("Technical")
		if b, ok := it.sub("Code"); ok {
			_, fs := splitFences(b)
			if len(fs) > 0 {
				t.Code = strings.TrimRight(fs[0].code, "\n") + "\n"
			}
		}
		if t.Simple == "" || t.Technical == "" {
			return nil, fmt.Errorf("term %s: needs Simple and Technical sections", t.Term)
		}
		out = append(out, t)
	}
	return out, nil
}
