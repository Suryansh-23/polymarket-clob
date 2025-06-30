import { WagmiWrapper } from "./wagmi";
import OrderForm from "./components/OrderForm";
import OrderBook from "./components/OrderBook";
import DepthChart from "./components/DepthChart";
import VolumeChart from "./components/VolumeChart";
import EventLog from "./components/EventLog";
import { tradingSimulator } from "./services/TradingSimulator";
import { useEffect, useState } from "react";

export default function App() {
  const [simulatorActive, setSimulatorActive] = useState(false);
  const [marketPrice, setMarketPrice] = useState(1.25);

  useEffect(() => {
    // Start the trading simulator with faster interval
    tradingSimulator.start(1000 + Math.random() * 1000); // Random interval 1-2 seconds (faster)
    setSimulatorActive(true);

    // Update market price periodically
    const priceInterval = setInterval(() => {
      setMarketPrice(tradingSimulator.getCurrentMarketPrice());
    }, 1000);

    return () => {
      tradingSimulator.stop();
      clearInterval(priceInterval);
    };
  }, []);

  const toggleSimulator = () => {
    if (simulatorActive) {
      tradingSimulator.stop();
      setSimulatorActive(false);
    } else {
      tradingSimulator.start(1000 + Math.random() * 1000); // Faster interval here too
      setSimulatorActive(true);
    }
  };

  return (
    <WagmiWrapper>
      <div
        style={{
          minHeight: "100vh",
          backgroundColor: "#0d1421",
          color: "#ffffff",
          fontFamily: "'Inter', -apple-system, BlinkMacSystemFont, sans-serif",
        }}
      >
        {/* Top Header */}
        <header
          style={{
            borderBottom: "1px solid #1e2329",
            padding: "12px 24px",
            backgroundColor: "#1e2329",
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
          }}
        >
          <div style={{ display: "flex", alignItems: "center" }}>
            <h1
              style={{
                margin: 0,
                color: "#f0b90b",
                fontSize: "24px",
                fontWeight: "600",
              }}
            >
              📈 Polymarket CLOB
            </h1>
            <span
              style={{
                marginLeft: "16px",
                color: "#848e9c",
                fontSize: "14px",
              }}
            >
              Decentralized Exchange
            </span>
          </div>
          <div
            style={{
              display: "flex",
              alignItems: "center",
              gap: "16px",
              fontSize: "12px",
              color: "#848e9c",
            }}
          >
            <div style={{ display: "flex", alignItems: "center", gap: "8px" }}>
              <div
                style={{
                  width: "8px",
                  height: "8px",
                  borderRadius: "50%",
                  backgroundColor: "#02c076",
                }}
              />
              <span>
                Connected to {import.meta.env.VITE_RPC_URL || "localhost:8545"}
              </span>
            </div>
            <div>Market: ${marketPrice.toFixed(4)}</div>
            <div>Block: #{new Date().getTime() % 100000}</div>
            <button
              onClick={toggleSimulator}
              style={{
                fontSize: "11px",
                padding: "4px 8px",
                backgroundColor: simulatorActive ? "#f84960" : "#02c076",
                color: "#ffffff",
                border: "1px solid transparent",
                borderRadius: "4px",
                cursor: "pointer",
              }}
            >
              {simulatorActive ? "Stop Bot" : "Start Bot"}
            </button>
          </div>
        </header>

        {/* Main Trading Interface */}
        <div
          style={{
            display: "grid",
            gridTemplateColumns: "280px 1fr 320px",
            height: "calc(100vh - 61px)", // Subtract header height
            gap: "0px",
            backgroundColor: "#1e2329",
          }}
        >
          {/* Left Sidebar - Order Form */}
          <div
            style={{
              backgroundColor: "#0d1421",
              padding: "20px",
              borderRight: "1px solid #1e2329",
              overflowY: "auto",
            }}
          >
            <OrderForm />
          </div>

          {/* Center Content */}
          <div
            style={{
              display: "grid",
              gridTemplateRows: "2fr 1fr", // Give more space to charts (2/3) and less to events (1/3)
              gap: "0px",
            }}
          >
            {/* Top Charts */}
            <div
              style={{
                backgroundColor: "#0d1421",
                padding: "20px",
                display: "grid",
                gridTemplateColumns: "1fr 1fr",
                gap: "20px",
                borderBottom: "1px solid #1e2329",
              }}
            >
              <DepthChart />
              <VolumeChart />
            </div>

            {/* Bottom Events */}
            <div
              style={{
                backgroundColor: "#0d1421",
                padding: "20px",
                overflowY: "auto",
              }}
            >
              <EventLog />
            </div>
          </div>

          {/* Right Sidebar - Order Book */}
          <div
            style={{
              backgroundColor: "#0d1421",
              padding: "20px",
              borderLeft: "1px solid #1e2329",
              overflowY: "auto",
            }}
          >
            <OrderBook />
          </div>
        </div>
      </div>
    </WagmiWrapper>
  );
}
