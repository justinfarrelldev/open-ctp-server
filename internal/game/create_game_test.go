package game

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func newGameSQLMock(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return sqlx.NewDb(db, "sqlmock"), mock
}

func TestCreateGame_AppliesCTP2DefaultsWhenOmitted(t *testing.T) {
	sqlxDB, mock := newGameSQLMock(t)

	// Provenance for these defaults:
	// - ctp2/ctp2_source_code/ctp2_code/gs/utility/gameinit.cpp uses one settler for setup slots.
	// - ctp2/ctp2_source_code/ctp2_code/gs/gameobj/Player.cpp reads starting gold from difficulty data.
	// - ctp2/ctp2_released_game/ctp2_data/default/gamedata/DiffDB.txt keeps human starting gold at 100.
	// - ctp2/ctp2_source_code/ctp2_code/ui/netshell/ns_gamesetup.cpp zero-initializes exclusions.
	body := map[string]any{
		"name":               "ctp2 default setup",
		"password_protected": false,
		"password":           "",
	}
	jsonBody, _ := json.Marshal(body)

	mock.ExpectQuery(regexp.QuoteMeta(insertGameQuery)).
		WithArgs(
			"ctp2 default setup",
			nil,
			nil,
			false,
			"",
			false,
			false,
			DefaultGameStyle,
			defaultRulesJSON,
			defaultExclusionsJSON,
			DefaultGameStatus,
			DefaultStartingSettlers,
			DefaultStartingGold,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(44))

	req := httptest.NewRequest(http.MethodPost, "/game/create_game", bytes.NewBuffer(jsonBody))
	rr := httptest.NewRecorder()

	GameHandler(rr, req, sqlxDB)

	if rr.Code != http.StatusCreated {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusCreated)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestCreateGame_RejectsNegativeStartingSettlers(t *testing.T) {
	sqlxDB, mock := newGameSQLMock(t)
	settlers := -1
	body := CreateGameArgs{StartingSettlers: &settlers}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/game/create_game", bytes.NewBuffer(jsonBody))
	rr := httptest.NewRecorder()

	GameHandler(rr, req, sqlxDB)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusBadRequest)
	}
	if rr.Body.String() != ErrorNegativeSettlers {
		t.Fatalf("got body %q want %q", rr.Body.String(), ErrorNegativeSettlers)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected sql execution: %v", err)
	}
}

func TestCreateGame_RejectsNegativeStartingGold(t *testing.T) {
	sqlxDB, mock := newGameSQLMock(t)
	gold := -1
	body := CreateGameArgs{StartingGold: &gold}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/game/create_game", bytes.NewBuffer(jsonBody))
	rr := httptest.NewRecorder()

	GameHandler(rr, req, sqlxDB)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusBadRequest)
	}
	if rr.Body.String() != ErrorNegativeGold {
		t.Fatalf("got body %q want %q", rr.Body.String(), ErrorNegativeGold)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected sql execution: %v", err)
	}
}

func TestCreateGame_RejectsInvalidRulesJSONValue(t *testing.T) {
	sqlxDB, mock := newGameSQLMock(t)
	jsonBody := []byte(`{"rules_json":"not-json-array"}`)

	req := httptest.NewRequest(http.MethodPost, "/game/create_game", bytes.NewBuffer(jsonBody))
	rr := httptest.NewRecorder()

	GameHandler(rr, req, sqlxDB)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusBadRequest)
	}
	if rr.Body.String() != ErrorRulesJSONInvalid {
		t.Fatalf("got body %q want %q", rr.Body.String(), ErrorRulesJSONInvalid)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected sql execution: %v", err)
	}
}

