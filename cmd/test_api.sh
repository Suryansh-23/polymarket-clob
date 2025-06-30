#!/bin/bash

# Test script for Polymarket CLOB Sequencer API

BASE_URL="http://localhost:8081"

echo "Testing Polymarket CLOB Sequencer API..."
echo "Make sure the sequencer is running with: go run main.go"
echo ""

# Test health endpoint
echo "1. Testing health endpoint..."
curl -s -X GET "$BASE_URL/health"
echo ""
echo ""

# Test order submission with a valid order
echo "2. Testing order submission (valid order)..."
curl -s -X POST "$BASE_URL/orders" \
  -H "Content-Type: application/json" \
  -d '{
    "maker": "0x742b35Cc6834C532532fa5A32b66F8d6C1F3b0B1",
    "takerAsset": "0x2e8a51B19f2bbE1FfA3d3F14D7E72F1C00E28Ef5",
    "makeAmount": "1000.0",
    "takeAmount": "500.0", 
    "price": 0.5,
    "timestamp": 1719734400,
    "signature": "0x1234567890abcdef"
  }' | jq .
echo ""

# Test order submission with another valid order to trigger matching
echo "3. Testing second order submission to trigger matching..."
curl -s -X POST "$BASE_URL/orders" \
  -H "Content-Type: application/json" \
  -d '{
    "maker": "0x123b35Cc6834C532532fa5A32b66F8d6C1F3b0B2",
    "takerAsset": "0x2e8a51B19f2bbE1FfA3d3F14D7E72F1C00E28Ef5",
    "makeAmount": "600.0",
    "takeAmount": "300.0",
    "price": 0.6,
    "timestamp": 1719734500,
    "signature": "0xabcdef1234567890"
  }' | jq .
echo ""

# Test order submission with invalid data
echo "4. Testing order submission (invalid order)..."
curl -s -X POST "$BASE_URL/orders" \
  -H "Content-Type: application/json" \
  -d '{
    "maker": "",
    "takerAsset": "0x2e8a51B19f2bbE1FfA3d3F14D7E72F1C00E28Ef5",
    "makeAmount": "invalid",
    "takeAmount": "500.0",
    "price": -1,
    "timestamp": 0,
    "signature": ""
  }' | jq .
echo ""

echo "API tests completed!"
