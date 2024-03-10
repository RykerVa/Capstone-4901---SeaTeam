package loadbalancer

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestLoadBalancer ensures that the load balancer distributes requests among servers.
func TestRoundRobinLoadBalancer(t *testing.T) {
    lb := NewRoundRobinLoadBalancer(
        []string{"server1", "server2", "server3"},
    )

    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        lb.ServeHTTP(w, r)
    })

    // Add assertions for load balancing behavior with multiple requests
    for i := 0; i < 10; i++ {
        request, _ := http.NewRequest("GET", "/", nil)
        recorder := httptest.NewRecorder()
        handler.ServeHTTP(recorder, request)

        // Add assertions here to test load balancer behavior with multiple requests
        if expectedServer := "server1"; recorder.Body.String() != expectedServer {
            t.Errorf("Expected load balancer to distribute requests to %s, got %s", expectedServer, recorder.Body.String())
        }
    }
}

/*func TestLeastConnectionsLoadBalancer(t *testing.T) {
    lb := NewLeastConnectionsLoadBalancer([]string{"server1", "server2", "server3"})

    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        lb.ServeHTTP(w, r)
    })
}
*/