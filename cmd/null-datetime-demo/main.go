package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ashkenazi1/gopocketbaseclient"
)

// Invoice demonstrates different approaches to handling nullable datetime fields
type Invoice struct {
	gopocketbaseclient.BaseRecord
	Number        string                            `json:"number"`
	Amount        float64                           `json:"amount"`
	DueDate       time.Time                         `json:"due_date"`       // Regular time.Time - will be zero time if null
	PaidDate      gopocketbaseclient.PocketBaseTime `json:"paid_date"`      // PocketBaseTime - handles null gracefully
	CancelledDate gopocketbaseclient.NullableTime   `json:"cancelled_date"` // NullableTime - explicit null handling
}

func main() {
	fmt.Println("=== Null DateTime Handling Demo ===")

	// Simulate PocketBase JSON response with null/empty datetime fields
	jsonData := `{
		"items": [
			{
				"id": "inv001",
				"collectionId": "invoices",
				"collectionName": "invoices",
				"created": "2025-01-20 10:00:00.000Z",
				"updated": "2025-01-20 10:00:00.000Z",
				"number": "INV-001",
				"amount": 1500.50,
				"due_date": "2025-02-20 00:00:00.000Z",
				"paid_date": "",
				"cancelled_date": null
			},
			{
				"id": "inv002",
				"collectionId": "invoices", 
				"collectionName": "invoices",
				"created": "2025-01-20 11:00:00.000Z",
				"updated": "2025-01-20 11:00:00.000Z",
				"number": "INV-002",
				"amount": 750.25,
				"due_date": "2025-02-15 00:00:00.000Z",
				"paid_date": "2025-01-25 14:30:00.000Z",
				"cancelled_date": "n/a"
			},
			{
				"id": "inv003",
				"collectionId": "invoices",
				"collectionName": "invoices", 
				"created": "2025-01-20 12:00:00.000Z",
				"updated": "2025-01-20 12:00:00.000Z",
				"number": "INV-003",
				"amount": 2200.00,
				"due_date": "",
				"paid_date": null,
				"cancelled_date": "2025-01-22 09:15:00.000Z"
			}
		]
	}`

	fmt.Println("\n1. Testing null datetime handling with UnmarshalPocketBaseJSON:")

	// Parse the JSON using the library's enhanced unmarshaling
	var jsonItems gopocketbaseclient.JSONItems
	err := json.Unmarshal([]byte(jsonData), &jsonItems)
	if err != nil {
		fmt.Printf("Failed to unmarshal JSON: %v\n", err)
		return
	}

	var invoices []Invoice
	err = gopocketbaseclient.UnmarshalPocketBaseJSON(jsonItems.Items, &invoices)
	if err != nil {
		fmt.Printf("Failed to unmarshal invoice details: %v\n", err)
		return
	}

	fmt.Printf("✅ Successfully parsed %d invoices!\n", len(invoices))

	// Display the results
	for i, invoice := range invoices {
		fmt.Printf("\n--- Invoice %d: %s ---\n", i+1, invoice.Number)
		fmt.Printf("Amount: $%.2f\n", invoice.Amount)

		// Regular time.Time field
		if invoice.DueDate.IsZero() {
			fmt.Printf("Due Date: <not set>\n")
		} else {
			fmt.Printf("Due Date: %s\n", invoice.DueDate.Format("2006-01-02"))
		}

		// PocketBaseTime field
		if invoice.PaidDate.IsZero() {
			fmt.Printf("Paid Date: <not paid>\n")
		} else {
			fmt.Printf("Paid Date: %s\n", invoice.PaidDate.Format("2006-01-02 15:04"))
		}

		// NullableTime field
		if !invoice.CancelledDate.Valid {
			fmt.Printf("Cancelled Date: <not cancelled>\n")
		} else {
			fmt.Printf("Cancelled Date: %s\n", invoice.CancelledDate.Time.Format("2006-01-02 15:04"))
		}
	}

	fmt.Println("\n2. Testing direct parsing of problematic datetime strings:")

	// Test cases that would previously cause errors
	testCases := []string{
		`""`,                         // Empty string
		`"null"`,                     // Null value
		`"n/a"`,                      // N/A value
		`"N/A"`,                      // N/A uppercase
		`"2025-01-20 15:30:00.000Z"`, // Valid datetime
	}

	for _, testCase := range testCases {
		fmt.Printf("\nTesting: %s\n", testCase)

		// Test with PocketBaseTime
		var pbt gopocketbaseclient.PocketBaseTime
		err := pbt.UnmarshalJSON([]byte(testCase))
		if err != nil {
			fmt.Printf("  PocketBaseTime: ERROR - %v\n", err)
		} else if pbt.IsZero() {
			fmt.Printf("  PocketBaseTime: <zero time>\n")
		} else {
			fmt.Printf("  PocketBaseTime: %s\n", pbt.Format("2006-01-02 15:04:05"))
		}

		// Test with NullableTime
		var nt gopocketbaseclient.NullableTime
		err = nt.UnmarshalJSON([]byte(testCase))
		if err != nil {
			fmt.Printf("  NullableTime: ERROR - %v\n", err)
		} else if !nt.Valid {
			fmt.Printf("  NullableTime: <null>\n")
		} else {
			fmt.Printf("  NullableTime: %s\n", nt.Time.Format("2006-01-02 15:04:05"))
		}
	}

	fmt.Println("\n✅ All null datetime scenarios handled successfully!")
	fmt.Println("\n📋 Summary of Solutions:")
	fmt.Println("  • Use `time.Time` for regular datetime fields (zero time for null)")
	fmt.Println("  • Use `gopocketbaseclient.PocketBaseTime` for better PocketBase compatibility")
	fmt.Println("  • Use `gopocketbaseclient.NullableTime` when you need to distinguish null vs zero time")
	fmt.Println("  • The library automatically handles empty strings, 'null', and 'n/a' values")
}
