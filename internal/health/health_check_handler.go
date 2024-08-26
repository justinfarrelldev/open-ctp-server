package health

import "net/http"

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	if err := HealthCheck(w, r); err != nil {
		// Handle the error, e.g., log it and send an appropriate response to the client
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
