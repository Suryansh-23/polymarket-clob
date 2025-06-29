# Sequencer Service

The Polymarket CLOB Sequencer service provides a decentralized order matching and batch settlement system built on EigenLayer's AVS infrastructure.

## Overview

The sequencer ingests EIP-712 signed orders, performs price-time priority matching, builds Merkle trees of fills, aggregates BLS signatures from operators, and submits batches to the on-chain BatchSettlement contract.

## Components

- **server.go**: HTTP server that accepts order submissions on port 8081
- **matcher.go**: Order matching engine with price-time priority and Merkle tree construction
- **submitter.go**: Ethereum transaction submission to BatchSettlement contract
- **main.go**: Service entrypoint with graceful shutdown handling

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
go run *.go
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

## Order Matching Logic

1. Orders are sorted by descending price, then ascending timestamp (price-time priority)
2. Top two orders are matched if available
3. Fill quantity is calculated as minimum of makeAmount and takeAmount
4. Merkle tree is built over fills
5. BLS signatures are aggregated from operators
6. Batch is submitted to the BatchSettlement contract

## Development

The service is designed to work with the Hourglass AVS template and integrates with:

- EigenLayer operator infrastructure
- BLS signature aggregation
- On-chain dispute resolution via DisputeGame contract

For production deployment, ensure proper operator key management and BLS signature collection from the operator quorum.
