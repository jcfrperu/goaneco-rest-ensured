package jsonschema_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/jsonschema"
)

const userSchemaJSON = `{
	"$schema": "http://json-schema.org/draft-07/schema#",
	"title": "User",
	"type": "object",
	"properties": {
		"id": {
			"type": "integer",
			"minimum": 1
		},
		"name": {
			"type": "string",
			"minLength": 2
		},
		"email": {
			"type": "string",
			"format": "email"
		},
		"roles": {
			"type": "array",
			"items": {
				"type": "string"
			}
		}
	},
	"required": ["id", "name", "email"]
}`

func TestMatchesJsonSchema_ValidDocuments(t *testing.T) {
	t.Parallel()
	must := require.New(t)

	validator := jsonschema.MatchesJsonSchema(userSchemaJSON)
	must.NotNil(validator)

	tests := []struct {
		name string
		json string
	}{
		{
			name: "complete valid user",
			json: `{"id": 1, "name": "Alice Smith", "email": "alice@example.com", "roles": ["admin"]}`,
		},
		{
			name: "minimal valid user without optional roles",
			json: `{"id": 10, "name": "Bob Jones", "email": "bob@example.com"}`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)
			err := validator.Validate(tt.json)
			is.NoError(err)

			errBytes := validator.ValidateBytes([]byte(tt.json))
			is.NoError(errBytes)

			errReader := validator.ValidateReader(bytes.NewReader([]byte(tt.json)))
			is.NoError(errReader)
		})
	}
}

func TestMatchesJsonSchema_InvalidDocuments(t *testing.T) {
	t.Parallel()

	validator := jsonschema.MatchesJsonSchema(userSchemaJSON)

	tests := []struct {
		name        string
		json        string
		expectedErr string
	}{
		{
			name:        "missing required email field",
			json:        `{"id": 1, "name": "Alice"}`,
			expectedErr: "email is required",
		},
		{
			name:        "invalid id type (string instead of integer)",
			json:        `{"id": "abc", "name": "Alice", "email": "alice@example.com"}`,
			expectedErr: "Invalid type",
		},
		{
			name:        "name below min length",
			json:        `{"id": 1, "name": "A", "email": "alice@example.com"}`,
			expectedErr: "String length must be greater than or equal to 2",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)

			err := validator.Validate(tt.json)
			is.Error(err)
			is.ErrorIs(err, jsonschema.ErrValidationFailed)
			is.Contains(err.Error(), tt.expectedErr)
		})
	}
}

func TestMatchesJsonSchema_InvalidSchemas(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	// Empty schema string
	vEmpty := jsonschema.MatchesJsonSchema("")
	is.Error(vEmpty.Validate(`{}`))

	// Malformed JSON schema
	vMalformed := jsonschema.MatchesJsonSchema("{ not a valid json schema")
	is.Error(vMalformed.Validate(`{}`))
}

func TestMatchesJsonSchemaFile(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	// Create temporary schema file
	tempDir := t.TempDir()
	schemaFilePath := filepath.Join(tempDir, "user_schema.json")
	err := os.WriteFile(schemaFilePath, []byte(userSchemaJSON), 0600)
	must.NoError(err)

	validator := jsonschema.MatchesJsonSchemaFile(schemaFilePath)
	must.NotNil(validator)

	validJSON := `{"id": 42, "name": "Charlie", "email": "charlie@example.com"}`
	is.NoError(validator.Validate(validJSON))

	invalidJSON := `{"id": 42, "name": "Charlie"}`
	is.Error(validator.Validate(invalidJSON))

	// Nonexistent file returns error on validation
	nonExistentValidator := jsonschema.MatchesJsonSchemaFile(filepath.Join(tempDir, "missing.json"))
	is.Error(nonExistentValidator.Validate(validJSON))
}

func TestMatchesJsonSchemaURI(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(userSchemaJSON))
	}))
	t.Cleanup(ts.Close)

	validator := jsonschema.MatchesJsonSchemaURI(ts.URL)
	must.NotNil(validator)

	validJSON := `{"id": 77, "name": "Evan", "email": "evan@example.com"}`
	is.NoError(validator.Validate(validJSON))

	// Invalid URI
	invalidValidator := jsonschema.MatchesJsonSchemaURI("http://127.0.0.1:1/nonexistent.json")
	is.Error(invalidValidator.Validate(validJSON))
}

func TestMatchesJsonSchemaReader(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	validator, err := jsonschema.MatchesJsonSchemaReader(bytes.NewReader([]byte(userSchemaJSON)))
	must.NoError(err)
	must.NotNil(validator)

	validJSON := `{"id": 100, "name": "Dave", "email": "dave@example.com"}`
	is.NoError(validator.Validate(validJSON))

	// Nil reader
	_, err = jsonschema.MatchesJsonSchemaReader(nil)
	is.Error(err)

	// ValidateReader with nil reader
	is.Error(validator.ValidateReader(nil))
}

func TestValidationError_Formatting(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	ve1 := jsonschema.ValidationError{
		Field:       "user.age",
		Description: "must be at least 18",
		Value:       15,
	}
	is.Equal("user.age: must be at least 18", ve1.String())

	ve2 := jsonschema.ValidationError{
		Field:       "",
		Description: "invalid root format",
	}
	is.Equal("invalid root format", ve2.String())

	sve := &jsonschema.SchemaValidationError{
		Errors: []jsonschema.ValidationError{ve1, ve2},
	}
	is.Contains(sve.Error(), "2 violations")
	is.Contains(sve.Error(), "user.age: must be at least 18")
	is.ErrorIs(sve, jsonschema.ErrValidationFailed)
}

func TestValidator_NilSafety(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	var nilValidator *jsonschema.Validator

	is.ErrorIs(nilValidator.Validate(`{}`), jsonschema.ErrNilValidator)
	is.ErrorIs(nilValidator.ValidateBytes([]byte(`{}`)), jsonschema.ErrNilValidator)
	is.ErrorIs(nilValidator.ValidateReader(bytes.NewReader([]byte(`{}`))), jsonschema.ErrNilValidator)
	is.ErrorIs(nilValidator.ValidateLoader(nil), jsonschema.ErrNilValidator)

	var nilErr *jsonschema.SchemaValidationError
	is.Contains(nilErr.Error(), "unknown error")
}
