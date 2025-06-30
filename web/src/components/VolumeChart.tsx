import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts";
import { useState, useEffect } from "react";
import axios from "axios";

interface VolumeData {
  time: string;
  volume: number;
  value: number;
}

interface VolumeResponse {
  hourlyVolume: VolumeData[];
  totalVolume: number;
  timestamp: number;
}

export default function VolumeChart() {
  const [volumeData, setVolumeData] = useState<VolumeData[]>([]);
  const [loading, setLoading] = useState(true);
  const [totalVolume, setTotalVolume] = useState(0);
  const [error, setError] = useState<string | null>(null);

  const generateMockVolumeData = (): VolumeData[] => {
    const now = new Date();
    const mockVolumeData: VolumeData[] = [];

    for (let i = 23; i >= 0; i--) {
      const time = new Date(now.getTime() - i * 60 * 60 * 1000);
      const volume = Math.floor(Math.random() * 10000) + 1000;
      const avgPrice = 1.25 + (Math.random() - 0.5) * 0.1;

      mockVolumeData.push({
        time: time.toLocaleTimeString("en-US", {
          hour: "2-digit",
          minute: "2-digit",
        }),
        volume: volume,
        value: volume * avgPrice,
      });
    }

    return mockVolumeData;
  };

  const fetchVolumeData = async () => {
    try {
      setError(null);
      const apiUrl = import.meta.env.VITE_API_URL || "http://localhost:8081";

      // Try to fetch precomputed volume data
      const response = await axios.get<VolumeResponse>(`${apiUrl}/volume`, {
        timeout: 5000,
      });

      setVolumeData(response.data.hourlyVolume);
      setTotalVolume(response.data.totalVolume);
      setLoading(false);
    } catch (err) {
      console.error("Failed to fetch volume data:", err);
      setError("Failed to load volume data");

      // Fallback to mock data
      const mockData = generateMockVolumeData();
      setVolumeData(mockData);
      setTotalVolume(mockData.reduce((sum, item) => sum + item.volume, 0));
      setLoading(false);
    }
  };

  useEffect(() => {
    // Initial fetch
    fetchVolumeData();

    // Set up polling every 5 seconds (less frequent for volume data)
    const interval = setInterval(fetchVolumeData, 5000);

    return () => clearInterval(interval);
  }, []);

  if (loading && volumeData.length === 0) {
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
          📊 Trading Volume
        </h3>
        <div
          style={{
            width: "100%",
            height: "220px",
            display: "flex",
            flexDirection: "column",
            gap: "8px",
          }}
        >
          <div
            className="skeleton skeleton-line"
            style={{ height: "20px", width: "70%" }}
          ></div>
          <div
            className="skeleton skeleton-line"
            style={{ height: "180px", width: "100%" }}
          ></div>
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
        <h3 style={{ margin: 0, color: "#333" }}>📊 Trading Volume (24h)</h3>
        <div
          style={{
            display: "flex",
            flexDirection: "column",
            alignItems: "flex-end",
            fontSize: "12px",
          }}
        >
          <div style={{ color: "#666" }}>
            Total: {totalVolume.toLocaleString()} units
          </div>
          <div style={{ fontSize: "11px", marginTop: "2px" }}>
            {error ? (
              <span style={{ color: "#d32f2f" }}>⚠️ Using mock data</span>
            ) : (
              <span style={{ color: "#388e3c" }}>🟢 Live</span>
            )}
          </div>
        </div>
      </div>

      <ResponsiveContainer width="100%" height="85%">
        <BarChart data={volumeData || []}>
          <CartesianGrid strokeDasharray="3 3" />
          <XAxis dataKey="time" fontSize={10} interval="preserveStartEnd" />
          <YAxis
            tickFormatter={(value) => `${(value / 1000).toFixed(0)}k`}
            fontSize={11}
          />
          <Tooltip
            formatter={(value: number, name: string) => [
              name === "volume"
                ? `${value.toLocaleString()} units`
                : `$${value.toLocaleString()}`,
              name === "volume" ? "Volume" : "Value",
            ]}
            labelFormatter={(label) => `Time: ${label}`}
          />
          <Bar
            dataKey="volume"
            fill="#1976d2"
            radius={[2, 2, 0, 0]}
            name="volume"
          />
        </BarChart>
      </ResponsiveContainer>
    </div>
  );
}
