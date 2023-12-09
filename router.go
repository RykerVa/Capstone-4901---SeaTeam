package main

import (
    api "Capstone-4901---SeaTeam/api"
    config "Capstone-4901---SeaTeam/config"
    lb "Capstone-4901---SeaTeam/loadbalancer"
    "fmt"
    "io"
    "log"
    "net/http"
    "strconv"
    "strings"
    "time"
)

// Router handles incoming HTTP requests and routes them to the appropriate backend.
type Router struct {
    Timeout      time.Duration
    LoadBalancer lb.LoadBalancer
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
        backendURL := sr.LoadBalancer.NextEndpoint()
        if backendURL == "" {
            http.NotFound(w, r)
            return
        }
        forwardRequest(w, r, backendURL)
    }
}

func (sr *Router) determineBackendURL(r *http.Request, endpointIndex int) string {
    urlPath := r.URL.Path
    routeConfig := sr.Config.StaticResources.Listeners[0].FilterChains[0].Filters[0].TypedConfig.RouteConfig
    virtualHost := routeConfig.VirtualHosts[0]

    for _, route := range virtualHost.Routes {
        if strings.HasPrefix(urlPath, route.Match.Prefix) {
            suffix := urlPath[len(route.Match.Prefix):]
            port := sr.Config.StaticResources.Clusters[0].LoadAssignment.Endpoints[0].LbEndpoints[endpointIndex].Endpoint.Address.SocketAddress.PortValue
            address := sr.Config.StaticResources.Clusters[0].LoadAssignment.Endpoints[0].LbEndpoints[endpointIndex].Endpoint.Address.SocketAddress.Address
            backendURL := fmt.Sprintf("http://%s:%d%s", address, port, suffix)
            return backendURL
        }
    }
    return ""
}

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

func handleError(w http.ResponseWriter, message string, statusCode int) {
    fmt.Println(message)
    http.Error(w, message, statusCode)
}
