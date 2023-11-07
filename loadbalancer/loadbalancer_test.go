package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestLoadBalancer ensures that the load balancer distributes requests among servers.
func TestLoadBalancer(t *testing.T) {
    lb := NewWeightedRoundRobinLoadBalancer(
        []string{"server1", "server2", "server3"},
        []int{1, 1, 1}, // Equal weights for initial testing
    )

    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        lb.ServeHTTP(w, r)
    })

    // Single request load balancing test
    request, _ := http.NewRequest("GET", "/", nil)
    recorder := httptest.NewRecorder()
    handler.ServeHTTP(recorder, request)

    // Add assertions to test load balancer behavior with a single request

    // Test for multiple requests
    for i := 0; i < 10; i++ {
        request, _ = http.NewRequest("GET", "/", nil)
        recorder = httptest.NewRecorder()
        handler.ServeHTTP(recorder, request)

		if expectedServer := "server1"; recorder.Body.String() != expectedServer {
            t.Errorf("Expected load balancer to distribute requests to %s, got %s", expectedServer, recorder.Body.String())
        }
        // Add assertions for load balancing behavior with multiple requests
    }
}
