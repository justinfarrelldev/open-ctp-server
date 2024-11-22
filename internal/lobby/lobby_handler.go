package lobby

import (
	"net/http"

	"github.com/jmoiron/sqlx"
)

func CreateLobbyHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	if err := CreateLobby(w, r, db); err != nil {
		// Handle the error, e.g., log it and send an appropriate response to the client
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func GetLobbyHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	if err := GetLobby(w, r, db); err != nil {
		// Handle the error, e.g., log it and send an appropriate response to the client
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// func UpdateLobbyHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
// 	if err := UpdateLobby(w, r, db); err != nil {
// 		// Handle the error, e.g., log it and send an appropriate response to the client
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// }
