package main

import (
	"encoding/json"
	"net/http"
)

func main() {
	// Start the HTTP server
	http.HandleFunc("/health", healthCheckHandler)
	http.ListenAndServe(":8080", nil)
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	
	//Need to fill in more check logic once the router is more coded
	healthStatus := "healthy"
	message := "The proxy/router is operating normally."

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
