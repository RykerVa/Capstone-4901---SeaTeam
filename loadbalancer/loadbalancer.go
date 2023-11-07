// Aldo M.
// Simple Round Robin Load Balancer
package main

import (
    "io"
    "net/http"
    "sync"
)

type RoundRobinLoadBalancer struct {
    servers   []string
    weights   []int
    current   int
    mutex     sync.Mutex
    serverLen int
}

func NewWeightedRoundRobinLoadBalancer(servers []string, weights []int) *RoundRobinLoadBalancer {
    if len(servers) != len(weights) {
        panic("Number of servers and weights must match")
    }

    return &RoundRobinLoadBalancer{
        servers:   servers,
        weights:   weights,
        current:   0,
        mutex:     sync.Mutex{},
        serverLen: len(servers),
    }
}

func (lb *RoundRobinLoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    lb.mutex.Lock()
    defer lb.mutex.Unlock()

    server := lb.getNextServer()

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

func (lb *RoundRobinLoadBalancer) getNextServer() string {
    totalWeight := 0

    for i := 0; i < lb.serverLen; i++ {
        totalWeight += lb.weights[i]
    }

    lb.current = (lb.current + 1) % totalWeight

    for i := 0; i < lb.serverLen; i++ {
        if lb.current < lb.weights[i] {
            return lb.servers[i]
        }
        lb.current -= lb.weights[i]
    }

    // This should not happen if the weights are correctly configured
    return lb.servers[0]
}