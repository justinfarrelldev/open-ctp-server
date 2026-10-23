package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine/turns"
)

func TestRejectionCodesNeverUsePrivateErrorText(t *testing.T) {
	for _, code := range []string{"queue_identity", "queue_shape", "queue_capacity", "content_unavailable", "production_prerequisite"} {
		public := PublicRejection(engineRejection(&engine.ProductionError{Code: code, Detail: errors.New("SECRET hidden wonder")}))
		raw, _ := json.Marshal(public)
		if public.Code != code || public.CorrelationID == "" || strings.Contains(string(raw), "SECRET") {
			t.Fatalf("unsafe/opaque error: %s", raw)
		}
	}
	if got := PublicRejection(engineRejection(&turns.Error{StatusCode: 422, Message: "private phase"})).Code; got != "phase_restricted" {
		t.Fatal(got)
	}
	for _, code := range []string{"revision_conflict", "idempotency_conflict", "controller_restricted"} {
		if got := PublicRejection(codedError(409, code, "SECRET")).Code; got != code {
			t.Fatal(got)
		}
	}
	if got := PublicRejection(&Error{Status: 422, Code: "SECRET", Message: "SECRET"}).Code; got != "command_invalid" {
		t.Fatal(got)
	}
}

func TestPrivateDiagnosticCorrelatesWithoutRequestCredentials(t *testing.T) {
	var output bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&output, nil)))
	defer slog.SetDefault(previous)
	request := CommandRequest{GameID: 9, CommandID: "SECRET-COMMAND", IdempotencyKey: "SECRET-KEY", Payload: []byte(`{"session_id":"SECRET-SESSION"}`)}
	err := attachDiagnostic(engineRejection(&engine.ProductionError{Code: "production_prerequisite", Detail: errors.New("city population is insufficient")}), request)
	public := PublicRejection(err)
	if !strings.Contains(output.String(), public.CorrelationID) || !strings.Contains(output.String(), "city population is insufficient") || !strings.Contains(output.String(), "command_sha256") || strings.Contains(output.String(), "SECRET") {
		t.Fatal(output.String())
	}
	// Arbitrary decoder/database errors can contain user-authored values; don't log them.
	output.Reset()
	_ = attachDiagnostic(errors.New("SECRET database password"), request)
	if strings.Contains(output.String(), "SECRET") {
		t.Fatal(output.String())
	}
}
