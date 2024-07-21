package main

import (
	"fmt"
	"log"

	mypocketbaseclient "github.com/ashkenazi1/gopocketbaseclient"
)

func main() {
	client := mypocketbaseclient.NewClient("https://xxx.pockethost.io", "your_jwt_token")

	// Create a new record
	// record := &mypocketbaseclient.Record{
	// 	Data: map[string]interface{}{
	// 		"field1": "value1",
	// 		"field2": "value2",
	// 	},
	// }
	// newRecord, err := client.CreateRecord("your_collection", record)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Printf("Created Record: %+v\n", newRecord)

	data, err := mypocketbaseclient.All(client, "collection_name")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Fetched Record: %s\n", data)
}
