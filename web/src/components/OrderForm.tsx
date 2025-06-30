import { useState } from "react";
import axios from "axios";
import { ethers } from "ethers";

interface OrderFormData {
  price: string;
  quantity: string;
  side: "buy" | "sell";
}

export default function OrderForm() {
  const [formData, setFormData] = useState<OrderFormData>({
    price: "",
    quantity: "",
    side: "buy",
  });
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState<{
    type: "success" | "error";
    text: string;
  } | null>(null);

  const handleInputChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>
  ) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
  };

  const createEIP712Order = (price: number, quantity: number, side: string) => {
    const now = Date.now();
    const demoKey =
      import.meta.env.VITE_DEMO_KEY ||
      "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80";
    const wallet = new ethers.Wallet(demoKey);

    // Create EIP-712 compatible order
    const order = {
      maker: wallet.address,
      takerAsset: "0x1234567890123456789012345678901234567890", // Mock asset address
      makeAmount: quantity.toString(),
      takeAmount: (quantity * price).toString(),
      price: price,
      side: side, // Use the side parameter
      timestamp: now,
      signature: "", // Will be filled after signing
    };

    // Create a simple signature (in production, this would be proper EIP-712)
    const orderHash = ethers.keccak256(
      ethers.toUtf8Bytes(
        `${order.maker}:${order.takerAsset}:${order.makeAmount}:${order.takeAmount}:${order.price}:${order.timestamp}`
      )
    );

    // Sign the hash
    order.signature = wallet.signMessageSync(orderHash);

    return order;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setMessage(null);

    try {
      const price = parseFloat(formData.price);
      const quantity = parseFloat(formData.quantity);

      if (isNaN(price) || isNaN(quantity) || price <= 0 || quantity <= 0) {
        throw new Error("Price and quantity must be positive numbers");
      }

      // Create and sign the order
      const order = createEIP712Order(price, quantity, formData.side);

      // Submit to sequencer
      const apiUrl = import.meta.env.VITE_API_URL || "http://localhost:8081";
      await axios.post(`${apiUrl}/orders`, order, {
        timeout: 5000,
        headers: {
          "Content-Type": "application/json",
        },
      });

      setMessage({ type: "success", text: "Order submitted successfully!" });
      setFormData({ price: "", quantity: "", side: "buy" });
    } catch (error: any) {
      console.error("Failed to submit order:", error);
      setMessage({
        type: "error",
        text:
          error.response?.data?.error ||
          error.message ||
          "Failed to submit order",
      });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="trading-panel" style={{ height: "fit-content" }}>
      <h3
        style={{
          margin: "0 0 16px 0",
          color: "#ffffff",
          fontSize: "16px",
          fontWeight: "600",
          borderBottom: "1px solid #2b3139",
          paddingBottom: "12px",
        }}
      >
        Place Order
      </h3>

      <form
        onSubmit={handleSubmit}
        style={{
          display: "flex",
          gap: "12px",
          alignItems: "end",
          flexWrap: "wrap",
        }}
      >
        <div
          style={{
            display: "flex",
            flexDirection: "column",
            minWidth: "120px",
          }}
        >
          <label
            style={{ fontSize: "12px", marginBottom: "4px", color: "#666" }}
          >
            Side
          </label>
          <select
            name="side"
            value={formData.side}
            onChange={handleInputChange}
            style={{
              padding: "8px",
              border: "1px solid #ddd",
              borderRadius: "4px",
              fontSize: "14px",
            }}
          >
            <option value="buy">Buy</option>
            <option value="sell">Sell</option>
          </select>
        </div>

        <div
          style={{
            display: "flex",
            flexDirection: "column",
            minWidth: "120px",
          }}
        >
          <label
            style={{ fontSize: "12px", marginBottom: "4px", color: "#666" }}
          >
            Price ($)
          </label>
          <input
            type="number"
            name="price"
            value={formData.price}
            onChange={handleInputChange}
            placeholder="1.25"
            step="0.001"
            min="0"
            required
            style={{
              padding: "8px",
              border: "1px solid #ddd",
              borderRadius: "4px",
              fontSize: "14px",
            }}
          />
        </div>

        <div
          style={{
            display: "flex",
            flexDirection: "column",
            minWidth: "120px",
          }}
        >
          <label
            style={{ fontSize: "12px", marginBottom: "4px", color: "#666" }}
          >
            Quantity
          </label>
          <input
            type="number"
            name="quantity"
            value={formData.quantity}
            onChange={handleInputChange}
            placeholder="1000"
            min="0"
            required
            style={{
              padding: "8px",
              border: "1px solid #ddd",
              borderRadius: "4px",
              fontSize: "14px",
            }}
          />
        </div>

        <button
          type="submit"
          disabled={loading}
          style={{
            padding: "8px 16px",
            backgroundColor: formData.side === "buy" ? "#388e3c" : "#d32f2f",
            color: "white",
            border: "none",
            borderRadius: "4px",
            fontSize: "14px",
            cursor: loading ? "not-allowed" : "pointer",
            opacity: loading ? 0.7 : 1,
          }}
        >
          {loading
            ? "Submitting..."
            : `Submit ${formData.side === "buy" ? "Buy" : "Sell"} Order`}
        </button>
      </form>

      {message && (
        <div
          style={{
            marginTop: "12px",
            padding: "8px",
            borderRadius: "4px",
            fontSize: "12px",
            backgroundColor: message.type === "success" ? "#e8f5e8" : "#ffebee",
            color: message.type === "success" ? "#388e3c" : "#d32f2f",
            border: `1px solid ${
              message.type === "success" ? "#388e3c" : "#d32f2f"
            }`,
          }}
        >
          {message.text}
        </div>
      )}
    </div>
  );
}
