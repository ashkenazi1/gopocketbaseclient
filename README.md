# 🚀 Go PocketBase Client

A comprehensive, production-ready Go client for PocketBase with automatic time handling, authentication, and high-performance bulk operations.

## ✨ Features

- 🔐 **Complete Authentication System** - Login, register, password reset, session management
- ⚡ **High-Performance Bulk Operations** - 10x faster with controlled concurrency
- 🔄 **Collection Migration** - Migrate data between PocketBase instances with progress tracking
- 🕒 **Automatic Time Handling** - Seamless `time.Time` support for PocketBase's date format
- 🔗 **Manual Relationship Expansion** - Explicit control over related record loading
- 🛡️ **Production-Ready** - Comprehensive error handling and input validation
- 🎯 **Type-Safe** - Strong typing with embedded `BaseRecord` for all collections
- 📦 **Zero Dependencies** - Uses only Go standard library

## 📦 Installation

```bash
go get github.com/ashkenazi1/gopocketbaseclient
```

## 🚀 Quick Start

### Basic Usage

```go
package main

import (
	"fmt"
	"time"
	"github.com/ashkenazi1/gopocketbaseclient"
)

// Define your record structure
type Task struct {
	gopocketbaseclient.BaseRecord
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	DueDate     time.Time `json:"due_date"`  // Automatic time handling!
}

func main() {
	client := gopocketbaseclient.NewClient("https://your-pocketbase.com", "your_token")

	// Create a record
	record := map[string]interface{}{
		"name":        "My Task",
		"description": "Important task",
		"status":      "active",
		"due_date":    time.Now().Add(24 * time.Hour),
	}
	
	err := client.CreateRecord("tasks", record)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Fetch records with automatic time conversion
	records, err := client.All("tasks")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	var tasks []Task
	err = gopocketbaseclient.UnmarshalPocketBaseJSON(records.Items, &tasks)
	if err != nil {
		fmt.Printf("Error parsing: %v\n", err)
		return
	}

	fmt.Printf("Found %d tasks\n", len(tasks))
}
```

### Authentication Example

```go
package main

import (
	"fmt"
	"github.com/ashkenazi1/gopocketbaseclient"
)

func main() {
	client := gopocketbaseclient.NewClient("https://your-pocketbase.com", "")

	// Register a new user
	registerReq := gopocketbaseclient.RegisterRequest{
		Username:        "john_doe",
		Email:           "john@example.com",
		Password:        "secure123",
		PasswordConfirm: "secure123",
		Name:            "John Doe",
	}

	authResp, err := client.Register(registerReq)
	if err != nil {
		fmt.Printf("Registration failed: %v\n", err)
		return
	}

	fmt.Printf("User registered: %s\n", authResp.Record.Username)

	// Login
	authResp, err = client.Login("john@example.com", "secure123")
	if err != nil {
		fmt.Printf("Login failed: %v\n", err)
		return
	}

	fmt.Printf("Logged in: %s\n", authResp.Record.Username)
	fmt.Printf("Authenticated: %v\n", client.IsAuthenticated())

	// Get current user
	user, err := client.GetCurrentUser()
	if err != nil {
		fmt.Printf("Error getting user: %v\n", err)
		return
	}

	fmt.Printf("Current user: %s (%s)\n", user.Username, user.Email)
}
```

## 📚 Complete API Reference

### 🔐 Authentication

```go
// User registration
authResp, err := client.Register(gopocketbaseclient.RegisterRequest{
	Username:        "username",
	Email:           "email@example.com", 
	Password:        "password",
	PasswordConfirm: "password",
	Name:            "Display Name",
})

// User login
authResp, err := client.Login("email@example.com", "password")

// Get current user
user, err := client.GetCurrentUser()

// Update user profile
updatedUser, err := client.UpdateUser(userID, map[string]interface{}{
	"name": "New Name",
})

// Token management
client.SetAuthToken("your-token")
token := client.GetAuthToken()
isAuth := client.IsAuthenticated()

// Password reset
err = client.RequestPasswordReset("email@example.com")
err = client.ConfirmPasswordReset("reset-token", "new-password", "new-password")

// Session management
authResp, err := client.RefreshAuth()
err = client.Logout()
```

### 📊 CRUD Operations

