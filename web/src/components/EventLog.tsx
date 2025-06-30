import { useState, useEffect } from "react";
import { useWatchContractEvent, useBlockNumber } from "wagmi";
import { batchSettlementABI, CONTRACT_ADDRESS } from "../contracts";

interface BlockchainEvent {
  id: string;
  type: "BatchSubmitted" | "Disputed" | "OrderMatched";
  txHash: string;
  blockNumber: number;
  timestamp: number;
  data: Record<string, any>;
}

export default function EventLog() {
  const [events, setEvents] = useState<BlockchainEvent[]>([]);
  const [loading, setLoading] = useState(true);
  const [isConnected, setIsConnected] = useState(false);

  // Get current block number to determine connection status
  const { data: blockNumber } = useBlockNumber({
    watch: true,
  });

  // Watch for BatchSubmitted events
  useWatchContractEvent({
    address: CONTRACT_ADDRESS,
    abi: batchSettlementABI,
    eventName: "BatchSubmitted",
    onLogs(logs) {
      console.log("BatchSubmitted events:", logs);

      const newEvents = logs.map((log) => ({
        id: `batch-${log.transactionHash}-${log.logIndex}`,
        type: "BatchSubmitted" as const,
        txHash: log.transactionHash || "",
        blockNumber: Number(log.blockNumber || 0),
        timestamp: Date.now(), // In real app, would fetch block timestamp
        data: {
          root: log.args?.root || "",
          submitter: log.args?.submitter || "",
        },
      }));

      setEvents((prev) => [...newEvents, ...prev].slice(0, 50)); // Keep last 50 events
    },
  });

  // Watch for Disputed events
  useWatchContractEvent({
    address: CONTRACT_ADDRESS,
    abi: batchSettlementABI,
    eventName: "Disputed",
    onLogs(logs) {
      console.log("Disputed events:", logs);

      const newEvents = logs.map((log) => ({
        id: `dispute-${log.transactionHash}-${log.logIndex}`,
        type: "Disputed" as const,
        txHash: log.transactionHash || "",
        blockNumber: Number(log.blockNumber || 0),
        timestamp: Date.now(),
        data: {
          root: log.args?.root || "",
          disputer: log.args?.disputer || "",
        },
      }));

      setEvents((prev) => [...newEvents, ...prev].slice(0, 50));
    },
  });

  useEffect(() => {
    // Set connection status based on block number availability
    if (blockNumber !== undefined) {
      setIsConnected(true);
      setLoading(false);
    } else {
      // If no block number after 5 seconds, show as disconnected but not loading
      const timer = setTimeout(() => {
        setLoading(false);
        setIsConnected(false);
      }, 5000);

      return () => clearTimeout(timer);
    }
  }, [blockNumber]);

  useEffect(() => {
    // Add some mock events initially for demo purposes
    if (!loading && events.length === 0) {
      const mockEvents: BlockchainEvent[] = [
        {
          id: "mock-1",
          type: "BatchSubmitted",
          txHash: "0x1234567890abcdef1234567890abcdef12345678",
          blockNumber: Number(blockNumber || 18500123),
          timestamp: Date.now() - 300000, // 5 minutes ago
          data: {
            root: "0xabcd1234...",
            submitter: "0x9876...5432",
          },
        },
        {
          id: "mock-2",
          type: "BatchSubmitted",
          txHash: "0x3456789012cdefgh3456789012cdefgh34567890",
          blockNumber: Number(blockNumber || 18500121) - 2,
          timestamp: Date.now() - 900000, // 15 minutes ago
          data: {
            root: "0xbeef8765...",
            submitter: "0x1111...2222",
          },
        },
      ];
      setEvents(mockEvents);
    }
  }, [loading, events.length, blockNumber]);

  const formatTimeAgo = (timestamp: number) => {
    const diff = Date.now() - timestamp;
    const minutes = Math.floor(diff / 60000);
    if (minutes < 1) return "Just now";
    if (minutes === 1) return "1 minute ago";
    if (minutes < 60) return `${minutes} minutes ago`;
    const hours = Math.floor(minutes / 60);
    if (hours === 1) return "1 hour ago";
    return `${hours} hours ago`;
  };

  const getEventIcon = (type: string) => {
    switch (type) {
      case "BatchSubmitted":
        return "📦";
      case "Disputed":
        return "⚠️";
      case "OrderMatched":
        return "✅";
      default:
        return "📋";
    }
  };

  const getEventColor = (type: string) => {
    switch (type) {
      case "BatchSubmitted":
        return "#1976d2";
      case "Disputed":
        return "#d32f2f";
      case "OrderMatched":
        return "#388e3c";
      default:
        return "#666";
    }
  };

  const getExplorerUrl = (txHash: string) => {
    const rpcUrl = import.meta.env.VITE_RPC_URL || "http://localhost:8545";
    if (rpcUrl.includes("localhost") || rpcUrl.includes("127.0.0.1")) {
      return `#`; // Local development, no explorer
    }
    return `https://explorer.testnet.io/tx/${txHash}`;
  };

  if (loading) {
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
        <h3 style={{ margin: "0 0 16px 0", color: "#333" }}>
          ⛓️ On-chain Events
        </h3>
        <div style={{ textAlign: "center", color: "#666", marginTop: "80px" }}>
          <div>🔄 Connecting to blockchain...</div>
          <div style={{ fontSize: "11px", marginTop: "4px" }}>
            Listening for BatchSubmitted events
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
        }}
      >
        <h3 style={{ margin: 0, color: "#333" }}>⛓️ On-chain Events</h3>
        <div
          style={{
            display: "flex",
            flexDirection: "column",
            alignItems: "flex-end",
            fontSize: "11px",
          }}
        >
          <div
            style={{
              display: "flex",
              alignItems: "center",
              color: isConnected ? "#388e3c" : "#d32f2f",
            }}
          >
            <div
              style={{
                width: "8px",
                height: "8px",
                borderRadius: "50%",
                backgroundColor: isConnected ? "#388e3c" : "#d32f2f",
                marginRight: "6px",
              }}
            />
            {isConnected ? "Connected" : "Disconnected"}
          </div>
          {blockNumber && (
            <div style={{ color: "#666", marginTop: "2px" }}>
              Block: {blockNumber.toString()}
            </div>
          )}
        </div>
      </div>

      <div
        style={{
          flex: 1,
          overflowY: "auto",
          maxHeight: "220px",
        }}
      >
        {events.length === 0 ? (
          <div
            style={{ textAlign: "center", color: "#666", marginTop: "40px" }}
          >
            <div>No events yet</div>
            <div style={{ fontSize: "11px", marginTop: "4px" }}>
              Waiting for BatchSubmitted events...
            </div>
          </div>
        ) : (
          events.map((event) => (
            <div
              key={event.id}
              style={{
                border: "1px solid #e0e0e0",
                borderRadius: "6px",
                padding: "12px",
                marginBottom: "8px",
                backgroundColor: "#fff",
                fontSize: "12px",
              }}
            >
              <div
                style={{
                  display: "flex",
                  justifyContent: "space-between",
                  alignItems: "flex-start",
                  marginBottom: "6px",
                }}
              >
                <div style={{ display: "flex", alignItems: "center" }}>
                  <span style={{ marginRight: "6px", fontSize: "14px" }}>
                    {getEventIcon(event.type)}
                  </span>
                  <span
                    style={{
                      fontWeight: "bold",
                      color: getEventColor(event.type),
                    }}
                  >
                    {event.type}
                  </span>
                </div>
                <span style={{ color: "#999", fontSize: "10px" }}>
                  {formatTimeAgo(event.timestamp)}
                </span>
              </div>

              <div style={{ marginBottom: "4px" }}>
                <strong>Block:</strong> {event.blockNumber.toLocaleString()}
              </div>

              <div style={{ marginBottom: "6px" }}>
                <strong>Tx:</strong>
                <a
                  href={getExplorerUrl(event.txHash)}
                  target="_blank"
                  rel="noopener noreferrer"
                  style={{
                    color: "#1976d2",
                    textDecoration: "none",
                    marginLeft: "4px",
                  }}
                >
                  {event.txHash.slice(0, 10)}...{event.txHash.slice(-6)}
                </a>
              </div>

              {Object.keys(event.data).length > 0 && (
                <div style={{ fontSize: "11px", color: "#666" }}>
                  {Object.entries(event.data).map(([key, value]) => (
                    <div key={key}>
                      <strong>{key}:</strong>{" "}
                      {typeof value === "string" && value.length > 20
                        ? `${value.slice(0, 20)}...`
                        : String(value)}
                    </div>
                  ))}
                </div>
              )}
            </div>
          ))
        )}
      </div>
    </div>
  );
}
