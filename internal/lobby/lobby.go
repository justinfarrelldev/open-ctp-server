package lobby

import "time"

const (
	lobbyEmptyInactivityTimeout    = 30 * time.Minute
	lobbyAbsoluteInactivityTimeout = 2 * time.Hour
)

// Lobby represents a player lobby.
//
// @Description Structure for representing a lobby, including ownership, visibility, mute state, and lifecycle timestamps.
type Lobby struct {
	ID             int64      `json:"id,omitempty" db:"id"`
	Name           string     `json:"name" db:"name"`
	OwnerName      string     `json:"owner_name" db:"owner_name"`
	OwnerAccountID int64      `json:"owner_account_id" db:"owner_account_id"`
	IsClosed       bool       `json:"is_closed" db:"is_closed"`
	IsMuted        bool       `json:"is_muted" db:"is_muted"`
	IsPublic       bool       `json:"is_public" db:"is_public"`
	CreatedAt      time.Time  `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at,omitempty" db:"updated_at"`
	LastActivityAt time.Time  `json:"last_activity_at,omitempty" db:"last_activity_at"`
	ClosedAt       *time.Time `json:"closed_at,omitempty" db:"closed_at"`
}

func (l Lobby) ShouldCloseForInactivity(now time.Time, connectedSubscriberCount int) bool {
	if l.ClosedAt != nil {
		return false
	}

	inactiveFor := now.Sub(l.LastActivityAt)
	if inactiveFor >= lobbyAbsoluteInactivityTimeout {
		return true
	}

	return connectedSubscriberCount == 0 && inactiveFor >= lobbyEmptyInactivityTimeout
}

// LobbyParam represents the create and update fields accepted for lobby lifecycle mutations.
//
// @Description Structure for lobby fields that can be created or updated.
type LobbyParam struct {
	Name     *string `json:"name,omitempty" db:"name"`
	IsMuted  *bool   `json:"is_muted,omitempty" db:"is_muted"`
	IsPublic *bool   `json:"is_public,omitempty" db:"is_public"`
}

type lobbyOwnerRow struct {
	AccountID int64     `db:"id"`
	Name      string    `db:"name"`
	JoinedAt  time.Time `db:"joined_at"`
}
