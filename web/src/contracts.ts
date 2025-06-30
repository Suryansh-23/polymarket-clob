// BatchSettlement contract ABI for event listening
export const batchSettlementABI = [
  {
    anonymous: true,
    inputs: [
      {
        indexed: true,
        name: "root",
        type: "bytes32",
      },
      {
        indexed: true,
        name: "submitter",
        type: "address",
      },
    ],
    name: "BatchSubmitted",
    type: "event",
  },
  {
    anonymous: true,
    inputs: [
      {
        indexed: true,
        name: "root",
        type: "bytes32",
      },
      {
        indexed: true,
        name: "disputer",
        type: "address",
      },
    ],
    name: "Disputed",
    type: "event",
  },
  {
    inputs: [
      { name: "root", type: "bytes32" },
      { name: "fills", type: "bytes" },
      { name: "aggSig", type: "bytes" },
    ],
    name: "submitBatch",
    outputs: [],
    stateMutability: "nonpayable",
    type: "function",
  },
] as const;

export const CONTRACT_ADDRESS = import.meta.env
  .VITE_CONTRACT_ADDRESS as `0x${string}`;
