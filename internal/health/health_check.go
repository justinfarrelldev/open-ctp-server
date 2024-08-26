package health

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Response struct {
	Status string `json:"status"`
}

func HealthCheck(w http.ResponseWriter, r *http.Request) error {
	fmt.Println("got health check request")
	resp := Response{Status: "OK"}
	jsonBody, _ := json.Marshal(resp)

	_, err := w.Write(jsonBody)
	return err
}
