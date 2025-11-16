package gopocketbaseclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"
)

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	Token      string
}

type BaseRecord struct {
	ID             string         `json:"id"`
	CollectionID   string         `json:"collectionId"`
	CollectionName string         `json:"collectionName"`
	Created        PocketBaseTime `json:"created"`
	Updated        PocketBaseTime `json:"updated"`
}

type Record struct {
	BaseRecord
	Data map[string]interface{} `json:"data"`
}

type JSONItems struct {
	Items json.RawMessage `json:"items"`
}

// PaginatedResponse represents a paginated response from PocketBase
type PaginatedResponse struct {
	Page       int             `json:"page"`
	PerPage    int             `json:"perPage"`
	TotalItems int             `json:"totalItems"`
	TotalPages int             `json:"totalPages"`
	Items      json.RawMessage `json:"items"`
}

// Authentication types
type User struct {
	BaseRecord
	Username        string `json:"username"`
	Email           string `json:"email"`
	Name            string `json:"name"`
	Avatar          string `json:"avatar"`
	Verified        bool   `json:"verified"`
	EmailVisibility bool   `json:"emailVisibility"`
}

type AuthResponse struct {
	Token  string `json:"token"`
	Record User   `json:"record"`
}

type LoginRequest struct {
	Identity string `json:"identity"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Username        string `json:"username"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	PasswordConfirm string `json:"passwordConfirm"`
	Name            string `json:"name"`
	Avatar          string `json:"avatar,omitempty"`
}

type PasswordResetRequest struct {
	Email string `json:"email"`
}

type PasswordResetConfirm struct {
	Token           string `json:"token"`
	Password        string `json:"password"`
	PasswordConfirm string `json:"passwordConfirm"`
}

// JWTValidationResponse represents the response from JWT validation
type JWTValidationResponse struct {
	Valid  bool   `json:"valid"`
	UserID string `json:"user_id,omitempty"`
	Email  string `json:"email,omitempty"`
	Error  string `json:"error,omitempty"`
}

// PocketBaseTime wraps time.Time to handle PocketBase's date format
type PocketBaseTime struct {
	time.Time
}

// UnmarshalJSON handles PocketBase's "2025-01-20 21:00:58.576Z" format and null values
func (pbt *PocketBaseTime) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), `"`)
	if str == "null" || str == "" || str == "n/a" || str == "N/A" {
		pbt.Time = time.Time{} // Set to zero time
		return nil
	}

	// Use the improved parsePocketBaseTime function
	t, err := parsePocketBaseTime(str)
	if err != nil {
		return fmt.Errorf("cannot parse time %q: %w", str, err)
	}

	pbt.Time = t
	return nil
}

// MarshalJSON formats time for PocketBase
func (pbt PocketBaseTime) MarshalJSON() ([]byte, error) {
	if pbt.Time.IsZero() {
		return []byte("null"), nil
	}
	formatted := pbt.Time.UTC().Format("2006-01-02 15:04:05.000Z")
	return json.Marshal(formatted)
}

// NullableTime represents a time that can be null
type NullableTime struct {
	Time  time.Time
	Valid bool // Valid is true if Time is not null
}

// UnmarshalJSON handles null datetime values gracefully
func (nt *NullableTime) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), `"`)
	if str == "null" || str == "" || str == "n/a" || str == "N/A" {
		nt.Time = time.Time{}
		nt.Valid = false
		return nil
	}

	t, err := parsePocketBaseTime(str)
	if err != nil {
		nt.Time = time.Time{}
		nt.Valid = false
		return fmt.Errorf("cannot parse time %q: %w", str, err)
	}

	nt.Time = t
	nt.Valid = true
	return nil
}

// MarshalJSON formats time for PocketBase or null if not valid
func (nt NullableTime) MarshalJSON() ([]byte, error) {
	if !nt.Valid || nt.Time.IsZero() {
		return []byte("null"), nil
	}
	formatted := nt.Time.UTC().Format("2006-01-02 15:04:05.000Z")
	return json.Marshal(formatted)
}

// Helper function to parse PocketBase time format
func parsePocketBaseTime(timeStr string) (time.Time, error) {
	// Handle null, empty, or "n/a" values
	if timeStr == "" || timeStr == "null" || timeStr == "n/a" || timeStr == "N/A" {
		return time.Time{}, nil
	}

	// Input validation - allow for more flexible length checking
	if len(timeStr) < 10 {
		return time.Time{}, fmt.Errorf("invalid time string length: %q", timeStr)
	}

	// Try different time formats that PocketBase might use
	formats := []string{
		"2006-01-02 15:04:05.000Z", // PocketBase format
		"2006-01-02T15:04:05.000Z", // RFC3339 with milliseconds
		time.RFC3339,               // Standard RFC3339
		"2006-01-02T15:04:05Z",     // RFC3339 without milliseconds
		"2006-01-02 15:04:05",      // Without timezone
		"2006-01-02",               // Date only
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("cannot parse time %q with any known format", timeStr)
}

// Helper function to format time for PocketBase
func formatPocketBaseTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format("2006-01-02 15:04:05.000Z")
}

