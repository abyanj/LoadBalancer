package main

import (
	"io"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type Server struct {
	URL     string
	Healthy bool
}

var backendServers = []*Server{
	{URL: "http://localhost:8081", Healthy: true},
	{URL: "http://localhost:8082", Healthy: true},
	{URL: "http://localhost:8083", Healthy: true},
}

var currentIndex uint32
var mu sync.RWMutex // Protects concurrent reads and writes to the backend server list

// getNextBackendServer returns the next healthy backend server in round-robin fashion
func getNextBackendServer() *Server {
	mu.RLock() // Acquire a read lock to safely access backendServers
	defer mu.RUnlock()

	for {
		index := atomic.AddUint32(&currentIndex, 1)
		server := backendServers[index%uint32(len(backendServers))]
		if server.Healthy {
			return server
		}
	}
}

// handleRequestAndRedirect forwards incoming requests to one of the backend servers
func handleRequestAndRedirect(w http.ResponseWriter, r *http.Request) {
	// Get the next healthy backend server to use
	backendServer := getNextBackendServer()

	// Forward the request to the chosen backend server
	resp, err := http.Get(backendServer.URL + r.URL.Path)
	if err != nil {
		log.Println("Error forwarding request:", err)
		http.Error(w, "Backend server not available", http.StatusServiceUnavailable)
		return
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Println("Error closing response body:", err)
		}
	}()

	// Copy the backend server's response to the client
	w.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Println("Error copying response body:", err)
		http.Error(w, "Error reading backend response", http.StatusInternalServerError)
	}
}

// healthCheckServer checks the health of a backend server
func healthCheckServer(server *Server) {
	resp, err := http.Get(server.URL + "/") // Replace "/" with a specific health check path if needed

	// Acquire a write lock before modifying the health status
	mu.Lock()
	defer mu.Unlock()

	if err != nil || resp.StatusCode != http.StatusOK {
		server.Healthy = false
		log.Printf("Server %s is unhealthy\n", server.URL)
		return
	}

	server.Healthy = true
	log.Printf("Server %s is healthy\n", server.URL)
}

// healthCheck periodically checks the health of all backend servers
func healthCheck() {
	for {
		for _, server := range backendServers {
			go healthCheckServer(server) // Run health check in a separate goroutine for each server
		}
		time.Sleep(10 * time.Second) // Wait for 10 seconds before checking again
	}
}

func init() {
	go healthCheck() // Start the health check goroutine when the server starts
}
