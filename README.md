# Polymarket CLOB Sequencer AVS

A decentralized order-matching sequencer for Polymarket's Central Limit Order Book (CLOB) built on EigenLayer's AVS infrastructure using the Hourglass framework.

## Overview

This project implements a slashable, stake-weighted operator set that replaces Polymarket's centralized order-matching server. The sequencer sorts EIP-712 orders by price-time priority, builds Merkle trees of matched fills, aggregates BLS signatures from operators, and submits batches on-chain for settlement and dispute resolution.

## Architecture

### Smart Contracts (`contracts/`)

- **BatchSettlement.sol**: Handles batch settlement with BLS signature verification for CLOB orders
- **DisputeGame.sol**: Dispute resolution for order sequencing violations with operator slashing
- **TaskAVSRegistrar.sol**: Operator registration and management (inherited from Hourglass)
- **AVSTaskHook.sol**: Task validation and fee markets (inherited from Hourglass)

### Sequencer Service (`cmd/`)

- **server.go**: HTTP API server for order ingestion on port 8081
- **matcher.go**: Price-time priority matching engine with Merkle tree construction
- **submitter.go**: Ethereum transaction submission to BatchSettlement contract
- **main.go**: Service entrypoint with graceful shutdown

## Quick Start

### Prerequisites

- [Docker (latest)](https://docs.docker.com/engine/install/)
- [Foundry (latest)](https://book.getfoundry.sh/getting-started/installation)
- [Go (v1.23.6)](https://go.dev/doc/install)
- [EigenLayer DevKit CLI](https://github.com/Layr-Labs/devkit-cli)

### 1. Setup Environment

```bash
# Copy environment template
cp .env.example .env

# Edit .env and set your Holesky RPC endpoints
# L1_FORK_URL=https://your-holesky-rpc-url
# L2_FORK_URL=https://your-holesky-rpc-url
```

### 2. Build and Deploy

```bash
# Build contracts and binaries
devkit avs build

# Start local devnet (deploys contracts, registers operators)
devkit avs devnet start
```

### 3. Run the Sequencer Service

```bash
# Navigate to sequencer directory
cd cmd

# Set environment variables for the sequencer
export PRIVATE_KEY="your_private_key_here"
export RPC_URL="http://localhost:8545"
export BATCH_SETTLEMENT_ADDRESS="contract_address_from_deployment"

# Install dependencies and run
go mod tidy
go run *.go
```

### 4. Test the API

```bash
# Make the test script executable and run it
chmod +x test_api.sh
./test_api.sh
```

## API Reference

### POST /orders

Submit EIP-712 signed orders to the sequencer.

**Request:**

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

Health check endpoint for monitoring.

## Order Matching Logic

1. **Price-Time Priority**: Orders sorted by descending price, then ascending timestamp
2. **Quantity Matching**: Fill quantity = min(makeAmount, takeAmount)
3. **Merkle Tree Construction**: Builds tree over (makerHash, takerHash, quantity) tuples
4. **BLS Aggregation**: Collects signatures from ≥2/3 stake-weighted operators
5. **Batch Submission**: Submits (root, fills, aggregatedSignature) to BatchSettlement contract

## Dispute Resolution

The DisputeGame contract enables fraud proofs for incorrect order sequencing:

1. **Dispute Submission**: Anyone can dispute a batch root with Merkle proof of misordering
2. **Validation**: Contract verifies proof and checks price-time priority violations
3. **Slashing**: Violating operators are slashed via AllocationManager integration
4. **Incentives**: Successful disputes are rewarded, false disputes are penalized

## Environment Variables

| Variable                   | Description                                   | Default                 |
| -------------------------- | --------------------------------------------- | ----------------------- |
| `PRIVATE_KEY`              | Ethereum private key for signing transactions | Required                |
| `RPC_URL`                  | Ethereum RPC endpoint                         | `http://localhost:8545` |
| `BATCH_SETTLEMENT_ADDRESS` | BatchSettlement contract address              | `0x5FbDB...`            |

## Docker Deployment

```bash
# Build Docker image
docker build -t polymarket-clob-sequencer ./cmd

# Run container
docker run -p 8081:8081 \
  -e PRIVATE_KEY=your_key \
  -e RPC_URL=your_rpc \
  -e BATCH_SETTLEMENT_ADDRESS=contract_addr \
  polymarket-clob-sequencer
```

## Development

### Testing Contracts

```bash
cd contracts
forge test
```

### Testing Sequencer

```bash
cd cmd
go test ./...
```

### Simulate AVS Tasks

```bash
devkit avs call --signature="(uint256,string)" args='(5,"hello")'
```

## EIP-712 Order Schema

Orders follow Polymarket's standard EIP-712 schema. See:

- [Polymarket Order Documentation](https://docs.polymarket.com/developers/CLOB/orders/orders)
- [EIP-712 Specification](https://eips.ethereum.org/EIPS/eip-712)

## Security Considerations

- **Operator Slashing**: Incorrect sequencing results in stake slashing
- **BLS Verification**: Requires ≥2/3 stake-weighted signatures for batch validity
- **Dispute Window**: Time-bounded dispute resolution for finality
- **Merkle Proofs**: Cryptographic verification of order inclusion and sequencing

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/new-feature`
3. Commit changes: `git commit -am 'Add new feature'`
4. Push to branch: `git push origin feature/new-feature`
5. Submit a Pull Request

## License

This project is licensed under the BUSL-1.1 License - see the [LICENSE](LICENSE) file for details.

---

## ⚠️ Warning: This is Alpha, non audited code ⚠️

This implementation is in active development and has not been audited. Use at your own risk in production environments.
