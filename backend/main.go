package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	// Define the port for the load balancer to listen on
	port := ":8080"

	// Set up the handler for incoming requests
	http.HandleFunc("/", handleRequestAndRedirect)

	// Start the server
	fmt.Println("Starting load balancer on port", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