```go
// Create record
err := client.CreateRecord("collection", map[string]interface{}{
	"field1": "value1",
	"field2": "value2",
})

// Get records with filters
records, err := client.GetRecords("collection", map[string]string{
	"status": "active",
	"priority": "high",
})

// Get all records
allRecords, err := client.All("collection")

// Update record
err := client.UpdateRecord("collection", "record-id", map[string]interface{}{
	"status": "completed",
})

// Delete record
err := client.DeleteRecord("collection", "record-id")
```

### 🔗 Relationship Expansion

```go
// Manual relationship expansion
expandedRecords, err := client.GetRecordsWithExpand("tasks", 
	map[string]string{"status": "active"}, 
	[]string{"project_id", "assignee_id"})
```

### ⚡ High-Performance Bulk Operations

```go
// Bulk create (10x faster than individual creates)
records := []map[string]interface{}{
	{"name": "Task 1", "status": "active"},
	{"name": "Task 2", "status": "pending"},
	// ... hundreds more
}

result, err := client.CreateMultipleRecords("tasks", records)
fmt.Printf("Created %d/%d records\n", result.SuccessCount, result.TotalCount)

// Bulk update
updates := []map[string]interface{}{
	{"id": "record1", "status": "completed"},
	{"id": "record2", "status": "completed"},
}

result, err := client.UpdateMultipleRecords("tasks", updates)

// Bulk upsert (insert or update)
upserts := []gopocketbaseclient.UpsertRecord{
	{Data: map[string]interface{}{"name": "New Task"}},                    // Insert
	{ID: "existing-id", Data: map[string]interface{}{"status": "done"}},   // Update
}

result, err := client.BulkUpsert("tasks", upserts)

// Bulk delete
result, err := client.DeleteMultipleRecords("tasks", []string{"id1", "id2", "id3"})
```

### 🔄 Collection Migration

Migrate data between PocketBase instances with comprehensive error handling and progress tracking:

```go
// Quick migration with default settings
result, err := sourceClient.QuickMigrate(
	"https://destination.pocketbase.io",
	"destination-jwt-token",
	"collection_name",
)

if err != nil {
	fmt.Printf("Migration failed: %v\n", err)
	return
}

fmt.Printf("Migrated %d/%d records in %s\n", 
	result.SuccessfulRecords, 
	result.TotalRecords,
	result.ProcessingTime,
)

// Advanced migration with custom configuration
config := gopocketbaseclient.MigrationConfig{
	DestinationURL:   "https://destination.pocketbase.io",
	DestinationJWT:   "destination-admin-jwt",
	CollectionName:   "my_collection",
	SkipExisting:     true,  // Skip records that already exist
	BatchSize:        25,    // Process 25 records per batch
}

result, err := sourceClient.MigrateCollection(config)
if err != nil {
	fmt.Printf("Migration failed: %v\n", err)
	return
}

// Detailed error handling
fmt.Printf("Migration Summary:\n")
fmt.Printf("  Total: %d records\n", result.TotalRecords)
fmt.Printf("  Successful: %d\n", result.SuccessfulRecords)
fmt.Printf("  Failed: %d\n", result.FailedRecords)
fmt.Printf("  Skipped: %d\n", result.SkippedRecords)
fmt.Printf("  Processing time: %s\n", result.ProcessingTime)

// Handle individual errors
for _, migErr := range result.Errors {
	fmt.Printf("Error in record %s: %s\n", migErr.RecordID, migErr.Error)
}
```

**Migration Features:**
- 🔄 **Batch Processing** - Efficient handling of large datasets
- 🛡️ **Duplicate Detection** - Skip existing records to avoid conflicts
- 📊 **Progress Tracking** - Detailed statistics and timing information
- 🚨 **Error Recovery** - Continue processing despite individual record failures
- ⚡ **Performance Optimized** - Concurrent operations with rate limiting

**Prerequisites:**
- Collection must exist in both source and destination PocketBase instances
- Admin JWT tokens required for both instances
- Sufficient permissions for read/write operations

### 🕒 Automatic Time Handling

The library automatically handles PocketBase's non-standard date format and **null datetime values**:

