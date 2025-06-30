package matcher

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/cbergoon/merkletree"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	"github.com/Layr-Labs/crypto-libs/pkg/signing"
)

// Global BLS private keys for operator signing
var (
	privKeys []signing.PrivateKey
)

// init loads BLS private keys from environment variable
func init() {
	// Load private keys from BLS_KEYS environment variable
	raw := os.Getenv("BLS_KEYS")
	if raw == "" {
		log.Printf("Warning: BLS_KEYS environment variable not set. Using mock BLS signing.")
		return
	}
	
	keys := strings.Split(raw, ",")
	privKeys = make([]signing.PrivateKey, 0, len(keys))
	
	for _, hexKey := range keys {
		hexKey = strings.TrimSpace(hexKey)
		if hexKey == "" {
			continue
		}
		
		keyBytes, err := hexutil.Decode(hexKey)
		if err != nil {
			log.Printf("Warning: Failed to decode BLS private key: %v", err)
			continue
		}
		
		// Use BN254 scheme to create private key from bytes
		scheme := bn254.NewScheme()
		privKey, err := scheme.NewPrivateKeyFromBytes(keyBytes)
		if err != nil {
			log.Printf("Warning: Failed to create BLS private key: %v", err)
			continue
		}
		
		privKeys = append(privKeys, privKey)
	}
	
	log.Printf("Loaded %d BLS private keys for operator signing", len(privKeys))
}

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

// parseAmount safely parses a string amount to float64
func parseAmount(amountStr string) (float64, error) {
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid amount format: %w", err)
	}
	if amount <= 0 {
		return 0, fmt.Errorf("amount must be positive, got: %f", amount)
	}
	return amount, nil
}

// formatAmount converts float64 back to string with consistent precision
func formatAmount(amount float64) string {
	return fmt.Sprintf("%.8f", amount)
}

// sortOrders sorts orders by price-time priority (descending price, ascending timestamp)
func sortOrders(orders []Order) {
	sort.Slice(orders, func(i, j int) bool {
		if orders[i].Price != orders[j].Price {
			return orders[i].Price > orders[j].Price // Descending price (higher prices first)
		}
		return orders[i].Timestamp < orders[j].Timestamp // Ascending timestamp (earlier first)
	})
}

// splitBidsAsks separates orders into bids (buyers) and asks (sellers)
// For simplicity, we assume higher priced orders are bids and lower are asks
// In a real implementation, this would be determined by order type field
func splitBidsAsks(orders []Order) (bids []Order, asks []Order) {
	if len(orders) == 0 {
		return bids, asks
	}

	// Sort by price first to determine bid/ask classification
	sorted := make([]Order, len(orders))
	copy(sorted, orders)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Price > sorted[j].Price
	})

	// Split at median price - top half are bids, bottom half are asks
	midpoint := len(sorted) / 2
	if len(sorted)%2 == 1 {
		midpoint++
	}

	for i := 0; i < midpoint && i < len(sorted); i++ {
		bids = append(bids, sorted[i])
	}
	for i := midpoint; i < len(sorted); i++ {
		asks = append(asks, sorted[i])
	}

	return bids, asks
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

