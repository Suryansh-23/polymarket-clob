# Polymarket CLOB - Web Dashboard

A real-time dashboard for monitoring the Polymarket CLOB sequencer, displaying live order book data, trading depth charts, volume metrics, and on-chain batch settlement events.

## Features

- **Order Placement Form**: Submit buy/sell orders with EIP-712 signing
- **Live Order Book**: Real-time bid/ask orders with price-time priority visualization
- **Depth Chart**: Market depth visualization showing liquidity at different price levels
- **Volume Chart**: Trading volume metrics and historical trends
- **Event Log**: Real-time on-chain BatchSubmitted events from the EigenLayer AVS
- **Batch Events Table**: Historical batch events with transaction links
- **Skeleton Loading**: Smooth loading states while fetching data

## Setup

```bash
cd web
npm install
```

## Environment Variables

Create a `.env` file in the `web/` directory:

```env
VITE_RPC_URL=http://localhost:8545
VITE_API_URL=http://localhost:8081
VITE_CONTRACT_ADDRESS=0x5FbDB2315678afecb367f032d93F642f64180aa3
VITE_DEMO_KEY=0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80
```

**Environment Variables:**

- `VITE_RPC_URL`: Ethereum RPC endpoint for blockchain connection
- `VITE_API_URL`: Sequencer HTTP API endpoint for order book data
- `VITE_CONTRACT_ADDRESS`: BatchSettlement contract address for events
- `VITE_DEMO_KEY`: Demo private key for order signing (test purposes only)

## Running

### Development Mode

```bash
# Start with default configuration
npm run dev

# With custom configuration
VITE_RPC_URL=https://rpc.testnet.io npm run dev
```

### Production Build

```bash
npm run build
npm run preview
```

## Using the Dashboard

### 1. Order Placement

- Use the order form at the top to submit buy/sell orders
- Enter price and quantity, select buy/sell side
- Orders are automatically signed with EIP-712 and submitted to sequencer
- Success/error messages will appear below the form

### 2. Real-time Monitoring

- **Order Book**: Shows live bids (green) and asks (red) with price-time priority
- **Depth Chart**: Visualizes cumulative liquidity at each price level
- **Volume Chart**: Displays 24-hour trading volume with hourly breakdown
- **Event Log**: Real-time blockchain events and historical batch table

### 3. Loading States

- Components show skeleton placeholders while loading data
- Smooth transitions when data becomes available
- Error states with fallback to mock data when API unavailable

## Architecture

- **React 18** with TypeScript and functional components
- **Vite** for fast development and building
- **Wagmi v2** for Ethereum integration with React Query
- **Recharts** for data visualization
- **Axios** for HTTP requests to sequencer API
- **Ethers v6** for wallet operations and signing

## API Integration

The dashboard connects to:

1. **Sequencer HTTP API** (localhost:8081):

   - `POST /orders` - Submit signed orders
   - `GET /book` - Fetch order book data
   - `GET /depth` - Market depth data
   - `GET /volume` - Trading volume metrics

2. **Ethereum RPC** for:

   - Real-time block monitoring
   - Contract event subscriptions
   - Historical event queries

3. **BatchSettlement Contract** for:
   - BatchSubmitted event monitoring
   - Transaction hash tracking
   - Block number references

## Development

```bash
# Start development server
npm run dev

# Type checking
npm run type-check

# Lint code
npm run lint

# Build for production
npm run build
```

The dashboard will be available at http://localhost:5173 with hot module replacement for rapid development.

## Demo Order Testing

The dashboard includes a demo order form that:

1. Uses the test private key from `VITE_DEMO_KEY`
2. Creates EIP-712 compatible order objects
3. Signs orders with ethers.js
4. Submits to the sequencer API
5. Updates the order book in real-time

**Note**: Only use the demo key for local testing. Never use real private keys in frontend code.
