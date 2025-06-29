package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strconv"

	"github.com/cbergoon/merkletree"
)

// Fill represents a matched order fill for the Merkle tree
type Fill struct {
	MakerHash string `json:"makerHash"`
	TakerHash string `json:"takerHash"`
	Quantity  string `json:"quantity"`
}

// CalculateHash implements merkletree.Content interface
func (f Fill) CalculateHash() ([]byte, error) {
	h := sha256.New()
	data := fmt.Sprintf("%s:%s:%s", f.MakerHash, f.TakerHash, f.Quantity)
	h.Write([]byte(data))
	return h.Sum(nil), nil
}

// Equals implements merkletree.Content interface
func (f Fill) Equals(other merkletree.Content) (bool, error) {
	otherFill, ok := other.(Fill)
	if !ok {
		return false, nil
	}
	return f.MakerHash == otherFill.MakerHash &&
		f.TakerHash == otherFill.TakerHash &&
		f.Quantity == otherFill.Quantity, nil
}

// orderHash creates a hash for an order
func orderHash(order Order) string {
	h := sha256.New()
	data := fmt.Sprintf("%s:%s:%s:%s:%.8f:%d:%s",
		order.Maker, order.TakerAsset, order.MakeAmount, order.TakeAmount,
		order.Price, order.Timestamp, order.Signature)
	h.Write([]byte(data))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// min returns the minimum of two float64 values
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// computeMerkleRoot builds a Merkle tree over fills and returns the root
func computeMerkleRoot(fills []Fill) (string, error) {
	if len(fills) == 0 {
		return "", fmt.Errorf("cannot compute merkle root for empty fills")
	}

	// Convert fills to merkletree.Content slice
	contents := make([]merkletree.Content, len(fills))
	for i, fill := range fills {
		contents[i] = fill
	}

	// Build the Merkle tree
	tree, err := merkletree.NewTree(contents)
	if err != nil {
		return "", fmt.Errorf("failed to create merkle tree: %w", err)
	}

	// Get the root hash
	root := tree.MerkleRoot()
	return fmt.Sprintf("%x", root), nil
}

// MatchAndBatch implements the core matching logic
func MatchAndBatch() (string, []byte, error) {
	orderMux.RLock()
	defer orderMux.RUnlock()

	// Check if we have enough orders to match
	if len(orderBook) < 2 {
		log.Printf("Not enough orders to match. Current orderbook size: %d", len(orderBook))
		return "", nil, nil
	}

	log.Printf("Starting matching process with %d orders", len(orderBook))

	// Create a copy of orders for sorting
	orders := make([]Order, len(orderBook))
	copy(orders, orderBook)

	// Sort orders by descending price, then ascending timestamp
	sort.Slice(orders, func(i, j int) bool {
		if orders[i].Price != orders[j].Price {
			return orders[i].Price > orders[j].Price // Descending price
		}
		return orders[i].Timestamp < orders[j].Timestamp // Ascending timestamp (earlier first)
	})

	log.Printf("Orders sorted by price-time priority")

	// For this MVP, we'll match the top two orders
	if len(orders) < 2 {
		return "", nil, nil
	}

	makerOrder := orders[0]
	takerOrder := orders[1]

	log.Printf("Matching orders - Maker: %s (Price: %.8f), Taker: %s (Price: %.8f)",
		makerOrder.Maker, makerOrder.Price, takerOrder.Maker, takerOrder.Price)

	// Parse amounts
	makeAmount, err := strconv.ParseFloat(makerOrder.MakeAmount, 64)
	if err != nil {
		return "", nil, fmt.Errorf("invalid makeAmount: %w", err)
	}

	takeAmount, err := strconv.ParseFloat(takerOrder.TakeAmount, 64)
	if err != nil {
		return "", nil, fmt.Errorf("invalid takeAmount: %w", err)
	}

	// Calculate fill quantity as minimum of available amounts
	quantity := min(makeAmount, takeAmount)
	quantityStr := fmt.Sprintf("%.8f", quantity)

	log.Printf("Fill quantity calculated: %s", quantityStr)

	// Create fill with order hashes
	fill := Fill{
		MakerHash: orderHash(makerOrder),
		TakerHash: orderHash(takerOrder),
		Quantity:  quantityStr,
	}

	fills := []Fill{fill}

	// Compute Merkle root
	root, err := computeMerkleRoot(fills)
	if err != nil {
		return "", nil, fmt.Errorf("failed to compute merkle root: %w", err)
	}

	log.Printf("Merkle root computed: %s", root)

	// Serialize fills to bytes
	fillsBytes, err := json.Marshal(fills)
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal fills: %w", err)
	}

	// Remove matched orders from orderbook
	// Note: In a production system, you'd want more sophisticated order management
	orderMux.RUnlock()
	orderMux.Lock()
	
	// Simple removal - remove first two orders
	if len(orderBook) >= 2 {
		orderBook = orderBook[2:]
		log.Printf("Removed matched orders from orderbook. Remaining orders: %d", len(orderBook))
	}
	
	orderMux.Unlock()
	orderMux.RLock()

	return root, fillsBytes, nil
}

// AggregateBLS creates a mock BLS signature for the root
// In a real implementation, this would collect signatures from operators
func AggregateBLS(root string) ([]byte, error) {
	log.Printf("Aggregating BLS signatures for root: %s", root)
	
	// Mock implementation - in production, this would:
	// 1. Send root to all operators in the quorum
	// 2. Collect BLS signatures from >= 2/3 stake weight
	// 3. Aggregate signatures using BLS aggregation
	
	// For now, return a mock signature
	mockSignature := fmt.Sprintf("mock_bls_signature_%s", root[:16])
	
	log.Printf("BLS signature aggregated (mock): %s", mockSignature)
	
	return []byte(mockSignature), nil
}