package gopocketbaseclient

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Compiled regex for better performance
var pocketBaseIDRegex = regexp.MustCompile("^[a-zA-Z0-9]{15}$")

// Sync pools for reusing objects and reducing allocations
var stringBuilderPool = sync.Pool{
	New: func() interface{} {
		return &strings.Builder{}
	},
}

// Buffer pool for JSON operations
var bufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, 1024) // Pre-allocate 1KB buffer
	},
}

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
	err = UnmarshalPocketBaseJSON(respBody, &authResponse)
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

	// Create user record first
	respBody, err := c.doRequest("POST", "/api/collections/users/records", req)
	if err != nil {
		return nil, fmt.Errorf("registration failed: %w", err)
	}

	var user User
	err = UnmarshalPocketBaseJSON(respBody, &user)
	if err != nil {
		return nil, fmt.Errorf("failed to parse registration response: %w", err)
	}

	// Try to authenticate the user to get the token (optional)
	authResp, err := c.Login(req.Email, req.Password)
	if err != nil {
		// If authentication fails, still return success with the user data
		// but no token (user was created successfully)
		return &AuthResponse{
			Token:  "",
			Record: user,
		}, nil
	}

	return authResp, nil
}

// CreateUser creates a new user account without authentication
func (c *Client) CreateUser(req RegisterRequest) (*User, error) {
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

	// Create user record only
	respBody, err := c.doRequest("POST", "/api/collections/users/records", req)
	if err != nil {
		return nil, fmt.Errorf("user creation failed: %w", err)
	}

	var user User
	err = UnmarshalPocketBaseJSON(respBody, &user)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user creation response: %w", err)
	}

	return &user, nil
}

// RefreshAuth refreshes the authentication token
func (c *Client) RefreshAuth() (*AuthResponse, error) {
	respBody, err := c.doRequest("POST", "/api/collections/users/auth-refresh", nil)
	if err != nil {
		return nil, fmt.Errorf("token refresh failed: %w", err)
	}

	var authResponse AuthResponse
	err = UnmarshalPocketBaseJSON(respBody, &authResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse refresh response: %w", err)
	}

	// Update client token
	c.Token = authResponse.Token

	return &authResponse, nil
}

// Logout invalidates the current authentication token
func (c *Client) Logout() error {
	// Clear client token (logout is primarily client-side)
	c.Token = ""

	// PocketBase doesn't have a server-side logout endpoint in all versions
	// Token invalidation is handled by clearing the client token
	return nil
}

