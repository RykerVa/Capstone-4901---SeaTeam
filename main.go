package main

import (
	api "Capstone-4901---SeaTeam/api"
	config "Capstone-4901---SeaTeam/config"
	lb "Capstone-4901---SeaTeam/loadbalancer"

	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)


// Router handles incoming HTTP requests and routes them to the appropriate backend.
type Router struct {
	Timeout      time.Duration
	LoadBalancer *lb.RoundRobinLoadBalancer
	ErrorLogger  *log.Logger
	Config       config.StaticBootstrap
	Routes map[string]http.Handler
}

// determineBackendURL determines the backend URL based on the request path.
//Old version, just in case
func determineBackendURL(r *http.Request) string {
	switch r.URL.Path {
	case "/service1":
		return "http://backend-service-1-url"
	case "/service2":
		return "http://backend-service-2-url"
	default:
		return ""
	}
}

func (sr *Router) AddRoute(path string, handler http.Handler) {
    if sr.Routes == nil {
        sr.Routes = make(map[string]http.Handler)
    }
    sr.Routes[path] = handler
}

// ServeHTTP implements the http.Handler interface for Router.
func (sr *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	backendURL := sr.determineBackendURL(r)
	if backendURL == "" {
		http.NotFound(w, r)
		return
	}

	forwardRequest(w, r, backendURL)
}

// determineBackendURL determines the backend URL based on the request.
func (sr *Router) determineBackendURL(r *http.Request) string {
	// Extract the URL path from the request
	urlPath := r.URL.Path

	// Extract route information from the request path and the configuration
	routeConfig := sr.Config.StaticResources.Listeners[0].FilterChains[0].Filters[0].TypedConfig.RouteConfig
	virtualHost := routeConfig.VirtualHosts[0]

	for _, route := range virtualHost.Routes {
		if strings.HasPrefix(urlPath, route.Match.Prefix) {
			// Extract the part of the URL path that follows the route's prefix
			suffix := urlPath

			// Check if the suffix is empty or starts with a "/"
			if suffix == "" || strings.HasPrefix(suffix, "/") {
				port := sr.Config.StaticResources.Clusters[0].LoadAssignment.Endpoints[0].LbEndpoints[0].Endpoint.Address.SocketAddress.PortValue
				address := sr.Config.StaticResources.Clusters[0].LoadAssignment.Endpoints[0].LbEndpoints[0].Endpoint.Address.SocketAddress.Address

				// Include the retrieved port and address in the backend URL
				backendURL := fmt.Sprintf("http://%s:%d/%s", address, port, suffix)
				fmt.Println("Determined backend URL:", backendURL)
				return backendURL
			}
		}
	}

	return ""
}

// forwardRequest forwards the HTTP request to the backend service.
func forwardRequest(w http.ResponseWriter, r *http.Request, backendURL string) {
	// Create a new HTTP request to the backend.
	req, err := http.NewRequest(r.Method, backendURL, r.Body)
	if err != nil {
		handleError(w, "Failed to create new request", http.StatusInternalServerError)
		return
	}

	// Copy original headers to the new request.
	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Forward the request to the backend.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		handleError(w, "Failed to forward request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Copy backend response headers to the original response writer.
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Copy backend response body to the original response writer.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		handleError(w, "Failed to read response body", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

// handleError logs an error message and writes an HTTP error response.
func handleError(w http.ResponseWriter, message string, statusCode int) {
	// Implement the error handling logic
	fmt.Println(message)
	http.Error(w, message, statusCode)
}

func main() {
	configuration := config.GetYAMLdata()
	r := &Router{
		Timeout: 10 * time.Second, // Example timeout value
		Config:  configuration,
	}
	// Serve the API endpoints
	http.Handle("/", r)
	http.HandleFunc("/health", api.HealthCheckHandler)
	http.HandleFunc("/endpoint1", api.Endpoint1Handler)
	http.HandleFunc("/endpoint2", api.Endpoint2Handler)

	// Serve frontend files
	fs := http.FileServer(http.Dir("./frontend"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	fmt.Println("Server started on :8000")
	http.ListenAndServe(":8000", nil)
}
