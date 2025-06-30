import axios from "axios";
import { ethers } from "ethers";

interface SimulatedOrder {
  price: number;
  quantity: number;
  side: "bid" | "ask"; // Changed from "buy"/"sell" to "bid"/"ask"
}

class TradingSimulator {
  private isRunning = false;
  private marketPrice = 1.25; // Starting market price
  private priceVolatility = 0.02; // 2% volatility
  private apiUrl: string;
  private intervalId?: NodeJS.Timeout;

  constructor(apiUrl: string = "http://localhost:8081") {
    this.apiUrl = apiUrl;
  }

  private createEIP712Order(price: number, quantity: number, side: string) {
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
      timestamp: now + Math.floor(Math.random() * 1000), // Add small random delay
      signature: "", // Will be filled after signing
    };

    // Create a simple signature
    const orderHash = ethers.keccak256(
      ethers.toUtf8Bytes(
        `${order.maker}:${order.takerAsset}:${order.makeAmount}:${order.takeAmount}:${order.price}:${order.timestamp}`
      )
    );

    order.signature = wallet.signMessageSync(orderHash);
    return order;
  }

  private generateRealisticOrder(): SimulatedOrder {
    // Simulate market movement
    const priceChange = (Math.random() - 0.5) * this.priceVolatility;
    this.marketPrice *= 1 + priceChange;

    // Keep price in reasonable range
    this.marketPrice = Math.max(0.5, Math.min(2.0, this.marketPrice));

    // Generate order around market price with some spread
    const spreadMultiplier = 0.995 + Math.random() * 0.01; // 0.5% spread
    const side = Math.random() > 0.5 ? "bid" : "ask"; // Changed from "buy"/"sell" to "bid"/"ask"

    const price =
      side === "bid"
        ? this.marketPrice * spreadMultiplier
        : this.marketPrice / spreadMultiplier;

    // Generate realistic quantities (favor smaller orders with occasional large ones)
    const baseQuantity = Math.random() > 0.9 ? 5000 : 500; // 10% chance of large order
    const quantity = Math.floor(
      baseQuantity + Math.random() * baseQuantity * 2
    );

    return {
      price: Math.round(price * 10000) / 10000, // Round to 4 decimals
      quantity,
      side,
    };
  }

  private async submitOrder(orderData: SimulatedOrder): Promise<boolean> {
    try {
      const order = this.createEIP712Order(
        orderData.price,
        orderData.quantity,
        orderData.side
      );

      await axios.post(`${this.apiUrl}/orders`, order, {
        timeout: 3000,
        headers: {
          "Content-Type": "application/json",
        },
      });

      console.log(
        `[TradingBot] Submitted ${orderData.side} order: ${orderData.quantity} @ ${orderData.price}`
      );
      return true;
    } catch (error) {
      console.warn(`[TradingBot] Failed to submit order:`, error);
      return false;
    }
  }

  private async simulationLoop() {
    if (!this.isRunning) return;

    try {
      // Generate 2-5 orders per cycle (increased from 1-3)
      const orderCount = Math.floor(Math.random() * 4) + 2;

      for (let i = 0; i < orderCount; i++) {
        const order = this.generateRealisticOrder();
        await this.submitOrder(order);

        // Small delay between orders (reduced)
        if (i < orderCount - 1) {
          await new Promise((resolve) =>
            setTimeout(resolve, 100 + Math.random() * 200)
          );
        }
      }
    } catch (error) {
      console.warn("[TradingBot] Error in simulation loop:", error);
    }
  }

  public start(intervalMs: number = 3000): void {
    if (this.isRunning) return;

    console.log(
      `[TradingBot] Starting trading simulation (interval: ${intervalMs}ms)`
    );
    this.isRunning = true;

    // Initial delay to let the UI load
    setTimeout(() => {
      this.simulationLoop();
      this.intervalId = setInterval(() => this.simulationLoop(), intervalMs);
    }, 2000);
  }

  public stop(): void {
    if (!this.isRunning) return;

    console.log("[TradingBot] Stopping trading simulation");
    this.isRunning = false;

    if (this.intervalId) {
      clearInterval(this.intervalId);
      this.intervalId = undefined;
    }
  }

  public getCurrentMarketPrice(): number {
    return this.marketPrice;
  }

  public isActive(): boolean {
    return this.isRunning;
  }
}

// Export singleton instance
export const tradingSimulator = new TradingSimulator(
  import.meta.env.VITE_API_URL || "http://localhost:8081"
);

export default TradingSimulator;
