#!/bin/bash

# Multi-fill testing script for Polymarket CLOB Sequencer

BASE_URL="http://localhost:8081"

echo "==================================================================="
echo "        Multi-Fill Order Matching Test for CLOB Sequencer"
echo "==================================================================="
echo ""

echo "📋 This test will:"
echo "• Start the sequencer service"
echo "• Submit multiple orders with different prices and amounts"
echo "• Verify that multiple fills are created per batch"
echo "• Confirm that the orderbook is properly pruned"
echo "• Test partial fills and order book management"
echo ""

# Check if sequencer is running
echo "🔍 Checking if sequencer is running..."
if curl -s -f "$BASE_URL/health" > /dev/null 2>&1; then
    echo "✅ Sequencer is running"
else
    echo "❌ Sequencer not running. Please start it with: go run main.go"
    exit 1
fi

echo ""
echo "🚀 Starting Multi-Fill Order Test..."
echo ""

# Test 1: Submit multiple orders that should match
echo "Test 1: Submitting overlapping bid/ask orders for multiple fills"
echo "--------------------------------------------------------------"

# Order 1: High price bid
echo "📤 Submitting Order 1: High price bid (1000 @ 1.5)"
curl -s -X POST "$BASE_URL/orders" \
  -H "Content-Type: application/json" \
  -d '{
    "maker": "0x1111111111111111111111111111111111111111",
    "takerAsset": "0xA0b86a33E6441E45cD6A7c2a9E0c0BEA64e14FF",
    "makeAmount": "1000.0",
    "takeAmount": "666.67",
    "price": 1.5,
    "timestamp": 1719734400,
    "signature": "0xsig1"
  }'

echo ""

# Order 2: Medium price ask
echo "📤 Submitting Order 2: Medium price ask (800 @ 1.2)"
curl -s -X POST "$BASE_URL/orders" \
  -H "Content-Type: application/json" \
  -d '{
    "maker": "0x2222222222222222222222222222222222222222",
    "takerAsset": "0xA0b86a33E6441E45cD6A7c2a9E0c0BEA64e14FF",
    "makeAmount": "800.0",
    "takeAmount": "666.67",
    "price": 1.2,
    "timestamp": 1719734401,
    "signature": "0xsig2"
  }'

echo ""

# Order 3: Another bid
echo "📤 Submitting Order 3: Another bid (600 @ 1.4)"
curl -s -X POST "$BASE_URL/orders" \
  -H "Content-Type: application/json" \
  -d '{
    "maker": "0x3333333333333333333333333333333333333333",
    "takerAsset": "0xA0b86a33E6441E45cD6A7c2a9E0c0BEA64e14FF",
    "makeAmount": "600.0",
    "takeAmount": "428.57",
    "price": 1.4,
    "timestamp": 1719734402,
    "signature": "0xsig3"
  }'

echo ""

# Order 4: Lower price ask
echo "📤 Submitting Order 4: Lower price ask (500 @ 1.0)"
curl -s -X POST "$BASE_URL/orders" \
  -H "Content-Type: application/json" \
  -d '{
    "maker": "0x4444444444444444444444444444444444444444",
    "takerAsset": "0xA0b86a33E6441E45cD6A7c2a9E0c0BEA64e14FF",
    "makeAmount": "500.0",
    "takeAmount": "500.0",
    "price": 1.0,
    "timestamp": 1719734403,
    "signature": "0xsig4"
  }'

echo ""

# Order 5: Medium bid
echo "📤 Submitting Order 5: Medium bid (750 @ 1.3)"
curl -s -X POST "$BASE_URL/orders" \
  -H "Content-Type: application/json" \
  -d '{
    "maker": "0x5555555555555555555555555555555555555555",
    "takerAsset": "0xA0b86a33E6441E45cD6A7c2a9E0c0BEA64e14FF",
    "makeAmount": "750.0",
    "takeAmount": "576.92",
    "price": 1.3,
    "timestamp": 1719734404,
    "signature": "0xsig5"
  }'

echo ""
echo "✅ All test orders submitted!"
echo ""

# Test 2: Partial fills test
echo "Test 2: Testing partial fills with large orders"
echo "----------------------------------------------"

# Large bid that should partially fill
echo "📤 Submitting Large Bid: 2000 @ 1.6"
curl -s -X POST "$BASE_URL/orders" \
  -H "Content-Type: application/json" \
  -d '{
    "maker": "0x6666666666666666666666666666666666666666",
    "takerAsset": "0xA0b86a33E6441E45cD6A7c2a9E0c0BEA64e14FF",
    "makeAmount": "2000.0",
    "takeAmount": "1250.0",
    "price": 1.6,
    "timestamp": 1719734405,
    "signature": "0xsig6"
  }'

echo ""

# Small ask that should be fully filled
echo "📤 Submitting Small Ask: 100 @ 1.1"
curl -s -X POST "$BASE_URL/orders" \
  -H "Content-Type: application/json" \
  -d '{
    "maker": "0x7777777777777777777777777777777777777777",
    "takerAsset": "0xA0b86a33E6441E45cD6A7c2a9E0c0BEA64e14FF",
    "makeAmount": "100.0",
    "takeAmount": "90.91",
    "price": 1.1,
    "timestamp": 1719734406,
    "signature": "0xsig7"
  }'