// GetCurrentUser returns the current authenticated user
func (c *Client) GetCurrentUser() (*User, error) {
	if c.Token == "" {
		return nil, fmt.Errorf("no authentication token available")
	}

	// Use auth-refresh to get current user info
	respBody, err := c.doRequest("POST", "/api/collections/users/auth-refresh", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	var authResponse AuthResponse
	err = UnmarshalPocketBaseJSON(respBody, &authResponse)
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

// ValidateJWT validates a JWT token by making a request to PocketBase
// This is a package-level function that can be used in applications to verify client tokens
// Since we don't have access to the JWT secret, we validate by attempting to get user info
// Creates a completely separate client instance to avoid affecting the original client's token
func ValidateJWT(c *Client, token string) (*JWTValidationResponse, error) {
	if token == "" {
		return &JWTValidationResponse{
			Valid: false,
			Error: "token cannot be empty",
		}, fmt.Errorf("token cannot be empty")
	}

	// Create a completely separate client instance for validation
	// This ensures we don't interfere with the original client's admin token
	validationClient := NewClient(c.BaseURL, token)

	// Try to get current user info to validate the token
	user, err := validationClient.GetCurrentUser()
	if err != nil {
		return &JWTValidationResponse{
			Valid: false,
			Error: fmt.Sprintf("token validation failed: %v", err),
		}, nil // Don't return error here, validation failed but function succeeded
	}

	return &JWTValidationResponse{
		Valid:  true,
		UserID: user.ID,
		Email:  user.Email,
	}, nil
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
	// PocketBase IDs are exactly 15-character alphanumeric strings
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

// GetRecordsWithExpand fetches records with explicit relationship expansion
func (c *Client) GetRecordsWithExpand(collection string, filters map[string]interface{}, expandFields []string) (*JSONItems, error) {
	// Use string builder from pool for better performance
	builder := stringBuilderPool.Get().(*strings.Builder)
	builder.Reset()
	defer stringBuilderPool.Put(builder)

	// Build filter string with proper type handling
	var filterParts []string
	for column, value := range filters {
		formattedValue := formatFilterValue(value)
		filterParts = append(filterParts, fmt.Sprintf("%s=%s", column, formattedValue))
	}

	// Build base endpoint
	endpoint := fmt.Sprintf("/api/collections/%s/records", collection)

	// Add query parameters
	params := url.Values{}

	// Disable pagination by fetching all records
	params.Add("perPage", "-1")

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
	err = UnmarshalPocketBaseJSON(respBody, &records)
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
	err = UnmarshalPocketBaseJSON(respBody, &createdRecord)
	if err != nil {
		return fmt.Errorf("failed to unmarshal create record response: %w", err)
	}

	return nil
}

// formatFilterValue properly formats a value for PocketBase filter queries
func formatFilterValue(value interface{}) string {
	if value == nil {
		return "null"
	}

	switch v := value.(type) {
	case string:
		// Escape single quotes and wrap in single quotes
		escaped := strings.ReplaceAll(v, "'", "\\'")
		return fmt.Sprintf("'%s'", escaped)
	case bool:
		return fmt.Sprintf("%t", v)
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%g", v)
	default:
		// For other types, convert to string and treat as string
		str := fmt.Sprintf("%v", v)
		escaped := strings.ReplaceAll(str, "'", "\\'")
		return fmt.Sprintf("'%s'", escaped)
	}
}

func (c *Client) GetRecords(collection string, filters map[string]interface{}) (*JSONItems, error) {
	if collection == "" {
		return nil, fmt.Errorf("collection name cannot be empty")
	}

	// Use string builder from pool for better performance
	builder := stringBuilderPool.Get().(*strings.Builder)
	builder.Reset()
	defer stringBuilderPool.Put(builder)

	// Build filter string efficiently
	builder.WriteByte('(')
	first := true
	for column, value := range filters {
		if !first {
			builder.WriteString(" && ")
		}
		builder.WriteString(column)
		builder.WriteByte('=')
		builder.WriteString(formatFilterValue(value))
		first = false
	}
	builder.WriteByte(')')

	encodedFilterString := url.QueryEscape(builder.String())

	endpoint := fmt.Sprintf("/api/collections/%s/records?perPage=-1&filter=%s", collection, encodedFilterString)
	respBody, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var records JSONItems
	err = UnmarshalPocketBaseJSON(respBody, &records)
	if err != nil {
		return nil, err
	}

	if len(records.Items) == 0 {
		return nil, fmt.Errorf("no records found")
	}

	return &records, nil
}

func (c *Client) All(collection string) (*JSONItems, error) {
	endpoint := "/api/collections/" + collection + "/records?perPage=-1"
	respBody, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var data JSONItems
	err = UnmarshalPocketBaseJSON(respBody, &data)
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
	err = UnmarshalPocketBaseJSON(respBody, &updatedRecord)
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
				if unmarshalErr := UnmarshalPocketBaseJSON(respBody, &createdRecord); unmarshalErr == nil {
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
						if unmarshalErr := UnmarshalPocketBaseJSON(respBody, &createdRecord); unmarshalErr == nil {
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
					if unmarshalErr := UnmarshalPocketBaseJSON(respBody, &createdRecord); unmarshalErr == nil {
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

// Migration functions

// MigrateCollection migrates all records from the current PocketBase collection to another PocketBase instance
func (c *Client) MigrateCollection(config MigrationConfig) (*MigrationResult, error) {
	startTime := time.Now()

	// Validate configuration
	if err := validateMigrationConfig(config); err != nil {
		return nil, fmt.Errorf("invalid migration config: %w", err)
	}

	// Set default batch size if not specified
	if config.BatchSize <= 0 {
		config.BatchSize = 50
	}

	// Create destination client
	destClient := NewClient(config.DestinationURL, config.DestinationJWT)

	// Test destination connection
	if err := testDestinationConnection(destClient, config.CollectionName); err != nil {
		return nil, fmt.Errorf("destination connection failed: %w", err)
	}

	// Get all records from source collection
	sourceRecords, err := c.getAllRecordsForMigration(config.CollectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch source records: %w", err)
	}

	if len(sourceRecords) == 0 {
		return &MigrationResult{
			SourceCollection:      config.CollectionName,
			DestinationCollection: config.CollectionName,
			TotalRecords:          0,
			SuccessfulRecords:     0,
			FailedRecords:         0,
			SkippedRecords:        0,
			ProcessingTime:        time.Since(startTime).String(),
			Errors:                []MigrationError{},
			Summary:               "No records found in source collection",
		}, nil
	}

	// Migrate records in batches
	result, err := c.migrateRecordsBatch(destClient, sourceRecords, config)
	if err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	// Update result with timing and summary
	result.ProcessingTime = time.Since(startTime).String()
	result.SourceCollection = config.CollectionName
	result.DestinationCollection = config.CollectionName
	result.Summary = fmt.Sprintf("Migration completed: %d/%d records successfully migrated",
		result.SuccessfulRecords, result.TotalRecords)

	return result, nil
}

// validateMigrationConfig validates the migration configuration
func validateMigrationConfig(config MigrationConfig) error {
	if config.DestinationURL == "" {
		return fmt.Errorf("destination URL cannot be empty")
	}
	if config.DestinationJWT == "" {
		return fmt.Errorf("destination JWT cannot be empty")
	}
	if config.CollectionName == "" {
		return fmt.Errorf("collection name cannot be empty")
	}
	if config.BatchSize < 0 {
		return fmt.Errorf("batch size cannot be negative")
	}
	return nil
}

// testDestinationConnection tests if the destination PocketBase is accessible and the collection exists
func testDestinationConnection(destClient *Client, collectionName string) error {
	// Try to get records from the destination collection to verify it exists
	_, err := destClient.All(collectionName)
	if err != nil {
		return fmt.Errorf("cannot access destination collection '%s': %w", collectionName, err)
	}
	return nil
}

// getAllRecordsForMigration fetches all records from the source collection
func (c *Client) getAllRecordsForMigration(collection string) ([]MigrationRecord, error) {
	// Get all records from the collection
	jsonItems, err := c.All(collection)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch records: %w", err)
	}

	// Parse the JSON items
	var records []map[string]interface{}
	if err := UnmarshalPocketBaseJSON(jsonItems.Items, &records); err != nil {
		return nil, fmt.Errorf("failed to parse records: %w", err)
	}

	// Convert to migration records
	migrationRecords := make([]MigrationRecord, 0, len(records))
	for _, record := range records {
		migRecord := MigrationRecord{
			Data: make(map[string]interface{}),
		}

		// Extract special fields
		if id, exists := record["id"]; exists {
			migRecord.SourceID = fmt.Sprintf("%v", id)
		}
		if created, exists := record["created"]; exists {
			migRecord.Created = fmt.Sprintf("%v", created)
		}
		if updated, exists := record["updated"]; exists {
			migRecord.Updated = fmt.Sprintf("%v", updated)
		}

		// Copy all data fields except system fields
		for key, value := range record {
			if !isSystemField(key) {
				migRecord.Data[key] = value
			}
		}

		migrationRecords = append(migrationRecords, migRecord)
	}

	return migrationRecords, nil
}

// isSystemField checks if a field is a system field that shouldn't be migrated
func isSystemField(fieldName string) bool {
	systemFields := []string{"id", "created", "updated", "collectionId", "collectionName"}
	for _, field := range systemFields {
		if fieldName == field {
			return true
		}
	}
	return false
}

// migrateRecordsBatch migrates records in batches to the destination
func (c *Client) migrateRecordsBatch(destClient *Client, records []MigrationRecord, config MigrationConfig) (*MigrationResult, error) {
	result := &MigrationResult{
		TotalRecords:      len(records),
		SuccessfulRecords: 0,
		FailedRecords:     0,
		SkippedRecords:    0,
		Errors:            []MigrationError{},
	}

	// Process records in batches
	for i := 0; i < len(records); i += config.BatchSize {
		end := i + config.BatchSize
		if end > len(records) {
			end = len(records)
		}

		batch := records[i:end]
		batchResult, err := c.migrateBatch(destClient, batch, config, i)
		if err != nil {
			return nil, fmt.Errorf("batch migration failed: %w", err)
		}

		// Aggregate results
		result.SuccessfulRecords += batchResult.SuccessfulRecords
		result.FailedRecords += batchResult.FailedRecords
		result.SkippedRecords += batchResult.SkippedRecords
		result.Errors = append(result.Errors, batchResult.Errors...)
	}

	return result, nil
}

// migrateBatch migrates a single batch of records
func (c *Client) migrateBatch(destClient *Client, batch []MigrationRecord, config MigrationConfig, startIndex int) (*MigrationResult, error) {
	result := &MigrationResult{
		SuccessfulRecords: 0,
		FailedRecords:     0,
		SkippedRecords:    0,
		Errors:            []MigrationError{},
	}

	for i, record := range batch {
		recordIndex := startIndex + i

		// Check if record already exists if skip_existing is enabled
		if config.SkipExisting {
			exists, err := c.recordExistsInDestination(destClient, config.CollectionName, record)
			if err != nil {
				result.Errors = append(result.Errors, MigrationError{
					RecordID:    record.SourceID,
					RecordIndex: recordIndex,
					Operation:   "existence_check",
					Error:       err.Error(),
				})
				result.FailedRecords++
				continue
			}
			if exists {
				result.SkippedRecords++
				continue
			}
		}

		// Create record in destination
		err := destClient.CreateRecord(config.CollectionName, record.Data)
		if err != nil {
			result.Errors = append(result.Errors, MigrationError{
				RecordID:    record.SourceID,
				RecordIndex: recordIndex,
				Operation:   "create",
				Error:       err.Error(),
			})
			result.FailedRecords++
		} else {
			result.SuccessfulRecords++
		}
	}

	return result, nil
}

// recordExistsInDestination checks if a record with similar data already exists in the destination
func (c *Client) recordExistsInDestination(destClient *Client, collectionName string, record MigrationRecord) (bool, error) {
	// This is a simple implementation that checks for exact data matches
	// In a real-world scenario, you might want to implement more sophisticated matching logic

	// Get all records from destination and compare
	jsonItems, err := destClient.All(collectionName)
	if err != nil {
		return false, err
	}

	var destRecords []map[string]interface{}
	if err := UnmarshalPocketBaseJSON(jsonItems.Items, &destRecords); err != nil {
		return false, err
	}

	// Simple comparison - check if any record has the same data
	for _, destRecord := range destRecords {
		if recordsMatch(record.Data, destRecord) {
			return true, nil
		}
	}

	return false, nil
}

// recordsMatch compares two records to see if they match (for duplicate detection)
func recordsMatch(record1 map[string]interface{}, record2 map[string]interface{}) bool {
	// Simple field-by-field comparison excluding system fields
	for key, value1 := range record1 {
		if isSystemField(key) {
			continue
		}

		value2, exists := record2[key]
		if !exists {
			return false
		}

		// Convert both values to strings for comparison
		str1 := fmt.Sprintf("%v", value1)
		str2 := fmt.Sprintf("%v", value2)

		if str1 != str2 {
			return false
		}
	}

	return true
}

// QuickMigrate provides a simple way to migrate a collection with default settings
func (c *Client) QuickMigrate(destinationURL, destinationJWT, collectionName string) (*MigrationResult, error) {
	config := MigrationConfig{
		DestinationURL: destinationURL,
		DestinationJWT: destinationJWT,
		CollectionName: collectionName,
		SkipExisting:   true,
		BatchSize:      50,
	}

	return c.MigrateCollection(config)
}
