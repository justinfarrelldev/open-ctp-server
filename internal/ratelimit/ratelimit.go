package ratelimit

import (
	"encoding/json"
	"time"

	"github.com/didip/tollbooth/v7"
	"github.com/didip/tollbooth/v7/limiter"
)

const defaultMessageContentType = "application/json"

type Message struct {
	Status string `json:"status"`
	Body   string `json:"body"`
}

var defaultMessage = Message{
	Status: "Request Failed",
	Body:   "The API is at capacity, try again later.",
}

func newLimiter(max float64) *limiter.Limiter {
	lmt := tollbooth.NewLimiter(max, nil)
	lmt.SetMessageContentType(defaultMessageContentType)
	lmt.SetMessage(defaultMessageJSON())
	lmt.SetTokenBucketExpirationTTL(time.Hour)
	lmt.SetBasicAuthExpirationTTL(time.Hour)
	lmt.SetHeaderEntryExpirationTTL(time.Hour)
	return lmt
}

func defaultMessageJSON() string {
	message, _ := json.Marshal(defaultMessage)
	return string(message)
}

// NewAuthWriteLimiter returns the limiter used for auth-sensitive writes.
func NewAuthWriteLimiter() *limiter.Limiter {
	return newLimiter(10.0 / 60.0)
}

// NewAccountLimiter returns the limiter used for account read/update/delete routes.
func NewAccountLimiter() *limiter.Limiter {
	return newLimiter(60.0 / 60.0)
}

// NewLobbyGameLimiter returns the limiter used for lobby and game standard actions.
func NewLobbyGameLimiter() *limiter.Limiter {
	return newLimiter(120.0 / 60.0)
}

// NewLobbySSELimiter returns the limiter used for lobby SSE connections.
func NewLobbySSELimiter() *limiter.Limiter {
	return newLimiter(20.0 / 60.0)
}

// NewHealthLimiter returns the limiter used for the health endpoint.
func NewHealthLimiter() *limiter.Limiter {
	return newLimiter(5)
}
