package gopocketbaseclient

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// Compiled regex for better performance
var pocketBaseIDRegex = regexp.MustCompile("^[a-zA-Z0-9]+$")

// Authentication methods

// Login authenticates a user with email/username and password
func (c *Client) Login(identity, password string) (*AuthResponse, error) {
	if identity == "" {
		return nil, fmt.Errorf("identity cannot be empty")
	}
	if password == "" {
		return nil, fmt.Errorf("password cannot be empty")
	}

	loginData := LoginRequest{
		Identity: identity,
		Password: password,
	}

	respBody, err := c.doRequest("POST", "/api/collections/users/auth-with-password", loginData)
	if err != nil {
		return nil, fmt.Errorf("login failed: %w", err)
	}

	var authResponse AuthResponse
	err = json.Unmarshal(respBody, &authResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse login response: %w", err)
	}

	// Update client token
	c.Token = authResponse.Token

	return &authResponse, nil
}

// Register creates a new user account
func (c *Client) Register(req RegisterRequest) (*AuthResponse, error) {
	if req.Username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}
	if req.Email == "" {
		return nil, fmt.Errorf("email cannot be empty")
	}
	if req.Password == "" {
		return nil, fmt.Errorf("password cannot be empty")
	}
	if req.Password != req.PasswordConfirm {
		return nil, fmt.Errorf("password and password confirmation do not match")
	}

	respBody, err := c.doRequest("POST", "/api/collections/users/records", req)
	if err != nil {
		return nil, fmt.Errorf("registration failed: %w", err)
	}

	var authResponse AuthResponse
	err = json.Unmarshal(respBody, &authResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse registration response: %w", err)
	}

	return &authResponse, nil
}

// RefreshAuth refreshes the authentication token
func (c *Client) RefreshAuth() (*AuthResponse, error) {
	respBody, err := c.doRequest("POST", "/api/collections/users/auth-refresh", nil)
	if err != nil {
		return nil, fmt.Errorf("token refresh failed: %w", err)
	}

	var authResponse AuthResponse
	err = json.Unmarshal(respBody, &authResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse refresh response: %w", err)
	}

	// Update client token
	c.Token = authResponse.Token

	return &authResponse, nil
}

// Logout invalidates the current authentication token
func (c *Client) Logout() error {
	_, err := c.doRequest("POST", "/api/collections/users/auth-logout", nil)
	if err != nil {
		return fmt.Errorf("logout failed: %w", err)
	}

	// Clear client token
	c.Token = ""

	return nil
}

// GetCurrentUser returns the current authenticated user
func (c *Client) GetCurrentUser() (*User, error) {
	respBody, err := c.doRequest("GET", "/api/collections/users/auth-refresh", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	var authResponse AuthResponse
	err = json.Unmarshal(respBody, &authResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user response: %w", err)
	}

	return &authResponse.Record, nil
}

// RequestPasswordReset sends a password reset email
func (c *Client) RequestPasswordReset(email string) error {
	resetData := PasswordResetRequest{
		Email: email,
	}

	_, err := c.doRequest("POST", "/api/collections/users/request-password-reset", resetData)
	if err != nil {
		return fmt.Errorf("password reset request failed: %w", err)
	}

	return nil
}

// ConfirmPasswordReset confirms password reset with token
func (c *Client) ConfirmPasswordReset(token, password, passwordConfirm string) error {
	confirmData := PasswordResetConfirm{
		Token:           token,
		Password:        password,
		PasswordConfirm: passwordConfirm,
	}

	_, err := c.doRequest("POST", "/api/collections/users/confirm-password-reset", confirmData)
	if err != nil {
		return fmt.Errorf("password reset confirmation failed: %w", err)
	}

	return nil
}

// IsAuthenticated checks if the client has a valid token
func (c *Client) IsAuthenticated() bool {
	return c.Token != ""
}

// SetAuthToken manually sets the authentication token
func (c *Client) SetAuthToken(token string) {
	c.Token = token
}

// GetAuthToken returns the current authentication token
func (c *Client) GetAuthToken() string {
	return c.Token
}

// GetUser gets a user by ID (requires admin token or same user)
func (c *Client) GetUser(userID string) (*User, error) {
	endpoint := fmt.Sprintf("/api/collections/users/records/%s", userID)
	respBody, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	var user User
	err = UnmarshalPocketBaseJSON(respBody, &user)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user: %w", err)
	}

	return &user, nil
}

