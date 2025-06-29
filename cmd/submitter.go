package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

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

// BatchSettlement contract interface
type BatchSettlement struct {
	client   *ethclient.Client
	contract *bind.BoundContract
	address  common.Address
}

// NewBatchSettlement creates a new BatchSettlement contract instance
func NewBatchSettlement(address common.Address, client *ethclient.Client) (*BatchSettlement, error) {
	parsedABI, err := abi.JSON(strings.NewReader(batchSettlementABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}

	contract := bind.NewBoundContract(address, parsedABI, client, client, client)

	return &BatchSettlement{
		client:   client,
		contract: contract,
		address:  address,
	}, nil
}

// getEthClient creates an Ethereum client connection
func getEthClient() (*ethclient.Client, error) {
	rpcURL := os.Getenv("RPC_URL")
	if rpcURL == "" {
		// Default to a local devnet URL
		rpcURL = "http://localhost:8545"
	}

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum client: %w", err)
	}

	return client, nil
}

// getPrivateKey loads the private key from environment
func getPrivateKey() (*ecdsa.PrivateKey, error) {
	privateKeyHex := os.Getenv("PRIVATE_KEY")
	if privateKeyHex == "" {
		return nil, fmt.Errorf("PRIVATE_KEY environment variable not set")
	}

	// Remove 0x prefix if present
	if strings.HasPrefix(privateKeyHex, "0x") {
		privateKeyHex = privateKeyHex[2:]
	}

	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return privateKey, nil
}

// getContractAddress gets the BatchSettlement contract address
func getContractAddress() (common.Address, error) {
	contractAddr := os.Getenv("BATCH_SETTLEMENT_ADDRESS")
	if contractAddr == "" {
		// Try to read from devnet deployment artifacts
		contractAddr = "0x5FbDB2315678afecb367f032d93F642f64180aa3" // Default local deployment address
		log.Printf("Using default contract address: %s", contractAddr)
	}

	if !common.IsHexAddress(contractAddr) {
		return common.Address{}, fmt.Errorf("invalid contract address: %s", contractAddr)
	}

	return common.HexToAddress(contractAddr), nil
}

// SubmitBatch submits a batch to the BatchSettlement contract
func SubmitBatch(root string, fills []byte, aggSig []byte) (string, error) {
	log.Printf("Submitting batch - Root: %s, Fills length: %d, Signature length: %d",
		root, len(fills), len(aggSig))

	// Get Ethereum client
	client, err := getEthClient()
	if err != nil {
		return "", fmt.Errorf("failed to get eth client: %w", err)
	}
	defer client.Close()

	// Get private key
	privateKey, err := getPrivateKey()
	if err != nil {
		return "", fmt.Errorf("failed to get private key: %w", err)
	}

	// Get contract address
	contractAddr, err := getContractAddress()
	if err != nil {
		return "", fmt.Errorf("failed to get contract address: %w", err)
	}

	// Create contract instance
	batchContract, err := NewBatchSettlement(contractAddr, client)
	if err != nil {
		return "", fmt.Errorf("failed to create contract instance: %w", err)
	}

	// Get chain ID for transaction signing
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return "", fmt.Errorf("failed to get chain ID: %w", err)
	}

	// Create auth object
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return "", fmt.Errorf("failed to create auth: %w", err)
	}

	// Set gas limit and gas price
	auth.GasLimit = uint64(500000) // 500k gas limit
	
	// Get suggested gas price
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Printf("Failed to get suggested gas price, using default: %v", err)
		gasPrice = big.NewInt(20000000000) // 20 gwei
	}
	auth.GasPrice = gasPrice

	// Convert root string to bytes32
	rootBytes := common.HexToHash(root)

	log.Printf("Calling submitBatch with gas limit: %d, gas price: %s",
		auth.GasLimit, gasPrice.String())

	// Call submitBatch function
	tx, err := batchContract.contract.Transact(auth, "submitBatch", rootBytes, fills, aggSig)
	if err != nil {
		return "", fmt.Errorf("failed to submit batch transaction: %w", err)
	}

	log.Printf("Transaction submitted: %s", tx.Hash().Hex())

	// Wait for transaction to be mined (optional)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	receipt, err := bind.WaitMined(ctx, client, tx)
	if err != nil {
		log.Printf("Warning: failed to wait for transaction to be mined: %v", err)
		return tx.Hash().Hex(), nil // Return hash even if we can't wait for mining
	}

	if receipt.Status == 0 {
		return "", fmt.Errorf("transaction failed with status 0")
	}

	log.Printf("Transaction mined successfully in block %d", receipt.BlockNumber.Uint64())

	return tx.Hash().Hex(), nil
}