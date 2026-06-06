package game

import (
	"net/http"

	"github.com/jmoiron/sqlx"
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

func GameHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	trackedWriter := &statusTrackingResponseWriter{ResponseWriter: w}
	err := CreateGame(trackedWriter, r, db)
	handleGameHandlerError(trackedWriter, err)
}
