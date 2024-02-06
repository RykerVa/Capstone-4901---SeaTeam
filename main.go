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

	// Import the necessary OpenTelemetry packages
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func traceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tracer := otel.Tracer("myApp")
		ctx, span := tracer.Start(r.Context(), "http.request")
		// Set attributes after starting the span
		span.SetAttributes(
			attribute.String("method", r.Method),
			attribute.String("path", r.URL.Path),
		)
		defer span.End()

		// Continue with the request
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func initTracing() {
	// Configure the Zipkin exporter to send traces to a Zipkin backend
	exporter, err := zipkin.New("http://localhost:9411/api/v2/spans")
	if err != nil {
		log.Fatalf("Failed to create Zipkin exporter: %v", err)
	}

	// Create a new tracer provider with a batch span processor and the Zipkin exporter
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		// Add a resource to the tracer provider to identify this application
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("SeaTeamSVC"),
		)),
	)

	// Set the created tracer provider as the global provider
	otel.SetTracerProvider(tp)
}

func main() {

	initTracing()
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
	http.Handle("/", traceMiddleware(r))
	http.Handle("/health", traceMiddleware(http.HandlerFunc(api.HealthCheckHandler)))
	http.Handle("/endpoint1", traceMiddleware(http.HandlerFunc(api.Endpoint1Handler)))
	http.Handle("/endpoint2", traceMiddleware(http.HandlerFunc(api.Endpoint2Handler)))

	// Serve frontend files
	fs := http.FileServer(http.Dir("./frontend"))
	http.Handle("/static/", http.StripPrefix("/static/", traceMiddleware(fs)))

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
