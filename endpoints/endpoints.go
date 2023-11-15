package main

import (
	"fmt"
	"net/http"
)

func main() {
	// Handler for the first endpoint (listening on port 1234)
	mux1 := http.NewServeMux()
	mux1.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Response from mockbackend 1 - Root Path")
	})
	mux1.HandleFunc("/service1", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Response from mockbackend 1 - Service1")
	})

	// Start the server on port 1234
	go func() {
		http.ListenAndServe(":1234", mux1)
	}()

	// Handler for the second endpoint (listening on port 5678)
	mux2 := http.NewServeMux()
	mux2.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Response from mockbackend 2 - Root Path")
	})
	mux2.HandleFunc("/service2", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Response from mockbackend 2 - Service2")
	})

	// Start the server on port 5678
	go func() {
		http.ListenAndServe(":5678", mux2)
	}()

	//Handler for the third endpoint (listening on port 9876)
	mux3 := http.NewServeMux()
	mux3.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Response from mockbackend 3 - Root Path")
	})
	mux3.HandleFunc("/service3", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Response from mockbackend 3 - Service3")
	})

	// Start the server on port 5678
	go func() {
		http.ListenAndServe(":9876", mux3)
	}()

	// Add more handlers for additional endpoints...
	// Repeat the pattern with different ports and handlers as needed.

	select {}
}
