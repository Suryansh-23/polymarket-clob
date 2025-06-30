import { useState, useEffect } from "react";
import axios from "axios";

interface Order {
  id: string;
  price: number;
  amount: number;
  timestamp: number;
  side: "bid" | "ask";
}

interface OrderBookResponse {
  bids: Order[];
  asks: Order[];
  timestamp: number;
}

export default function OrderBook() {
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [lastUpdate, setLastUpdate] = useState<Date>(new Date());

  const fetchOrderBook = async () => {
    try {
      setError(null);
      const apiUrl = import.meta.env.VITE_API_URL || "http://localhost:8081";
      const response = await axios.get<OrderBookResponse>(`${apiUrl}/book`, {
        timeout: 5000,
      });

      const allOrders = [...response.data.bids, ...response.data.asks];
      setOrders(allOrders);
      setLastUpdate(new Date());
      setLoading(false);
    } catch (err) {
      console.error("Failed to fetch order book:", err);
      setError("Failed to load order book");
      setLoading(false);

      // Fallback to mock data if API is not available
      const mockOrders: Order[] = [
        {
          id: "1",
          price: 1.25,
          amount: 1000,
          timestamp: Date.now(),
          side: "bid",
        },
        {
          id: "2",
          price: 1.24,
          amount: 500,
          timestamp: Date.now(),
          side: "bid",
        },
        {
          id: "3",
          price: 1.26,
          amount: 750,
          timestamp: Date.now(),
          side: "ask",
        },
        {
          id: "4",
          price: 1.27,
          amount: 300,
          timestamp: Date.now(),
          side: "ask",
        },
      ];
      setOrders(mockOrders);
    }
  };

  useEffect(() => {
    // Initial fetch
    fetchOrderBook();

    // Set up polling every 1 second
    const interval = setInterval(fetchOrderBook, 1000);

    return () => clearInterval(interval);
  }, []);

  const bids = orders
    .filter((order) => order.side === "bid")
    .sort((a, b) => b.price - a.price || a.timestamp - b.timestamp); // Price desc, then time asc

  const asks = orders
    .filter((order) => order.side === "ask")
    .sort((a, b) => a.price - b.price || a.timestamp - b.timestamp); // Price asc, then time asc

  if (loading && orders.length === 0) {
    return (
      <div
        style={{
          border: "1px solid #e0e0e0",
          borderRadius: "8px",
          padding: "16px",
          backgroundColor: "#f9f9f9",
        }}
      >
        <h3 style={{ margin: "0 0 16px 0", color: "#333" }}>📊 Order Book</h3>
        <div style={{ textAlign: "center", color: "#666" }}>
          <div>🔄 Loading orders...</div>
          <div style={{ fontSize: "11px", marginTop: "4px" }}>
            Connecting to{" "}
            {import.meta.env.VITE_API_URL || "http://localhost:8081"}
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
        <h3 style={{ margin: 0, color: "#333" }}>📊 Order Book</h3>
        <div
          style={{ display: "flex", alignItems: "center", fontSize: "11px" }}
        >
          {error ? (
            <span style={{ color: "#d32f2f" }}>⚠️ {error}</span>
          ) : (
            <span style={{ color: "#388e3c" }}>🟢 Live</span>
          )}
        </div>
      </div>

      <div
        style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "16px" }}
      >
        {/* Asks */}
        <div>
          <h4
            style={{ margin: "0 0 8px 0", color: "#d32f2f", fontSize: "14px" }}
          >
            ASKS (Sell Orders)
          </h4>
          <div style={{ maxHeight: "200px", overflowY: "auto" }}>
            {asks.length === 0 ? (
              <div
                style={{
                  textAlign: "center",
                  color: "#999",
                  padding: "20px",
                  fontSize: "12px",
                }}
              >
                No ask orders
              </div>
            ) : (
              asks.map((order) => (
                <div
                  key={order.id}
                  style={{
                    display: "flex",
                    justifyContent: "space-between",
                    padding: "4px 8px",
                    backgroundColor: "#ffebee",
                    borderRadius: "4px",
                    marginBottom: "2px",
                    fontSize: "12px",
                  }}
                >
                  <span style={{ color: "#d32f2f", fontWeight: "bold" }}>
                    ${order.price.toFixed(3)}
                  </span>
                  <span style={{ color: "#666" }}>
                    {order.amount.toLocaleString()}
                  </span>
                </div>
              ))
            )}
          </div>
        </div>

        {/* Bids */}
        <div>
          <h4
            style={{ margin: "0 0 8px 0", color: "#388e3c", fontSize: "14px" }}
          >
            BIDS (Buy Orders)
          </h4>
          <div style={{ maxHeight: "200px", overflowY: "auto" }}>
            {bids.length === 0 ? (
              <div
                style={{
                  textAlign: "center",
                  color: "#999",
                  padding: "20px",
                  fontSize: "12px",
                }}
              >
                No bid orders
              </div>
            ) : (
              bids.map((order) => (
                <div
                  key={order.id}
                  style={{
                    display: "flex",
                    justifyContent: "space-between",
                    padding: "4px 8px",
                    backgroundColor: "#e8f5e8",
                    borderRadius: "4px",
                    marginBottom: "2px",
                    fontSize: "12px",
                  }}
                >
                  <span style={{ color: "#388e3c", fontWeight: "bold" }}>
                    ${order.price.toFixed(3)}
                  </span>
                  <span style={{ color: "#666" }}>
                    {order.amount.toLocaleString()}
                  </span>
                </div>
              ))
            )}
          </div>
        </div>
      </div>

      <div
        style={{
          marginTop: "12px",
          fontSize: "11px",
          color: "#999",
          textAlign: "center",
        }}
      >
        Last updated: {lastUpdate.toLocaleTimeString()}
      </div>
    </div>
  );
}
