package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Server struct {
	URL     string
	Healthy bool
}

var backendServers = []*Server{
	{URL: "http://localhost:8081", Healthy: false}, // Start with servers offline
	{URL: "http://localhost:8082", Healthy: false},
	{URL: "http://localhost:8083", Healthy: false},
}

var currentIndex uint32
var mu sync.RWMutex

func getNextBackendServer() (*Server, []string) {
	mu.RLock()
	defer mu.RUnlock()

	var skippedServers []string
	healthyServerFound := false
	healthyServer := &Server{}
	attemptCount := 0

	for attemptCount < len(backendServers) {
		index := atomic.AddUint32(&currentIndex, 1)
		server := backendServers[index%uint32(len(backendServers))]

		if server.Healthy {
			healthyServerFound = true
			healthyServer = server
			break
		} else {

			skippedServers = append(skippedServers, server.URL)
			attemptCount++
		}
	}

	if !healthyServerFound {
		log.Println("All servers unhealthy. No healthy servers available.")
		return backendServers[0], skippedServers
	}

	return healthyServer, skippedServers
}

// handleRequestAndRedirect forwards requests to backend servers
func handleRequestAndRedirect(w http.ResponseWriter, r *http.Request) {
	backendServer, skippedServers := getNextBackendServer()
	timestamp := time.Now().Format(time.RFC3339)

	skippedServersString := ""
	if len(skippedServers) > 0 {
		skippedServersString = strings.Join(skippedServers, ", ")
	}

	// Broadcast request forwarding and skipped servers
	broadcastUpdate(backendServer.URL, true, "request", timestamp, skippedServersString)

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
	w.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Println("Error copying response body:", err)
		http.Error(w, "Error reading backend response", http.StatusInternalServerError)
	}
}

func startServerHandler(w http.ResponseWriter, r *http.Request) {
	serverID := r.URL.Query().Get("server")
	if serverID == "" {
		http.Error(w, "Missing server ID", http.StatusBadRequest)
		return
	}

	var cmd *exec.Cmd
	if serverID == "8081" {
		cmd = exec.Command("python", "-m", "http.server", "8081")
	} else if serverID == "8082" {
		cmd = exec.Command("python", "-m", "http.server", "8082")
	} else if serverID == "8083" {
		cmd = exec.Command("python", "-m", "http.server", "8083")
	} else {
		http.Error(w, "Invalid server ID", http.StatusBadRequest)
		return
	}

	err := cmd.Start()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to start server: %v", err), http.StatusInternalServerError)
		return
	}

	markServerAsHealthy(serverID)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Server %s started successfully", serverID)
}

func markServerAsHealthy(serverID string) {
	mu.Lock()
	defer mu.Unlock()

	for _, server := range backendServers {
		if server.URL == fmt.Sprintf("http://localhost:%s", serverID) {
			server.Healthy = true
			broadcastUpdate(server.URL, true, "status", "", "")
		}
	}
}

func healthCheckServer(server *Server) {
	resp, err := http.Get(server.URL + "/")
	mu.Lock()
	defer mu.Unlock()

	previousHealth := server.Healthy
	if err != nil || resp.StatusCode != http.StatusOK {
		server.Healthy = false
		log.Printf("Server %s is unhealthy\n", server.URL)
	} else {
		server.Healthy = true
		log.Printf("Server %s is healthy\n", server.URL)
	}

	if server.Healthy != previousHealth {
		broadcastUpdate(server.URL, server.Healthy, "status", "", "")
	}
}

func healthCheck() {
	for {
		for _, server := range backendServers {
			go healthCheckServer(server)
		}
		time.Sleep(10 * time.Second)
	}
}

func init() {
	go healthCheck()
}