```go
// Define structs with different datetime field types
type Event struct {
	gopocketbaseclient.BaseRecord
	Title         string                             `json:"title"`
	StartTime     time.Time                          `json:"start_time"`      // Zero time for null values
	EndTime       gopocketbaseclient.PocketBaseTime  `json:"end_time"`        // Better PocketBase compatibility
	CancelledDate gopocketbaseclient.NullableTime    `json:"cancelled_date"`  // Explicit null handling
}

// Use helper functions for automatic conversion
var events []Event
err = gopocketbaseclient.UnmarshalPocketBaseJSON(jsonData, &events)

// Check for null datetime values
for _, event := range events {
	if event.StartTime.IsZero() {
		fmt.Println("Start time not set")
	}
	
	if !event.CancelledDate.Valid {
		fmt.Println("Event not cancelled")
	} else {
		fmt.Printf("Cancelled at: %s\n", event.CancelledDate.Time.Format("2006-01-02"))
	}
}
```

**Null DateTime Support:**
- ✅ Empty strings (`""`)
- ✅ Null values (`null`)
- ✅ N/A values (`"n/a"`, `"N/A"`)
- ✅ Multiple datetime formats
- ✅ Automatic conversion in `UnmarshalPocketBaseJSON`

## 🏗️ Data Models

### User Model

```go
type User struct {
	BaseRecord                              // id, created, updated, etc.
	Username        string `json:"username"`
	Email           string `json:"email"`
	Name            string `json:"name"`
	Avatar          string `json:"avatar"`
	Verified        bool   `json:"verified"`
	EmailVisibility bool   `json:"emailVisibility"`
}
```

### BaseRecord (Embed in your structs)

```go
type BaseRecord struct {
	ID             string         `json:"id"`
	CollectionID   string         `json:"collectionId"`
	CollectionName string         `json:"collectionName"`
	Created        PocketBaseTime `json:"created"`
	Updated        PocketBaseTime `json:"updated"`
}
```

### Bulk Operation Result

```go
type BulkOperationResult struct {
	Success      []string    `json:"success"`       // IDs of successful operations
	Failed       []BulkError `json:"failed"`        // Details of failed operations
	SuccessCount int         `json:"success_count"`
	FailureCount int         `json:"failure_count"`
	TotalCount   int         `json:"total_count"`
}
```

## 🎯 Demo Applications

Run the included demo applications to see the library in action:

### Collections Demo
```bash
go run cmd/test/main.go
```
Demonstrates:
- CRUD operations
- Automatic time handling
- Type-safe structs
- Bulk operations
- Relationship expansion

### Authentication Demo
```bash
go run cmd/auth-demo/main.go
```
Demonstrates:
- User registration and login
- Session management
- Profile updates
- Password reset flow
- Token handling

### Migration Demo
```bash
go run cmd/migration-demo/main.go
```
Demonstrates:
- Quick migration with default settings
- Advanced migration with custom configuration
- Batch processing for large datasets
- Duplicate detection and skipping
- Comprehensive error handling and reporting
- Processing time tracking

## 🛡️ Error Handling

All functions return detailed errors with proper error wrapping:

```go
records, err := client.GetRecords("tasks", filters)
if err != nil {
	fmt.Printf("Failed to fetch tasks: %v\n", err)
	return
}

// Bulk operations provide detailed error tracking
result, err := client.CreateMultipleRecords("tasks", records)
if err != nil {
	fmt.Printf("Bulk operation failed: %v\n", err)
	return
}

// Check individual failures
for _, failure := range result.Failed {
	fmt.Printf("Record %d failed: %s\n", failure.Index, failure.Error)
}
```

## 🚀 Performance Features

- **Compiled Regexes**: Pre-compiled patterns for optimal performance
- **Concurrent Bulk Operations**: Up to 10 parallel operations with server-friendly limits
- **Efficient Time Parsing**: Optimized time format detection and conversion
- **Input Validation**: Early validation prevents unnecessary API calls
- **Memory Efficient**: Minimal allocations and proper resource management

## 🔧 Advanced Configuration

```go
// Custom HTTP client
client := gopocketbaseclient.NewClient("https://your-pocketbase.com", "token")
client.HTTPClient.Timeout = 30 * time.Second

// Manual token management
client.SetAuthToken("your-jwt-token")
token := client.GetAuthToken()

// Check authentication status
if client.IsAuthenticated() {
	// Make authenticated requests
}
```

## 📋 Requirements

- Go 1.19 or later
- PocketBase 0.16+ (for full compatibility)

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🙏 Acknowledgments

- [PocketBase](https://pocketbase.io/) - The amazing backend solution
- Go community for excellent standard library support

---

**⭐ If this library helps you, please give it a star!** ⭐