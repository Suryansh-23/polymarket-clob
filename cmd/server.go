package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
)

// Order represents a polymarket CLOB order with EIP-712 signature
type Order struct {
	Maker      string  `json:"maker"`
	TakerAsset string  `json:"takerAsset"`
	MakeAmount string  `json:"makeAmount"`
	TakeAmount string  `json:"takeAmount"`
	Price      float64 `json:"price"`
	Timestamp  int64   `json:"timestamp"`
	Signature  string  `json:"signature"`
}

// Global orderbook protected by mutex
var (
	orderBook []Order
	orderMux  sync.RWMutex
)

// Response types
type SuccessResponse struct {
	Success bool `json:"success"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// validateOrder validates an incoming order
func validateOrder(order Order) error {
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

// handleOrderSubmission handles POST /orders endpoint
func handleOrderSubmission(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var order Order
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&order); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid JSON format"})
		return
	}

	// Validate the order
	if err := validateOrder(order); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid order"})
		return
	}

	// Add order to the global orderbook
	orderMux.Lock()
	orderBook = append(orderBook, order)
	orderBookLen := len(orderBook)
	orderMux.Unlock()

	log.Printf("Order added to orderbook. Total orders: %d", orderBookLen)

	// Try to match and batch orders after each submission
	go func() {
		root, fillsBytes, err := MatchAndBatch()
		if err != nil {
			log.Printf("Error in MatchAndBatch: %v", err)
			return
		}
		if root == "" {
			log.Printf("No matches found, waiting for more orders")
			return
		}

		aggSig, err := AggregateBLS(root)
		if err != nil {
			log.Printf("Error aggregating BLS signatures: %v", err)
			return
		}

		txHash, err := SubmitBatch(root, fillsBytes, aggSig)
		if err != nil {
			log.Printf("Error submitting batch: %v", err)
			return
		}

		log.Printf("Batch submitted successfully! Transaction hash: %s", txHash)
	}()

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{Success: true})
}

// setupRoutes configures HTTP routes
func setupRoutes() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/orders", handleOrderSubmission).Methods("POST")
	
	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	}).Methods("GET")

	return router
}

// startHTTPServer starts the HTTP server on port 8081
func startHTTPServer() {
	router := setupRoutes()
	
	log.Println("Starting Polymarket CLOB Sequencer HTTP server on port 8081...")
	log.Fatal(http.ListenAndServe(":8081", router))
}