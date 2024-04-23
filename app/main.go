package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	// Import the necessary OpenTelemetry packages
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	api "seateam/api"
	config "seateam/config"
	loadbalancer "seateam/loadbalancer"
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
	exporter, err := zipkin.New("http://zipkin:9411/api/v2/spans")
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

var (
	// Counter for each endpoint
	endpointRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "endpoint_requests_total",
			Help: "Total number of requests to each endpoint.",
		},
		[]string{"endpoint"},
	)

	// Total number of endpoints
	totalEndpoints = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "total_endpoints",
			Help: "Total number of endpoints.",
		},
	)

	totalRequests = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "total_requests",
			Help: "Total number of requests across all endpoints.",
		},
	)

	// Mutex to ensure safe access to connectionCount
	mu sync.Mutex
)

func init() {
	prometheus.MustRegister(endpointRequests)
	prometheus.MustRegister(totalEndpoints)
	prometheus.MustRegister(totalRequests)
}

func incrementTotalEndpoints(count int) {
	mu.Lock()
	defer mu.Unlock()
	totalEndpoints.Set(float64(count))
}

func main() {

	initTracing()
	// Handler for the first endpoint (listening on port 1234)
	mux1 := http.NewServeMux()
	mux1.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		endpointRequests.WithLabelValues("endpoint_1").Inc()
		totalRequests.Inc()
		html := `
			<!DOCTYPE html>
			<html lang="en">
			<head>
				<meta charset="UTF-8">
				<meta name="viewport" content="width=device-width, initial-scale=1.0">
				<title>Envoy Router</title>
				<link rel="stylesheet" type="text/css" href="styles.css">
				<style>
					body {
						font-family: Arial, sans-serif;
						background-color: #f0f0f0;
						color: #333;
						margin: 20px;
					}

					h1 {
						color: #3498db;
					}

					button {
						display: block;
						padding: 15px 40px;
						color: #fff;
						text-decoration: none;
						background-color: #3498db;
						border-radius: 8px;
						transition: background-color 0.3s, color 0.3s;
						border: none; /* Remove default button border */
						cursor: pointer;
					}

					button:hover {
						background-color: #267bb5;
					}
				</style>
			</head>
			<body>
				<h1>Response from mockbackend 1 - Root Path</h1>
				<button onclick="sendNewRequest()">New Request</button>

				<script>
					function sendNewRequest() {
						// You can customize this URL based on your requirements
						window.location.href = "/?endpoint=lb";
					}
				</script>
			</body>
			</html>
		`
		fmt.Fprint(w, html)
	})
	mux1.HandleFunc("/service1", func(w http.ResponseWriter, r *http.Request) {
		html := `
			<!DOCTYPE html>
			<html lang="en">
			<head>
				<meta charset="UTF-8">
				<meta name="viewport" content="width=device-width, initial-scale=1.0">
				<title>Envoy Router</title>
				<link rel="stylesheet" type="text/css" href="styles.css">
			</head>
			<body>
				<h1>Response from mockbackend 1 - Service1</h1>
				<button onclick="sendNewRequest()">New Request</button>

				<script>
					function sendNewRequest() {
						// You can customize this URL based on your requirements
						window.location.href = "/?endpoint=lb";
					}
				</script>
			</body>
			</html>
		`
		fmt.Fprint(w, html)
	})

	// Start the server on port 1234
	go func() {
		http.ListenAndServe(":1234", mux1)
	}()

	// Handler for the second endpoint (listening on port 5678)
	mux2 := http.NewServeMux()
	mux2.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		endpointRequests.WithLabelValues("endpoint_2").Inc()
		totalRequests.Inc()
		html := `
			<!DOCTYPE html>
			<html lang="en">
			<head>
				<meta charset="UTF-8">
				<meta name="viewport" content="width=device-width, initial-scale=1.0">
				<title>Envoy Router</title>
				<link rel="stylesheet" type="text/css" href="styles.css">
				<style>
					body {
						font-family: Arial, sans-serif;
						background-color: #f0f0f0;
						color: #333;
						margin: 20px;
					}

					h1 {
						color: #3498db;
					}

					button {
						display: block;
						padding: 15px 40px;
						color: #fff;
						text-decoration: none;
						background-color: #00FF00;
						border-radius: 8px;
						transition: background-color 0.3s, color 0.3s;
						border: none; /* Remove default button border */
						cursor: pointer;
					}

					button:hover {
						background-color: #008000;
					}
				</style>
			</head>
			<body>
				<h1>Response from mockbackend 2 - Root Path</h1>
				<button onclick="sendNewRequest()">New Request</button>

				<script>
					function sendNewRequest() {
						// You can customize this URL based on your requirements
						window.location.href = "/?endpoint=lb";
					}
				</script>
			</body>
			</html>
		`
		fmt.Fprint(w, html)
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
		endpointRequests.WithLabelValues("endpoint_3").Inc()
		totalRequests.Inc()
		html := `
			<!DOCTYPE html>
			<html lang="en">
			<head>
				<meta charset="UTF-8">
				<meta name="viewport" content="width=device-width, initial-scale=1.0">
				<title>Envoy Router</title>
				<link rel="stylesheet" type="text/css" href="styles.css">
				<style>
					body {
						font-family: Arial, sans-serif;
						background-color: #f0f0f0;
						color: #333;
						margin: 20px;
					}

					h1 {
						color: #3498db;
					}

					button {
						display: block;
						padding: 15px 40px;
						color: #fff;
						text-decoration: none;
						background-color: #FF0000;
						border-radius: 8px;
						transition: background-color 0.3s, color 0.3s;
						border: none; /* Remove default button border */
						cursor: pointer;
					}

					button:hover {
						background-color: #CC0000;
					}
				</style>
			</head>
			<body>
				<h1>Response from mockbackend 3 - Root Path</h1>
				<button onclick="sendNewRequest()">New Request</button>

				<script>
					function sendNewRequest() {
						// You can customize this URL based on your requirements
						window.location.href = "/?endpoint=lb";
					}
				</script>
			</body>
			</html>
		`
		fmt.Fprint(w, html)
	})
	mux3.HandleFunc("/service3", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Response from mockbackend 3 - Service3")
	})

	// Start the server on port 5678
	go func() {
		http.ListenAndServe(":9876", mux3)
	}()

	//Handler for the fourth endpoint (listening on port 5544)
	mux4 := http.NewServeMux()
	mux4.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		endpointRequests.WithLabelValues("endpoint_4").Inc()
		totalRequests.Inc()
		html := `
			<!DOCTYPE html>
			<html lang="en">
			<head>
				<meta charset="UTF-8">
				<meta name="viewport" content="width=device-width, initial-scale=1.0">
				<title>Envoy Router</title>
				<link rel="stylesheet" type="text/css" href="styles.css">
				<style>
					body {
						font-family: Arial, sans-serif;
						background-color: #f0f0f0;
						color: #333;
						margin: 20px;
					}

					h1 {
						color: #3498db;
					}

					button {
						display: block;
						padding: 15px 40px;
						color: #fff;
						text-decoration: none;
						background-color: #7851A9;
						border-radius: 8px;
						transition: background-color 0.3s, color 0.3s;
						border: none; /* Remove default button border */
						cursor: pointer;
					}

					button:hover {
						background-color: #663399;
					}
				</style>
			</head>
			<body>
				<h1>Response from mockbackend 4 - Root Path</h1>
				<button onclick="sendNewRequest()">New Request</button>

				<script>
					function sendNewRequest() {
						// You can customize this URL based on your requirements
						window.location.href = "/?endpoint=lb";
					}
				</script>
			</body>
			</html>
		`
		fmt.Fprint(w, html)
	})
	mux4.HandleFunc("/service4", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Response from mockbackend 4 - Service4")
	})

	// Start the server on port 5544
	go func() {
		http.ListenAndServe(":5544", mux4)
	}()

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":8081", nil)
	}()

	incrementTotalEndpoints(4)

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
	fmt.Println("Configuration loaded successfully:", configuration)
	fmt.Println("Backend servers:", servers)
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
