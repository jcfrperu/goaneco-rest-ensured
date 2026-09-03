package rest

import "errors"

var (
	// ErrRequestFailed indicates that the HTTP request failed to execute.
	ErrRequestFailed = errors.New("restassured: http request execution failed")

	// ErrInvalidURL indicates that a URL or path could not be parsed.
	ErrInvalidURL = errors.New("restassured: invalid url or path")

	// ErrAssertionFailed indicates that one or more response assertions failed.
	ErrAssertionFailed = errors.New("restassured: assertion failed")

	// ErrJSONParsing indicates that a JSON response could not be parsed.
	ErrJSONParsing = errors.New("restassured: json parsing error")

	// ErrNilResponse indicates that an operation was attempted on a nil response.
	ErrNilResponse = errors.New("restassured: response is nil")
)
