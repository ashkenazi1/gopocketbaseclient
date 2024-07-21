package gopocketbaseclient

import "net/http"

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
	BaseRecord BaseRecord
	Data       map[string]interface{} `json:"data"`
}
