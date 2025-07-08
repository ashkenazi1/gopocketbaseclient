package main

import (
	"fmt"
	"time"

	"github.com/ashkenazi1/gopocketbaseclient"
)

func main() {
	sourceURL := ""
	sourceJWT := ""

	destinationURL := ""
	destinationJWT := ""

	collectionName := ""

	fmt.Printf("🔗 Source: %s\n", sourceURL)
	fmt.Printf("🎯 Destination: %s\n", destinationURL)
	fmt.Printf("📦 Collection: %s\n", collectionName)

	sourceClient := gopocketbaseclient.NewClient(sourceURL, sourceJWT)
	destClient := gopocketbaseclient.NewClient(destinationURL, destinationJWT)

	fmt.Println("1. Testing source connection...")
	_, err := sourceClient.All(collectionName)
	if err != nil {
		fmt.Printf("   ❌ Source failed: %v\n", err)
		return
	}
	fmt.Printf("   ✅ Source connection successful\n")

	_, err = destClient.All(collectionName)
	if err != nil {
		fmt.Printf("   ❌ Destination failed: %v\n", err)
		fmt.Println("   💡 Issue: You're using a USER JWT, not ADMIN JWT for destination")
		fmt.Println("   💡 Migration requires ADMIN JWT tokens for both source and destination")
		return
	}
	fmt.Printf("   ✅ Destination connection successful\n")

	fmt.Println("🚀 Starting migration...")

	config := gopocketbaseclient.MigrationConfig{
		DestinationURL: destinationURL,
		DestinationJWT: destinationJWT,
		CollectionName: collectionName,
		SkipExisting:   false,
		BatchSize:      10,
	}

	result, err := sourceClient.MigrateCollection(config)
	if err != nil {
		fmt.Printf("❌ Migration failed: %v\n", err)
		return
	}

	fmt.Println("✅ Migration completed successfully!")
	fmt.Printf("📊 Summary: %s\n", result.Summary)
	fmt.Printf("📈 Records: %d total, %d successful, %d failed, %d skipped\n",
		result.TotalRecords,
		result.SuccessfulRecords,
		result.FailedRecords,
		result.SkippedRecords,
	)
	fmt.Printf("⏱️  Processing time: %s\n", result.ProcessingTime)

	if len(result.Errors) > 0 {
		fmt.Printf("⚠️  Errors encountered: %d\n", len(result.Errors))
		for i, migErr := range result.Errors {
			if i >= 3 { // Show first 3 errors
				fmt.Printf("   ... and %d more errors\n", len(result.Errors)-3)
				break
			}
			fmt.Printf("   • Record %s: %s\n", migErr.RecordID, migErr.Error)
		}
	}

	fmt.Println("\n🎉 Migration test complete!")
}
