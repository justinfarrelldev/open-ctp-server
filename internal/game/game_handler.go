package game

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	auth "github.com/justinfarrelldev/open-ctp-server/internal/auth"
)

type statusTrackingResponseWriter struct {
	http.ResponseWriter
	statusCode  int
	wroteHeader bool
	wroteBody   bool
}

func (w *statusTrackingResponseWriter) WriteHeader(statusCode int) {
	if !w.wroteHeader {
		w.statusCode = statusCode
		w.wroteHeader = true
	}
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *statusTrackingResponseWriter) Write(body []byte) (int, error) {
	if !w.wroteHeader {
		w.statusCode = http.StatusOK
		w.wroteHeader = true
	}
	w.wroteBody = true
	return w.ResponseWriter.Write(body)
}

func handleGameHandlerError(w *statusTrackingResponseWriter, err error) {
	if err == nil || w.wroteBody {
		return
	}

	if w.wroteHeader {
		_, _ = w.ResponseWriter.Write([]byte(err.Error()))
		return
	}

	http.Error(w.ResponseWriter, err.Error(), http.StatusInternalServerError)
}

func writeGameStatusError(w http.ResponseWriter, status int, message string) error {
	w.WriteHeader(status)
	return &statusError{message: message}
}

type statusError struct {
	message string
}

func (e *statusError) Error() string {
	return e.message
}

func CreateGameHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) {
	trackedWriter := &statusTrackingResponseWriter{ResponseWriter: w}
	err := CreateGame(trackedWriter, r, db, store)
	handleGameHandlerError(trackedWriter, err)
}

func GetGameHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) {
	trackedWriter := &statusTrackingResponseWriter{ResponseWriter: w}
	err := GetGame(trackedWriter, r, db, store)
	handleGameHandlerError(trackedWriter, err)
}

func UpdateGameHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) {
	trackedWriter := &statusTrackingResponseWriter{ResponseWriter: w}
	err := UpdateGame(trackedWriter, r, db, store)
	handleGameHandlerError(trackedWriter, err)
}

func DeleteGameHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) {
	trackedWriter := &statusTrackingResponseWriter{ResponseWriter: w}
	err := DeleteGame(trackedWriter, r, db, store)
	handleGameHandlerError(trackedWriter, err)
}

func RenameGameHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) {
	trackedWriter := &statusTrackingResponseWriter{ResponseWriter: w}
	err := RenameGame(trackedWriter, r, db, store)
	handleGameHandlerError(trackedWriter, err)
}

func LockGameHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) {
	trackedWriter := &statusTrackingResponseWriter{ResponseWriter: w}
	err := LockGame(trackedWriter, r, db, store)
	handleGameHandlerError(trackedWriter, err)
}

func SetGamePublicHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) {
	trackedWriter := &statusTrackingResponseWriter{ResponseWriter: w}
	err := SetPublic(trackedWriter, r, db, store)
	handleGameHandlerError(trackedWriter, err)
}

func UpdateGameStyleHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) {
	trackedWriter := &statusTrackingResponseWriter{ResponseWriter: w}
	err := UpdateStyle(trackedWriter, r, db, store)
	handleGameHandlerError(trackedWriter, err)
}

func UpdateGameRulesHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) {
	trackedWriter := &statusTrackingResponseWriter{ResponseWriter: w}
	err := UpdateRules(trackedWriter, r, db, store)
	handleGameHandlerError(trackedWriter, err)
}

func UpdateGameExclusionsHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) {
	trackedWriter := &statusTrackingResponseWriter{ResponseWriter: w}
	err := UpdateExclusions(trackedWriter, r, db, store)
	handleGameHandlerError(trackedWriter, err)
}

func AddAIHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) {
	trackedWriter := &statusTrackingResponseWriter{ResponseWriter: w}
	err := AddAI(trackedWriter, r, db, store)
	handleGameHandlerError(trackedWriter, err)
}