// MatchAndBatch implements enhanced multi-fill matching logic with order book pruning
func MatchAndBatch(orders []Order, maxBatch int) (string, []byte, []Order, error) {
	// Check if we have enough orders to match
	if len(orders) < 2 {
		log.Printf("Not enough orders to match. Current orderbook size: %d", len(orders))
		return "", nil, orders, nil
	}

	log.Printf("Starting multi-fill matching process with %d orders, maxBatch: %d", len(orders), maxBatch)

	// Initialize variables for the matching loop
	fills := []Fill{}
	remainingOrders := make([]Order, 0, len(orders))

	// 1. Sort orders by descending Price, ascending Timestamp
	sortedOrders := make([]Order, len(orders))
	copy(sortedOrders, orders)
	sortOrders(sortedOrders)

	// 2. Split into bids and asks
	bids, asks := splitBidsAsks(sortedOrders)
	
	log.Printf("Split orders: %d bids, %d asks", len(bids), len(asks))

	// Create working copies to modify during matching
	workingBids := make([]Order, len(bids))
	workingAsks := make([]Order, len(asks))
	copy(workingBids, bids)
	copy(workingAsks, asks)

	i, j := 0, 0

	// Multi-fill matching loop
	for len(fills) < maxBatch && i < len(workingBids) && j < len(workingAsks) {
		bid := &workingBids[i]
		ask := &workingAsks[j]

		// Check if orders can cross (bid price >= ask price)
		if bid.Price < ask.Price {
			log.Printf("No more crossable orders: bid price %.8f < ask price %.8f", bid.Price, ask.Price)
			break
		}

		// Parse amounts for calculation
		bidMakeAmount, err := parseAmount(bid.MakeAmount)
		if err != nil {
			log.Printf("Error parsing bid makeAmount: %v", err)
			i++
			continue
		}

		askTakeAmount, err := parseAmount(ask.TakeAmount)
		if err != nil {
			log.Printf("Error parsing ask takeAmount: %v", err)
			j++
			continue
		}

		// 2. Compute fillQty = min(bid.makeAmount, ask.takeAmount)
		fillQty := min(bidMakeAmount, askTakeAmount)

		log.Printf("Matching bid %.8f @ %.8f with ask %.8f @ %.8f, fill quantity: %.8f",
			bidMakeAmount, bid.Price, askTakeAmount, ask.Price, fillQty)

		// Create fill record
		fill := Fill{
			MakerHash: orderHash(*bid),
			TakerHash: orderHash(*ask),
			Quantity:  formatAmount(fillQty),
		}
		fills = append(fills, fill)

		// 3. Reduce bid.MakeAmount and ask.TakeAmount by fillQty
		bidMakeAmount -= fillQty
		askTakeAmount -= fillQty

		bid.MakeAmount = formatAmount(bidMakeAmount)
		ask.TakeAmount = formatAmount(askTakeAmount)

		// 4. Advance or keep pointers based on leftover
		if bidMakeAmount <= 0.00000001 { // Use epsilon for floating point comparison
			log.Printf("Bid fully filled, advancing to next bid")
			i++
		}
		if askTakeAmount <= 0.00000001 { // Use epsilon for floating point comparison
			log.Printf("Ask fully filled, advancing to next ask")
			j++
		}
	}

	log.Printf("Matching complete: %d fills created", len(fills))

	// 5. Build remaining orders list - append unmatched bids and asks
	// Add unmatched bids
	for idx := i; idx < len(workingBids); idx++ {
		if amount, err := parseAmount(workingBids[idx].MakeAmount); err == nil && amount > 0.00000001 {
			remainingOrders = append(remainingOrders, workingBids[idx])
		}
	}

	// Add unmatched asks  
	for idx := j; idx < len(workingAsks); idx++ {
		if amount, err := parseAmount(workingAsks[idx].TakeAmount); err == nil && amount > 0.00000001 {
			remainingOrders = append(remainingOrders, workingAsks[idx])
		}
	}

	// Add partially filled orders if they have remaining amounts
	if i > 0 && i <= len(workingBids) {
		if bid := workingBids[i-1]; i-1 < len(workingBids) {
			if amount, err := parseAmount(bid.MakeAmount); err == nil && amount > 0.00000001 {
				remainingOrders = append(remainingOrders, bid)
			}
		}
	}
	if j > 0 && j <= len(workingAsks) {
		if ask := workingAsks[j-1]; j-1 < len(workingAsks) {
			if amount, err := parseAmount(ask.TakeAmount); err == nil && amount > 0.00000001 {
				remainingOrders = append(remainingOrders, ask)
			}
		}
	}

	log.Printf("Remaining orders after matching: %d (started with %d)", len(remainingOrders), len(orders))

	// If no fills were created, return original orders
	if len(fills) == 0 {
		log.Printf("No matches found, returning original orders")
		return "", nil, orders, nil
	}

	// Compute Merkle root for fills
	root, err := computeMerkleRoot(fills)
	if err != nil {
		return "", nil, remainingOrders, fmt.Errorf("failed to compute merkle root: %w", err)
	}

	log.Printf("Merkle root computed: %s", root)

	// Serialize fills to bytes
	fillsBytes, err := json.Marshal(fills)
	if err != nil {
		return "", nil, remainingOrders, fmt.Errorf("failed to marshal fills: %w", err)
	}

	return root, fillsBytes, remainingOrders, nil
}

// AggregateBLS creates a real BLS aggregate signature for the batch root
// Uses loaded operator private keys to sign and aggregate signatures
func AggregateBLS(root string) ([]byte, error) {
	log.Printf("Aggregating BLS signatures for root: %s", root)
	
	// If no real keys loaded, fall back to mock
	if len(privKeys) == 0 {
		log.Printf("No BLS private keys loaded, using mock signature")
		mockSignature := fmt.Sprintf("mock_bls_signature_%s", root[:16])
		return []byte(mockSignature), nil
	}
	
	// 1. Hash root into a BLS message
	msg := common.FromHex(root)
	if len(msg) == 0 {
		return nil, fmt.Errorf("invalid root hex string: %s", root)
	}
	
	// Use SHA256 hash of the root as the message to sign
	hasher := sha256.New()
	hasher.Write(msg)
	messageHash := hasher.Sum(nil)
	
	log.Printf("Message hash for signing: %s", hex.EncodeToString(messageHash))
	
	// 2. Each operator signs
	var sigs []signing.Signature
	for i, sk := range privKeys {
		s, err := sk.Sign(messageHash)
		if err != nil {
			log.Printf("Error signing with private key %d: %v", i, err)
			continue
		}
		sigs = append(sigs, s)
		log.Printf("Operator %d signed successfully", i)
	}
	
	if len(sigs) == 0 {
		return nil, fmt.Errorf("no valid signatures collected from %d operators", len(privKeys))
	}
	
	log.Printf("Collected %d valid signatures from operators", len(sigs))
	
	// 3. Aggregate signatures using BN254 scheme
	scheme := bn254.NewScheme()
	aggSig, err := scheme.AggregateSignatures(sigs)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate BLS signatures: %w", err)
	}
	
	// Serialize the aggregated signature to bytes
	aggSigBytes := aggSig.Bytes()
	
	log.Printf("BLS signature aggregated successfully: %s (length: %d)", 
		hex.EncodeToString(aggSigBytes), len(aggSigBytes))
	
	return aggSigBytes, nil
}
