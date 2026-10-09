package runtime

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/didip/tollbooth/v7"
	"github.com/didip/tollbooth/v7/limiter"
	"github.com/jmoiron/sqlx"
)

func TestRuntimeExposesAutonomousAIWorker(t *testing.T) {
	db, mock := newRuntimeSQLMock(t)
	defer db.Close()
	runtime := New(db, Config{NewLimiter: func(string) *limiter.Limiter { return tollbooth.NewLimiter(1000000, nil) }})
	if runtime.AIWorker == nil {
		t.Fatal("runtime does not expose the autonomous AI worker")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected SQL execution: %v", err)
	}
}

func TestRouteSurfaceContract(t *testing.T) {
	t.Run("rejects a mux without the lobby chat surface", func(t *testing.T) {
		incorrect := http.NewServeMux()
		incorrect.Handle("GET /world/{game_id}", http.NotFoundHandler())
		incorrect.Handle("GET /world/{game_id}/space", http.NotFoundHandler())
		if err := assertRouteSurface(incorrect); err == nil {
			t.Fatal("route contract accepted a mux without /lobby/send_chat")
		}
	})

	db, mock := newRuntimeSQLMock(t)
	defer db.Close()
	runtime := New(db, Config{NewLimiter: func(string) *limiter.Limiter { return tollbooth.NewLimiter(1000000, nil) }})
	if err := assertRouteSurface(runtime.Handler); err != nil {
		t.Fatalf("production route contract: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected SQL execution: %v", err)
	}
}

func assertRouteSurface(handler http.Handler) error {
	for _, route := range []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/world/44"},
		{method: http.MethodGet, path: "/world/44/space"},
		{method: http.MethodPost, path: "/lobby/send_chat"},
	} {
		status := routeStatus(handler, route.method, route.path)
		if status == http.StatusNotFound || status == http.StatusMethodNotAllowed {
			return fmt.Errorf("required route %s %s returned %d", route.method, route.path, status)
		}
	}
	for _, route := range []struct {
		method string
		path   string
	}{
		{method: http.MethodPost, path: "/game/chat"},
		{method: http.MethodPost, path: "/world/chat"},
		{method: http.MethodGet, path: "/diagnostic"},
		{method: http.MethodPost, path: "/puppet"},
		{method: http.MethodGet, path: "/game/44/diagnostic"},
		{method: http.MethodPost, path: "/game/44/puppet"},
		{method: http.MethodGet, path: "/game/44/events/diagnostic"},
		{method: http.MethodGet, path: "/game/44/events/puppet"},
	} {
		status := routeStatus(handler, route.method, route.path)
		if status != http.StatusNotFound && status != http.StatusMethodNotAllowed {
			return fmt.Errorf("forbidden chat route %s %s returned %d", route.method, route.path, status)
		}
	}
	return nil
}

func routeStatus(handler http.Handler, method, path string) int {
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(method, path, nil))
	return recorder.Code
}

func newRuntimeSQLMock(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	t.Helper()
	database, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	return sqlx.NewDb(database, "sqlmock"), mock
}
