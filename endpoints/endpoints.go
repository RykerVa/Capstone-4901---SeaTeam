package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

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

	// Mutex to ensure safe access to connectionCount
	mu sync.Mutex
	// Total number of connections
	connectionCount = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "total_connections",
			Help: "Total number of connections across all endpoints.",
		},
	)
)

func CustomPrometheusHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a filtered registry
		filteredRegistry := prometheus.NewRegistry()

		// Register only the metrics you created
		filteredRegistry.MustRegister(endpointRequests)
		filteredRegistry.MustRegister(totalEndpoints)
		filteredRegistry.MustRegister(connectionCount)

		// Serve metrics from the filtered registry
		h := promhttp.HandlerFor(filteredRegistry, promhttp.HandlerOpts{})
		h.ServeHTTP(w, r)
	})
}

func incrementTotalEndpoints(count int) {
	mu.Lock()
	defer mu.Unlock()
	totalEndpoints.Set(float64(count))
}

func IncrementConnectionCount() {
	mu.Lock()
	defer mu.Unlock()
	connectionCount.Inc()
}

func DecrementConnectionCount() {
	mu.Lock()
	defer mu.Unlock()
	connectionCount.Dec()
}

func main() {
	// Handler for the first endpoint (listening on port 1234)
	mux1 := http.NewServeMux()
	mux1.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		endpointRequests.WithLabelValues("endpoint_1").Inc()
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
		http.Handle("/metrics", CustomPrometheusHandler())
		http.ListenAndServe(":8080", nil)
	}()

	incrementTotalEndpoints(4)
	select {}
}