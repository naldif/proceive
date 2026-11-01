package response_test

import (
	"encoding/json"
	"io"
	"testing"

	"proceive/pkg/response"
)

func TestFormatValidationErrors(t *testing.T) {
	// Test EOF
	errMap := response.FormatValidationErrors(io.EOF)
	if _, exists := errMap["body"]; !exists {
		t.Errorf("Expected 'body' key for EOF, got %v", errMap)
	}

	// Test UnmarshalTypeError
	var dest struct {
		Age int `json:"age"`
	}
	err := json.Unmarshal([]byte(`{"age": "bukan_angka"}`), &dest)
	if err != nil {
		errMap2 := response.FormatValidationErrors(err)
		if _, exists := errMap2["age"]; !exists {
			t.Errorf("Expected 'age' key for UnmarshalTypeError, got %v", errMap2)
		}
	}
}
