package lobby

// Lobby represents a player lobby.
//
// @Description Structure for representing a player lobby.
type Lobby struct {
	// ID is the unique identifier for the lobby.
	ID string `json:"id"`

	// Name is the name of the lobby.
	Name string `json:"name"`

	// OwnerName is the name of the lobby owner.
	OwnerName string `json:"owner_name"`

	// IsClosed indicates if the lobby is closed.
	IsClosed bool `json:"is_closed"`

	// IsMuted indicates if the lobby is muted.
	IsMuted bool `json:"is_muted"`

	// IsPublic indicates if the lobby is public.
	IsPublic bool `json:"is_public"`
}

// LobbyParam represents a player lobby with non-required fields.
//
// @Description Structure for representing a player lobby with non-required fields.
type LobbyParam struct {
	// ID is the unique identifier for the lobby.
	ID *string `json:"id,omitempty"`

	// Name is the name of the lobby.
	Name *string `json:"name,omitempty"`

	// OwnerName is the name of the lobby owner.
	OwnerName *string `json:"owner_name,omitempty"`

	// IsClosed indicates if the lobby is closed.
	IsClosed *bool `json:"is_closed,omitempty"`

	// IsMuted indicates if the lobby is muted.
	IsMuted *bool `json:"is_muted,omitempty"`

	// IsPublic indicates if the lobby is public.
	IsPublic *bool `json:"is_public,omitempty"`
}
