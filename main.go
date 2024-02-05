package main

import (
	api "Capstone-4901---SeaTeam/api"
	config "Capstone-4901---SeaTeam/config"
	"Capstone-4901---SeaTeam/loadbalancer"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
)

func main() {
	// Initial config load
	configuration, backendServers := loadConfig()

	lbPolicy := configuration.StaticResources.Clusters[0].LbPolicy

	// Grabbing all endpoints from the config
	var backendAddresses []string
	for _, server := range backendServers {
		backendAddresses = append(backendAddresses, fmt.Sprintf("http://%s:%d/", server.Address, server.Port))
	}

	// Create loadbalancer with grabbed endpoints
	var loadBalancer loadbalancer.LoadBalancer
	switch lbPolicy {
	case "ROUND_ROBIN":
		loadBalancer = loadbalancer.NewRoundRobinLoadBalancer(backendAddresses)
	case "LEAST_CONNECTIONS":
		loadBalancer = loadbalancer.NewLeastConnectionsLoadBalancer(backendAddresses)
	default:
		// Default to Round Robin if lbPolicy is not recognized
		loadBalancer = loadbalancer.NewRoundRobinLoadBalancer(backendAddresses)
	}

	r := &Router{
		Timeout:      10 * time.Second, // Example timeout value
		Config:       configuration,
		LoadBalancer: loadBalancer,
		ErrorLogger:  log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile),
	}

	// Watch for changes in the config file
	go watchConfigFile("config/static.yaml", func() {
		configuration, backendServers = loadConfig()
		lbPolicy = configuration.StaticResources.Clusters[0].LbPolicy
		var updatedBackendAddresses []string
		for _, server := range backendServers {
			updatedBackendAddresses = append(updatedBackendAddresses, fmt.Sprintf("http://%s:%d/", server.Address, server.Port))
		}

		var updatedLoadBalancer loadbalancer.LoadBalancer
		switch lbPolicy {
		case "ROUND_ROBIN":
			updatedLoadBalancer = loadbalancer.NewRoundRobinLoadBalancer(updatedBackendAddresses)
		case "LEAST_CONNECTIONS":
			updatedLoadBalancer = loadbalancer.NewLeastConnectionsLoadBalancer(updatedBackendAddresses)
		default:
			updatedLoadBalancer = loadbalancer.NewRoundRobinLoadBalancer(updatedBackendAddresses)
		}

		r.LoadBalancer = updatedLoadBalancer
		r.Config = configuration
	})

	// Serve the API endpoints
	http.Handle("/", r)
	http.HandleFunc("/health", api.HealthCheckHandler)
	http.HandleFunc("/endpoint1", api.Endpoint1Handler)
	http.HandleFunc("/endpoint2", api.Endpoint2Handler)

	// Serve frontend files
	fs := http.FileServer(http.Dir("./frontend"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Println("Server started on :8000")
	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	// Wait for termination signals
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)
	<-signalChannel
	log.Println("Shutting down the server...")
}

func loadConfig() (config.StaticBootstrap, []config.BackendServer) {
	// Load configuration from file
	configuration, servers := config.GetYAMLdata()
	// Update existing structs using atomic operations or mutexes
	// ...
	fmt.Println("Config reloaded successfully")
	return configuration, servers
}

func watchConfigFile(filePath string, reloadFunc func()) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Error creating file watcher: %v", err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("Config file modified. Reloading...")
					reloadFunc()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("Error watching config file: %v", err)
			}
		}
	}()

	err = watcher.Add(filePath)
	if err != nil {
		log.Fatalf("Error adding config file to watcher: %v", err)
	}

	<-done // This will block until the watcher is closed (which never happens in this code)
}
