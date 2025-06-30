package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/Layr-Labs/hourglass-avs-template/cmd/matcher"
)

func main() {
	log.Println("Testing multi-fill matcher...")

	// Create test orders
	orders := []matcher.Order{
		{
			Maker:      "0x1111",
			TakerAsset: "0xTOKEN1",
			MakeAmount: "1000.0",
			TakeAmount: "600.0",
			Price:      0.60,
			Timestamp:  1719734400,
			Signature:  "0xbid1",
			IsBid:      true,
		},
		{
			Maker:      "0x2222",
			TakerAsset: "0xTOKEN1",
			MakeAmount: "300.0",
			TakeAmount: "500.0",
			Price:      0.50,
			Timestamp:  1719734401,
			Signature:  "0xask1",
			IsBid:      false,
		},
		{
			Maker:      "0x3333",
			TakerAsset: "0xTOKEN1",
			MakeAmount: "800.0",
			TakeAmount: "480.0",
			Price:      0.55,
			Timestamp:  1719734402,
			Signature:  "0xbid2",
			IsBid:      true,
		},
	}

	fmt.Printf("Starting with %d orders\n", len(orders))

	// Test matching
	root, fillsBytes, remaining, err := matcher.MatchAndBatch(orders, 10)
	if err != nil {
		log.Fatalf("Error in matching: %v", err)
	}

	fmt.Printf("Matching result:\n")
	fmt.Printf("  Root: %s\n", root)
	fmt.Printf("  Fills bytes length: %d\n", len(fillsBytes))
	fmt.Printf("  Remaining orders: %d\n", len(remaining))

	if len(fillsBytes) > 0 {
		var fills []matcher.Fill
		json.Unmarshal(fillsBytes, &fills)
		fmt.Printf("  Fills created: %d\n", len(fills))
		for i, fill := range fills {
			fmt.Printf("    Fill %d: %s\n", i+1, fill.Quantity)
		}
	}

	fmt.Println("Multi-fill matcher test completed successfully!")
}
