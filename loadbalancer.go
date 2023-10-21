// Aldo M.
// Simple Round Robin Load Balancer
package main

import (
	"fmt"
	"io"
	"net/http"
	"sync"
)

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

func main() {

	servers := []string{
		"server1:8080",
		"server2:8080",
		"server3:8080",
	}

	lb := NewRoundRobinLoadBalancer(servers)

	http.Handle("/", lb)
	fmt.Println("Load balancer started on :8080")
	http.ListenAndServe(":8080", nil)
}
