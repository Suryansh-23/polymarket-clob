package submitter

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Contract ABI for BatchSettlement.submitBatch function
const batchSettlementABI = `[
	{
		"inputs": [
			{"name": "root", "type": "bytes32"},
			{"name": "fills", "type": "bytes"},
			{"name": "aggSig", "type": "bytes"}
		],
		"name": "submitBatch",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	}
]`

// Global configuration variables
var (
	ethClient    *ethclient.Client
	contractAddr common.Address
	contractABI  abi.ABI
	privateKey   *ecdsa.PrivateKey
	maxRetries   int
	backoffMS    int
	
	// Durable queue for failed batches
	failedBatches []FailedBatch
	failedMutex   sync.RWMutex
)

// FailedBatch represents a batch that failed to submit
type FailedBatch struct {
	Root      string
	Fills     []byte
	Sig       []byte
	Timestamp time.Time
	Attempts  int
}

// init initializes the submitter package with environment configuration
func init() {
	var err error
	
	// Initialize Ethereum client
	rpcURL := os.Getenv("RPC_URL")
	if rpcURL == "" {
		rpcURL = "http://localhost:8545"
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	ethClient, err = ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		log.Fatalf("Failed to connect to Ethereum client at %s: %v", rpcURL, err)
	}
	
	// Parse contract address
	contractAddrStr := os.Getenv("CONTRACT_ADDRESS")
	if contractAddrStr == "" {
		// Fallback to legacy env var name
		contractAddrStr = os.Getenv("BATCH_SETTLEMENT_ADDRESS")
	}
	if contractAddrStr == "" {
		contractAddrStr = "0x5FbDB2315678afecb367f032d93F642f64180aa3" // Default local
		log.Printf("Warning: CONTRACT_ADDRESS not set, using default: %s", contractAddrStr)
	}
	
	if !common.IsHexAddress(contractAddrStr) {
		log.Fatalf("Invalid CONTRACT_ADDRESS: %s", contractAddrStr)
	}
	contractAddr = common.HexToAddress(contractAddrStr)
	
	// Parse contract ABI
	contractABI, err = abi.JSON(strings.NewReader(batchSettlementABI))
	if err != nil {
		log.Fatalf("Failed to parse contract ABI: %v", err)
	}
	
	// Load private key
	privateKeyHex := os.Getenv("PRIVATE_KEY")
	if privateKeyHex == "" {
		log.Fatalf("PRIVATE_KEY environment variable not set")
	}
	
	privateKeyHex = strings.TrimPrefix(privateKeyHex, "0x")
	
	privateKey, err = crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		log.Fatalf("Failed to parse PRIVATE_KEY: %v", err)
	}
	
	// Parse retry configuration
	maxRetriesStr := os.Getenv("MAX_RETRIES")
	if maxRetriesStr == "" {
		maxRetries = 5
	} else {
		maxRetries, err = strconv.Atoi(maxRetriesStr)
		if err != nil || maxRetries < 1 {
			log.Fatalf("Invalid MAX_RETRIES: %s (must be positive integer)", maxRetriesStr)
		}
	}
	
	backoffMSStr := os.Getenv("BACKOFF_MS")
	if backoffMSStr == "" {
		backoffMS = 200
	} else {
		backoffMS, err = strconv.Atoi(backoffMSStr)
		if err != nil || backoffMS < 50 {
			log.Fatalf("Invalid BACKOFF_MS: %s (must be >= 50)", backoffMSStr)
		}
	}
	
	log.Printf("Submitter initialized - RPC: %s, Contract: %s, MaxRetries: %d, Backoff: %dms",
		rpcURL, contractAddr.Hex(), maxRetries, backoffMS)
}

