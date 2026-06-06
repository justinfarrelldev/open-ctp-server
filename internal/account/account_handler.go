package account

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

func handleAccountHandlerError(w *statusTrackingResponseWriter, err error) {
	if err == nil || w.wroteBody {
		return
	}

	if w.wroteHeader {
		_, _ = w.ResponseWriter.Write([]byte(err.Error()))
		return
	}

	http.Error(w.ResponseWriter, err.Error(), http.StatusInternalServerError)
}

func CreateAccountHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) {
	trackedWriter := &statusTrackingResponseWriter{ResponseWriter: w}
	_, err := CreateAccount(trackedWriter, r, db, store)
	handleAccountHandlerError(trackedWriter, err)
}

func GetAccountHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) {
	trackedWriter := &statusTrackingResponseWriter{ResponseWriter: w}
	err := GetAccount(trackedWriter, r, db, store)
	handleAccountHandlerError(trackedWriter, err)
}

func UpdateAccountHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) {
	trackedWriter := &statusTrackingResponseWriter{ResponseWriter: w}
	err := UpdateAccount(trackedWriter, r, db, store)
	handleAccountHandlerError(trackedWriter, err)
}

func DeleteAccountHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) {
	trackedWriter := &statusTrackingResponseWriter{ResponseWriter: w}
	err := DeleteAccount(trackedWriter, r, db, store)
	handleAccountHandlerError(trackedWriter, err)
}

func LoginAccountHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) {
	trackedWriter := &statusTrackingResponseWriter{ResponseWriter: w}
	err := LoginAccount(trackedWriter, r, db, store)
	handleAccountHandlerError(trackedWriter, err)
}

func ResetPasswordHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB, store *auth.SessionStore) {
	trackedWriter := &statusTrackingResponseWriter{ResponseWriter: w}
	err := ResetPassword(trackedWriter, r, db, store)
	handleAccountHandlerError(trackedWriter, err)
}
