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
var mu sync.RWMutex

func getNextBackendServer() (*Server, string) {
	mu.RLock()
	defer mu.RUnlock()

	var skippedServer string

	for {
		index := atomic.AddUint32(&currentIndex, 1)
		server := backendServers[index%uint32(len(backendServers))]
		if server.Healthy {
			return server, skippedServer
		} else {
			skippedServer = server.URL
		}
	}
}

func handleRequestAndRedirect(w http.ResponseWriter, r *http.Request) {
	backendServer, skippedServer := getNextBackendServer()
	timestamp := time.Now().Format(time.RFC3339)
	broadcastUpdate(backendServer.URL, true, "request", timestamp, skippedServer)
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
