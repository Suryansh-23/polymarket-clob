package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/Layr-Labs/hourglass-avs-template/cmd/submitter"
)

func main() {
	action := flag.String("action", "status", "Action to perform: status, retry, or clear")
	flag.Parse()

	fmt.Println("=================================================================")
	fmt.Println("              POLYMARKET CLOB - FAILED BATCH MANAGER")
	fmt.Println("=================================================================")
	fmt.Println()

	switch *action {
	case "status":
		showStatus()
	case "retry":
		retryBatches()
	case "clear":
		clearBatches()
	default:
		fmt.Printf("Unknown action: %s\n", *action)
		fmt.Println("Available actions: status, retry, clear")
		os.Exit(1)
	}
}

func showStatus() {
	count := submitter.GetFailedBatchesCount()
	batches := submitter.GetFailedBatches()

	fmt.Printf("📊 Failed Batch Queue Status\n")
	fmt.Printf("Total failed batches: %d\n\n", count)

	if count == 0 {
		fmt.Println("✅ No failed batches in queue")
		return
	}

	fmt.Println("Failed Batches:")
	fmt.Println("---------------")

	for i, batch := range batches {
		fmt.Printf("%d. Root: %s\n", i+1, batch.Root)
		fmt.Printf("   Attempts: %d\n", batch.Attempts)
		fmt.Printf("   Timestamp: %v\n", batch.Timestamp.Format("2006-01-02 15:04:05"))
		fmt.Printf("   Fills Size: %d bytes\n", len(batch.Fills))
		fmt.Printf("   Signature Size: %d bytes\n\n", len(batch.Sig))
	}

	// Export as JSON for external tools
	jsonData, err := json.MarshalIndent(batches, "", "  ")
	if err == nil {
		fmt.Println("JSON Export:")
		fmt.Println("------------")
		fmt.Println(string(jsonData))
	}
}

func retryBatches() {
	count := submitter.GetFailedBatchesCount()
	
	if count == 0 {
		fmt.Println("✅ No failed batches to retry")
		return
	}

	fmt.Printf("🔄 Retrying %d failed batches...\n\n", count)

	err := submitter.RetryFailedBatches()
	if err != nil {
		fmt.Printf("❌ Some batches failed to retry: %v\n", err)
		
		remainingCount := submitter.GetFailedBatchesCount()
		if remainingCount > 0 {
			fmt.Printf("⚠️  %d batches still in failed queue\n", remainingCount)
		}
		os.Exit(1)
	}

	fmt.Println("✅ All failed batches retried successfully!")
	
	remainingCount := submitter.GetFailedBatchesCount()
	fmt.Printf("📊 Remaining failed batches: %d\n", remainingCount)
}

func clearBatches() {
	count := submitter.GetFailedBatchesCount()
	
	if count == 0 {
		fmt.Println("✅ No failed batches to clear")
		return
	}

	fmt.Printf("⚠️  About to clear %d failed batches from queue\n", count)
	fmt.Print("Are you sure? This action cannot be undone (y/N): ")

	var response string
	fmt.Scanln(&response)

	if response != "y" && response != "Y" && response != "yes" {
		fmt.Println("❌ Clear operation cancelled")
		return
	}

	cleared := submitter.ClearFailedBatches()
	fmt.Printf("🗑️  Cleared %d failed batches from queue\n", cleared)
	
	log.Printf("Failed batch queue cleared by retry utility")
}
