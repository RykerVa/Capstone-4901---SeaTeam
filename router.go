package main

import (
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
	Routes       map[string]http.Handler
}

// AddRoute adds a new route to the router.
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
	urlPath := r.URL.Path
	routeConfig := sr.Config.StaticResources.Listeners[0].FilterChains[0].Filters[0].TypedConfig.RouteConfig
	virtualHost := routeConfig.VirtualHosts[0]

	for _, route := range virtualHost.Routes {
		if strings.HasPrefix(urlPath, route.Match.Prefix) {
			suffix := urlPath[len(route.Match.Prefix):]
			if suffix == "" || strings.HasPrefix(suffix, "/") {
				port := sr.Config.StaticResources.Clusters[0].LoadAssignment.Endpoints[0].LbEndpoints[0].Endpoint.Address.SocketAddress.PortValue
				address := sr.Config.StaticResources.Clusters[0].LoadAssignment.Endpoints[0].LbEndpoints[0].Endpoint.Address.SocketAddress.Address
				backendURL := fmt.Sprintf("http://%s:%d%s", address, port, suffix)
				return backendURL
			}
		}
	}
	return ""
}

// forwardRequest forwards the HTTP request to the backend service.
func forwardRequest(w http.ResponseWriter, r *http.Request, backendURL string) {
	req, err := http.NewRequest(r.Method, backendURL, r.Body)
	if err != nil {
		handleError(w, "Failed to create new request", http.StatusInternalServerError)
		return
	}

	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		handleError(w, "Failed to forward request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

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
	log.Println(message)
	http.Error(w, message, statusCode)
}
