package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/Layr-Labs/hourglass-avs-template/cmd/matcher"
	"github.com/Layr-Labs/hourglass-avs-template/cmd/submitter"
	"github.com/joho/godotenv"
)

// Global variables
var (
	orderBook    []matcher.Order
	mu           sync.Mutex
	volumeData   []VolumeEntry
	volumeMu     sync.Mutex
	totalVolume  float64
)

// Frontend-compatible data structures
type FrontendOrder struct {
	ID        string  `json:"id"`
	Price     float64 `json:"price"`
	Amount    float64 `json:"amount"`
	Timestamp int64   `json:"timestamp"`
	Side      string  `json:"side"`
}

type OrderBookResponse struct {
	Bids      []FrontendOrder `json:"bids"`
	Asks      []FrontendOrder `json:"asks"`
	Timestamp int64           `json:"timestamp"`
}

type DepthData struct {
	Price    float64 `json:"price"`
	BidDepth float64 `json:"bidDepth"`
	AskDepth float64 `json:"askDepth"`
}

type DepthResponse struct {
	Depths    []DepthData `json:"depths"`
	Timestamp int64       `json:"timestamp"`
}

type VolumeEntry struct {
	Time   string  `json:"time"`
	Volume float64 `json:"volume"`
	Value  float64 `json:"value"`
}

type VolumeResponse struct {
	HourlyVolume []VolumeEntry `json:"hourlyVolume"`
	TotalVolume  float64       `json:"totalVolume"`
	Timestamp    int64         `json:"timestamp"`
}

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

// convertToFrontendOrder converts internal matcher.Order to frontend-compatible format
func convertToFrontendOrder(order matcher.Order, side string) (FrontendOrder, error) {
	amount, err := strconv.ParseFloat(order.MakeAmount, 64)
	if err != nil {
		return FrontendOrder{}, fmt.Errorf("invalid makeAmount: %v", err)
	}

	// Generate unique ID from order hash
	id := fmt.Sprintf("%s_%d", order.Maker[:8], order.Timestamp)

	return FrontendOrder{
		ID:        id,
		Price:     order.Price,
		Amount:    amount,
		Timestamp: order.Timestamp,
		Side:      side,
	}, nil
}

// classifyOrderSide determines if an order is a bid or ask based on price relative to market
func classifyOrderSide(order matcher.Order, allOrders []matcher.Order) string {
	if len(allOrders) == 0 {
		return "bid" // Default to bid if no comparison available
	}

	// Calculate median price to determine bid/ask classification
	prices := make([]float64, len(allOrders))
	for i, o := range allOrders {
		prices[i] = o.Price
	}

	// Simple classification: above median = bid, below median = ask
	var sum float64
	for _, p := range prices {
		sum += p
	}
	avgPrice := sum / float64(len(prices))

	if order.Price >= avgPrice {
		return "bid"
	}
	return "ask"
}

// trackVolume adds volume data for a completed trade
func trackVolume(price, quantity float64) {
	volumeMu.Lock()
	defer volumeMu.Unlock()

	now := time.Now()
	timeStr := now.Format("15:04")
	value := price * quantity

	// Add to total volume
	totalVolume += quantity

	// Add to hourly volume data (keep last 24 hours)
	entry := VolumeEntry{
		Time:   timeStr,
		Volume: quantity,
		Value:  value,
	}

	volumeData = append(volumeData, entry)

	// Keep only last 24 entries (24 hours if updated hourly)
	if len(volumeData) > 24 {
		volumeData = volumeData[1:]
	}
}

// generateMockVolumeData creates sample volume data for demonstration
func generateMockVolumeData() []VolumeEntry {
	data := make([]VolumeEntry, 24)
	now := time.Now()

	for i := 0; i < 24; i++ {
		t := now.Add(time.Duration(-23+i) * time.Hour)
		volume := float64(1000 + (i*50)) // Increasing volume throughout day
		price := 1.25 + (float64(i%8)-4)*0.01 // Price variation

		data[i] = VolumeEntry{
			Time:   t.Format("15:04"),
			Volume: volume,
			Value:  volume * price,
		}
	}

	return data
}

