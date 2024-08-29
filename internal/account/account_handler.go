package account

import (
	"database/sql"
	"net/http"
)

func AccountHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if err := CreateAccount(w, r, db); err != nil {
		// Handle the error, e.g., log it and send an appropriate response to the client
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
