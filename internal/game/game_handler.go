package game

import (
	"database/sql"
	"net/http"
)

func GameHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if err := CreateGame(w, r, db); err != nil {
		// Handle the error, e.g., log it and send an appropriate response to the client
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
