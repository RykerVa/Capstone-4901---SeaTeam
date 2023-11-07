package main 

import (
    "testing"
)

// MockService represents a mock service for testing service discovery
type MockService struct {
    Name     string
    Address  string
    IsHealthy bool
}

// MockServiceDiscovery is a mock implementation of a ServiceDiscovery interface for testing
type MockServiceDiscovery struct {
    services map[string]MockService
}

// NewMockServiceDiscovery creates and returns a new instance of MockServiceDiscovery
func NewMockServiceDiscovery(services map[string]MockService) *MockServiceDiscovery {
    return &MockServiceDiscovery{services: services}
}

// FindService mocks the behavior of looking up a service by name
func (msd *MockServiceDiscovery) FindService(name string) (MockService, bool) {
    service, exists := msd.services[name]
    return service, exists
}

// TestFindService tests the service discovery's ability to find registered services
func TestFindService(t *testing.T) {
    services := map[string]MockService{
        "auth-service": {Name: "auth-service", Address: "http://auth.local", IsHealthy: true},
        "user-service": {Name: "user-service", Address: "http://user.local", IsHealthy: true},
    }

    sd := NewMockServiceDiscovery(services)

    tests := []struct {
        serviceName string
        expected    bool
    }{
        {"auth-service", true},
        {"user-service", true},
        {"payment-service", false}, // not registered, should return false
    }

    for _, tt := range tests {
        _, found := sd.FindService(tt.serviceName)
        if found != tt.expected {
            t.Errorf("FindService(%s) expected %v, got %v", tt.serviceName, tt.expected, found)
        }
    }
}

// TestHealthCheck verifies that the service discovery checks for service health correctly
func TestHealthCheck(t *testing.T) {
    services := map[string]MockService{
        "auth-service": {Name: "auth-service", Address: "http://auth.local", IsHealthy: true},
        "user-service": {Name: "user-service", Address: "http://user.local", IsHealthy: false}, // unhealthy service
    }

    sd := NewMockServiceDiscovery(services)

    if service, found := sd.FindService("user-service"); found && service.IsHealthy {
        t.Errorf("Service %s should be unhealthy", service.Name)
    }

    if service, found := sd.FindService("auth-service"); found && !service.IsHealthy {
        t.Errorf("Service %s should be healthy", service.Name)
    }
}

// Add more tests as needed for all the functionalities the service discovery component.
