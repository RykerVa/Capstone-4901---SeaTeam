package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Response from the Root Path")
	})

	http.HandleFunc("/service1", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Response from Service 1")
	})

	http.HandleFunc("/service2", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Response from Service 2")
	})

	http.ListenAndServe(":1234", nil)
}
