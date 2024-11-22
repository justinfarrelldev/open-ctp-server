package game

import (
	"net/http"

	"github.com/jmoiron/sqlx"
)

func GameHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	if err := CreateGame(w, r, db); err != nil {
		// Handle the error, e.g., log it and send an appropriate response to the client
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
