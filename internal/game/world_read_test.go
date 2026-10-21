package game

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	auth "github.com/justinfarrelldev/open-ctp-server/internal/auth"
)

// Authenticated membership is not permission to download hidden geography, the
// generation seed or foreign spawns. No snapshot query may run on this endpoint.
func TestGetWorldRejectsCanonicalMapDownload(t *testing.T) {
	sqlxDB, mock := newGameSQLMock(t)
	store := newGameSessionStore(sqlxDB)
	expectReadableWorldAccess(mock, 55, 7, gameRowWithLobby(44, 7, nil, "started", "World Game", nil, false, false, DefaultGameStyle, defaultRulesJSON, defaultExclusionsJSON, 2, 250))
	req := httptest.NewRequest(http.MethodGet, "/world/44?session_id=55", nil)
	req.SetPathValue("game_id", "44")
	rr := httptest.NewRecorder()
	GetWorldHandler(rr, req, sqlxDB, store)
	if rr.Code != http.StatusGone || !strings.Contains(rr.Body.String(), "viewer-safe") {
		t.Fatalf("canonical map endpoint remains accessible: %d %s", rr.Code, rr.Body.String())
	}
	for _, secret := range []string{"\"tiles\"", "\"seed\"", "\"spawns\""} {
		if strings.Contains(rr.Body.String(), secret) {
			t.Fatalf("canonical data leaked: %s", rr.Body.String())
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetWorldRequiresAuthAndMembership(t *testing.T) {
	t.Run("requires auth", func(t *testing.T) {
		sqlxDB, mock := newGameSQLMock(t)
		store := newGameSessionStore(sqlxDB)

		req := httptest.NewRequest(http.MethodGet, "/world/44", nil)
		req.SetPathValue("game_id", "44")
		rr := httptest.NewRecorder()
		GetWorldHandler(rr, req, sqlxDB, store)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("got status %d want %d body=%q", rr.Code, http.StatusUnauthorized, rr.Body.String())
		}
		if strings.TrimSpace(rr.Body.String()) != auth.ErrSessionIDRequired.Error() {
			t.Fatalf("got body %q want %q", rr.Body.String(), auth.ErrSessionIDRequired.Error())
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet sql expectations: %v", err)
		}
	})

	t.Run("requires membership", func(t *testing.T) {
		sqlxDB, mock := newGameSQLMock(t)
		store := newGameSessionStore(sqlxDB)

		expectGameSessionLookup(mock, 55, 7)
		mock.ExpectQuery(regexp.QuoteMeta(gameSelectByID)).
			WithArgs(int64(44)).
			WillReturnRows(gameRowWithLobby(44, 7, nil, "started", "World Game", nil, false, false, DefaultGameStyle, defaultRulesJSON, defaultExclusionsJSON, 2, 250))
		mock.ExpectQuery(regexp.QuoteMeta(gamePlayerByAccountQuery)).
			WithArgs(int64(44), int64(7)).
			WillReturnError(sql.ErrNoRows)

		req := httptest.NewRequest(http.MethodGet, "/world/44?session_id=55", nil)
		req.SetPathValue("game_id", "44")
		rr := httptest.NewRecorder()
		GetWorldHandler(rr, req, sqlxDB, store)

		if rr.Code != http.StatusForbidden {
			t.Fatalf("got status %d want %d body=%q", rr.Code, http.StatusForbidden, rr.Body.String())
		}
		if strings.TrimSpace(rr.Body.String()) != ErrorGamePlayerRequired {
			t.Fatalf("got body %q want %q", rr.Body.String(), ErrorGamePlayerRequired)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet sql expectations: %v", err)
		}
	})
}

func TestRetiredWorldEndpointDoesNotProbeSnapshotExistence(t *testing.T) {
	sqlxDB, mock := newGameSQLMock(t)
	store := newGameSessionStore(sqlxDB)

	expectReadableWorldAccess(mock, 55, 7, gameRowWithLobby(44, 7, nil, "started", "World Game", nil, false, false, DefaultGameStyle, defaultRulesJSON, defaultExclusionsJSON, 2, 250))
	req := httptest.NewRequest(http.MethodGet, "/world/44?session_id=55", nil)
	req.SetPathValue("game_id", "44")
	rr := httptest.NewRecorder()
	GetWorldHandler(rr, req, sqlxDB, store)

	if rr.Code != http.StatusGone {
		t.Fatalf("got status %d want %d body=%q", rr.Code, http.StatusGone, rr.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestWorldSpaceLayerPlaceholder(t *testing.T) {
	sqlxDB, mock := newGameSQLMock(t)
	store := newGameSessionStore(sqlxDB)

	expectReadableWorldAccess(mock, 55, 7, gameRowWithLobby(44, 7, nil, "started", "World Game", nil, false, false, DefaultGameStyle, defaultRulesJSON, defaultExclusionsJSON, 2, 250))

	req := httptest.NewRequest(http.MethodGet, "/world/44/space?session_id=55", nil)
	req.SetPathValue("game_id", "44")
	rr := httptest.NewRecorder()
	GetWorldSpaceLayerHandler(rr, req, sqlxDB, store)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("got status %d want %d body=%q", rr.Code, http.StatusNotFound, rr.Body.String())
	}
	if contentType := rr.Header().Get("Content-Type"); !strings.Contains(contentType, "application/json") {
		t.Fatalf("got content type %q want application/json", contentType)
	}
	if strings.TrimSpace(rr.Body.String()) != `{"error":"space layer is not supported for CTP2 worlds"}` {
		t.Fatalf("got body %q want %q", rr.Body.String(), `{"error":"space layer is not supported for CTP2 worlds"}`)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func expectReadableWorldAccess(mock sqlmock.Sqlmock, sessionID int64, accountID int, gameRows *sqlmock.Rows) {
	expectGameSessionLookup(mock, sessionID, accountID)
	mock.ExpectQuery(regexp.QuoteMeta(gameSelectByID)).
		WithArgs(int64(44)).
		WillReturnRows(gameRows)
	mock.ExpectQuery(regexp.QuoteMeta(gamePlayerByAccountQuery)).
		WithArgs(int64(44), int64(accountID)).
		WillReturnRows(gamePlayerRows(44,
			struct {
				id        int64
				accountID int64
				country   string
				ping      int
			}{id: 1, accountID: int64(accountID), country: "Greeks", ping: 0},
		))
}
