package main

import (
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
    r := &Router{}
    http.ListenAndServe(":8000", r)
}

// Router handles incoming HTTP requests and routes them to the appropriate backend.
type Router struct {
    Timeout      time.Duration
    LoadBalancer LoadBalancer
    ErrorLogger  *log.Logger
}

// ServeHTTP implements the http.Handler interface for Router.
func (sr *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    backendURL := determineBackendURL(r)
    if backendURL == "" {
        http.NotFound(w, r)
        return
    }

    forwardRequest(w, r, backendURL)
}

// determineBackendURL determines the backend URL based on the request.
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
    body, err := ioutil.ReadAll(resp.Body)
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
