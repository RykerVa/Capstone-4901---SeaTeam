package router // Replace with  actual package name

import (
    "net/http"
    "net/http/httptest"
    "testing"
)

// Define a mock handler to use for testing
func mockHandler() http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte("mock response"))
    })
}

// TestRouter tests the basic routing logic
func TestRouter(t *testing.T) {
    // Initialize router with some routes for testing
    r := NewRouter()
    r.AddRoute("/test", mockHandler())

    // Create a new HTTP request to test the routing
    req, err := http.NewRequest("GET", "/test", nil)
    if err != nil {
        t.Fatal(err)
    }

    // Record the HTTP response using httptest
    rr := httptest.NewRecorder()
    handler := http.Handler(r)

    // Dispatch the request to the handler
    handler.ServeHTTP(rr, req)

    // Check the status code is what we expect.
    if status := rr.Code; status != http.StatusOK {
        t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
    }

    // Check the response body is what we expect.
    expected := `mock response`
    if rr.Body.String() != expected {
        t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
    }
}

// TestRouter_NoRoute tests the behavior when no route is matched
func TestRouter_NoRoute(t *testing.T) {
    // Initialize router without any routes
    r := NewRouter()

    // Create a request for a route that does not exist
    req, err := http.NewRequest("GET", "/not-found", nil)
    if err != nil {
        t.Fatal(err)
    }

    // Record the HTTP response
    rr := httptest.NewRecorder()
    handler := http.Handler(r)

    // Dispatch the request
    handler.ServeHTTP(rr, req)

    // Check that the status code reflects a not found error
    if rr.Code != http.StatusNotFound {
        t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusNotFound)
    }
}

// I should add more tests to cover all functionality of your router,
// such as different HTTP methods, path parameters, query parameters, etc.

