package lobby

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/justinfarrelldev/open-ctp-server/internal/auth"
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

func (w *statusTrackingResponseWriter) Flush() {
	flusher, ok := w.ResponseWriter.(http.Flusher)
	if !ok {
		return
	}
	if !w.wroteHeader {
		w.statusCode = http.StatusOK
		w.wroteHeader = true
	}
	flusher.Flush()
}

func handleLobbyHandlerError(w *statusTrackingResponseWriter, err error) {
	if err == nil || w.wroteBody {
		return
	}

	if w.wroteHeader {
		_, _ = w.ResponseWriter.Write([]byte(err.Error()))
		return
	}

	http.Error(w.ResponseWriter, err.Error(), http.StatusInternalServerError)
}

func writeLobbyStatusError(w http.ResponseWriter, status int, message string) error {
	w.WriteHeader(status)
	return &statusError{message: message}
}

type statusError struct {
	message string
}

func (e *statusError) Error() string {
	return e.message
}

func CreateLobbyHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) {
	trackedWriter := &statusTrackingResponseWriter{ResponseWriter: w}
	err := CreateLobby(trackedWriter, r, db, store)
	handleLobbyHandlerError(trackedWriter, err)
}

func GetLobbyHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) {
	trackedWriter := &statusTrackingResponseWriter{ResponseWriter: w}
	err := GetLobby(trackedWriter, r, db, store)
	handleLobbyHandlerError(trackedWriter, err)
}

func UpdateLobbyHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) {
	trackedWriter := &statusTrackingResponseWriter{ResponseWriter: w}
	err := UpdateLobby(trackedWriter, r, db, store)
	handleLobbyHandlerError(trackedWriter, err)
}

func DeleteLobbyHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) {
	trackedWriter := &statusTrackingResponseWriter{ResponseWriter: w}
	err := DeleteLobby(trackedWriter, r, db, store)
	handleLobbyHandlerError(trackedWriter, err)
}

func JoinLobbyHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) {
	trackedWriter := &statusTrackingResponseWriter{ResponseWriter: w}
	err := JoinLobby(trackedWriter, r, db, store)
	handleLobbyHandlerError(trackedWriter, err)
}

func LeaveLobbyHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) {
	trackedWriter := &statusTrackingResponseWriter{ResponseWriter: w}
	err := LeaveLobby(trackedWriter, r, db, store)
	handleLobbyHandlerError(trackedWriter, err)
}

func ConnectLobbyHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) {
	trackedWriter := &statusTrackingResponseWriter{ResponseWriter: w}
	err := ConnectLobby(trackedWriter, r, db, store)
	handleLobbyHandlerError(trackedWriter, err)
}

func SendChatHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) {
	trackedWriter := &statusTrackingResponseWriter{ResponseWriter: w}
	err := SendChat(trackedWriter, r, db, store)
	handleLobbyHandlerError(trackedWriter, err)
}
