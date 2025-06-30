# -----------------------------------------------------------------------------
# This Makefile is used for building your AVS application.
#
# It contains basic targets for building the application, installing dependencies,
# and building a Docker container.
#
# Modify each target as needed to suit your application's requirements.
# -----------------------------------------------------------------------------

GO = $(shell which go)
OUT = ./bin

build: deps
	@mkdir -p $(OUT) || true
	@echo "Building binaries..."
	go build -o $(OUT)/performer ./cmd/main.go

deps:
	GOPRIVATE=github.com/Layr-Labs/* go mod tidy

build/container:
	./.hourglass/scripts/buildContainer.sh

test:
	go test ./... -v -p 1

# Deploy AVS contracts using the Solidity deployment script
deploy: check-yq
	@echo "Deploying Polymarket CLOB AVS contracts..."
	@mkdir -p contracts/script/devnet/output || true
	@cd contracts && forge script script/DeployPolymarketAVS.s.sol:DeployPolymarketAVS \
		--rpc-url http://localhost:8545 \
		--broadcast \
		--sig "run(string,string)" \
		"devnet" \
		"$$(yq eval -o=json '../config/contexts/devnet.yaml')"
	@echo "✅ Deployment completed! Check contracts/script/devnet/output/deploy_polymarket_avs_output.json"

# Deploy to a specific environment (devnet, testnet, mainnet)
deploy-env: check-yq
	@if [ -z "$(ENV)" ]; then \
		echo "❌ Error: ENV variable not set. Usage: make deploy-env ENV=devnet"; \
		exit 1; \
	fi
	@if [ ! -f "config/contexts/$(ENV).yaml" ]; then \
		echo "❌ Error: Config file not found: config/contexts/$(ENV).yaml"; \
		exit 1; \
	fi
	@echo "Deploying Polymarket CLOB AVS contracts to $(ENV)..."
	@mkdir -p contracts/script/$(ENV)/output || true
	@cd contracts && forge script script/DeployPolymarketAVS.s.sol:DeployPolymarketAVS \
		--rpc-url $$(yq eval '.context.chains.l1.rpc_url' '../config/contexts/$(ENV).yaml') \
		--broadcast \
		--sig "run(string,string)" \
		"$(ENV)" \
		"$$(yq eval -o=json '../config/contexts/$(ENV).yaml')"
	@echo "✅ Deployment to $(ENV) completed!"
check-yq:
	@which yq > /dev/null || (echo "❌ Error: yq is not installed. Install with: brew install yq" && exit 1)
	@which forge > /dev/null || (echo "❌ Error: forge is not installed. Install Foundry first." && exit 1)

# Verify deployment by checking deployed contracts
verify-deployment:
	@if [ ! -f "contracts/script/devnet/output/deploy_polymarket_avs_output.json" ]; then \
		echo "❌ Error: Deployment output not found. Run 'make deploy' first."; \
		exit 1; \
	fi
	@echo "🔍 Verifying deployed contracts..."
	@BATCH_SETTLEMENT=$$(jq -r '.addresses.BatchSettlement' contracts/script/devnet/output/deploy_polymarket_avs_output.json) && \
	DISPUTE_GAME=$$(jq -r '.addresses.DisputeGame' contracts/script/devnet/output/deploy_polymarket_avs_output.json) && \
	echo "BatchSettlement: $$BATCH_SETTLEMENT" && \
	echo "DisputeGame: $$DISPUTE_GAME" && \
	cast call $$BATCH_SETTLEMENT "totalBatchesSubmitted()" --rpc-url http://localhost:8545 && \
	cast call $$DISPUTE_GAME "totalDisputes()" --rpc-url http://localhost:8545
	@echo "✅ Contract verification completed!"
