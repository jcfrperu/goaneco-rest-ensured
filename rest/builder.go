package rest

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// RequestSpec holds a reusable specification of request settings.
type RequestSpec struct {
	BaseURI            string
	BasePath           string
	Port               int
	PortSet            bool // explicit flag so Port=0 can clear a previously set port
	Headers            http.Header
	Cookies            []*http.Cookie
	QueryParams        url.Values
	FormParams         url.Values
	PathParams         map[string]string
	Body               []byte
	ContentType        string
	Accept             string
	Auth               AuthScheme
	Filters            []Filter
	Config             *Config
	InsecureSkipVerify bool
}

// RequestSpecBuilder builds reusable RequestSpec instances.
type RequestSpecBuilder struct {
	spec *RequestSpec
}

// NewRequestSpecBuilder creates a new RequestSpecBuilder.
func NewRequestSpecBuilder() *RequestSpecBuilder {
	return &RequestSpecBuilder{
		spec: &RequestSpec{
			Headers:     make(http.Header),
			Cookies:     make([]*http.Cookie, 0),
			QueryParams: make(url.Values),
			FormParams:  make(url.Values),
			PathParams:  make(map[string]string),
			Filters:     make([]Filter, 0),
		},
	}
}

// SetBaseURI sets the base URI.
func (b *RequestSpecBuilder) SetBaseURI(uri string) *RequestSpecBuilder {
	b.spec.BaseURI = uri
	return b
}

// SetBasePath sets the base path.
func (b *RequestSpecBuilder) SetBasePath(path string) *RequestSpecBuilder {
	b.spec.BasePath = path
	return b
}

// SetPort sets the port. A value of 0 clears any previously inherited port.
func (b *RequestSpecBuilder) SetPort(port int) *RequestSpecBuilder {
	b.spec.Port = port
	b.spec.PortSet = true
	return b
}

// AddHeader adds an HTTP header.
func (b *RequestSpecBuilder) AddHeader(name, value string) *RequestSpecBuilder {
	b.spec.Headers.Add(name, value)
	return b
}

// AddHeaders adds multiple HTTP headers.
func (b *RequestSpecBuilder) AddHeaders(headers map[string]string) *RequestSpecBuilder {
	for k, v := range headers {
		b.spec.Headers.Add(k, v)
	}
	return b
}

// AddCookie adds a cookie.
func (b *RequestSpecBuilder) AddCookie(name, value string) *RequestSpecBuilder {
	b.spec.Cookies = append(b.spec.Cookies, &http.Cookie{Name: name, Value: value})
	return b
}

// AddQueryParam adds a query parameter.
func (b *RequestSpecBuilder) AddQueryParam(name string, values ...any) *RequestSpecBuilder {
	for _, val := range values {
		b.spec.QueryParams.Add(name, fmt.Sprintf("%v", val))
	}
	return b
}

// AddFormParam adds a form parameter.
func (b *RequestSpecBuilder) AddFormParam(name string, values ...any) *RequestSpecBuilder {
	for _, val := range values {
		b.spec.FormParams.Add(name, fmt.Sprintf("%v", val))
	}
	return b
}

// AddPathParam adds a path parameter.
func (b *RequestSpecBuilder) AddPathParam(name string, value any) *RequestSpecBuilder {
	b.spec.PathParams[name] = fmt.Sprintf("%v", value)
	return b
}

// SetContentType sets the Content-Type.
func (b *RequestSpecBuilder) SetContentType(ct ContentType) *RequestSpecBuilder {
	b.spec.ContentType = string(ct)
	return b
}

// SetAccept sets the Accept header.
func (b *RequestSpecBuilder) SetAccept(ct ContentType) *RequestSpecBuilder {
	b.spec.Accept = string(ct)
	return b
}

// SetAuth sets the authentication scheme.
func (b *RequestSpecBuilder) SetAuth(auth AuthScheme) *RequestSpecBuilder {
	b.spec.Auth = auth
	return b
}

// AddFilter adds filters.
func (b *RequestSpecBuilder) AddFilter(filters ...Filter) *RequestSpecBuilder {
	b.spec.Filters = append(b.spec.Filters, filters...)
	return b
}

// SetConfig sets the configuration.
func (b *RequestSpecBuilder) SetConfig(cfg *Config) *RequestSpecBuilder {
	b.spec.Config = cfg
	return b
}

// SetBody sets the raw string body.
func (b *RequestSpecBuilder) SetBody(body string) *RequestSpecBuilder {
	b.spec.Body = []byte(body)
	return b
}

