package gopocketbaseclient

import (
	"encoding/json"
	"fmt"
	"log"
)

func (c *Client) CreateRecord(collection string, record map[string]interface{}) error {
	endpoint := "/api/collections/" + collection + "/records"
	respBody, err := c.doRequest("POST", endpoint, record)
	if err != nil {
		return err
	}

	var createdRecord map[string]interface{}
	err = json.Unmarshal(respBody, &createdRecord)
	if err != nil {
		log.Println("Error unmarshaling response:", err)
		return err
	}

	return nil
}

func (c *Client) GetRecords(collection, column string, value string) (*JSONItems, error) {
	endpoint := fmt.Sprintf("/api/collections/%s/records/?filter=(%s='%s')", collection, column, value)
	respBody, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var records JSONItems
	err = json.Unmarshal(respBody, &records)
	if err != nil {
		return nil, err
	}

	return &records, nil
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
	endpoint := "/api/collections/" + collection + "/records"
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
