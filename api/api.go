// api.go

package api

import (
    "encoding/json"
    "net/http"
)

// HealthCheckHandler handles health check requests.
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
    // Need to fill in more check logic once the router is more coded
    healthStatus := "healthy"
    message := "The proxy/router is operating normally"

    // Create a response JSON
    response := struct {
        Status  string `json:"status"`
        Message string `json:"message"`
    }{
        Status:  healthStatus,
        Message: message,
    }

    // Marshal the response into JSON
    responseJSON, err := json.Marshal(response)
    if err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    // Set the Content-Type header and write the response
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write(responseJSON)
}

// Endpoint1Handler handles requests for the first endpoint.
func Endpoint1Handler(w http.ResponseWriter, r *http.Request) {
    data := map[string]string{
        "message": "This is Endpoint 1",
    }

    responseJSON, err := json.Marshal(data)
    if err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    // Set header and write the response
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write(responseJSON)
}

// Endpoint2Handler handles requests for the second endpoint.
func Endpoint2Handler(w http.ResponseWriter, r *http.Request) {
    data := map[string]string{
        "message": "This is Endpoint 2",
    }

    responseJSON, err := json.Marshal(data)
    if err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    // Set header and write the response
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write(responseJSON)
}