// UpdateUser updates user information
func (c *Client) UpdateUser(userID string, updates map[string]interface{}) (*User, error) {
	endpoint := fmt.Sprintf("/api/collections/users/records/%s", userID)
	respBody, err := c.doRequest("PATCH", endpoint, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	var user User
	err = UnmarshalPocketBaseJSON(respBody, &user)
	if err != nil {
		return nil, fmt.Errorf("failed to parse updated user: %w", err)
	}

	return &user, nil
}

// Smart relationship detection functions

// isPocketBaseID checks if a string looks like a PocketBase record ID
func isPocketBaseID(s string) bool {
	if len(s) != 15 {
		return false
	}
	// PocketBase IDs are 15-character alphanumeric strings
	return pocketBaseIDRegex.MatchString(s)
}

// isArrayOfPocketBaseIDs checks if an array contains PocketBase IDs
func isArrayOfPocketBaseIDs(arr []interface{}) bool {
	if len(arr) == 0 {
		return false
	}

	// Check first few elements to determine if it's an ID array
	checkCount := len(arr)
	if checkCount > 3 {
		checkCount = 3 // Check max 3 elements for performance
	}

	for i := 0; i < checkCount; i++ {
		if str, ok := arr[i].(string); ok {
			if !isPocketBaseID(str) {
				return false
			}
		} else {
			return false
		}
	}
	return true
}

// Removed automatic relationship detection functions as they relied on unreliable pattern matching

// GetRecordsWithExpand fetches records with explicit relationship expansion
func (c *Client) GetRecordsWithExpand(collection string, filters map[string]string, expandFields []string) (*JSONItems, error) {
	// Build filter string
	var filterParts []string
	for column, value := range filters {
		filterParts = append(filterParts, fmt.Sprintf("%s='%s'", column, value))
	}

	// Build base endpoint
	endpoint := fmt.Sprintf("/api/collections/%s/records", collection)

	// Add query parameters
	params := url.Values{}

	// Add filters if any
	if len(filterParts) > 0 {
		filterString := strings.Join(filterParts, " && ")
		params.Add("filter", fmt.Sprintf("(%s)", filterString))
	}

	// Add expand parameter
	if len(expandFields) > 0 {
		params.Add("expand", strings.Join(expandFields, ","))
	}

	// Build final URL
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

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

// Existing CRUD methods

func (c *Client) CreateRecord(collection string, record map[string]interface{}) error {
	if collection == "" {
		return fmt.Errorf("collection name cannot be empty")
	}
	if record == nil {
		return fmt.Errorf("record cannot be nil")
	}

	endpoint := "/api/collections/" + collection + "/records"
	respBody, err := c.doRequest("POST", endpoint, record)
	if err != nil {
		return fmt.Errorf("failed to create record: %w", err)
	}

	var createdRecord map[string]interface{}
	err = json.Unmarshal(respBody, &createdRecord)
	if err != nil {
		return fmt.Errorf("failed to unmarshal create record response: %w", err)
	}

	return nil
}

func (c *Client) GetRecords(collection string, filters map[string]string) (*JSONItems, error) {
	if collection == "" {
		return nil, fmt.Errorf("collection name cannot be empty")
	}

	var filterParts []string
	for column, value := range filters {
		filterParts = append(filterParts, fmt.Sprintf("%s='%s'", column, value))
	}
	filterString := strings.Join(filterParts, " && ")
	encodedFilterString := url.QueryEscape(fmt.Sprintf("(%s)", filterString))

	endpoint := fmt.Sprintf("/api/collections/%s/records?filter=%s", collection, encodedFilterString)
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

func (c *Client) UpdateRecord(collection, id string, record map[string]interface{}) error {
	if collection == "" {
		return fmt.Errorf("collection name cannot be empty")
	}
	if id == "" {
		return fmt.Errorf("record ID cannot be empty")
	}
	if record == nil {
		return fmt.Errorf("record cannot be nil")
	}

	endpoint := "/api/collections/" + collection + "/records/" + id
	respBody, err := c.doRequest("PATCH", endpoint, record)
	if err != nil {
		return err
	}

	var updatedRecord map[string]interface{}
	err = json.Unmarshal(respBody, &updatedRecord)
	if err != nil {
		return fmt.Errorf("failed to unmarshal update record response: %w", err)
	}

	return nil
}

func (c *Client) DeleteRecord(collection, id string) error {
	if collection == "" {
		return fmt.Errorf("collection name cannot be empty")
	}
	if id == "" {
		return fmt.Errorf("record ID cannot be empty")
	}

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

// Bulk Operations

// BulkOperationResult represents the result of a bulk operation
type BulkOperationResult struct {
	Success      []string    `json:"success"` // IDs of successful operations
	Failed       []BulkError `json:"failed"`  // Details of failed operations
	SuccessCount int         `json:"success_count"`
	FailureCount int         `json:"failure_count"`
	TotalCount   int         `json:"total_count"`
}

// BulkError represents an error in a bulk operation
type BulkError struct {
	Index int    `json:"index"` // Index in the original slice
	ID    string `json:"id"`    // Record ID (if available)
	Error string `json:"error"` // Error message
}

// UpsertRecord represents a record for bulk upsert operations
type UpsertRecord struct {
	ID   string                 `json:"id,omitempty"` // If provided, will attempt update first
	Data map[string]interface{} `json:"data"`         // The record data
}

// CreateMultipleRecords inserts many records at once using concurrent operations
func (c *Client) CreateMultipleRecords(collection string, records []map[string]interface{}) (*BulkOperationResult, error) {
	if collection == "" {
		return nil, fmt.Errorf("collection name cannot be empty")
	}
	if records == nil {
		return nil, fmt.Errorf("records cannot be nil")
	}
	if len(records) == 0 {
		return &BulkOperationResult{}, nil
	}

	// Limit concurrency to avoid overwhelming the server
	const maxConcurrency = 10
	semaphore := make(chan struct{}, maxConcurrency)

	type result struct {
		index int
		id    string
		err   error
	}

	results := make(chan result, len(records))

	// Process records concurrently
	for i, record := range records {
		go func(idx int, rec map[string]interface{}) {
			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			// Create the record
			endpoint := "/api/collections/" + collection + "/records"
			respBody, err := c.doRequest("POST", endpoint, rec)

			var id string
			if err == nil {
				// Extract ID from response
				var createdRecord map[string]interface{}
				if unmarshalErr := json.Unmarshal(respBody, &createdRecord); unmarshalErr == nil {
					if recordID, exists := createdRecord["id"]; exists {
						id = fmt.Sprintf("%v", recordID)
					}
				}
			}

			results <- result{index: idx, id: id, err: err}
		}(i, record)
	}

	// Collect results
	bulkResult := &BulkOperationResult{
		Success:    []string{},
		Failed:     []BulkError{},
		TotalCount: len(records),
	}

	for i := 0; i < len(records); i++ {
		res := <-results
		if res.err != nil {
			bulkResult.Failed = append(bulkResult.Failed, BulkError{
				Index: res.index,
				ID:    res.id,
				Error: res.err.Error(),
			})
			bulkResult.FailureCount++
		} else {
			bulkResult.Success = append(bulkResult.Success, res.id)
			bulkResult.SuccessCount++
		}
	}

	return bulkResult, nil
}

// UpdateMultipleRecords performs bulk updates with conditions
func (c *Client) UpdateMultipleRecords(collection string, updates []map[string]interface{}) (*BulkOperationResult, error) {
	if len(updates) == 0 {
		return &BulkOperationResult{}, nil
	}

	// Each update map must contain an "id" field
	const maxConcurrency = 10
	semaphore := make(chan struct{}, maxConcurrency)

	type result struct {
		index int
		id    string
		err   error
	}

	results := make(chan result, len(updates))

	// Process updates concurrently
	for i, update := range updates {
		go func(idx int, upd map[string]interface{}) {
			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			// Extract ID from update data
			idVal, exists := upd["id"]
			if !exists {
				results <- result{index: idx, err: fmt.Errorf("update record at index %d missing 'id' field", idx)}
				return
			}

			id := fmt.Sprintf("%v", idVal)

			// Remove ID from update data to avoid sending it in the body
			updateData := make(map[string]interface{})
			for k, v := range upd {
				if k != "id" {
					updateData[k] = v
				}
			}

			// Update the record
			endpoint := fmt.Sprintf("/api/collections/%s/records/%s", collection, id)
			_, err := c.doRequest("PATCH", endpoint, updateData)

			results <- result{index: idx, id: id, err: err}
		}(i, update)
	}

	// Collect results
	bulkResult := &BulkOperationResult{
		Success:    []string{},
		Failed:     []BulkError{},
		TotalCount: len(updates),
	}

	for i := 0; i < len(updates); i++ {
		res := <-results
		if res.err != nil {
			bulkResult.Failed = append(bulkResult.Failed, BulkError{
				Index: res.index,
				ID:    res.id,
				Error: res.err.Error(),
			})
			bulkResult.FailureCount++
		} else {
			bulkResult.Success = append(bulkResult.Success, res.id)
			bulkResult.SuccessCount++
		}
	}

	return bulkResult, nil
}

// DeleteMultipleRecords deletes multiple records efficiently
func (c *Client) DeleteMultipleRecords(collection string, ids []string) (*BulkOperationResult, error) {
	if len(ids) == 0 {
		return &BulkOperationResult{}, nil
	}

	const maxConcurrency = 10
	semaphore := make(chan struct{}, maxConcurrency)

	type result struct {
		index int
		id    string
		err   error
	}

	results := make(chan result, len(ids))

	// Process deletions concurrently
	for i, id := range ids {
		go func(idx int, recordID string) {
			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			// Delete the record
			endpoint := fmt.Sprintf("/api/collections/%s/records/%s", collection, recordID)
			_, err := c.doRequest("DELETE", endpoint, nil)

			results <- result{index: idx, id: recordID, err: err}
		}(i, id)
	}

	// Collect results
	bulkResult := &BulkOperationResult{
		Success:    []string{},
		Failed:     []BulkError{},
		TotalCount: len(ids),
	}

	for i := 0; i < len(ids); i++ {
		res := <-results
		if res.err != nil {
			bulkResult.Failed = append(bulkResult.Failed, BulkError{
				Index: res.index,
				ID:    res.id,
				Error: res.err.Error(),
			})
			bulkResult.FailureCount++
		} else {
			bulkResult.Success = append(bulkResult.Success, res.id)
			bulkResult.SuccessCount++
		}
	}

	return bulkResult, nil
}

// BulkUpsert inserts or updates records based on existence
func (c *Client) BulkUpsert(collection string, records []UpsertRecord) (*BulkOperationResult, error) {
	if len(records) == 0 {
		return &BulkOperationResult{}, nil
	}

	const maxConcurrency = 10
	semaphore := make(chan struct{}, maxConcurrency)

	type result struct {
		index int
		id    string
		err   error
	}

	results := make(chan result, len(records))

	// Process upserts concurrently
	for i, record := range records {
		go func(idx int, rec UpsertRecord) {
			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			var finalID string
			var err error

			// If ID is provided, try update first
			if rec.ID != "" {
				endpoint := fmt.Sprintf("/api/collections/%s/records/%s", collection, rec.ID)
				_, updateErr := c.doRequest("PATCH", endpoint, rec.Data)

				if updateErr == nil {
					// Update succeeded
					finalID = rec.ID
				} else {
					// Update failed, try create
					createEndpoint := "/api/collections/" + collection + "/records"
					respBody, createErr := c.doRequest("POST", createEndpoint, rec.Data)

					if createErr != nil {
						err = fmt.Errorf("update failed: %v, create failed: %v", updateErr, createErr)
					} else {
						// Extract ID from create response
						var createdRecord map[string]interface{}
						if unmarshalErr := json.Unmarshal(respBody, &createdRecord); unmarshalErr == nil {
							if recordID, exists := createdRecord["id"]; exists {
								finalID = fmt.Sprintf("%v", recordID)
							}
						}
					}
				}
			} else {
				// No ID provided, just create
				endpoint := "/api/collections/" + collection + "/records"
				respBody, createErr := c.doRequest("POST", endpoint, rec.Data)

				if createErr != nil {
					err = createErr
				} else {
					// Extract ID from response
					var createdRecord map[string]interface{}
					if unmarshalErr := json.Unmarshal(respBody, &createdRecord); unmarshalErr == nil {
						if recordID, exists := createdRecord["id"]; exists {
							finalID = fmt.Sprintf("%v", recordID)
						}
					}
				}
			}

			results <- result{index: idx, id: finalID, err: err}
		}(i, record)
	}

	// Collect results
	bulkResult := &BulkOperationResult{
		Success:    []string{},
		Failed:     []BulkError{},
		TotalCount: len(records),
	}

	for i := 0; i < len(records); i++ {
		res := <-results
		if res.err != nil {
			bulkResult.Failed = append(bulkResult.Failed, BulkError{
				Index: res.index,
				ID:    res.id,
				Error: res.err.Error(),
			})
			bulkResult.FailureCount++
		} else {
			bulkResult.Success = append(bulkResult.Success, res.id)
			bulkResult.SuccessCount++
		}
	}

	return bulkResult, nil
}
