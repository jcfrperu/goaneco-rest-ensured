package rest

import "strings"

// ContentType represents HTTP Content-Type header values.
type ContentType string

const (
	ContentTypeJSON       ContentType = "application/json"
	ContentTypeXML        ContentType = "application/xml"
	ContentTypeHTML       ContentType = "text/html"
	ContentTypeText       ContentType = "text/plain"
	ContentTypeURLEncoded ContentType = "application/x-www-form-urlencoded"
	ContentTypeMultipart  ContentType = "multipart/form-data"
	ContentTypeBinary     ContentType = "application/octet-stream"
	ContentTypeAny        ContentType = "*/*"
)

// knownContentTypes lists all defined ContentType constants for lookup.
var knownContentTypes = []ContentType{
	ContentTypeJSON, ContentTypeXML, ContentTypeHTML, ContentTypeText,
	ContentTypeURLEncoded, ContentTypeMultipart, ContentTypeBinary, ContentTypeAny,
}

// WithCharset returns the content type string with "; charset=<charset>" appended.
func (ct ContentType) WithCharset(charset string) string {
	return string(ct) + "; charset=" + charset
}

// Matches reports whether other equals this content type, case-insensitively.
// Returns false when other is empty.
func (ct ContentType) Matches(other string) bool {
	if other == "" {
		return false
	}
	return strings.EqualFold(string(ct), other)
}

// FromContentType returns the ContentType constant matching s (case-insensitive).
// Returns ("", false) if no known ContentType matches.
func FromContentType(s string) (ContentType, bool) {
	for _, ct := range knownContentTypes {
		if strings.EqualFold(string(ct), s) {
			return ct, true
		}
	}
	return "", false
}

// Method represents an HTTP method.
type Method string

const (
	MethodGet     Method = "GET"
	MethodPost    Method = "POST"
	MethodPut     Method = "PUT"
	MethodDelete  Method = "DELETE"
	MethodHead    Method = "HEAD"
	MethodPatch   Method = "PATCH"
	MethodOptions Method = "OPTIONS"
)

// LogDetail defines what information is logged during request/response inspection.
type LogDetail int

const (
	LogDetailNone LogDetail = iota
	LogDetailStatus
	LogDetailHeaders
	LogDetailCookies
	LogDetailBody
	LogDetailAll
	LogDetailIfError
	LogDetailIfValidationFails
)
