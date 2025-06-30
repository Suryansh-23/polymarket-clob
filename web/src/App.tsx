import { WagmiWrapper } from "./wagmi";
import OrderBook from "./components/OrderBook";
import DepthChart from "./components/DepthChart";
import VolumeChart from "./components/VolumeChart";
import EventLog from "./components/EventLog";

export default function App() {
  return (
    <WagmiWrapper>
      <div
        style={{
          minHeight: "100vh",
          backgroundColor: "#f5f5f5",
          padding: "20px",
        }}
      >
        <header
          style={{
            textAlign: "center",
            marginBottom: "24px",
            padding: "20px",
            backgroundColor: "#fff",
            borderRadius: "12px",
            boxShadow: "0 2px 8px rgba(0,0,0,0.1)",
          }}
        >
          <h1
            style={{
              margin: "0 0 8px 0",
              color: "#1976d2",
              fontSize: "28px",
              fontWeight: "bold",
            }}
          >
            📈 Polymarket CLOB Dashboard
          </h1>
          <p
            style={{
              margin: 0,
              color: "#666",
              fontSize: "16px",
            }}
          >
            Real-time monitoring of order book, trading activity, and on-chain
            events
          </p>
        </header>

        <div
          style={{
            display: "grid",
            gridTemplateColumns: "repeat(auto-fit, minmax(400px, 1fr))",
            gap: "20px",
            maxWidth: "1400px",
            margin: "0 auto",
          }}
        >
          <OrderBook />
          <DepthChart />
          <VolumeChart />
          <EventLog />
        </div>

        <footer
          style={{
            textAlign: "center",
            marginTop: "24px",
            padding: "16px",
            color: "#999",
            fontSize: "14px",
          }}
        >
          <div>
            🔗 Connected to:{" "}
            {import.meta.env.VITE_RPC_URL || "http://localhost:8545"}
          </div>
          <div style={{ marginTop: "4px" }}>
            Last updated: {new Date().toLocaleString()}
          </div>
        </footer>
      </div>
    </WagmiWrapper>
  );
}
