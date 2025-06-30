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

      // Handle null/undefined arrays from API
      const allOrders = [
        ...(response.data.bids || []),
        ...(response.data.asks || []),
      ];
      setOrders(allOrders);
      setLastUpdate(new Date());
      setLoading(false);
    } catch (err) {
      console.error("Failed to fetch order book:", err);
      setError("API Error");
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchOrderBook();
    const interval = setInterval(fetchOrderBook, 1000);
    return () => clearInterval(interval);
  }, []);

  const bids = orders
    .filter((order) => order.side === "bid")
    .sort((a, b) => b.price - a.price);
  const asks = orders
    .filter((order) => order.side === "ask")
    .sort((a, b) => a.price - b.price);

  return (
    <div
      style={{
        height: "100%",
        backgroundColor: "#0d1421",
        color: "#ffffff",
        display: "flex",
        flexDirection: "column",
      }}
    >
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          marginBottom: "16px",
          paddingBottom: "8px",
          borderBottom: "1px solid #1e2329",
        }}
      >
        <h3
          style={{
            margin: 0,
            color: "#ffffff",
            fontSize: "16px",
            fontWeight: "600",
          }}
        >
          📊 Order Book
        </h3>
        <div
          style={{
            display: "flex",
            alignItems: "center",
            fontSize: "11px",
            gap: "6px",
          }}
        >
          <div
            style={{
              width: "6px",
              height: "6px",
              borderRadius: "50%",
              backgroundColor: error ? "#f84960" : "#02c076",
            }}
          />
          <span style={{ color: error ? "#f84960" : "#848e9c" }}>
            {error ? "Error" : "Live"}
          </span>
        </div>
      </div>

      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          marginBottom: "12px",
          fontSize: "11px",
          color: "#848e9c",
          textTransform: "uppercase",
          fontWeight: "500",
        }}
      >
        <span>Price (USDC)</span>
        <span>Size</span>
      </div>

      <div style={{ flex: 1, display: "flex", flexDirection: "column" }}>
        <div style={{ marginBottom: "12px" }}>
          <div
            style={{
              fontSize: "11px",
              color: "#848e9c",
              marginBottom: "8px",
              textTransform: "uppercase",
              fontWeight: "500",
            }}
          >
            Asks ({asks.length})
          </div>
          <div style={{ maxHeight: "150px", overflowY: "auto" }}>
            {loading && orders.length === 0 ? (
              [...Array(3)].map((_, i) => (
                <div
                  key={i}
                  style={{
                    display: "flex",
                    justifyContent: "space-between",
                    padding: "3px 0",
                  }}
                >
                  <div
                    className="skeleton"
                    style={{ width: "60px", height: "14px" }}
                  ></div>
                  <div
                    className="skeleton"
                    style={{ width: "80px", height: "14px" }}
                  ></div>
                </div>
              ))
            ) : asks.length === 0 ? (
              <div
                style={{
                  textAlign: "center",
                  color: "#848e9c",
                  padding: "20px 0",
                  fontSize: "12px",
                }}
              >
                No asks
              </div>
            ) : (
              asks
                .slice()
                .reverse()
                .map((order) => (
                  <div
                    key={order.id}
                    style={{
                      display: "flex",
                      justifyContent: "space-between",
                      padding: "3px 0",
                      fontSize: "12px",
                      fontFamily: "monospace",
                    }}
                  >
                    <span style={{ color: "#f84960", fontWeight: "500" }}>
                      {order.price.toFixed(4)}
                    </span>
                    <span style={{ color: "#848e9c" }}>
                      {order.amount.toLocaleString()}
                    </span>
                  </div>
                ))
            )}
          </div>
        </div>

        <div
          style={{
            textAlign: "center",
            padding: "8px 0",
            backgroundColor: "#1e2329",
            borderRadius: "4px",
            fontSize: "11px",
            color: "#848e9c",
            marginBottom: "12px",
          }}
        >
          Spread ↔ USDC
        </div>

        <div>
          <div
            style={{
              fontSize: "11px",
              color: "#848e9c",
              marginBottom: "8px",
              textTransform: "uppercase",
              fontWeight: "500",
            }}
          >
            Bids ({bids.length})
          </div>
          <div style={{ maxHeight: "150px", overflowY: "auto" }}>
            {loading && orders.length === 0 ? (
              [...Array(3)].map((_, i) => (
                <div
                  key={i}
                  style={{
                    display: "flex",
                    justifyContent: "space-between",
                    padding: "3px 0",
                  }}
                >
                  <div
                    className="skeleton"
                    style={{ width: "60px", height: "14px" }}
                  ></div>
                  <div
                    className="skeleton"
                    style={{ width: "80px", height: "14px" }}
                  ></div>
                </div>
              ))
            ) : bids.length === 0 ? (
              <div
                style={{
                  textAlign: "center",
                  color: "#848e9c",
                  padding: "20px 0",
                  fontSize: "12px",
                }}
              >
                No bids
              </div>
            ) : (
              bids.map((order) => (
                <div
                  key={order.id}
                  style={{
                    display: "flex",
                    justifyContent: "space-between",
                    padding: "3px 0",
                    fontSize: "12px",
                    fontFamily: "monospace",
                  }}
                >
                  <span style={{ color: "#02c076", fontWeight: "500" }}>
                    {order.price.toFixed(4)}
                  </span>
                  <span style={{ color: "#848e9c" }}>
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
          marginTop: "auto",
          fontSize: "10px",
          color: "#848e9c",
          textAlign: "center",
          paddingTop: "8px",
          borderTop: "1px solid #1e2329",
        }}
      >
        Updated: {lastUpdate.toLocaleTimeString()}
      </div>
    </div>
  );
}