// UnmarshalPocketBaseJSON unmarshals JSON with automatic time.Time field conversion
func UnmarshalPocketBaseJSON(data []byte, v interface{}) error {
	if len(data) == 0 {
		return fmt.Errorf("cannot unmarshal empty data")
	}
	if v == nil {
		return fmt.Errorf("cannot unmarshal into nil value")
	}

	// First unmarshal into map to handle time fields
	var rawData interface{}
	if err := json.Unmarshal(data, &rawData); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Convert time fields recursively
	convertedData := convertTimeFields(rawData)

	// Marshal back to JSON and unmarshal into target struct
	convertedJSON, err := json.Marshal(convertedData)
	if err != nil {
		return fmt.Errorf("failed to marshal converted data: %w", err)
	}

	if err := json.Unmarshal(convertedJSON, v); err != nil {
		return fmt.Errorf("failed to unmarshal into target struct: %w", err)
	}

	return nil
}

// MarshalPocketBaseJSON marshals with automatic time.Time field conversion
func MarshalPocketBaseJSON(v interface{}) ([]byte, error) {
	if v == nil {
		return []byte("null"), nil
	}

	// First marshal to get JSON representation
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input: %w", err)
	}

	// Unmarshal into interface{} to process
	var rawData interface{}
	if err := json.Unmarshal(data, &rawData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal for processing: %w", err)
	}

	// Convert time fields to PocketBase format
	convertedData := convertTimeFieldsForPocketBase(rawData)

	// Marshal final result
	result, err := json.Marshal(convertedData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal final result: %w", err)
	}

	return result, nil
}

// convertTimeFields recursively converts PocketBase time strings to time.Time
func convertTimeFields(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, value := range v {
			result[key] = convertTimeFields(value)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, value := range v {
			result[i] = convertTimeFields(value)
		}
		return result
	case string:
		// Handle null/empty datetime values first
		if v == "" || v == "null" || v == "n/a" || v == "N/A" {
			// Check if this might be a time field by context or format
			if isTimeString(v) || v == "" {
				// Return zero time for empty/null time fields
				return time.Time{}
			}
		}
		// Try to parse as time if it looks like PocketBase time format
		if isTimeString(v) {
			if t, err := parsePocketBaseTime(v); err == nil {
				return t
			}
		}
		return v
	default:
		return v
	}
}

// convertTimeFieldsForPocketBase recursively converts time.Time to PocketBase format
func convertTimeFieldsForPocketBase(data interface{}) interface{} {
	val := reflect.ValueOf(data)
	switch val.Kind() {
	case reflect.Map:
		result := make(map[string]interface{})
		for _, key := range val.MapKeys() {
			keyStr := fmt.Sprintf("%v", key.Interface())
			value := val.MapIndex(key).Interface()
			result[keyStr] = convertTimeFieldsForPocketBase(value)
		}
		return result
	case reflect.Slice:
		result := make([]interface{}, val.Len())
		for i := 0; i < val.Len(); i++ {
			result[i] = convertTimeFieldsForPocketBase(val.Index(i).Interface())
		}
		return result
	default:
		if t, ok := data.(time.Time); ok {
			return formatPocketBaseTime(t)
		}
		return data
	}
}

// isTimeString checks if a string looks like a PocketBase time format
func isTimeString(s string) bool {
	// Handle empty strings and null values
	if s == "" || s == "null" || s == "n/a" || s == "N/A" {
		return false
	}

	// Check minimum length for a date
	if len(s) < 10 {
		return false
	}

	// Check for basic date format patterns
	if len(s) >= 10 && s[4] == '-' && s[7] == '-' {
		// Could be a date: "2025-01-20" or datetime
		if len(s) == 10 {
			return true // Date only format
		}

		// Check for datetime patterns
		if len(s) >= 19 && (s[10] == ' ' || s[10] == 'T') &&
			len(s) < 35 && // Reasonable upper limit
			s[13] == ':' && s[16] == ':' {
			return true
		}
	}

	return false
}

// Migration types

// MigrationConfig holds configuration for migrating data between PocketBase instances
type MigrationConfig struct {
	DestinationURL string `json:"destination_url"` // URL of the destination PocketBase
	DestinationJWT string `json:"destination_jwt"` // JWT token for destination PocketBase
	CollectionName string `json:"collection_name"` // Name of collection to migrate
	SkipExisting   bool   `json:"skip_existing"`   // Whether to skip records that already exist
	BatchSize      int    `json:"batch_size"`      // Number of records to process in each batch
}

// MigrationResult contains results from a migration operation
type MigrationResult struct {
	SourceCollection      string           `json:"source_collection"`
	DestinationCollection string           `json:"destination_collection"`
	TotalRecords          int              `json:"total_records"`
	SuccessfulRecords     int              `json:"successful_records"`
	FailedRecords         int              `json:"failed_records"`
	SkippedRecords        int              `json:"skipped_records"`
	ProcessingTime        string           `json:"processing_time"`
	Errors                []MigrationError `json:"errors"`
	Summary               string           `json:"summary"`
}

// MigrationError represents an error that occurred during migration
type MigrationError struct {
	RecordID    string `json:"record_id"`
	RecordIndex int    `json:"record_index"`
	Operation   string `json:"operation"`
	Error       string `json:"error"`
}

// MigrationRecord represents a record during migration process
type MigrationRecord struct {
	SourceID string                 `json:"source_id"`
	Data     map[string]interface{} `json:"data"`
	Created  string                 `json:"created"`
	Updated  string                 `json:"updated"`
}
