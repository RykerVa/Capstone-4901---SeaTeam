package main

import (
	"net/http"
    "net/http/httptest"
	"testing"
    "time"
    lb "Capstone-4901---SeaTeam/loadbalancer"
)

// TestDetermineBackendURL checks if the correct backend URL is determined.
func TestDetermineBackendURL(t *testing.T) {
	request1, _ := http.NewRequest("GET", "/service1", nil)
	url1 := determineBackendURL(request1)
	expectedURL1 := "http://backend-service-1-url"
	if url1 != expectedURL1 {
		t.Errorf("Expected URL: %s, Got: %s", expectedURL1, url1)
	}

	request2, _ := http.NewRequest("GET", "/service2", nil)
	url2 := determineBackendURL(request2)
	expectedURL2 := "http://backend-service-2-url"
	if url2 != expectedURL2 {
		t.Errorf("Expected URL: %s, Got: %s", expectedURL2, url2)
}
}

// Define a mock handler to use for testing
func mockHandler() http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte("mock response"))
    })
}

// TestRouter tests the basic routing logic
func TestRouter(t *testing.T) {
    // Initialize router with some routes for testing
    r := NewRouter()
    r.AddRoute("/test", mockHandler())

    // Create a new HTTP request to test the routing
    req, err := http.NewRequest("GET", "/test", nil)
    if err != nil {
        t.Fatal(err)
    }

    // Record the HTTP response using httptest
    rr := httptest.NewRecorder()
    handler := http.Handler(r)

    // Dispatch the request to the handler
    handler.ServeHTTP(rr, req)

    // Check the status code is what we expect.
    if status := rr.Code; status != http.StatusOK {
        t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
    }

    // Check the response body is what we expect.
    expected := `mock response`
    if rr.Body.String() != expected {
        t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
    }
}


// TestRouter_NoRoute tests the behavior when no route is matched
func TestRouter_NoRoute(t *testing.T) {
    // Initialize router without any routes
    r := NewRouter()

    // Create a request for a route that does not exist
    req, err := http.NewRequest("GET", "/not-found", nil)
    if err != nil {
        t.Fatal(err)
    }

    // Record the HTTP response
    rr := httptest.NewRecorder()
    handler := http.Handler(r)

    // Dispatch the request
    handler.ServeHTTP(rr, req)

    // Check that the status code reflects a not found error
    if rr.Code != http.StatusNotFound {
        t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusNotFound)
    }
}

func NewRouter() *Router {
    // Initialize your load balancer with appropriate parameters
    loadBalancer := lb.NewRoundRobinLoadBalancer([]string{"backend-service-1-url", "backend-service-2-url"})
    
    return &Router{
        Timeout:      10 * time.Second,
        LoadBalancer: loadBalancer, // Note the dereference to make LoadBalancer a value, not a pointer
        // other initialization logic if needed
    }
}
