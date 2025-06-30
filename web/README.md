# Polymarket CLOB - Web Dashboard

A real-time dashboard for monitoring the Polymarket CLOB sequencer, displaying live order book data, trading depth charts, volume metrics, and on-chain batch settlement events.

## Features

- **Live Order Book**: Real-time bid/ask orders with price-time priority visualization
- **Depth Chart**: Market depth visualization showing liquidity at different price levels
- **Volume Chart**: Trading volume metrics and historical trends
- **Event Log**: On-chain BatchSubmitted events from the EigenLayer AVS

## Setup

```bash
cd web
npm install
```

## Running

### Development Mode

```bash
# Using default local RPC
npm run dev

# With custom RPC endpoint
VITE_RPC_URL=http://localhost:8545 npm run dev

# With testnet RPC
VITE_RPC_URL=https://rpc.testnet.io npm run dev
```

### Production Build

```bash
npm run build
npm run preview
```

## Environment Variables

- `VITE_RPC_URL`: Ethereum RPC endpoint (default: http://localhost:8545)

## Architecture

- **React 18** with TypeScript
- **Vite** for fast development and building
- **Wagmi** for Ethereum integration
- **Recharts** for data visualization
- **Axios** for HTTP requests to sequencer API

## API Integration

The dashboard connects to:

1. **Sequencer HTTP API** (http://localhost:8081) for order book data
2. **Ethereum RPC** for on-chain event monitoring
3. **BatchSettlement Contract** for batch submission events

## Development

```bash
# Start development server
npm run dev

# Type checking
npm run type-check

# Lint
npm run lint

# Build for production
npm run build
```

The dashboard will be available at http://localhost:5173 with hot module replacement for rapid development.
