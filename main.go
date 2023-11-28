package main

import (
	api "Capstone-4901---SeaTeam/api"
	config "Capstone-4901---SeaTeam/config"
	"fmt"
	"net/http"
	"time"
)

func main() {
	configuration := config.GetYAMLdata()
	r := &Router{
		Timeout:      10 * time.Second,
		Config:       configuration,
		LoadBalancer: nil, // LoadBalancer can be initialized here if needed
		ErrorLogger:  nil, // ErrorLogger can be initialized here if needed
	}

	http.Handle("/", r)
	http.HandleFunc("/health", api.HealthCheckHandler)
	http.HandleFunc("/endpoint1", api.Endpoint1Handler)
	http.HandleFunc("/endpoint2", api.Endpoint2Handler)

	fs := http.FileServer(http.Dir("./frontend"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	fmt.Println("Server started on :8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
