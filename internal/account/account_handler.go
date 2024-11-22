package account

import (
	"net/http"

	"github.com/jmoiron/sqlx"
)

func CreateAccountHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	if err := CreateAccount(w, r, db); err != nil {
		// Handle the error, e.g., log it and send an appropriate response to the client
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func GetAccountHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	if err := GetAccount(w, r, db); err != nil {
		// Handle the error, e.g., log it and send an appropriate response to the client
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func UpdateAccountHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	if err := UpdateAccount(w, r, db); err != nil {
		// Handle the error, e.g., log it and send an appropriate response to the client
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