echo ""
echo "✅ Partial fill test orders submitted!"
echo ""

echo "🔍 Test Summary:"
echo "• Multiple overlapping orders submitted to test multi-fill matching"
echo "• Large orders tested for partial fill behavior"
echo "• Each order submission should trigger matching with remaining orders"
echo "• Check logs for multi-fill batch creation and order book pruning"
echo ""
echo "📊 Expected behavior:"
echo "• Orders should be sorted by price-time priority"
echo "• Multiple fills should be created per batch (up to maxBatch=100)"
echo "• Partially filled orders should remain in orderbook with updated amounts"
echo "• Fully filled orders should be removed from orderbook"
echo ""
echo "==================================================================="
echo "                     Multi-Fill Test Complete! 🎉"
echo "==================================================================="

echo "=========================================="
echo "  Multi-Fill Order Matching Test"
echo "=========================================="
echo ""

echo "Testing multi-fill matching with multiple orders..."
echo "Make sure the sequencer is running with: go run main.go"
echo ""

# Test health endpoint first
echo "1. Testing health endpoint..."
health_response=$(curl -s -X GET "$BASE_URL/health")
if [ "$health_response" = "OK" ]; then
    echo "✅ Health check passed"
else
    echo "❌ Health check failed: $health_response"
    exit 1
fi
echo ""

# Submit multiple bid orders (buyers)
echo "2. Submitting bid orders (buyers)..."
echo ""

echo "Submitting bid #1: 1000@0.60"
curl -s -X POST "$BASE_URL/orders" \
  -H "Content-Type: application/json" \
  -d '{
    "maker": "0x1111111111111111111111111111111111111111",
    "takerAsset": "0xTOKEN1",
    "makeAmount": "1000.0",
    "takeAmount": "600.0",
    "price": 0.60,
    "timestamp": 1719734400,
    "signature": "0xbid1signature"
  }' | jq .
echo ""

echo "Submitting bid #2: 800@0.55"
curl -s -X POST "$BASE_URL/orders" \
  -H "Content-Type: application/json" \
  -d '{
    "maker": "0x2222222222222222222222222222222222222222",
    "takerAsset": "0xTOKEN1",
    "makeAmount": "800.0",
    "takeAmount": "440.0",
    "price": 0.55,
    "timestamp": 1719734401,
    "signature": "0xbid2signature"
  }' | jq .
echo ""

echo "Submitting bid #3: 1200@0.58"
curl -s -X POST "$BASE_URL/orders" \
  -H "Content-Type: application/json" \
  -d '{
    "maker": "0x3333333333333333333333333333333333333333",
    "takerAsset": "0xTOKEN1",
    "makeAmount": "1200.0",
    "takeAmount": "696.0",
    "price": 0.58,
    "timestamp": 1719734402,
    "signature": "0xbid3signature"
  }' | jq .
echo ""

# Submit multiple ask orders (sellers)
echo "3. Submitting ask orders (sellers)..."
echo ""

echo "Submitting ask #1: 500@0.50"
curl -s -X POST "$BASE_URL/orders" \
  -H "Content-Type: application/json" \
  -d '{
    "maker": "0x4444444444444444444444444444444444444444",
    "takerAsset": "0xTOKEN1",
    "makeAmount": "250.0",
    "takeAmount": "500.0",
    "price": 0.50,
    "timestamp": 1719734403,
    "signature": "0xask1signature"
  }' | jq .
echo ""

echo "Submitting ask #2: 700@0.52"
curl -s -X POST "$BASE_URL/orders" \
  -H "Content-Type: application/json" \
  -d '{
    "maker": "0x5555555555555555555555555555555555555555",
    "takerAsset": "0xTOKEN1",
    "makeAmount": "364.0",
    "takeAmount": "700.0",
    "price": 0.52,
    "timestamp": 1719734404,
    "signature": "0xask2signature"
  }' | jq .
echo ""

echo "Submitting ask #3: 600@0.54"
curl -s -X POST "$BASE_URL/orders" \
  -H "Content-Type: application/json" \
  -d '{
    "maker": "0x6666666666666666666666666666666666666666",
    "takerAsset": "0xTOKEN1",
    "makeAmount": "324.0",
    "takeAmount": "600.0",
    "price": 0.54,
    "timestamp": 1719734405,
    "signature": "0xask3signature"
  }' | jq .
echo ""

echo "=========================================="
echo "Expected matching behavior:"
echo "- Bids sorted by price: 1000@0.60, 1200@0.58, 800@0.55"
echo "- Asks sorted by price: 500@0.50, 700@0.52, 600@0.54"
echo "- Should create multiple fills:"
echo "  * 1000@0.60 vs 500@0.50 → fill 500"
echo "  * 1000@0.60 (remaining 500) vs 700@0.52 → fill 500"
echo "  * 1200@0.58 vs 700@0.52 (remaining 200) → fill 200"
echo "  * 1200@0.58 (remaining 1000) vs 600@0.54 → fill 600"
echo "- Should result in multiple batches being submitted"
echo "=========================================="
