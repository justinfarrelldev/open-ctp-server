package account

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAccountStatusTrackingResponseWriter_WriteSetsStatusAndBody(t *testing.T) {
	base := httptest.NewRecorder()
	tracked := &statusTrackingResponseWriter{ResponseWriter: base}

	n, err := tracked.Write([]byte("ok"))
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if n != len("ok") {
		t.Fatalf("Write() bytes = %d want %d", n, len("ok"))
	}
	if tracked.statusCode != http.StatusOK || !tracked.wroteHeader || !tracked.wroteBody {
		t.Fatalf("writer state = %+v want implicit ok with body tracking", tracked)
	}
	if body := base.Body.String(); body != "ok" {
		t.Fatalf("body = %q want %q", body, "ok")
	}
}

func TestHandleAccountHandlerError_WritesBodyWhenHeaderAlreadyWritten(t *testing.T) {
	base := httptest.NewRecorder()
	tracked := &statusTrackingResponseWriter{ResponseWriter: base}
	tracked.WriteHeader(http.StatusBadRequest)

	handleAccountHandlerError(tracked, errors.New("bad request detail"))

	if base.Code != http.StatusBadRequest {
		t.Fatalf("status = %d want %d", base.Code, http.StatusBadRequest)
	}
	if body := base.Body.String(); body != "bad request detail" {
		t.Fatalf("body = %q want %q", body, "bad request detail")
	}
}

func TestHandleAccountHandlerError_WritesInternalServerErrorWhenHeaderMissing(t *testing.T) {
	base := httptest.NewRecorder()
	tracked := &statusTrackingResponseWriter{ResponseWriter: base}

	handleAccountHandlerError(tracked, errors.New("boom"))

	if base.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d want %d", base.Code, http.StatusInternalServerError)
	}
	if body := strings.TrimSpace(base.Body.String()); !strings.Contains(body, "boom") {
		t.Fatalf("body = %q want substring %q", body, "boom")
	}
}
