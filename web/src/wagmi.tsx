import { createConfig, http } from "wagmi";
import { WagmiProvider } from "wagmi";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createPublicClient, defineChain } from "viem";
import type { ReactNode } from "react";

// Define local development chain
const localChain = defineChain({
  id: 1337,
  name: "LocalDev",
  nativeCurrency: {
    decimals: 18,
    name: "Ether",
    symbol: "ETH",
  },
  rpcUrls: {
    default: {
      http: [import.meta.env.VITE_RPC_URL || "http://localhost:8545"],
    },
  },
});

// Create Wagmi configuration
const config = createConfig({
  chains: [localChain],
  transports: {
    [localChain.id]: http(),
  },
});

// Create Query Client for React Query
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5, // 5 minutes
      retry: 3,
      refetchOnWindowFocus: false,
    },
  },
});

// Create public client for reading blockchain data
export const publicClient = createPublicClient({
  chain: localChain,
  transport: http(import.meta.env.VITE_RPC_URL || "http://localhost:8545"),
});

export { config, queryClient };

// Wagmi Provider wrapper component
export function WagmiWrapper({ children }: { children: ReactNode }) {
  return (
    <QueryClientProvider client={queryClient}>
      <WagmiProvider config={config}>{children}</WagmiProvider>
    </QueryClientProvider>
  );
}
