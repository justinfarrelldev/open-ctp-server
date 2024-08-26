package game

import "net/http"

func GameHandler(w http.ResponseWriter, r *http.Request) {
	if err := CreateGame(w, r); err != nil {
		// Handle the error, e.g., log it and send an appropriate response to the client
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
