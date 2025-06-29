package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// Main function starts the Polymarket CLOB Sequencer service
func main() {
	log.Println("Starting Polymarket CLOB Sequencer...")

	// Set up graceful shutdown
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal, stopping server...")
		cancel()
	}()

	// Initialize the global orderbook
	orderBook = make([]Order, 0)
	log.Println("Orderbook initialized")

	// Start the HTTP server
	// This call blocks until the server is stopped
	startHTTPServer()
}
