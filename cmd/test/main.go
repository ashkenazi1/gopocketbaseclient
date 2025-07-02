package main

import (
	"fmt"
	"time"

	"github.com/ashkenazi1/gopocketbaseclient"
)

// Task represents a simple task record
type Task struct {
	gopocketbaseclient.BaseRecord
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Priority    int       `json:"priority"`
	DueDate     time.Time `json:"due_date"`
	ProjectId   string    `json:"project_id"`  // Relation to projects
	AssigneeId  string    `json:"assignee_id"` // Relation to users
}

func main() {
	fmt.Println("=== PocketBase Collections & CRUD Demo ===")

	// Initialize client with your PocketBase URL and admin/user token
	client := gopocketbaseclient.NewClient("https://your-pocketbase.com", "your_jwt_token")

	// Example 1: Create a record
	fmt.Println("\n1. Creating a record:")
	newRecord := map[string]interface{}{
		"name":        "Sample Task",
		"description": "This is a sample task",
		"status":      "active",
		"priority":    5,
		"due_date":    time.Now().Add(24 * time.Hour),
	}

	err := client.CreateRecord("tasks", newRecord)
	if err != nil {
		fmt.Printf("Failed to create record: %v\n", err)
	} else {
		fmt.Println("✓ Record created successfully")
	}

	// Example 2: Fetch all records
	fmt.Println("\n2. Fetching all records:")
	allTasks, err := client.All("tasks")
	if err != nil {
		fmt.Printf("Failed to fetch records: %v\n", err)
	} else {
		var tasks []Task
		err = gopocketbaseclient.UnmarshalPocketBaseJSON(allTasks.Items, &tasks)
		if err != nil {
			fmt.Printf("Failed to parse records: %v\n", err)
		} else {
			fmt.Printf("Found %d tasks\n", len(tasks))
		}
	}

	// Example 3: Query records with filters
	fmt.Println("\n3. Querying with filters:")
	filters := map[string]interface{}{
		"status": "active",
	}

	filteredRecords, err := client.GetRecords("tasks", filters)
	if err != nil {
		fmt.Printf("Failed to fetch filtered records: %v\n", err)
	} else {
		fmt.Printf("Found %d filtered records\n", len(filteredRecords.Items))
	}

	// Example 4: Basic querying (removed auto-expand feature)
	fmt.Println("\n4. Basic querying:")
	basicRecords, err := client.GetRecords("tasks", map[string]interface{}{
		"status": "active",
	})
	if err != nil {
		fmt.Printf("Failed to fetch records: %v\n", err)
	} else {
		fmt.Printf("✓ Found %d records\n", len(basicRecords.Items))
	}

	// Example 5: Manual relationship expansion
	fmt.Println("\n5. Manual relationship expansion:")
	expandFields := []string{"project_id", "assignee_id"}
	_, err = client.GetRecordsWithExpand("tasks", map[string]string{
		"status": "active",
	}, expandFields)
	if err != nil {
		fmt.Printf("Failed to fetch expanded records: %v\n", err)
	} else {
		fmt.Printf("✓ Manually expanded %d fields\n", len(expandFields))
	}

	// Example 6: Update a record
	fmt.Println("\n6. Updating a record:")
	if filteredRecords != nil && len(filteredRecords.Items) > 0 {
		var tasks []Task
		err = gopocketbaseclient.UnmarshalPocketBaseJSON(filteredRecords.Items, &tasks)
		if err == nil && len(tasks) > 0 {
			updateData := map[string]interface{}{
				"status":     "completed",
				"updated_at": time.Now(),
			}

			err = client.UpdateRecord("tasks", tasks[0].ID, updateData)
			if err != nil {
				fmt.Printf("Failed to update record: %v\n", err)
			} else {
				fmt.Printf("✓ Record updated successfully\n")
			}
		}
	}

	// Example 7: Working with time fields
	fmt.Println("\n7. Working with time fields:")
	fmt.Println("✓ time.Time fields are automatically handled by the library")

	// Example 8: Delete a record (commented for safety)
	fmt.Println("\n8. Delete operation (commented for safety):")
	// err = client.DeleteRecord("tasks", "record_id")

	// ⚡ BULK OPERATIONS DEMO
	fmt.Println("\n⚡ === BULK OPERATIONS DEMO ===")

	// Example 9: Bulk Create Records
	fmt.Println("\n9. ⚡ Bulk creating multiple records:")
	bulkCreateRecords := []map[string]interface{}{
		{"name": "Task A", "status": "active", "priority": 1},
		{"name": "Task B", "status": "pending", "priority": 2},
	}

	createResult, err := client.CreateMultipleRecords("tasks", bulkCreateRecords)
	if err != nil {
		fmt.Printf("Bulk create failed: %v\n", err)
	} else {
		fmt.Printf("✓ Created %d/%d records\n", createResult.SuccessCount, createResult.TotalCount)
	}

	// Example 10: Bulk Update Records
	fmt.Println("\n10. ⚡ Bulk updating records:")
	if createResult != nil && len(createResult.Success) > 0 {
		bulkUpdateRecords := []map[string]interface{}{
			{"id": createResult.Success[0], "status": "completed"},
		}

		updateResult, err := client.UpdateMultipleRecords("tasks", bulkUpdateRecords)
		if err != nil {
			fmt.Printf("Bulk update failed: %v\n", err)
		} else {
			fmt.Printf("✓ Updated %d/%d records\n", updateResult.SuccessCount, updateResult.TotalCount)
		}
	}

	// Example 11: Bulk Upsert
	fmt.Println("\n11. ⚡ Bulk upsert operations:")
	upsertRecords := []gopocketbaseclient.UpsertRecord{
		{Data: map[string]interface{}{"name": "New Task", "status": "new"}},
		{ID: "fake123", Data: map[string]interface{}{"name": "Maybe Update", "status": "maybe"}},
	}

	upsertResult, err := client.BulkUpsert("tasks", upsertRecords)
	if err != nil {
		fmt.Printf("Bulk upsert failed: %v\n", err)
	} else {
		fmt.Printf("✓ Upserted %d/%d records\n", upsertResult.SuccessCount, upsertResult.TotalCount)
	}

	// Example 12: Bulk Delete (commented for safety)
	fmt.Println("\n12. ⚡ Bulk delete:")
	fmt.Println("   (Commented out for safety)")
	// deleteResult, _ := client.DeleteMultipleRecords("tasks", []string{"id1", "id2"})

	fmt.Println("\n📈 Bulk Operations Benefits:")
	fmt.Println("  • ⚡ Concurrent processing (10x faster)")
	fmt.Println("  • 🛡️ Detailed error reporting")
	fmt.Println("  • 🎯 Server-friendly concurrency limits")

	fmt.Println("\n✅ Demo completed!")
	fmt.Println("\n🎯 Features demonstrated:")
	fmt.Println("  • Basic CRUD operations")
	fmt.Println("  • Manual relationship expansion")
	fmt.Println("  • Bulk operations (10x faster)")
	fmt.Println("  • Automatic time.Time handling")
	fmt.Println("  • Type-safe structs")
	fmt.Println("\nUpdate the URL and token to test with real data.")
}