func TestCreateGame_RejectsRulesJSONThatIsNotAnArray(t *testing.T) {
	sqlxDB, mock := newGameSQLMock(t)
	jsonBody := []byte(`{"rules_json":{"pollution":true}}`)

	req := httptest.NewRequest(http.MethodPost, "/game/create_game", bytes.NewBuffer(jsonBody))
	rr := httptest.NewRecorder()

	GameHandler(rr, req, sqlxDB)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusBadRequest)
	}
	if rr.Body.String() != ErrorRulesJSONInvalid {
		t.Fatalf("got body %q want %q", rr.Body.String(), ErrorRulesJSONInvalid)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected sql execution: %v", err)
	}
}

func TestCreateGame_RejectsInvalidExclusionsJSONValue(t *testing.T) {
	sqlxDB, mock := newGameSQLMock(t)
	jsonBody := []byte(`{"exclusions_json":"not-json-object"}`)

	req := httptest.NewRequest(http.MethodPost, "/game/create_game", bytes.NewBuffer(jsonBody))
	rr := httptest.NewRecorder()

	GameHandler(rr, req, sqlxDB)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusBadRequest)
	}
	if rr.Body.String() != ErrorExclusionsJSONInvalid {
		t.Fatalf("got body %q want %q", rr.Body.String(), ErrorExclusionsJSONInvalid)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected sql execution: %v", err)
	}
}

func TestCreateGame_RejectsExclusionsJSONThatIsNotAnObject(t *testing.T) {
	sqlxDB, mock := newGameSQLMock(t)
	jsonBody := []byte(`{"exclusions_json":["wonder_a"]}`)

	req := httptest.NewRequest(http.MethodPost, "/game/create_game", bytes.NewBuffer(jsonBody))
	rr := httptest.NewRecorder()

	GameHandler(rr, req, sqlxDB)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusBadRequest)
	}
	if rr.Body.String() != ErrorExclusionsJSONInvalid {
		t.Fatalf("got body %q want %q", rr.Body.String(), ErrorExclusionsJSONInvalid)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected sql execution: %v", err)
	}
}

func TestCreateGame_PasswordValidation(t *testing.T) {
	tests := []struct {
		name       string
		body       CreateGameArgs
		wantStatus int
		wantBody   string
	}{
		{
			name:       "missing password when protected",
			body:       CreateGameArgs{PasswordProtected: true},
			wantStatus: http.StatusBadRequest,
			wantBody:   ErrorPasswordRequired,
		},
		{
			name:       "short password when protected",
			body:       CreateGameArgs{PasswordProtected: true, Password: "123"},
			wantStatus: http.StatusBadRequest,
			wantBody:   ErrorPasswordTooShort,
		},
		{
			name:       "password present when not protected",
			body:       CreateGameArgs{PasswordProtected: false, Password: "password123"},
			wantStatus: http.StatusBadRequest,
			wantBody:   ErrorPasswordUnexpected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sqlxDB, mock := newGameSQLMock(t)
			jsonBody, _ := json.Marshal(tt.body)

			req := httptest.NewRequest(http.MethodPost, "/game/create_game", bytes.NewBuffer(jsonBody))
			rr := httptest.NewRecorder()

			GameHandler(rr, req, sqlxDB)

			if rr.Code != tt.wantStatus {
				t.Fatalf("got status %d want %d", rr.Code, tt.wantStatus)
			}
			if rr.Body.String() != tt.wantBody {
				t.Fatalf("got body %q want %q", rr.Body.String(), tt.wantBody)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("unexpected sql execution: %v", err)
			}
		})
	}
}

func TestCreateGame_InvalidMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/game/create_game", nil)
	rr := httptest.NewRecorder()

	GameHandler(rr, req, nil)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusBadRequest)
	}
	if rr.Body.String() != "invalid request; request must be a POST request" {
		t.Fatalf("got body %q", rr.Body.String())
	}
}

func TestCreateGame_DecodeError(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/game/create_game", bytes.NewBufferString(`{"starting_gold":"oops"}`))
	rr := httptest.NewRecorder()

	GameHandler(rr, req, nil)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("got status %d want %d", rr.Code, http.StatusBadRequest)
	}
}
