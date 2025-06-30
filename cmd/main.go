package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/Layr-Labs/hourglass-avs-template/cmd/matcher"
	"github.com/Layr-Labs/hourglass-avs-template/cmd/submitter"
)

// Global variables
var (
	orderBook []matcher.Order
	mu        sync.Mutex
)

// validateOrder validates an incoming order
func validateOrder(order matcher.Order) error {
	if order.Maker == "" {
		return fmt.Errorf("maker address cannot be empty")
	}
	if order.Price <= 0 {
		return fmt.Errorf("price must be positive")
	}
	if order.Timestamp <= 0 {
		return fmt.Errorf("timestamp must be positive")
	}
	if order.Signature == "" {
		return fmt.Errorf("signature cannot be empty")
	}
	if order.TakerAsset == "" {
		return fmt.Errorf("takerAsset cannot be empty")
	}
	if order.MakeAmount == "" {
		return fmt.Errorf("makeAmount cannot be empty")
	}
	if order.TakeAmount == "" {
		return fmt.Errorf("takeAmount cannot be empty")
	}

	// Validate amounts are positive numbers
	if makeAmt, err := strconv.ParseFloat(order.MakeAmount, 64); err != nil || makeAmt <= 0 {
		return fmt.Errorf("makeAmount must be a positive number")
	}
	if takeAmt, err := strconv.ParseFloat(order.TakeAmount, 64); err != nil || takeAmt <= 0 {
		return fmt.Errorf("takeAmount must be a positive number")
	}

	return nil
}

// handleOrders handles POST /orders endpoint
func handleOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var o matcher.Order
	if err := json.NewDecoder(r.Body).Decode(&o); err != nil {
		http.Error(w, `{"error":"Invalid order"}`, http.StatusBadRequest)
		return
	}

	// Validate order fields
	if err := validateOrder(o); err != nil {
		http.Error(w, `{"error":"Invalid order"}`, http.StatusBadRequest)
		return
	}

	// Add order to orderbook
	mu.Lock()
	orderBook = append(orderBook, o)
	orderBookCopy := make([]matcher.Order, len(orderBook))
	copy(orderBookCopy, orderBook)
	mu.Unlock()

	log.Printf("Order added to orderbook. Total orders: %d", len(orderBookCopy))

	// Trigger multi-fill matching with batch size limit of 100
	mu.Lock()
	root, fillsBytes, updatedBook, err := matcher.MatchAndBatch(orderBook, 100)
	orderBook = updatedBook
	mu.Unlock()

	if err == nil && len(fillsBytes) > 0 {
		aggSig := matcher.AggregateBLS(root)
		if txHash, err2 := submitter.SubmitBatch(root, fillsBytes, aggSig); err2 == nil {
			log.Printf("Batch submitted: %s", txHash)
		} else {
			log.Printf("Error submitting batch: %v", err2)
		}
	} else if err != nil {
		log.Printf("Error in MatchAndBatch: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success":true}`))
}

// handleHealth handles GET /health endpoint
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// Main function starts the Polymarket CLOB Sequencer service
func main() {
	log.Println("Starting Polymarket CLOB Sequencer...")

	// Initialize the global orderbook
	orderBook = make([]matcher.Order, 0)
	log.Println("Orderbook initialized")

	// Setup HTTP routes
	http.HandleFunc("/orders", handleOrders)
	http.HandleFunc("/health", handleHealth)

	// Start the HTTP server on port 8081
	log.Println("Starting HTTP server on port 8081...")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
