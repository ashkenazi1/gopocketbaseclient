package gopocketbaseclient

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"
)

func (c *Client) CreateRecord(collection string, record map[string]interface{}) error {
	endpoint := "/api/collections/" + collection + "/records"
	respBody, err := c.doRequest("POST", endpoint, record)
	if err != nil {
		return fmt.Errorf("failed to create record: %w", err)
	}

	var createdRecord map[string]interface{}
	err = json.Unmarshal(respBody, &createdRecord)
	if err != nil {
		log.Println("Error unmarshaling create record response:", err)
		return fmt.Errorf("failed to unmarshal create record response: %w", err)
	}

	return nil
}

func (c *Client) GetRecords(collection string, filters map[string]interface{}) (*JSONItems, error) {
	var filterParts []string
	for column, value := range filters {
		switch v := value.(type) {
		case string:
			filterParts = append(filterParts, fmt.Sprintf("%s='%s'", column, v))
		default:
			filterParts = append(filterParts, fmt.Sprintf("%s=%v", column, v))
		}
	}
	filterString := strings.Join(filterParts, " && ")
	encodedFilterString := url.QueryEscape(fmt.Sprintf("(%s)", filterString))

	endpoint := fmt.Sprintf("/api/collections/%s/records?filter=%s&perPage=10000000", collection, encodedFilterString)
	respBody, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var records JSONItems
	err = json.Unmarshal(respBody, &records)
	if err != nil {
		return nil, err
	}

	if len(records.Items) == 0 {
		return nil, fmt.Errorf("no records found")
	}

	return &records, nil
}

func (c *Client) All(collection string) (*JSONItems, error) {
	endpoint := fmt.Sprintf("/api/collections/%s/records?perPage=10000000", collection)
	respBody, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var data JSONItems
	err = json.Unmarshal(respBody, &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (c *Client) UpdateRecord(collection, id string, record map[string]interface{}) error {
	endpoint := "/api/collections/" + collection + "/records/" + id
	respBody, err := c.doRequest("PATCH", endpoint, record)
	if err != nil {
		return err
	}

	var updatedRecord map[string]interface{}
	err = json.Unmarshal(respBody, &updatedRecord)
	if err != nil {
		log.Println("Error unmarshaling response:", err)
		return err
	}

	return nil
}

func (c *Client) DeleteRecord(collection, id string) error {
	endpoint := "/api/collections/" + collection + "/records/" + id
	_, err := c.doRequest("DELETE", endpoint, nil)
	return err
}

func All(c *Client, collection string) (*JSONItems, error) {
	endpoint := fmt.Sprintf("/api/collections/%s/records?perPage=10000000", collection)
	respBody, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var data JSONItems
	err = json.Unmarshal(respBody, &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}
