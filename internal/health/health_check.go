package health

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Response struct {
	Status string `json:"status" example:"OK"`
}

// HealthCheck handles the health check request.
//
// @Summary Health check endpoint
// @Description Returns the status of the service.
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} Response
// @Failure 500 {object} error
// @Router /health [get]
func HealthCheck(w http.ResponseWriter, r *http.Request) error {
	fmt.Println("got health check request")
	resp := Response{Status: "OK"}
	jsonBody, _ := json.Marshal(resp)

	_, err := w.Write(jsonBody)
	return err
}
