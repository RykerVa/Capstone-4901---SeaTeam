package main

import (
    "net"
)

// RouterFilter is a basic router filter that routes traffic based on destination IP addresses.
type RouterFilter struct {
    allowedIPs map[string]bool
}

// NewRouterFilter creates a new RouterFilter instance with a list of allowed destination IP addresses.
func NewRouterFilter(allowedIPs []string) *RouterFilter {
    allowedIPMap := make(map[string]bool)
    for _, ip := range allowedIPs {
        allowedIPMap[ip] = true
    }
    return &RouterFilter{
        allowedIPs: allowedIPMap,
    }
}

// Init initializes the router filter.
func (f *RouterFilter) Init(config map[string]interface{}) error {
    // No specific initialization needed for this example
    return nil
}

// OnNewConnection is called when a new client connection is established.
func (f *RouterFilter) OnNewConnection(clientConn net.Conn) (bool, error) {
    // You can implement access control logic here
    clientAddr := clientConn.RemoteAddr().(*net.TCPAddr).IP.String()
    if f.allowedIPs[clientAddr] {
        return true, nil // Connection allowed
    }
    return false, nil // Connection denied
}

// OnRequest is not used in this router filter, as it focuses on connection-level routing.

// OnResponse is not used in this router filter.

// Sample usage:
func main() {
    allowedIPs := []string{"192.168.1.1", "192.168.1.2"}
    routerFilter := NewRouterFilter(allowedIPs)
    
    // Simulate a new client connection
    clientConn, _ := net.Dial("tcp", "192.168.1.1:12345")
    
    // Apply the filter to the new connection
    if allow, err := routerFilter.OnNewConnection(clientConn); err != nil {
        panic(err)
    } else if !allow {
        // Close the connection because it was denied
        clientConn.Close()
        // Handle denied connection (e.g., log or reject)
    } else {
        // Connection is allowed; proceed with processing
        // ...
    }
}
