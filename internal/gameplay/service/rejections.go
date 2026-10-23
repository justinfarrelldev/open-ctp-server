package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log/slog"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine/turns"
)

// Rejection is safe for an authenticated caller. Error is retained for older
// consumers; neither Message nor Code is derived from private error text.
type Rejection struct {
	Error         string `json:"error"`
	Code          string `json:"code"`
	Message       string `json:"message"`
	CorrelationID string `json:"correlation_id"`
}

var rejectionMessages = map[string]string{
	"request_invalid":         "The command request is malformed.",
	"authentication_required": "Authentication is required.",
	"controller_restricted":   "This player cannot submit this command.",
	"phase_restricted":        "This command is not available in the current phase.",
	"not_found":               "The requested gameplay endpoint is unavailable.",
	"revision_conflict":       "The game revision has changed. Refresh before choosing a new action.",
	"idempotency_conflict":    "This command identity has already been used.",
	"state_conflict":          "The game is not accepting this command.",
	"queue_identity":          "A city or queue entry identity is invalid for this edit.",
	"queue_shape":             "The production queue edit has an invalid shape.",
	"queue_capacity":          "The production queue limit is 256 entries.",
	"content_unavailable":     "This production content is unavailable.",
	"production_prerequisite": "The proposed production queue is not currently buildable.",
	"command_invalid":         "This command cannot be applied to the current game state.",
	"request_too_large":       "The command request is too large.",
	"rate_limited":            "The request limit or deadline was exceeded.",
	"internal_error":          "The server could not process this request.",
}

func defaultRejectionCode(status int) string {
	switch status {
	case 400, 405:
		return "request_invalid"
	case 401:
		return "authentication_required"
	case 403:
		return "controller_restricted"
	case 404:
		return "not_found"
	case 409:
		return "state_conflict"
	case 413:
		return "request_too_large"
	case 422:
		return "command_invalid"
	case 429:
		return "rate_limited"
	default:
		return "internal_error"
	}
}

// PublicRejection only accepts allowlisted codes, never Error.Message. Correlation
// identity is server-generated and independent of credentials and command text.
func PublicRejection(err error) Rejection {
	code := defaultRejectionCode(StatusCode(err))
	correlation := ""
	var target *Error
	if errors.As(err, &target) {
		if _, ok := rejectionMessages[target.Code]; ok {
			code = target.Code
		}
		correlation = target.CorrelationID
	}
	if correlation == "" {
		correlation = newCorrelationID()
	}
	return Rejection{Error: "gameplay request rejected", Code: code, Message: rejectionMessages[code], CorrelationID: correlation}
}

func newCorrelationID() string {
	var id [16]byte
	// crypto/rand failure is process-fatal in modern Go; fail closed rather than
	// falling back to command IDs or credentials as diagnostic identities.
	if _, err := rand.Read(id[:]); err != nil {
		panic("gameplay correlation entropy unavailable")
	}
	return hex.EncodeToString(id[:])
}

func codedError(status int, code, message string) error {
	return &Error{Status: status, Code: code, Message: message}
}

func engineRejection(err error) error {
	code := "command_invalid"
	var production *engine.ProductionError
	var phase *turns.Error
	if errors.As(err, &production) {
		code = production.Code
	} else if errors.As(err, &phase) {
		code = "phase_restricted"
	}
	return &Error{Status: 422, Code: code, Message: "gameplay command is invalid: " + err.Error(), Cause: err}
}

// attachDiagnostic never logs request payloads, session/account credentials, raw
// command IDs or arbitrary errors (e.g. JSON fields or database connection text).
// Production reasons are server-authored constants. Other details stay private
// in the returned error chain, not in public responses or unredacted logs.
func attachDiagnostic(err error, request CommandRequest) error {
	if err == nil {
		return nil
	}
	var target *Error
	if !errors.As(err, &target) {
		target = &Error{Status: 500, Message: "gameplay request failed", Cause: err}
	}
	copy := *target
	copy.CorrelationID = newCorrelationID()
	public := PublicRejection(&copy)
	reason := public.Code
	var production *engine.ProductionError
	if errors.As(copy.Cause, &production) {
		reason = production.Error()
	}
	fingerprint := sha256.Sum256([]byte(request.CommandID))
	slog.Warn("gameplay command rejected", "command_sha256", hex.EncodeToString(fingerprint[:]), "game_id", request.GameID, "correlation_id", copy.CorrelationID, "code", public.Code, "expected_revision", request.ExpectedRevision, "reason", reason)
	return &copy
}
