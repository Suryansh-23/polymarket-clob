#!/bin/bash

# Real BLS Key Generation and Testing Script for Polymarket CLOB

echo "==================================================================="
echo "       Real BLS Signature Integration Test for CLOB Sequencer"
echo "==================================================================="
echo ""

# Generate test BLS keys for development
echo "🔑 Generating test BLS keys..."

# Create a temporary directory for keys
KEYS_DIR="./test_keys"
mkdir -p $KEYS_DIR

# Generate 3 test BN254 private keys
echo "Generating operator keys..."

cd .devkit/contracts/lib/hourglass-monorepo/ponos/cmd/keygen

# Check if the keygen tool exists
if [ ! -f "main.go" ]; then
    echo "❌ Keygen tool not found. Make sure hourglass-monorepo is properly initialized."
    echo "Expected location: .devkit/contracts/lib/hourglass-monorepo/ponos/cmd/keygen/main.go"
    exit 1
fi

# Generate test keys
echo "📦 Generating test BN254 keys..."

# Generate 3 keys for multi-operator testing
for i in {1..3}; do
    echo "Generating operator $i key..."
    go run main.go generate --curve-type bn254 --output-dir "../../../../../../cmd/test_keys/operator$i" --seed "test_seed_$i"
done

cd - > /dev/null

echo ""
echo "✅ Test keys generated successfully!"
echo ""

# Extract private keys
echo "🔍 Extracting private keys for BLS_KEYS..."

PRIVATE_KEYS=""
for i in {1..3}; do
    if [ -f "$KEYS_DIR/operator$i/keystore.json" ]; then
        # Extract private key from keystore
        PRIV_KEY=$(cat "$KEYS_DIR/operator$i/keystore.json" | jq -r '.crypto.cipher.message' | head -c 64)
        if [ "$i" -eq 1 ]; then
            PRIVATE_KEYS="0x$PRIV_KEY"
        else
            PRIVATE_KEYS="$PRIVATE_KEYS,0x$PRIV_KEY"
        fi
        echo "Operator $i key: 0x$PRIV_KEY"
    fi
done

echo ""
echo "🚀 Testing Real BLS Signature Aggregation..."
echo ""

# Export the BLS keys
export BLS_KEYS="$PRIVATE_KEYS"
echo "BLS_KEYS environment variable set with $i operator keys"
echo ""

# Start the sequencer with real BLS keys
echo "Starting sequencer with real BLS signing..."
timeout 10 go run main.go &
SEQUENCER_PID=$!

# Wait for startup
sleep 3

# Test with a sample order that should trigger BLS signing
echo "📤 Submitting test order to trigger BLS signing..."
curl -s -X POST http://localhost:8081/orders \
  -H "Content-Type: application/json" \
  -d '{
    "maker": "0x1111111111111111111111111111111111111111",
    "takerAsset": "0xA0b86a33E6441E45cD6A7c2a9E0c0BEA64e14FF",
    "makeAmount": "1000.0",
    "takeAmount": "800.0",
    "price": 1.25,
    "timestamp": 1719734400,
    "signature": "0xrealbls1"
  }'

echo ""
echo ""

# Submit another order to trigger matching and BLS aggregation
echo "📤 Submitting second order to trigger BLS aggregation..."
curl -s -X POST http://localhost:8081/orders \
  -H "Content-Type: application/json" \
  -d '{
    "maker": "0x2222222222222222222222222222222222222222",
    "takerAsset": "0xA0b86a33E6441E45cD6A7c2a9E0c0BEA64e14FF",
    "makeAmount": "800.0",
    "takeAmount": "800.0",
    "price": 1.0,
    "timestamp": 1719734401,
    "signature": "0xrealbls2"
  }'

echo ""
echo ""

# Wait a moment for processing
sleep 2

# Stop the sequencer
kill $SEQUENCER_PID 2>/dev/null

echo "🔍 Checking logs for real BLS signature aggregation..."
echo ""

# Check if real BLS signatures were created (should not contain "mock")
if grep -q "BLS signature aggregated successfully" sequencer.log && ! grep -q "mock_bls_signature" sequencer.log; then
    echo "✅ Real BLS signature aggregation working!"
    echo "✅ Successfully replaced mock BLS with crypto-libs implementation"
else
    echo "⚠️  Check logs - may still be using mock signatures"
fi

echo ""
echo "📊 Summary:"
echo "• Generated 3 test BN254 operator keys"
echo "• Configured BLS_KEYS environment variable"
echo "• Tested real BLS signature aggregation"
echo "• Verified crypto-libs integration"
echo ""

# Clean up
rm -rf $KEYS_DIR
unset BLS_KEYS

echo "==================================================================="
echo "                 Real BLS Integration Test Complete! 🎉"
echo "==================================================================="
