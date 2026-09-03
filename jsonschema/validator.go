// Package jsonschema provides JSON Schema validation utilities powered by gojsonschema.
package jsonschema

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

// ErrNilValidator is returned when calling validation on a nil Validator receiver.
var ErrNilValidator = errors.New("nil validator instance")

// ErrInvalidSchema is returned when the provided JSON schema is malformed.
var ErrInvalidSchema = errors.New("invalid json schema")

// ErrValidationFailed is returned when the JSON document does not match the schema.
var ErrValidationFailed = errors.New("json schema validation failed")

// ValidationError represents an individual failure in schema validation.
type ValidationError struct {
	Field       string
	Description string
	Value       any
}

func (ve ValidationError) String() string {
	if ve.Field != "" {
		return fmt.Sprintf("%s: %s", ve.Field, ve.Description)
	}
	return ve.Description
}

// SchemaValidationError wraps multiple schema validation errors.
type SchemaValidationError struct {
	Errors []ValidationError
}

func (sve *SchemaValidationError) Error() string {
	if sve == nil || len(sve.Errors) == 0 {
		return "schema validation failed with unknown error"
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s (%d violations):\n", ErrValidationFailed.Error(), len(sve.Errors)))
	for i, err := range sve.Errors {
		b.WriteString(fmt.Sprintf("  [%d] %s\n", i+1, err.String()))
	}
	return strings.TrimRight(b.String(), "\n")
}

// Unwrap allows errors.Is comparison with ErrValidationFailed.
func (sve *SchemaValidationError) Unwrap() error {
	return ErrValidationFailed
}

// Validator validates JSON strings, bytes, and readers against a JSON Schema.
type Validator struct {
	schemaLoader gojsonschema.JSONLoader
	schema       *gojsonschema.Schema
	err          error
}

// MatchesJsonSchema creates a Validator from a raw JSON schema string.
func MatchesJsonSchema(schemaJSON string) *Validator {
	if schemaJSON == "" {
		return &Validator{err: fmt.Errorf("%w: empty schema string", ErrInvalidSchema)}
	}
	loader := gojsonschema.NewStringLoader(schemaJSON)
	schema, err := gojsonschema.NewSchema(loader)
	if err != nil {
		return &Validator{err: fmt.Errorf("%w: %v", ErrInvalidSchema, err)}
	}
	return &Validator{
		schemaLoader: loader,
		schema:       schema,
	}
}

// MatchesJsonSchemaFile creates a Validator from a JSON schema file path on disk.
func MatchesJsonSchemaFile(filePath string) *Validator {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return &Validator{err: fmt.Errorf("%w: invalid file path %q: %v", ErrInvalidSchema, filePath, err)}
	}
	if _, err := os.Stat(absPath); err != nil {
		return &Validator{err: fmt.Errorf("%w: schema file not found %q: %v", ErrInvalidSchema, absPath, err)}
	}

	// Normalize file URL for gojsonschema
	fileURI := "file://" + filepath.ToSlash(absPath)
	if !strings.HasPrefix(fileURI, "file:///") && strings.HasPrefix(fileURI, "file://") {
		// Ensure three slashes for file URI on windows (e.g. file:///C:/...)
		fileURI = "file:///" + strings.TrimPrefix(fileURI, "file://")
	}

	loader := gojsonschema.NewReferenceLoader(fileURI)
	schema, err := gojsonschema.NewSchema(loader)
	if err != nil {
		return &Validator{err: fmt.Errorf("%w: compiling schema from file %q: %v", ErrInvalidSchema, filePath, err)}
	}

	return &Validator{
		schemaLoader: loader,
		schema:       schema,
	}
}

// MatchesJsonSchemaURI creates a Validator from a schema URL or URI.
func MatchesJsonSchemaURI(schemaURI string) *Validator {
	if schemaURI == "" {
		return &Validator{err: fmt.Errorf("%w: empty schema URI", ErrInvalidSchema)}
	}
	if _, err := url.Parse(schemaURI); err != nil {
		return &Validator{err: fmt.Errorf("%w: invalid schema URI %q: %v", ErrInvalidSchema, schemaURI, err)}
	}
	loader := gojsonschema.NewReferenceLoader(schemaURI)
	schema, err := gojsonschema.NewSchema(loader)
	if err != nil {
		return &Validator{err: fmt.Errorf("%w: compiling schema from URI %q: %v", ErrInvalidSchema, schemaURI, err)}
	}
	return &Validator{
		schemaLoader: loader,
		schema:       schema,
	}
}

// MatchesJsonSchemaReader creates a Validator by reading schema content from an io.Reader.
func MatchesJsonSchemaReader(r io.Reader) (*Validator, error) {
	if r == nil {
		return nil, fmt.Errorf("%w: nil reader", ErrInvalidSchema)
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading schema from reader: %w", err)
	}
	v := MatchesJsonSchema(string(data))
	if v.err != nil {
		return nil, v.err
	}
	return v, nil
}

// Validate validates a JSON string against the schema.
func (v *Validator) Validate(jsonStr string) error {
	if v == nil {
		return ErrNilValidator
	}
	if v.err != nil {
		return v.err
	}
	return v.ValidateLoader(gojsonschema.NewStringLoader(jsonStr))
}

// ValidateBytes validates raw JSON bytes against the schema.
func (v *Validator) ValidateBytes(data []byte) error {
	if v == nil {
		return ErrNilValidator
	}
	if v.err != nil {
		return v.err
	}
	return v.ValidateLoader(gojsonschema.NewBytesLoader(data))
}

// ValidateReader validates JSON content read from an io.Reader against the schema.
func (v *Validator) ValidateReader(r io.Reader) error {
	if v == nil {
		return ErrNilValidator
	}
	if v.err != nil {
		return v.err
	}
	if r == nil {
		return fmt.Errorf("%w: nil reader to validate", ErrValidationFailed)
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("reading json document to validate: %w", err)
	}
	return v.ValidateBytes(data)
}

// ValidateLoader validates a generic gojsonschema.JSONLoader against the compiled schema.
func (v *Validator) ValidateLoader(loader gojsonschema.JSONLoader) error {
	if v == nil {
		return ErrNilValidator
	}
	if v.err != nil {
		return v.err
	}
	if v.schema == nil {
		return fmt.Errorf("%w: uncompiled schema", ErrInvalidSchema)
	}

	result, err := v.schema.Validate(loader)
	if err != nil {
		return fmt.Errorf("schema validation execution: %w", err)
	}

	if result.Valid() {
		return nil
	}

	failures := make([]ValidationError, 0, len(result.Errors()))
	for _, resErr := range result.Errors() {
		failures = append(failures, ValidationError{
			Field:       resErr.Field(),
			Description: resErr.Description(),
			Value:       resErr.Value(),
		})
	}

	return &SchemaValidationError{
		Errors: failures,
	}
}
