package account

type ExperienceLevel int

const (
	Beginner ExperienceLevel = iota
	Easy
	Medium
	Hard
	Very_Hard
	Impossible
)

// Account represents a player account.
//
// @Description Structure for representing a player account.
type Account struct {
	// Name is the name of the player.
	Name string `json:"name"`

	// Info contains additional information about the player.
	Info string `json:"info"`

	// Location indicates the player's real-life location.
	Location string `json:"location"`

	// Email is the email address of the player.
	Email string `json:"email"`

	// ExperienceLevel represents the player's experience level (0=beginner, 1=easy, 2=medium, 3=hard, 4=very hard, 5=impossible)
	ExperienceLevel ExperienceLevel `json:"experience_level"`
}

// AccountParam represents a player account with non-required fields.
//
// @Description Structure for representing a player account with non-required fields.
type AccountParam struct {
	// Name is the name of the player.
	Name *string `json:"name,omitempty"`

	// Info contains additional information about the player.
	Info *string `json:"info,omitempty"`

	// Location indicates the player's real-life location.
	Location *string `json:"location,omitempty"`

	// Email is the email address of the player.
	Email *string `json:"email,omitempty"`

	// ExperienceLevel represents the player's experience level (0=beginner, 1=easy, 2=medium, 3=hard, 4=very hard, 5=impossible)
	ExperienceLevel *ExperienceLevel `json:"experience_level,omitempty"`
}