// SubmitBatch submits a batch to the BatchSettlement contract with retry logic
func SubmitBatch(root string, fills []byte, aggSig []byte) (string, error) {
	log.Printf("Submitting batch - Root: %s, Fills length: %d, Signature length: %d",
		root, len(fills), len(aggSig))

	for attempt := 1; attempt <= maxRetries; attempt++ {
		txHash, err := attemptSubmitBatch(root, fills, aggSig)
		if err == nil {
			log.Printf("✅ Batch submitted successfully on attempt %d: https://explorer.testnet.io/tx/%s", 
				attempt, txHash)
			return txHash, nil
		}
		
		log.Printf("❌ Attempt %d/%d failed: %v", attempt, maxRetries, err)
		
		if attempt < maxRetries {
			backoffDuration := time.Duration(backoffMS*attempt) * time.Millisecond
			log.Printf("⏳ Waiting %v before retry %d...", backoffDuration, attempt+1)
			time.Sleep(backoffDuration)
		}
	}
	
	// All retries failed - add to durable queue
	failedBatch := FailedBatch{
		Root:      root,
		Fills:     fills,
		Sig:       aggSig,
		Timestamp: time.Now(),
		Attempts:  maxRetries,
	}
	
	failedMutex.Lock()
	failedBatches = append(failedBatches, failedBatch)
	queueLength := len(failedBatches)
	failedMutex.Unlock()
	
	log.Printf("🚨 Batch submission failed after %d attempts. Root: %s, Queue length: %d", 
		maxRetries, root, queueLength)
	
	return "", fmt.Errorf("batch submission failed after %d attempts", maxRetries)
}

// attemptSubmitBatch makes a single attempt to submit a batch
func attemptSubmitBatch(root string, fills []byte, aggSig []byte) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	
	// Convert root to bytes32
	rootHash := common.HexToHash(root)
	
	// Pack transaction data for gas estimation
	data, err := contractABI.Pack("submitBatch", rootHash, fills, aggSig)
	if err != nil {
		return "", fmt.Errorf("failed to pack transaction data: %w", err)
	}
	
	// Get pending nonce for the account
	fromAddr := crypto.PubkeyToAddress(privateKey.PublicKey)
	nonce, err := ethClient.PendingNonceAt(ctx, fromAddr)
	if err != nil {
		return "", fmt.Errorf("failed to get pending nonce: %w", err)
	}
	
	// Estimate gas for the transaction
	gasEstimate, err := ethClient.EstimateGas(ctx, ethereum.CallMsg{
		From: fromAddr,
		To:   &contractAddr,
		Data: data,
	})
	if err != nil {
		return "", fmt.Errorf("failed to estimate gas: %w", err)
	}
	
	// Apply 20% buffer to gas estimate
	gasLimit := uint64(float64(gasEstimate) * 1.2)
	
	// Get suggested gas price
	gasPrice, err := ethClient.SuggestGasPrice(ctx)
	if err != nil {
		log.Printf("Failed to get suggested gas price, using default: %v", err)
		gasPrice = big.NewInt(20000000000) // 20 gwei fallback
	}
	
	// Get chain ID for transaction signing
	chainID, err := ethClient.NetworkID(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get chain ID: %w", err)
	}
	
	// Create auth object with all parameters
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return "", fmt.Errorf("failed to create auth: %w", err)
	}
	
	auth.Nonce = big.NewInt(int64(nonce))
	auth.GasLimit = gasLimit
	auth.GasPrice = gasPrice
	auth.Context = ctx
	
	log.Printf("📤 Submitting transaction - Nonce: %d, Gas: %d, Price: %s wei", 
		nonce, gasLimit, gasPrice.String())
	
	// Create bound contract and submit transaction
	contract := bind.NewBoundContract(contractAddr, contractABI, ethClient, ethClient, ethClient)
	tx, err := contract.Transact(auth, "submitBatch", rootHash, fills, aggSig)
	if err != nil {
		return "", fmt.Errorf("failed to submit transaction: %w", err)
	}
	
	log.Printf("🚀 Transaction sent: %s (nonce: %d)", tx.Hash().Hex(), nonce)
	
	// Wait for transaction to be mined with timeout
	receipt, err := bind.WaitMined(ctx, ethClient, tx)
	if err != nil {
		// Return the tx hash even if we can't wait for confirmation
		log.Printf("⚠️  Transaction submitted but couldn't wait for confirmation: %v", err)
		return tx.Hash().Hex(), nil
	}
	
	if receipt.Status == 0 {
		return "", fmt.Errorf("transaction failed with status 0 (reverted)")
	}
	
	log.Printf("⛏️  Transaction mined in block %d, gas used: %d", 
		receipt.BlockNumber.Uint64(), receipt.GasUsed)
	
	return tx.Hash().Hex(), nil
}

