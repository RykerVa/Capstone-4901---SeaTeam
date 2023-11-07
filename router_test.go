package main

import (
	"net/http"
	"testing"
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