package lobby

type ExperienceLevel int

const (
	Beginner ExperienceLevel = iota
	Easy
	Medium
	Hard
	Very_Hard
	Impossible
)

type Player struct {
	name             string
	info             string
	location         string
	email            string
	experience_level ExperienceLevel
}

type Lobby struct {
	players []Player
}