// RetryFailedBatches attempts to resubmit all failed batches
func RetryFailedBatches() error {
	failedMutex.Lock()
	if len(failedBatches) == 0 {
		failedMutex.Unlock()
		log.Printf("No failed batches to retry")
		return nil
	}
	
	// Make a copy to avoid holding the lock during network calls
	batchesToRetry := make([]FailedBatch, len(failedBatches))
	copy(batchesToRetry, failedBatches)
	failedMutex.Unlock()
	
	log.Printf("🔄 Retrying %d failed batches...", len(batchesToRetry))
	
	var successCount, failCount int
	var successfulIndices []int
	
	for i, batch := range batchesToRetry {
		log.Printf("Retrying batch %d/%d (Root: %s, Previous attempts: %d)", 
			i+1, len(batchesToRetry), batch.Root, batch.Attempts)
		
		// Try to submit with exponential backoff
		for attempt := 1; attempt <= maxRetries; attempt++ {
			txHash, err := attemptSubmitBatch(batch.Root, batch.Fills, batch.Sig)
			if err == nil {
				log.Printf("✅ Retry successful for batch %s: https://explorer.testnet.io/tx/%s", 
					batch.Root, txHash)
				successfulIndices = append(successfulIndices, i)
				successCount++
				break
			}
			
			log.Printf("❌ Retry attempt %d/%d failed for batch %s: %v", 
				attempt, maxRetries, batch.Root, err)
			
			if attempt < maxRetries {
				backoffDuration := time.Duration(backoffMS*attempt) * time.Millisecond
				time.Sleep(backoffDuration)
			}
		}
		
		if len(successfulIndices) == 0 || successfulIndices[len(successfulIndices)-1] != i {
			failCount++
		}
	}
	
	// Remove successful batches from the failed queue
	if len(successfulIndices) > 0 {
		failedMutex.Lock()
		// Remove in reverse order to maintain indices
		for i := len(successfulIndices) - 1; i >= 0; i-- {
			idx := successfulIndices[i]
			if idx < len(failedBatches) {
				failedBatches = append(failedBatches[:idx], failedBatches[idx+1:]...)
			}
		}
		failedMutex.Unlock()
	}
	
	log.Printf("🔄 Retry completed - Success: %d, Failed: %d, Remaining in queue: %d", 
		successCount, failCount, len(failedBatches))
	
	if failCount > 0 {
		return fmt.Errorf("failed to retry %d batches", failCount)
	}
	
	return nil
}

// GetFailedBatchesCount returns the number of batches in the failed queue
func GetFailedBatchesCount() int {
	failedMutex.RLock()
	defer failedMutex.RUnlock()
	return len(failedBatches)
}

// GetFailedBatches returns a copy of all failed batches for inspection
func GetFailedBatches() []FailedBatch {
	failedMutex.RLock()
	defer failedMutex.RUnlock()
	
	batches := make([]FailedBatch, len(failedBatches))
	copy(batches, failedBatches)
	return batches
}

// ClearFailedBatches removes all failed batches from the queue (use with caution)
func ClearFailedBatches() int {
	failedMutex.Lock()
	defer failedMutex.Unlock()
	
	count := len(failedBatches)
	failedBatches = failedBatches[:0]
	
	log.Printf("🗑️  Cleared %d failed batches from queue", count)
	return count
}
