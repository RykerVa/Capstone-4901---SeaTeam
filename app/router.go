package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"seateam/config"
	"seateam/loadbalancer"
)

// Router struct and its methods are defined here, reflecting the original design and functionality.
type Router struct {
	Timeout      time.Duration
	LoadBalancer loadbalancer.LoadBalancer
	ErrorLogger  *log.Logger
	Config       config.StaticBootstrap
	Routes       map[string]http.Handler
}

func (sr *Router) AddRoute(path string, handler http.Handler) {
	if sr.Routes == nil {
		sr.Routes = make(map[string]http.Handler)
	}
	sr.Routes[path] = handler
}

func handleError(w http.ResponseWriter, message string, statusCode int) {
	// Implement the error handling logic
	fmt.Println(message)
	http.Error(w, message, statusCode)
}

// ServeHTTP implements the http.Handler interface for Router.
func (sr *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	endpointIndexStr := r.URL.Query().Get("endpoint")
	if endpointIndexStr != "lb" {
		endpointIndex, err := strconv.Atoi(endpointIndexStr)
		if err != nil {
			http.Error(w, "Invalid endpoint index", http.StatusBadRequest)
			return
		}
		backendURL := sr.determineBackendURL(r, endpointIndex)
		if backendURL == "" {
			http.NotFound(w, r)
			return
		}
		forwardRequest(w, r, backendURL)
	} else {
		// No specific endpoint index provided, use the load balancer to determine the backend
		backendURL := sr.LoadBalancer.NextEndpoint()
		if backendURL == "" {
			http.NotFound(w, r)
			return
		}
		forwardRequest(w, r, backendURL)
	}
}

// determineBackendURL determines the backend URL based on the request.
func (sr *Router) determineBackendURL(r *http.Request, endpointIndex int) string {
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
				port := sr.Config.StaticResources.Clusters[0].LoadAssignment.Endpoints[0].LbEndpoints[endpointIndex].Endpoint.Address.SocketAddress.PortValue
				address := sr.Config.StaticResources.Clusters[0].LoadAssignment.Endpoints[0].LbEndpoints[endpointIndex].Endpoint.Address.SocketAddress.Address

				// Include the retrieved port and address in the backend URL
				backendURL := fmt.Sprintf("http://%s:%d/%s", address, port, suffix)
				fmt.Println("Determined backend URL:", backendURL)
				return backendURL
			}
		}
	}

	return ""
}
//Good ^


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
