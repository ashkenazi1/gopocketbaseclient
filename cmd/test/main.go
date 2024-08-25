package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/ashkenazi1/gopocketbaseclient"
)

type Job struct {
	CountryCode         string         `json:"country_code"`
	IncludeExclude      IncludeExclude `json:"include_exclude"`
	OptimizeImpressions int            `json:"optimize_impressions"`
	QueryMinImpressions int            `json:"query_min_impressions"`
	ReportOnly          bool           `json:"report_only"`
	TrafficSourceID     int            `json:"traffic_source_id"`
	TrafficSourceName   string         `json:"traffic_source_name"`
	ZoneIdSpend         int            `json:"zone_id_spend"`
}
type IncludeExclude struct {
	Include []int `json:"include"`
	Exclude []int `json:"exclude"`
}

func main() {

	client := gopocketbaseclient.NewClient("https://xxx.pockethost.io", "your_jwt_token")

	// Create a new record
	row := map[string]interface{}{
		"item1": "1",
		"item2": "2",
	}

	client.CreateRecord("traffic_optimizer", row)

	// Get record/s with filter
	filters := map[string]string{
		"name":      "John Doe",
		"status":    "active",
		"createdAt": "2023-08-25",
	}

	// Fetch records with the specified filters
	records, err := client.GetRecords("users", filters)
	if err != nil {
		fmt.Println("Error fetching records:", err)
		return
	}

	fmt.Printf("Fetched Record: %v\n", records)

	// All records
	data, err := gopocketbaseclient.All(client, "traffic_optimizer")
	if err != nil {
		log.Fatal(err)
	}

	var jobs []Job
	err = json.Unmarshal([]byte(data.Items), &jobs)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Fetched Record: %v\n", jobs)
}
