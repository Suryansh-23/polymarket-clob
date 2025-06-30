import {
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Area,
  AreaChart,
} from "recharts";
import { useState, useEffect } from "react";
import axios from "axios";

interface DepthData {
  price: number;
  bidDepth: number;
  askDepth: number;
}

interface DepthResponse {
  depths: DepthData[];
  timestamp: number;
}

interface Order {
  id: string;
  price: number;
  amount: number;
  timestamp: number;
  side: "bid" | "ask";
}

export default function DepthChart() {
  const [depthData, setDepthData] = useState<DepthData[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const computeDepthFromOrders = (orders: Order[]): DepthData[] => {
    const bids = orders
      .filter((o) => o.side === "bid")
      .sort((a, b) => b.price - a.price);
    const asks = orders
      .filter((o) => o.side === "ask")
      .sort((a, b) => a.price - b.price);

    const depthMap = new Map<number, { bidDepth: number; askDepth: number }>();

    // Calculate cumulative bid depth
    let cumulativeBidDepth = 0;
    for (const bid of bids) {
      cumulativeBidDepth += bid.amount;
      depthMap.set(bid.price, {
        bidDepth: cumulativeBidDepth,
        askDepth: depthMap.get(bid.price)?.askDepth || 0,
      });
    }

    // Calculate cumulative ask depth
    let cumulativeAskDepth = 0;
    for (const ask of asks) {
      cumulativeAskDepth += ask.amount;
      const existing = depthMap.get(ask.price);
      depthMap.set(ask.price, {
        bidDepth: existing?.bidDepth || 0,
        askDepth: cumulativeAskDepth,
      });
    }

    // Convert to array and sort by price
    return Array.from(depthMap.entries())
      .map(([price, depths]) => ({
        price,
        bidDepth: depths.bidDepth,
        askDepth: depths.askDepth,
      }))
      .sort((a, b) => a.price - b.price);
  };

  const fetchDepthData = async () => {
    try {
      setError(null);
      const apiUrl = import.meta.env.VITE_API_URL || "http://localhost:8081";

      // Try to fetch precomputed depth data first
      try {
        const depthResponse = await axios.get<DepthResponse>(
          `${apiUrl}/depth`,
          {
            timeout: 3000,
          }
        );
        setDepthData(depthResponse.data.depths);
        setLoading(false);
        return;
      } catch (depthError) {
        // If depth endpoint doesn't exist, compute from order book
        console.log("Depth endpoint not available, computing from order book");
      }

      // Fallback: fetch order book and compute depth
      const orderBookResponse = await axios.get<{
        bids: Order[];
        asks: Order[];
      }>(`${apiUrl}/book`, {
        timeout: 5000,
      });

      const allOrders = [
        ...orderBookResponse.data.bids,
        ...orderBookResponse.data.asks,
      ];
      const computedDepth = computeDepthFromOrders(allOrders);
      setDepthData(computedDepth);
      setLoading(false);
    } catch (err) {
      console.error("Failed to fetch depth data:", err);
      setError("Failed to load depth data");

      // Fallback to mock data
      const mockDepthData: DepthData[] = [
        { price: 1.2, bidDepth: 15000, askDepth: 0 },
        { price: 1.21, bidDepth: 12000, askDepth: 0 },
        { price: 1.22, bidDepth: 8000, askDepth: 0 },
        { price: 1.23, bidDepth: 5000, askDepth: 0 },
        { price: 1.24, bidDepth: 2000, askDepth: 0 },
        { price: 1.25, bidDepth: 0, askDepth: 2500 },
        { price: 1.26, bidDepth: 0, askDepth: 5500 },
        { price: 1.27, bidDepth: 0, askDepth: 8200 },
        { price: 1.28, bidDepth: 0, askDepth: 11000 },
        { price: 1.29, bidDepth: 0, askDepth: 14500 },
      ];
      setDepthData(mockDepthData);
      setLoading(false);
    }
  };

  useEffect(() => {
    // Initial fetch
    fetchDepthData();

    // Set up polling every 2 seconds (less frequent than order book)
    const interval = setInterval(fetchDepthData, 2000);

    return () => clearInterval(interval);
  }, []);

  if (loading && depthData.length === 0) {
    return (
      <div
        style={{
          border: "1px solid #e0e0e0",
          borderRadius: "8px",
          padding: "16px",
          backgroundColor: "#f9f9f9",
          height: "300px",
        }}
      >
        <h3 style={{ margin: "0 0 16px 0", color: "#333" }}>📈 Market Depth</h3>
        <div style={{ textAlign: "center", color: "#666", marginTop: "80px" }}>
          <div>🔄 Loading depth chart...</div>
          <div style={{ fontSize: "11px", marginTop: "4px" }}>
            Computing from order book data
          </div>
        </div>
      </div>
    );
  }

  return (
    <div
      style={{
        border: "1px solid #e0e0e0",
        borderRadius: "8px",
        padding: "16px",
        backgroundColor: "#f9f9f9",
        height: "300px",
      }}
    >
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          marginBottom: "16px",
        }}
      >
        <h3 style={{ margin: 0, color: "#333" }}>📈 Market Depth</h3>
        <div style={{ fontSize: "11px" }}>
          {error ? (
            <span style={{ color: "#d32f2f" }}>⚠️ {error}</span>
          ) : (
            <span style={{ color: "#388e3c" }}>🟢 Live</span>
          )}
        </div>
      </div>

      <ResponsiveContainer width="100%" height="85%">
        <AreaChart data={depthData}>
          <CartesianGrid strokeDasharray="3 3" />
          <XAxis
            dataKey="price"
            tickFormatter={(value) => `$${value.toFixed(2)}`}
            fontSize={11}
          />
          <YAxis
            tickFormatter={(value) => `${(value / 1000).toFixed(0)}k`}
            fontSize={11}
          />
          <Tooltip
            formatter={(value: number, name: string) => [
              `${(value / 1000).toFixed(1)}k`,
              name === "bidDepth" ? "Bid Depth" : "Ask Depth",
            ]}
            labelFormatter={(label) => `Price: $${label}`}
          />
          <Area
            type="stepAfter"
            dataKey="bidDepth"
            stackId="1"
            stroke="#388e3c"
            fill="#388e3c"
            fillOpacity={0.6}
            name="Bid Depth"
          />
          <Area
            type="stepBefore"
            dataKey="askDepth"
            stackId="2"
            stroke="#d32f2f"
            fill="#d32f2f"
            fillOpacity={0.6}
            name="Ask Depth"
          />
        </AreaChart>
      </ResponsiveContainer>
    </div>
  );
}
