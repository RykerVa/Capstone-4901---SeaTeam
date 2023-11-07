package main

import (
	"encoding/json"
	"net/http"
)

func main() {
	// Start the HTTP server
	http.HandleFunc("/health", healthCheckHandler)
	http.HandleFunc("/endpoint1", endpoint1Handler)
	http.HandleFunc("/endpoint2", endpoint2handler)
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

//for right now these are just basic endpoints and we can implement specific logic in the future.
func endpoint1Handler(w http.ResponseWriter, r *http.Request){
	
	data := map[string]string{
		"message": "This is Endpoint 1",
	}

	responseJSON, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	//set header and write the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
}

//for right now these are just basic endpoints and we can implement specific logic in the future.
func endpoint2handler(w http.ResponseWriter, r *http.Request){
	
	data := map[string]string{
		"message": "This is Endpoint 2",
	}

	responseJSON, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	//set header and write the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
}
