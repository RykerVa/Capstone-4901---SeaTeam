// Aldo M.
// Simple Round Robin Load Balancer
package loadbalancer

import (
	"io"
	"net/http"
	"sync"
)

// LoadBalancer is the interface that defines the methods for a load balancer.
type LoadBalancer interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	NextEndpoint() string
	UpdateEndpoints(newServers []string)
}

type RoundRobinLoadBalancer struct {
	servers   []string
	current   int
	mutex     sync.Mutex
	serverLen int
}

func NewRoundRobinLoadBalancer(servers []string) *RoundRobinLoadBalancer {
	return &RoundRobinLoadBalancer{
		servers:   servers,
		current:   0,
		mutex:     sync.Mutex{},
		serverLen: len(servers),
	}
}

func (lb *RoundRobinLoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	server := lb.servers[lb.current]
	lb.current = (lb.current + 1) % lb.serverLen

	proxyRequest, err := http.NewRequest(r.Method, "http://"+server+r.URL.String(), r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	proxyRequest.Header = make(http.Header)
	for key, values := range r.Header {
		proxyRequest.Header[key] = values
	}

	client := &http.Client{}
	resp, err := client.Do(proxyRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		w.Header()[key] = values
	}

	w.WriteHeader(resp.StatusCode)

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// NextEndpoint returns the next backend server based on round-robin logic.
func (lb *RoundRobinLoadBalancer) NextEndpoint() string {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	server := lb.servers[lb.current]
	lb.current = (lb.current + 1) % lb.serverLen

	return server
}

func (lb *RoundRobinLoadBalancer) UpdateEndpoints(newServers []string) {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	// Create a map to check for existing servers
	existingServers := make(map[string]bool)
	for _, server := range lb.servers {
		existingServers[server] = true
	}

	// Remove servers that are no longer present
	var updatedServers []string
	for _, server := range lb.servers {
		if existingServers[server] && contains(newServers, server) {
			updatedServers = append(updatedServers, server)
		}
	}

	// Add new servers that are not already present
	for _, server := range newServers {
		if !existingServers[server] {
			updatedServers = append(updatedServers, server)
		}
	}

	// Update the list of servers and length
	lb.servers = updatedServers
	lb.serverLen = len(updatedServers)

	// If the current index is out of bounds, reset to 0
	if lb.current >= lb.serverLen {
		lb.current = 0
	}
}

func contains(servers []string, target string) bool {
	for _, server := range servers {
		if server == target {
			return true
		}
	}
	return false
}
