# My PocketBase Client

A simple Go client for interacting with the PocketBase API.

## Installation

```sh
go get github.com/ashkenazi1/gopocketbaseclient
```

## Usage

Here's a quick example of how to use the PocketBase client:

```go
package main

import (
	"fmt"
	"log"

	"github.com/ashkenazi1/gopocketbaseclient"
)

func main() {
	client := gopocketbaseclient.NewClient("https://your-pocketbase-url.com", "your-jwt-token")

	// Create a new record
	record := map[string]interface{}{
		"field1": "value1",
		"field2": "value2",
	}
	err := client.CreateRecord("your-collection", record)
	if err != nil {
		log.Fatalf("Error creating record: %v", err)
	}

	// Get a specific record
	rec, err := client.GetRecord("your-collection", "record-id")
	if err != nil {
		log.Fatalf("Error getting record: %v", err)
	}
	fmt.Printf("Retrieved Record: %+v\n", rec)

	// Get all records
	allRecords, err := client.All("your-collection")
	if err != nil {
		log.Fatalf("Error getting all records: %v", err)
	}
	fmt.Printf("All Records: %+v\n", allRecords)
}

```

## Features
- Create, read, update, and delete records in PocketBase.
- Simple and intuitive API for interacting with the PocketBase API.

## Error Handling
Errors are returned as part of the method signatures, allowing you to handle them appropriately in your application.

## Contributing
Contributions are welcome! Please feel free to submit a pull request or open an issue for any suggestions or improvements.