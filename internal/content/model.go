// Package content loads the curriculum (modules, lessons, challenges, projects,
// glossary) from embedded Markdown files and serves it from memory.
package content

type Status string

const (
	StatusDraft     Status = "draft"
	StatusReview    Status = "review"
	StatusPublished Status = "published"
	StatusArchived  Status = "archived"
	// StatusPlanned marks a lesson that is on the roadmap but not written yet.
	StatusPlanned Status = "planned"
)

type Module struct {
	ID       string    `json:"id"`
	Number   string    `json:"number"`
	Title    string    `json:"title"`
	Summary  string    `json:"summary"`
	Track    string    `json:"track"` // core, advanced, devops, architecture
	Paths    []string  `json:"paths"` // beginner, pro
	Skill    string    `json:"skill"`
	Requires []string  `json:"requires"`
	Project  string    `json:"project,omitempty"`
	Lessons  []*Lesson `json:"lessons"`
}

type Question struct {
	ID          string   `json:"id"`
	Type        string   `json:"type"` // mcq, tf, output, debug, multi, short
	Prompt      string   `json:"prompt"`
	Code        string   `json:"code,omitempty"`
	Options     []string `json:"options,omitempty"`
	Explanation string   `json:"explanation"`
	// Answers are the correct option indexes (mcq/tf/output/debug/multi) or
	// accepted strings (short). Never sent to clients before an attempt.
	Correct []int    `json:"-"`
	Accept  []string `json:"-"`
}

type Exercise struct {
	Prompt   string `json:"prompt"`
	Starter  string `json:"starter"`
	Expected string `json:"expected"`
	Solution string `json:"-"`
}

type Lesson struct {
	ID         string     `json:"id"` // "<module>/<slug>"
	Slug       string     `json:"slug"`
	ModuleID   string     `json:"moduleId"`
	Title      string     `json:"title"`
	Status     Status     `json:"status"`
	Minutes    int        `json:"minutes"`
	Skill      string     `json:"skill"`
	Objectives []string   `json:"objectives"`
	Concept    string     `json:"concept"`
	Examples   []Example  `json:"examples"`
	Exercise   *Exercise  `json:"exercise,omitempty"`
	Quiz       []Question `json:"quiz,omitempty"`
	Challenges []string   `json:"challenges,omitempty"` // challenge IDs
	Takeaways  []string   `json:"takeaways"`
	Order      int        `json:"order"` // global position among published lessons
}

type Example struct {
	Lang     string `json:"lang"`
	Code     string `json:"code"`
	Runnable bool   `json:"runnable"`
	Caption  string `json:"caption,omitempty"`
}

type Challenge struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Difficulty  string   `json:"difficulty"` // beginner, intermediate, advanced, expert
	ModuleID    string   `json:"moduleId"`
	Lesson      string   `json:"lesson,omitempty"` // lesson ID
	Skill       string   `json:"skill"`
	Problem     string   `json:"problem"`
	Input       string   `json:"input,omitempty"`
	Expected    string   `json:"expectedOutput,omitempty"`
	Constraints []string `json:"constraints"`
	Starter     string   `json:"starter"`
	Hints       []string `json:"hints"`
	Explanation string   `json:"explanation"`
	TestNames   []string `json:"testNames"`
	Tests       string   `json:"-"`
	Solution    string   `json:"-"`
}

type ProjectTask struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Notes string `json:"notes,omitempty"`
}

type Project struct {
	ID           string        `json:"id"`
	Title        string        `json:"title"`
	Level        string        `json:"level"` // beginner, intermediate, advanced, capstone
	ModuleID     string        `json:"moduleId,omitempty"`
	Skills       []string      `json:"skills"`
	Hours        int           `json:"hours"`
	Summary      string        `json:"summary"`
	Description  string        `json:"description"`
	Requirements []string      `json:"requirements"`
	Architecture string        `json:"architecture,omitempty"`
	Tasks        []ProjectTask `json:"tasks"`
	Stretch      []string      `json:"stretch,omitempty"`
}

type GlossaryTerm struct {
	ID        string   `json:"id"`
	Term      string   `json:"term"`
	Simple    string   `json:"simple"`
	Technical string   `json:"technical"`
	Code      string   `json:"code,omitempty"`
	Related   []string `json:"related"` // lesson IDs
}

type Skill struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Order int    `json:"order"`
}