// enableCORS adds CORS headers to allow frontend access
func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// handleOrders handles POST /orders endpoint
func handleOrders(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

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

	// Track volume for completed fills
	if err == nil && len(fillsBytes) > 0 {
		// Parse fills to track volume
		var fills []matcher.Fill
		if jsonErr := json.Unmarshal(fillsBytes, &fills); jsonErr == nil {
			for _, fill := range fills {
				if quantity, parseErr := strconv.ParseFloat(fill.Quantity, 64); parseErr == nil {
					// Use average price for volume tracking
					avgPrice := o.Price
					trackVolume(avgPrice, quantity)
				}
			}
		}
	}
	mu.Unlock()

	if err == nil && len(fillsBytes) > 0 {
		aggSig, err := matcher.AggregateBLS(root)
		if err != nil {
			log.Printf("BLS aggregate error: %v", err)
		} else {
			if txHash, err2 := submitter.SubmitBatch(root, fillsBytes, aggSig); err2 == nil {
				log.Printf("Batch submitted: %s", txHash)
			} else {
				log.Printf("Error submitting batch: %v", err2)
			}
		}
	} else if err != nil {
		log.Printf("Error in MatchAndBatch: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success":true}`))
}

// handleOrderBook handles GET /book endpoint
func handleOrderBook(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	mu.Lock()
	orderBookCopy := make([]matcher.Order, len(orderBook))
	copy(orderBookCopy, orderBook)
	mu.Unlock()

	// Convert to frontend format and classify as bids/asks
	var bids, asks []FrontendOrder

	for _, order := range orderBookCopy {
		side := classifyOrderSide(order, orderBookCopy)
		frontendOrder, err := convertToFrontendOrder(order, side)
		if err != nil {
			log.Printf("Error converting order: %v", err)
			continue
		}

		if side == "bid" {
			bids = append(bids, frontendOrder)
		} else {
			asks = append(asks, frontendOrder)
		}
	}

	// Sort bids by price (descending) and asks by price (ascending)
	for i := 0; i < len(bids)-1; i++ {
		for j := i + 1; j < len(bids); j++ {
			if bids[i].Price < bids[j].Price {
				bids[i], bids[j] = bids[j], bids[i]
			}
		}
	}

	for i := 0; i < len(asks)-1; i++ {
		for j := i + 1; j < len(asks); j++ {
			if asks[i].Price > asks[j].Price {
				asks[i], asks[j] = asks[j], asks[i]
			}
		}
	}

	response := OrderBookResponse{
		Bids:      bids,
		Asks:      asks,
		Timestamp: time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleDepth handles GET /depth endpoint
func handleDepth(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	mu.Lock()
	orderBookCopy := make([]matcher.Order, len(orderBook))
	copy(orderBookCopy, orderBook)
	mu.Unlock()

	// Group orders by price and compute cumulative depth
	priceMap := make(map[float64]struct {
		bidAmount float64
		askAmount float64
	})

	for _, order := range orderBookCopy {
		side := classifyOrderSide(order, orderBookCopy)
		amount, err := strconv.ParseFloat(order.MakeAmount, 64)
		if err != nil {
			continue
		}

		entry := priceMap[order.Price]
		if side == "bid" {
			entry.bidAmount += amount
		} else {
			entry.askAmount += amount
		}
		priceMap[order.Price] = entry
	}

	// Convert to sorted depth data with cumulative amounts
	var depths []DepthData
	var prices []float64

	for price := range priceMap {
		prices = append(prices, price)
	}

	// Sort prices
	for i := 0; i < len(prices)-1; i++ {
		for j := i + 1; j < len(prices); j++ {
			if prices[i] > prices[j] {
				prices[i], prices[j] = prices[j], prices[i]
			}
		}
	}

	// Calculate cumulative depths
	var cumulativeBidDepth, cumulativeAskDepth float64

	for _, price := range prices {
		entry := priceMap[price]
		cumulativeBidDepth += entry.bidAmount
		cumulativeAskDepth += entry.askAmount

		depths = append(depths, DepthData{
			Price:    price,
			BidDepth: cumulativeBidDepth,
			AskDepth: cumulativeAskDepth,
		})
	}

	response := DepthResponse{
		Depths:    depths,
		Timestamp: time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleVolume handles GET /volume endpoint
func handleVolume(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	volumeMu.Lock()
	var hourlyVolume []VolumeEntry
	currentTotalVolume := totalVolume

	if len(volumeData) == 0 {
		// Generate mock data if no real data exists
		hourlyVolume = generateMockVolumeData()
		currentTotalVolume = 45000 // Mock total
	} else {
		hourlyVolume = make([]VolumeEntry, len(volumeData))
		copy(hourlyVolume, volumeData)
	}
	volumeMu.Unlock()

	response := VolumeResponse{
		HourlyVolume: hourlyVolume,
		TotalVolume:  currentTotalVolume,
		Timestamp:    time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleHealth handles GET /health endpoint
func handleHealth(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// Main function starts the Polymarket CLOB Sequencer service
func main() {
	log.Println("Starting Polymarket CLOB Sequencer...")

	// Load environment variables from .env file
	if err := godotenv.Load(".env"); err != nil {
		// Try to load from cmd directory if not found in current directory
		if err2 := godotenv.Load("cmd/.env"); err2 != nil {
			log.Printf("Warning: Could not load .env file: %v", err)
		} else {
			log.Println("Loaded environment variables from cmd/.env")
		}
	} else {
		log.Println("Loaded environment variables from .env")
	}

	// Validate required environment variables
	requiredVars := []string{"RPC_URL", "PRIVATE_KEY", "BATCH_SETTLEMENT_ADDRESS"}
	for _, envVar := range requiredVars {
		if os.Getenv(envVar) == "" {
			log.Printf("Warning: %s environment variable not set", envVar)
		}
	}

	// Initialize the global orderbook and volume tracking
	orderBook = make([]matcher.Order, 0)
	volumeData = make([]VolumeEntry, 0)
	totalVolume = 0
	log.Println("Orderbook and volume tracking initialized")

	// Setup HTTP routes
	http.HandleFunc("/orders", handleOrders)
	http.HandleFunc("/book", handleOrderBook)
	http.HandleFunc("/depth", handleDepth) 
	http.HandleFunc("/volume", handleVolume)
	http.HandleFunc("/health", handleHealth)

	// Start the HTTP server on port 8081
	log.Println("Starting HTTP server on port 8081...")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
