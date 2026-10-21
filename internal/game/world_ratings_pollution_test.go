package game

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/justinfarrelldev/open-ctp-server/internal/auth"
)

// Legacy aggregates count undiscovered terrain (recalculateRatings uses all land;
// recalculatePollution weights all terrain). They are not viewer-safe statistics.
func TestLegacyWorldAggregatesDoNotBypassViewerProjection(t *testing.T) {
	for _, endpoint := range []struct {
		suffix  string
		handler func(http.ResponseWriter, *http.Request, *sqlx.DB, *auth.SessionStore)
	}{{"ratings", GetWorldRatingsHandler}, {"pollution", GetWorldPollutionHandler}} {
		t.Run(endpoint.suffix, func(t *testing.T) {
			for _, scenario := range []string{"joined", "setup", "nonmember", "unauthenticated", "invalid-method"} {
				t.Run(scenario, func(t *testing.T) {
					db, mock := newGameSQLMock(t)
					store := newGameSessionStore(db)
					status := http.StatusGone
					method := http.MethodGet
					query := "?session_id=55"
					switch scenario {
					case "joined", "setup":
						gameStatus := "started"
						if scenario == "setup" {
							gameStatus = DefaultGameStatus
							status = http.StatusConflict
						}
						expectReadableWorldAccess(mock, 55, 7, gameRowWithLobby(44, 7, nil, gameStatus, "World Game", nil, false, false, DefaultGameStyle, defaultRulesJSON, defaultExclusionsJSON, 2, 250))
					case "nonmember":
						status = http.StatusForbidden
						expectGameSessionLookup(mock, 55, 7)
						mock.ExpectQuery(regexp.QuoteMeta(gameSelectByID)).WithArgs(int64(44)).WillReturnRows(gameRowWithLobby(44, 7, nil, "started", "World Game", nil, false, false, DefaultGameStyle, defaultRulesJSON, defaultExclusionsJSON, 2, 250))
						mock.ExpectQuery(regexp.QuoteMeta(gamePlayerByAccountQuery)).WithArgs(int64(44), int64(7)).WillReturnError(sql.ErrNoRows)
					case "unauthenticated":
						status = http.StatusUnauthorized
						query = ""
					case "invalid-method":
						status = http.StatusBadRequest
						method = http.MethodPost
					}
					request := httptest.NewRequest(method, "/world/44/"+endpoint.suffix+query, nil)
					request.SetPathValue("game_id", "44")
					response := httptest.NewRecorder()
					endpoint.handler(response, request, db, store)
					if response.Code != status {
						t.Fatalf("status=%d want=%d body=%s", response.Code, status, response.Body.String())
					}
					if scenario == "joined" && !strings.Contains(response.Body.String(), "viewer-safe") {
						t.Fatal("retired aggregate needs a safe migration path")
					}
					for _, privateField := range []string{`"economic"`, `"global_pollution"`, `"tile_pollution_count"`} {
						if strings.Contains(response.Body.String(), privateField) {
							t.Fatal("hidden-world aggregate leaked")
						}
					}
					if err := mock.ExpectationsWereMet(); err != nil {
						t.Fatal(err)
					}
				})
			}
		})
	}
}
