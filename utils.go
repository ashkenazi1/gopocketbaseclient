package gopocketbaseclient

import (
	"encoding/json"
)

type JSONItems struct {
	Items json.RawMessage `json:"items"`
}

func (c *Client) CreateRecord(collection string, record *Record) (*Record, error) {
	endpoint := "/api/collections/" + collection + "/records"
	respBody, err := c.doRequest("POST", endpoint, record)
	if err != nil {
		return nil, err
	}

	var createdRecord Record
	err = json.Unmarshal(respBody, &createdRecord)
	if err != nil {
		return nil, err
	}

	return &createdRecord, nil
}

func (c *Client) GetRecord(collection, id string) (*Record, error) {
	endpoint := "/api/collections/" + collection + "/records/" + id
	respBody, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var record Record
	err = json.Unmarshal(respBody, &record)
	if err != nil {
		return nil, err
	}

	return &record, nil
}

func (c *Client) UpdateRecord(collection, id string, record *Record) (*Record, error) {
	endpoint := "/api/collections/" + collection + "/records/" + id
	respBody, err := c.doRequest("PATCH", endpoint, record)
	if err != nil {
		return nil, err
	}

	var updatedRecord Record
	err = json.Unmarshal(respBody, &updatedRecord)
	if err != nil {
		return nil, err
	}

	return &updatedRecord, nil
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