// Build returns a copy of the configured RequestSpec.
// Mutations on the returned value do not affect the builder.
func (b *RequestSpecBuilder) Build() *RequestSpec {
	cpy := *b.spec
	cpy.Headers = b.spec.Headers.Clone()
	cpy.Cookies = append([]*http.Cookie{}, b.spec.Cookies...)
	cpy.QueryParams = cloneValues(b.spec.QueryParams)
	cpy.FormParams = cloneValues(b.spec.FormParams)
	cpy.PathParams = cloneStringMap(b.spec.PathParams)
	cpy.Filters = append([]Filter{}, b.spec.Filters...)
	if len(b.spec.Body) > 0 {
		cpy.Body = append([]byte{}, b.spec.Body...)
	}
	return &cpy
}

// ResponseSpec holds reusable expectations for validating responses.
type ResponseSpec struct {
	ExpectedStatusCode  int
	ExpectedHeaders     map[string]string
	ExpectedCookies     map[string]string
	ExpectedContentType ContentType
	ExpectedBodies      map[string]any
	ExpectedSchema      string
	ExpectedSchemaFile  string
	MaxResponseTime     time.Duration
	RootPath            string
}

// GetRootPath returns the current root path prefix for body assertions.
func (s *ResponseSpec) GetRootPath() string {
	return s.RootPath
}

// ResponseSpecBuilder builds reusable ResponseSpec instances.
type ResponseSpecBuilder struct {
	spec *ResponseSpec
}

// NewResponseSpecBuilder creates a new ResponseSpecBuilder.
func NewResponseSpecBuilder() *ResponseSpecBuilder {
	return &ResponseSpecBuilder{
		spec: &ResponseSpec{
			ExpectedHeaders: make(map[string]string),
			ExpectedCookies: make(map[string]string),
			ExpectedBodies:  make(map[string]any),
		},
	}
}

// ExpectStatusCode sets the expected HTTP status code.
func (b *ResponseSpecBuilder) ExpectStatusCode(code int) *ResponseSpecBuilder {
	b.spec.ExpectedStatusCode = code
	return b
}

// ExpectHeader sets an expected response header.
func (b *ResponseSpecBuilder) ExpectHeader(name, value string) *ResponseSpecBuilder {
	b.spec.ExpectedHeaders[name] = value
	return b
}

// ExpectCookie sets an expected cookie.
func (b *ResponseSpecBuilder) ExpectCookie(name, value string) *ResponseSpecBuilder {
	b.spec.ExpectedCookies[name] = value
	return b
}

// ExpectContentType sets an expected Content-Type.
func (b *ResponseSpecBuilder) ExpectContentType(ct ContentType) *ResponseSpecBuilder {
	b.spec.ExpectedContentType = ct
	return b
}

// ExpectBody adds an expected body assertion at path.
func (b *ResponseSpecBuilder) ExpectBody(path string, expected any) *ResponseSpecBuilder {
	b.spec.ExpectedBodies[path] = expected
	return b
}

// ExpectBodyMatchesSchema adds a JSON Schema validation expectation from a raw JSON schema string.
func (b *ResponseSpecBuilder) ExpectBodyMatchesSchema(schemaJSON string) *ResponseSpecBuilder {
	b.spec.ExpectedSchema = schemaJSON
	return b
}

// ExpectBodyMatchesSchemaFile adds a JSON Schema validation expectation loaded from a file path.
func (b *ResponseSpecBuilder) ExpectBodyMatchesSchemaFile(filePath string) *ResponseSpecBuilder {
	b.spec.ExpectedSchemaFile = filePath
	return b
}

// ExpectResponseTimeLessThan sets the maximum response duration.
func (b *ResponseSpecBuilder) ExpectResponseTimeLessThan(d time.Duration) *ResponseSpecBuilder {
	b.spec.MaxResponseTime = d
	return b
}

// RootPath sets the root path prefix for body assertions.
// Optional args are applied via fmt.Sprintf(path, args...).
func (b *ResponseSpecBuilder) RootPath(path string, args ...any) *ResponseSpecBuilder {
	if len(args) > 0 {
		path = fmt.Sprintf(path, args...)
	}
	b.spec.RootPath = path
	return b
}

// AppendRootPath appends a sub-path to the current root path.
// Optional args are applied via fmt.Sprintf(subPath, args...).
func (b *ResponseSpecBuilder) AppendRootPath(subPath string, args ...any) *ResponseSpecBuilder {
	if len(args) > 0 {
		subPath = fmt.Sprintf(subPath, args...)
	}
	if b.spec.RootPath == "" {
		b.spec.RootPath = subPath
	} else {
		b.spec.RootPath = b.spec.RootPath + "." + strings.TrimPrefix(subPath, ".")
	}
	return b
}

