import { useState, useEffect } from "react";

interface Event {
  id: string;
  type: "OrderPlaced" | "OrderMatched" | "BatchSubmitted";
  timestamp: number;
  data: any;
}

export default function EventLog() {
  const [events, setEvents] = useState<Event[]>([]);

  useEffect(() => {
    // Simulate events
    const mockEvents: Event[] = [
      {
        id: "1",
        type: "OrderPlaced",
        timestamp: Date.now() - 60000,
        data: { price: "1.2412", amount: "1074", side: "bid" },
      },
      {
        id: "2",
        type: "BatchSubmitted",
        timestamp: Date.now() - 30000,
        data: { root: "0xabcd...", fills: 5 },
      },
    ];
    setEvents(mockEvents);

    // Add new events periodically
    const interval = setInterval(() => {
      if (Math.random() > 0.7) {
        const newEvent: Event = {
          id: Date.now().toString(),
          type: Math.random() > 0.5 ? "OrderPlaced" : "OrderMatched",
          timestamp: Date.now(),
          data: {
            price: (1.2 + Math.random() * 0.1).toFixed(4),
            amount: Math.floor(Math.random() * 2000 + 500).toString(),
          },
        };
        setEvents((prev) => [newEvent, ...prev.slice(0, 9)]);
      }
    }, 3000);

    return () => clearInterval(interval);
  }, []);

  const getEventIcon = (type: string) => {
    switch (type) {
      case "BatchSubmitted":
        return "📦";
      case "OrderMatched":
        return "🤝";
      case "OrderPlaced":
        return "📋";
      default:
        return "⚡";
    }
  };

  const getEventColor = (type: string) => {
    switch (type) {
      case "BatchSubmitted":
        return "#02c076";
      case "OrderMatched":
        return "#f0b90b";
      case "OrderPlaced":
        return "#1890ff";
      default:
        return "#848e9c";
    }
  };

  const formatTimeAgo = (timestamp: number) => {
    const seconds = Math.floor((Date.now() - timestamp) / 1000);
    if (seconds < 60) return `${seconds}s ago`;
    if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`;
    return `${Math.floor(seconds / 3600)}h ago`;
  };

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
          ⛓️ Real-time Events
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
              backgroundColor: "#02c076",
            }}
          />
          <span style={{ color: "#848e9c" }}>Live</span>
        </div>
      </div>

      <div style={{ flex: 1, overflowY: "auto" }}>
        {events.length === 0 ? (
          <div
            style={{ textAlign: "center", color: "#848e9c", padding: "40px 0" }}
          >
            <div>No events yet</div>
            <div style={{ fontSize: "11px", marginTop: "4px" }}>
              Waiting for activity...
            </div>
          </div>
        ) : (
          events.map((event) => (
            <div
              key={event.id}
              style={{
                border: "1px solid #1e2329",
                borderRadius: "6px",
                padding: "12px",
                marginBottom: "8px",
                backgroundColor: "#1e2329",
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
                <span style={{ color: "#848e9c", fontSize: "10px" }}>
                  {formatTimeAgo(event.timestamp)}
                </span>
              </div>

              {event.data && (
                <div style={{ fontSize: "11px", color: "#848e9c" }}>
                  {Object.entries(event.data).map(([key, value]) => (
                    <div key={key}>
                      <strong>{key}:</strong> {String(value)}
                    </div>
                  ))}
                </div>
              )}
            </div>
          ))
        )}
      </div>

      <div
        style={{
          marginTop: "16px",
          paddingTop: "16px",
          borderTop: "1px solid #1e2329",
        }}
      >
        <h4
          style={{
            margin: "0 0 12px 0",
            color: "#ffffff",
            fontSize: "14px",
            fontWeight: "600",
          }}
        >
          📊 Recent Batches
        </h4>
        <div style={{ overflowX: "auto" }}>
          <table
            style={{
              width: "100%",
              fontSize: "11px",
              borderCollapse: "collapse",
            }}
          >
            <thead>
              <tr style={{ backgroundColor: "#1e2329" }}>
                <th
                  style={{
                    padding: "6px",
                    textAlign: "left",
                    border: "1px solid #2b3139",
                    color: "#848e9c",
                  }}
                >
                  Root
                </th>
                <th
                  style={{
                    padding: "6px",
                    textAlign: "left",
                    border: "1px solid #2b3139",
                    color: "#848e9c",
                  }}
                >
                  Block
                </th>
                <th
                  style={{
                    padding: "6px",
                    textAlign: "left",
                    border: "1px solid #2b3139",
                    color: "#848e9c",
                  }}
                >
                  Time
                </th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td
                  style={{
                    padding: "6px",
                    border: "1px solid #2b3139",
                    color: "#ffffff",
                    fontFamily: "monospace",
                  }}
                >
                  0xabcd...1234
                </td>
                <td
                  style={{
                    padding: "6px",
                    border: "1px solid #2b3139",
                    color: "#ffffff",
                  }}
                >
                  12,345,678
                </td>
                <td
                  style={{
                    padding: "6px",
                    border: "1px solid #2b3139",
                    color: "#ffffff",
                  }}
                >
                  2 min ago
                </td>
              </tr>
              <tr>
                <td
                  style={{
                    padding: "6px",
                    border: "1px solid #2b3139",
                    color: "#ffffff",
                    fontFamily: "monospace",
                  }}
                >
                  0xbeef...5678
                </td>
                <td
                  style={{
                    padding: "6px",
                    border: "1px solid #2b3139",
                    color: "#ffffff",
                  }}
                >
                  12,345,675
                </td>
                <td
                  style={{
                    padding: "6px",
                    border: "1px solid #2b3139",
                    color: "#ffffff",
                  }}
                >
                  5 min ago
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
