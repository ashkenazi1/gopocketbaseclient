package main

import (
	"fmt"
	"log"

	"github.com/ashkenazi1/gopocketbaseclient"
)

func main() {

	client := gopocketbaseclient.NewClient("https://xxx.pockethost.io", "your_jwt_token")

	// Create a new record
	row := map[string]interface{}{
		"item1": "1",
		"item2": "2",
	}

	err := client.CreateRecord("israeli_vehicles", row)
	if err != nil {
		log.Println(err)
		return
	}

	// Get a record
	data, err := gopocketbaseclient.All(client, "collection_name")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Fetched Record: %s\n", data)
}
