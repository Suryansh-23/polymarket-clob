---
applyTo: "**"
---

AVS Development Instructions

These instructions guide the coding agent in building the Decentralized CLOB Sequencer AVS MVP using the Hourglass EigenLayer template (via devkit-cli). Refer to this file for context, links, and conventions for all development tasks.

⸻

1. Project Overview
   • Goal: Replace Polymarket’s centralized order-matching server with a slashable, stake-weighted operator set that sorts EIP-712 orders by price→time priority and submits matched batches on-chain.
   • Scaffold: Generated via Layr Labs devkit-cli using the Hourglass AVS template:
   • Repo: https://github.com/Layr-Labs/devkit-cli
   • Command used: devkit avs create poly-sequencer .
   • Key Components:

   1. Contracts (contracts/):
      • BatchSettlement.sol (stores Merkle root + BLS sig)
      • DisputeGame.sol (fraud proofs & slashing)
   2. Sequencer Service (cmd/sequencer/):
      • In-memory orderbook
      • Price-time matcher
      • Merkle tree builder
      • BLS aggregator & batch submitter
   3. Configuration (config/config.yaml):
      • AVS name, BLS quorum settings, network endpoints

2. Coding Conventions & Context
   • Language: Go or TypeScript for the sequencer service; Solidity (Foundry) for contracts.
   • EIP-712 Orders: Use Polymarket’s signed-order schema. See:
   • Polymarket order spec: https://github.com/polymarket/polymarket-docs#eip-712-order-schema
   • BLS Aggregation: Leverage Hourglass AVS helper libraries bundled in the template.
   • Merkle Trees: Build over tuples of (makerHash, takerHash, fillQuantity); use a standard merkle implementation (e.g., merkletreejs for TS or go-merkle).

3. Development Workflow

   1. Implement Contracts:
      • Add BatchSettlement.sol and DisputeGame.sol under contracts/.
      • Ensure inheritance from the template’s base AVS core.
   2. Write Sequencer Service:
      • Create cmd/sequencer/main.go or index.ts.
      • Implement:
      • Order ingestion API stub (e.g., /submitOrder).
      • Matching loop with price→time ordering.
      • Merkle builder and BLS signing.
      • RPC call to BatchSettlement.submitBatch(...).
   3. Local Devkit Testing:
      • Run devkit avs devnet up to start a local devnet.
      • Use devkit avs call submitBatch --args ./fixtures/batch.json to simulate batch submission.
   4. Context for Each Prompt:
      • Always remind the agent of the file path (contracts/ or cmd/sequencer/).
      • Reference this instructions file by path: ./instructions.md.

4. Useful Links & References
   • devkit-cli (Hourglass template): https://github.com/Layr-Labs/devkit-cli
   • Polymarket Docs (CLOB intro): https://polymarket.gitbook.io/docs/technical-overview/order-book
   • EIP-712 Order Schema: https://github.com/polymarket/polymarket-docs#eip-712-order-schema
   • EigenLayer AVS Core Reference: (in scaffold under pkg/avs)

⸻

Remember: Keep prompts concise, reference file paths, and rely on this markdown as the single source of truth for context.
