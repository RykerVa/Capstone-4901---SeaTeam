package main

import "net"

// Filter interface defines the methods that a custom filter should implement.
type Filter interface {
    // Init initializes the filter with configuration parameters.
    Init(config map[string]interface{}) error

    // OnNewConnection is called when a new client connection is established.
    OnNewConnection(clientConn net.Conn) (bool, error)

    // OnRequest is called when a new request is received from the client.
    OnRequest(request []byte) ([]byte, error)

    // OnResponse is called when a response is received from the backend service.
    OnResponse(response []byte) ([]byte, error)
}
