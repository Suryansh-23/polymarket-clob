# Sequencer Service

The Polymarket CLOB Sequencer service provides a decentralized order matching and batch settlement system built on EigenLayer's AVS infrastructure.

## Overview

The sequencer ingests EIP-712 signed orders, performs price-time priority matching, builds Merkle trees of fills, aggregates BLS signatures from operators, and submits batches to the on-chain BatchSettlement contract.

## Components

- **main.go**: HTTP server entrypoint that accepts order submissions on port 8081
- **matcher/**: Order matching engine package with price-time priority and Merkle tree construction
- **submitter/**: Ethereum transaction submission package for BatchSettlement contract

## Quick Start

```bash
# Navigate to the cmd directory
cd cmd

# Install dependencies
go mod tidy

# Set environment variables
export PRIVATE_KEY="your_private_key_here"
export RPC_URL="http://localhost:8545"  # or your RPC endpoint
export BATCH_SETTLEMENT_ADDRESS="0x5FbDB2315678afecb367f032d93F642f64180aa3"  # contract address

# Run the sequencer
go run main.go
```

## API Endpoints

### POST /orders

Submit a new order to the orderbook.

**Request Body:**

```json
{
  "maker": "0x742b35Cc6834C532532fa5A32b66F8d6C1F3b0B1",
  "takerAsset": "0x2e8a51B19f2bbE1FfA3d3F14D7E72F1C00E28Ef5",
  "makeAmount": "1000.0",
  "takeAmount": "500.0",
  "price": 0.5,
  "timestamp": 1719734400,
  "signature": "0x..."
}
```

**Response:**

```json
{
  "success": true
}
```

### GET /health

Health check endpoint.

**Response:**

```json
{
  "status": "healthy"
}
```

## Environment Variables

- `PRIVATE_KEY`: Ethereum private key for transaction signing (required)
- `RPC_URL`: Ethereum RPC endpoint (default: http://localhost:8545)
- `BATCH_SETTLEMENT_ADDRESS`: BatchSettlement contract address
- `BLS_KEYS`: Comma-separated list of BLS private keys in hex format for operator signing (optional)

### BLS Key Configuration

The sequencer supports **real BLS signature aggregation** using operator private keys from the EigenLayer crypto-libs. Configure with:

```bash
export BLS_KEYS="0x123abc...,0x456def...,0x789ghi..."
```

**Key Generation**: Use the Hourglass BN254 keygen tool to generate test keys:

```bash
# Generate a BN254 private key for testing
cd .devkit/contracts/lib/hourglass-monorepo/ponos/cmd/keygen
go run main.go generate --curve-type bn254 --output-dir ./keys

# Extract hex private key for BLS_KEYS environment variable
go run main.go info --key-file ./keys/bn254-0.key
```

**Production Keys**: In production, use proper operator key management:

```bash
# Generate multiple operator keys for multi-signature
go run main.go generate --curve-type bn254 --output-dir ./operator1
go run main.go generate --curve-type bn254 --output-dir ./operator2
go run main.go generate --curve-type bn254 --output-dir ./operator3

# Combine hex private keys into BLS_KEYS
export BLS_KEYS="0xkey1...,0xkey2...,0xkey3..."
```

**Fallback Behavior**: If `BLS_KEYS` is not set, the system automatically falls back to mock BLS signatures for development and testing.

## Order Matching Logic

The enhanced matching engine supports **multiple fills per batch** and **partial order fills**:

1. **Multi-Fill Processing**: Orders are sorted by descending price, then ascending timestamp (price-time priority)
2. **Bid/Ask Separation**: Orders are automatically classified as bids (buyers) or asks (sellers)
3. **Cross-Price Matching**: Bids and asks are matched when bid price ≥ ask price
4. **Partial Fills**: Orders can be partially filled across multiple batches
5. **Batch Size Limiting**: Maximum number of fills per batch (default: 100)
6. **Order Book Pruning**: Fully filled orders are removed, partially filled orders remain with updated amounts

### Multi-Fill Algorithm:

```
1. Sort all orders by price-time priority
2. Split into bids (buyers) and asks (sellers)
3. Match bids vs asks while bid_price ≥ ask_price:
   - Calculate fill_qty = min(bid.makeAmount, ask.takeAmount)
   - Create fill record with order hashes
   - Reduce both order amounts by fill_qty
   - Advance to next order if current order is fully filled
4. Build Merkle tree over all fills in batch
5. Return remaining orders with updated amounts
```

### Enhanced MatchAndBatch Signature:

```go
func MatchAndBatch(orders []Order, maxBatch int) (root string, fillsBytes []byte, remaining []Order, err error)
```

- **orders**: Input order book
- **maxBatch**: Maximum fills per batch (prevents large batches)
- **root**: Merkle root of all fills in batch
- **fillsBytes**: JSON-serialized fill records
- **remaining**: Updated order book after matching
- **err**: Any matching errors

## Development

The service is designed to work with the Hourglass AVS template and integrates with:

- EigenLayer operator infrastructure
- BLS signature aggregation
- On-chain dispute resolution via DisputeGame contract

For production deployment, ensure proper operator key management and BLS signature collection from the operator quorum.
