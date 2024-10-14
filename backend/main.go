package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

type ServerUpdate struct {
	Server      string `json:"Server"`
	Healthy     bool   `json:"Healthy"`
	MessageType string `json:"MessageType"`
	Timestamp   string `json:"Timestamp,omitempty"`
	Skipped     string `json:"Skipped,omitempty"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[*websocket.Conn]bool)

func handleWebSocketConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading to WebSocket:", err)
		return
	}
	defer func() {
		err := ws.Close()
		if err != nil {
			log.Println("Error closing WebSocket:", err)
		}
	}()

	clients[ws] = true

	broadcastInitialStatus(ws)

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ticker.C:
				if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
					log.Println("Ping failed:", err)
					ws.Close()
					delete(clients, ws)
					return
				}
			}
		}
	}()

	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			delete(clients, ws)
			break
		}
	}
}

func broadcastInitialStatus(ws *websocket.Conn) {
	mu.RLock()
	defer mu.RUnlock()

	for _, server := range backendServers {
		update := ServerUpdate{
			Server:      server.URL,
			Healthy:     server.Healthy,
			MessageType: "status",
		}
		message, err := json.Marshal(update)
		if err != nil {
			log.Printf("Error marshalling JSON: %v", err)
			return
		}
		ws.WriteMessage(websocket.TextMessage, message)
	}
}

func broadcastUpdate(serverURL string, healthy bool, messageType, timestamp, skippedServer string) {
	update := ServerUpdate{
		Server:      serverURL,
		Healthy:     healthy,
		MessageType: messageType,
		Timestamp:   timestamp,
		Skipped:     skippedServer,
	}

	message, err := json.Marshal(update)
	if err != nil {
		log.Printf("Error marshalling JSON: %v", err)
		return
	}

	for client := range clients {
		err := client.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("Error sending message to client: %v", err)
			client.Close()
			delete(clients, client)
		}
	}
}

func main() {
	port := ":8080"
	http.HandleFunc("/", handleRequestAndRedirect)
	http.HandleFunc("/ws", handleWebSocketConnections)
	fmt.Println("Starting load balancer with WebSocket on port", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