// NoRootPath resets the root path to empty string.
func (b *ResponseSpecBuilder) NoRootPath() *ResponseSpecBuilder {
	b.spec.RootPath = ""
	return b
}

// DetachRootPath removes the given suffix from the current root path.
func (b *ResponseSpecBuilder) DetachRootPath(suffix string) *ResponseSpecBuilder {
	suffix = strings.TrimPrefix(suffix, ".")
	if b.spec.RootPath == suffix {
		b.spec.RootPath = ""
		return b
	}
	b.spec.RootPath = strings.TrimSuffix(b.spec.RootPath, "."+suffix)
	return b
}

// AddResponseSpec merges all expectations from other into this builder, overwriting existing values where set.
func (b *ResponseSpecBuilder) AddResponseSpec(other *ResponseSpec) *ResponseSpecBuilder {
	if other == nil {
		return b
	}
	if other.ExpectedStatusCode > 0 {
		b.spec.ExpectedStatusCode = other.ExpectedStatusCode
	}
	for k, v := range other.ExpectedHeaders {
		b.spec.ExpectedHeaders[k] = v
	}
	for k, v := range other.ExpectedCookies {
		b.spec.ExpectedCookies[k] = v
	}
	if other.ExpectedContentType != "" {
		b.spec.ExpectedContentType = other.ExpectedContentType
	}
	for k, v := range other.ExpectedBodies {
		b.spec.ExpectedBodies[k] = v
	}
	if other.ExpectedSchema != "" {
		b.spec.ExpectedSchema = other.ExpectedSchema
	}
	if other.ExpectedSchemaFile != "" {
		b.spec.ExpectedSchemaFile = other.ExpectedSchemaFile
	}
	if other.MaxResponseTime > 0 {
		b.spec.MaxResponseTime = other.MaxResponseTime
	}
	if other.RootPath != "" {
		b.spec.RootPath = other.RootPath
	}
	return b
}

// Build returns a copy of the configured ResponseSpec.
// Mutations on the returned value do not affect the builder.
func (b *ResponseSpecBuilder) Build() *ResponseSpec {
	cpy := *b.spec
	cpy.ExpectedHeaders = cloneStringMap(b.spec.ExpectedHeaders)
	cpy.ExpectedCookies = cloneStringMap(b.spec.ExpectedCookies)
	cpy.ExpectedBodies = make(map[string]any, len(b.spec.ExpectedBodies))
	for k, v := range b.spec.ExpectedBodies {
		cpy.ExpectedBodies[k] = v
	}
	return &cpy
}

// cloneValues returns a deep copy of src so that callers cannot mutate the original
// slice backing each key's value list.
func cloneValues(src url.Values) url.Values {
	dst := make(url.Values, len(src))
	for k, vs := range src {
		dst[k] = append([]string{}, vs...)
	}
	return dst
}

// cloneStringMap returns a shallow copy of src. Values are copied by assignment, so
// this is safe for string, numeric, and other non-pointer value types.
func cloneStringMap[V any](src map[string]V) map[string]V {
	dst := make(map[string]V, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// Validate applies all ResponseSpec assertions to a ValidatableResponse.
func (s *ResponseSpec) Validate(v *ValidatableResponse) *ValidatableResponse {
	if s.ExpectedStatusCode > 0 {
		v.StatusCode(s.ExpectedStatusCode)
	}
	for name, val := range s.ExpectedHeaders {
		v.Header(name, val)
	}
	for name, val := range s.ExpectedCookies {
		v.Cookie(name, val)
	}
	if s.ExpectedContentType != "" {
		v.ContentType(s.ExpectedContentType)
	}
	if s.RootPath != "" {
		prevRootPath := v.rootPath
		v.RootPath(s.RootPath)
		for path, val := range s.ExpectedBodies {
			v.Body(path, val)
		}
		// Restore the caller's root path instead of clearing it unconditionally.
		v.rootPath = prevRootPath
	} else {
		for path, val := range s.ExpectedBodies {
			v.Body(path, val)
		}
	}
	if s.ExpectedSchema != "" {
		v.BodyMatchesSchema(s.ExpectedSchema)
	}
	if s.ExpectedSchemaFile != "" {
		v.BodyMatchesSchemaFile(s.ExpectedSchemaFile)
	}
	if s.MaxResponseTime > 0 {
		v.TimeLessThan(s.MaxResponseTime)
	}
	return v
}
