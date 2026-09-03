// Package rest provides a fluent DSL for building, executing, and validating HTTP requests,
// inspired by the REST-Assured library for Java.
//
// The typical usage pattern follows the Given-When-Then style familiar from
// behavior-driven development:
//
//	rest.Given().
//	    Header("Accept", "application/json").
//	    Auth().Basic("user", "pass").
//	    When().
//	    Get("/api/users/1").
//	    Then().
//	    StatusCode(200).
//	    Body("name", "Alice").
//	    AssertAll()
//
// Global defaults (base URI, port, headers, filters, auth) can be configured once via
// the package-level functions (BaseURI, GlobalHeader, GlobalFilter, etc.) and are
// applied to every request built with Given(), When(), or With(). Call Reset() between
// tests to restore all defaults.
//
// The filter chain (Filter interface) lets you intercept requests and responses for
// cross-cutting concerns: logging, cookie jar management, CSRF token injection, session
// tracking, and timing. Filters run in registration order (or by Order() if they
// implement OrderedFilter) before the actual HTTP dispatch.
package rest

import (
	"fmt"
	"net/http"
	"sync"
)

var (
	globalMu                 sync.RWMutex
	globalBaseURI            = "http://localhost"
	globalPort               = 0
	globalBasePath           = ""
	globalURLEncodingEnabled = true
	globalInsecureSkipVerify = false
	globalRootPath           = ""
	globalHeaders            = make(map[string][]string)
	globalQueryParams        = make(map[string][]string)
	globalCookies            = make([]*http.Cookie, 0)
	globalFilters            = make([]Filter, 0)
	globalAuth               AuthScheme
	globalConfig             = DefaultConfig()
)

// Given starts building a new HTTP request specification.
func Given() *Request {
	return NewRequest()
}

// When is syntactic sugar for Given().
func When() *Request {
	return Given()
}

// With is syntactic sugar for Given().
func With() *Request {
	return Given()
}

// Get executes a GET request using global defaults.
func Get(path string, pathParams ...any) *Response {
	return Given().Get(path, pathParams...)
}

// Post executes a POST request using global defaults.
func Post(path string, pathParams ...any) *Response {
	return Given().Post(path, pathParams...)
}

// Put executes a PUT request using global defaults.
func Put(path string, pathParams ...any) *Response {
	return Given().Put(path, pathParams...)
}

// Delete executes a DELETE request using global defaults.
func Delete(path string, pathParams ...any) *Response {
	return Given().Delete(path, pathParams...)
}

// Head executes a HEAD request using global defaults.
func Head(path string, pathParams ...any) *Response {
	return Given().Head(path, pathParams...)
}

// Patch executes a PATCH request using global defaults.
func Patch(path string, pathParams ...any) *Response {
	return Given().Patch(path, pathParams...)
}

// Options executes an OPTIONS request using global defaults.
func Options(path string, pathParams ...any) *Response {
	return Given().Options(path, pathParams...)
}

// Reset restores all global configuration to default initial values.
func Reset() {
	globalMu.Lock()
	defer globalMu.Unlock()

	globalBaseURI = "http://localhost"
	globalPort = 0
	globalBasePath = ""
	globalURLEncodingEnabled = true
	globalInsecureSkipVerify = false
	globalRootPath = ""
	globalHeaders = make(map[string][]string)
	globalQueryParams = make(map[string][]string)
	globalCookies = make([]*http.Cookie, 0)
	globalFilters = make([]Filter, 0)
	globalAuth = nil
	globalConfig = DefaultConfig()
}

// BaseURI sets the default base URI (e.g. "http://localhost" or "https://api.example.com").
func BaseURI(uri string) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalBaseURI = uri
}

// GetBaseURI returns the current default base URI.
func GetBaseURI() string {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return globalBaseURI
}

// BasePath sets the default base path prefix (e.g. "/v1" or "/api").
func BasePath(path string) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalBasePath = path
}

// GetBasePath returns the current default base path.
func GetBasePath() string {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return globalBasePath
}

// Port sets the default port.
func Port(port int) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalPort = port
}

// GetPort returns the current default port.
func GetPort() int {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return globalPort
}

// RootPath sets the global root path for JSON body evaluations.
func RootPath(path string) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalRootPath = path
}

// GetRootPath returns the current global root path.
func GetRootPath() string {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return globalRootPath
}

// URLEncodingEnabled enables or disables automatic URL query encoding globally.
func URLEncodingEnabled(enabled bool) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalURLEncodingEnabled = enabled
}

// IsURLEncodingEnabled returns whether URL encoding is enabled globally.
func IsURLEncodingEnabled() bool {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return globalURLEncodingEnabled
}

// GlobalHeader sets a default header that will be included in all requests.
func GlobalHeader(name, value string) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalHeaders[name] = append(globalHeaders[name], value)
}

// GlobalHeaders sets multiple default headers.
func GlobalHeaders(headers map[string]string) {
	globalMu.Lock()
	defer globalMu.Unlock()
	for k, v := range headers {
		globalHeaders[k] = append(globalHeaders[k], v)
	}
}

// SetGlobalHeader replaces (rather than appends to) the global default header with the given name.
func SetGlobalHeader(name, value string) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalHeaders[name] = []string{value}
}

// SetGlobalHeaders replaces (rather than appends to) each of the given global default headers.
func SetGlobalHeaders(headers map[string]string) {
	globalMu.Lock()
	defer globalMu.Unlock()
	for k, v := range headers {
		globalHeaders[k] = []string{v}
	}
}

// GlobalQueryParam adds a default query parameter included in all requests.
func GlobalQueryParam(name string, values ...any) {
	globalMu.Lock()
	defer globalMu.Unlock()
	for _, v := range values {
		globalQueryParams[name] = append(globalQueryParams[name], fmt.Sprintf("%v", v))
	}
}

// GlobalQueryParams adds multiple default query parameters included in all requests.
func GlobalQueryParams(params map[string]any) {
	globalMu.Lock()
	defer globalMu.Unlock()
	for k, v := range params {
		globalQueryParams[k] = append(globalQueryParams[k], fmt.Sprintf("%v", v))
	}
}

// GlobalCookie sets a default cookie that will be included in all requests.
func GlobalCookie(name, value string) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalCookies = append(globalCookies, &http.Cookie{
		Name:  name,
		Value: value,
	})
}

// GlobalFilter adds one or more default filters applied to all requests.
func GlobalFilter(filters ...Filter) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalFilters = append(globalFilters, filters...)
}

// GlobalAuth sets default authentication scheme for all requests.
func GlobalAuth(auth AuthScheme) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalAuth = auth
}

// GlobalConfig sets default configuration for all requests.
func GlobalConfig(cfg *Config) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalConfig = cfg
}
