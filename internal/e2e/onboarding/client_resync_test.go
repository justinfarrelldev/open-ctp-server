package onboarding

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type clientResyncTransport func(*http.Request) (*http.Response, error)

func (transport clientResyncTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	return transport(request)
}

func TestReadProjectedWorldUsesViewerSafeCurrentResync(t *testing.T) {
	client := HTTPClient{BaseURL: "https://stub.invalid", Client: &http.Client{Transport: clientResyncTransport(func(request *http.Request) (*http.Response, error) {
		if request.URL.Path != "/game/9/events" || request.Header.Get("X-Gameplay-Resync") != "current" || request.URL.Query().Get("session_id") != "22" {
			t.Errorf("read bypassed current viewer projection: %s %v", request.URL, request.Header)
		}
		return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"text/event-stream"}}, Body: io.NopCloser(strings.NewReader("event: resync\ndata: {\"current_revision\":7,\"snapshot\":{\"revision\":7,\"tiles\":[]}}\n\n"))}, nil
	})}}
	projection, err := client.ReadProjectedWorld(context.Background(), Actor{Session: 22}, 9)
	if err != nil || string(projection) != `{"revision":7,"tiles":[]}` {
		t.Fatalf("current snapshot not extracted exactly: %s err=%v", projection, err)
	}
}
