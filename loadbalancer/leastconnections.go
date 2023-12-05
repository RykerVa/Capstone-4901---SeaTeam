package loadbalancer

import (
	"fmt"
	"io"
	"net/http"
	"sync"
)

type LeastConnectionsLoadBalancer struct {
	servers         []string
	connectionCount map[string]int
	mutex           sync.Mutex
}

func NewLeastConnectionsLoadBalancer(servers []string) *LeastConnectionsLoadBalancer {
	connectionCount := make(map[string]int)
	for _, server := range servers {
		connectionCount[server] = 0
	}

	return &LeastConnectionsLoadBalancer{
		servers:         servers,
		connectionCount: connectionCount,
	}
}

func (lb *LeastConnectionsLoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	server := lb.getLeastConnectionsServer()
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

func (lb *LeastConnectionsLoadBalancer) getLeastConnectionsServer() string {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	leastConnectionsServer := lb.servers[0]
	minConnections := lb.connectionCount[leastConnectionsServer]

	fmt.Printf("Current connection counts:\n")
	for _, server := range lb.servers {
		fmt.Printf("%s: %d connections\n", server, lb.connectionCount[server])
	}
	for _, server := range lb.servers {
		if lb.connectionCount[server] < minConnections {
			leastConnectionsServer = server
			minConnections = lb.connectionCount[server]
		}
	}

	lb.connectionCount[leastConnectionsServer]++
	return leastConnectionsServer
}

// NextEndpoint returns the next backend server based on round-robin logic.
func (lb *LeastConnectionsLoadBalancer) NextEndpoint() string {
	server := lb.getLeastConnectionsServer()

	return server
}

func (lb *LeastConnectionsLoadBalancer) UpdateEndpoints(newServers []string) {
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
}

// UpdateConnectionCount updates the connection count for a specific server.
func (lb *LeastConnectionsLoadBalancer) UpdateConnectionCount(server string) {
	lb.updateConnectionCount(server)
}

func (lb *LeastConnectionsLoadBalancer) updateConnectionCount(server string) {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()
	lb.connectionCount[server]++
}
