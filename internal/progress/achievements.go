package progress

import "github.com/valentrahq/golearn/internal/content"

type achievementContext struct {
	cat            *content.Catalog
	in             Input
	sum            *Summary
	done           func(string) bool
	projectDone    map[string]bool
	perfectQuizzes int
}

func (c achievementContext) moduleDone(id string) bool {
	for _, m := range c.sum.Modules {
		if m.ModuleID == id {
			return m.State == Completed
		}
	}
	return false
}

func (c achievementContext) exercisesPassed() int {
	n := 0
	for _, lp := range c.in.Lessons {
		if lp.ExercisePassed {
			n++
		}
	}
	return n
}

type achievementDef struct {
	ID, Title, Description, Category string
	Check                            func(achievementContext) bool
}

func moduleAch(id, title, desc, module string) achievementDef {
	return achievementDef{ID: id, Title: title, Description: desc, Category: "skill",
		Check: func(c achievementContext) bool { return c.moduleDone(module) }}
}

var achievementDefs = []achievementDef{
	{"first-program", "First Go Program", "Run an exercise that produces the expected output.", "milestone",
		func(c achievementContext) bool { return c.exercisesPassed() >= 1 }},
	{"first-lesson", "First Lesson", "Complete your first lesson.", "milestone",
		func(c achievementContext) bool { return c.sum.LessonsCompleted >= 1 }},
	{"lessons-25", "25 Lessons", "Complete 25 lessons.", "milestone",
		func(c achievementContext) bool { return c.sum.LessonsCompleted >= 25 }},
	{"lessons-100", "100 Lessons", "Complete 100 lessons.", "milestone",
		func(c achievementContext) bool { return c.sum.LessonsCompleted >= 100 }},
	{"challenges-10", "10 Challenges", "Pass 10 coding challenges.", "milestone",
		func(c achievementContext) bool { return c.sum.ChallengesCompleted >= 10 }},
	{"challenges-25", "25 Challenges", "Pass 25 coding challenges.", "milestone",
		func(c achievementContext) bool { return c.sum.ChallengesCompleted >= 25 }},
	{"challenges-50", "50 Challenges", "Pass 50 coding challenges.", "milestone",
		func(c achievementContext) bool { return c.sum.ChallengesCompleted >= 50 }},
	{"advanced-challenge", "Advanced Problem Solver", "Pass an advanced or expert challenge.", "milestone",
		func(c achievementContext) bool {
			for id, cs := range c.in.Challenges {
				if ch := c.cat.Challenge(id); ch != nil && cs.Passed && (ch.Difficulty == "advanced" || ch.Difficulty == "expert") {
					return true
				}
			}
			return false
		}},
	{"quiz-ace", "Quiz Ace", "Score 100% on five different knowledge checks.", "milestone",
		func(c achievementContext) bool { return c.perfectQuizzes >= 5 }},
	{"streak-3", "3-Day Streak", "Learn on three consecutive days.", "consistency",
		func(c achievementContext) bool { return c.sum.LongestStreak >= 3 }},
	{"streak-7", "7-Day Streak", "Learn on seven consecutive days.", "consistency",
		func(c achievementContext) bool { return c.sum.LongestStreak >= 7 }},
	{"streak-30", "30-Day Streak", "Learn on thirty consecutive days.", "consistency",
		func(c achievementContext) bool { return c.sum.LongestStreak >= 30 }},
	{"time-10h", "10 Hours of Learning", "Spend ten hours in lessons and challenges.", "consistency",
		func(c achievementContext) bool { return c.sum.LearningSeconds >= 10*3600 }},
	{"first-project", "First Project", "Complete every task of a project.", "milestone",
		func(c achievementContext) bool { return c.sum.ProjectsCompleted >= 1 }},
	{"projects-5", "Builder", "Complete five projects.", "milestone",
		func(c achievementContext) bool { return c.sum.ProjectsCompleted >= 5 }},
	moduleAch("fundamentals-complete", "Fundamentals Complete", "Finish the Go Fundamentals module.", "fundamentals"),
	moduleAch("data-structures-complete", "Data Structures", "Finish the Data Structures module.", "data-structures"),
	moduleAch("interfaces-complete", "Interface Designer", "Finish the Interfaces module.", "interfaces"),
	moduleAch("concurrency-master", "Concurrency Master", "Finish the Concurrency module.", "concurrency"),
	moduleAch("api-builder", "API Builder", "Finish the REST API Development module.", "rest-api"),
	moduleAch("testing-specialist", "Testing Specialist", "Finish the Testing module.", "testing"),
	moduleAch("kubernetes-developer", "Kubernetes Developer", "Finish the Kubernetes with Go module.", "kubernetes"),
	moduleAch("production-go-developer", "Production Go Developer", "Finish the Production Go module.", "production"),
	moduleAch("security-aware", "Security Aware", "Finish the Go Security module.", "security"),
}
