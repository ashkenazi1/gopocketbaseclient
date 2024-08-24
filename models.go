package gopocketbaseclient

import (
	"encoding/json"
	"net/http"
)

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	Token      string
}

type BaseRecord struct {
	ID             string `json:"id"`
	CollectionID   string `json:"collectionId"`
	CollectionName string `json:"collectionName"`
	Created        string `json:"created"`
	Updated        string `json:"updated"`
}

type Record struct {
	BaseRecord
	Data map[string]interface{} `json:"data"`
}

type JSONItems struct {
	Items json.RawMessage `json:"items"`
}
