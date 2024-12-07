package lobby

// Lobby represents a player lobby.
//
// @Description Structure for representing a player lobby.
type Lobby struct {
	// ID is the unique identifier for the lobby.
	ID int64 `json:"id,omitempty" db:"id"`

	// Name is the name of the lobby.
	Name string `json:"name" db:"name"`

	// OwnerName is the name of the lobby owner.
	OwnerName string `json:"owner_name" db:"owner_name"`

	// OwnerAccountId is the account ID of the lobby owner.
	OwnerAccountId int64 `json:"owner_account_id" db:"owner_account_id"`

	// IsClosed indicates if the lobby is closed.
	IsClosed bool `json:"is_closed" db:"is_closed"`

	// IsMuted indicates if the lobby is muted.
	IsMuted bool `json:"is_muted" db:"is_muted"`

	// IsPublic indicates if the lobby is public.
	IsPublic bool `json:"is_public" db:"is_public"`
}

// LobbyParam represents a player lobby with non-required fields.
//
// @Description Structure for representing a player lobby with non-required fields.
type LobbyParam struct {
	// ID is the unique identifier for the lobby.
	ID *int64 `json:"id,omitempty" db:"id"`

	// Name is the name of the lobby.
	Name *string `json:"name,omitempty" db:"name"`

	// OwnerName is the name of the lobby owner.
	OwnerName *string `json:"owner_name,omitempty" db:"owner_name"`

	// OwnerAccountId is the account ID of the lobby owner.
	OwnerAccountId *int64 `json:"owner_account_id,omitempty" db:"owner_account_id"`

	// IsClosed indicates if the lobby is closed.
	IsClosed *bool `json:"is_closed,omitempty" db:"is_closed"`

	// IsMuted indicates if the lobby is muted.
	IsMuted *bool `json:"is_muted,omitempty" db:"is_muted"`

	// IsPublic indicates if the lobby is public.
	IsPublic *bool `json:"is_public,omitempty" db:"is_public"`
}
