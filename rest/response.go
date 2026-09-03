package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/tidwall/gjson"

	"github.com/jcfrperu/goaneco-rest-ensured/jsonpath"
	"github.com/jcfrperu/goaneco-rest-ensured/xmlpath"
)

// Response contains the HTTP response data, headers, cookies, timing, and associated request configuration.
type Response struct {
	raw         *http.Response
	req         *http.Request
	config      *Config
	body        []byte
	reqBody     []byte // original request body bytes, stored for failure logging
	statusCode  int
	statusLine  string
	headers     http.Header
	cookies     []*http.Cookie
	contentType string
	elapsed     time.Duration
	err         error
	rootPath    string
}

// AsString returns the response body as a string.
func (r *Response) AsString() string {
	if r == nil {
		return ""
	}
	return string(r.body)
}

// AsBytes returns the response body as a byte slice.
func (r *Response) AsBytes() []byte {
	if r == nil {
		return []byte{}
	}
	return append([]byte{}, r.body...)
}

// As unmarshals the response JSON body into the provided pointer value.
func (r *Response) As(v any) error {
	if r == nil {
		return ErrNilResponse
	}
	if r.err != nil {
		return r.err
	}
	if len(r.body) == 0 {
		return fmt.Errorf("%w: empty body", ErrJSONParsing)
	}
	if err := json.Unmarshal(r.body, v); err != nil {
		return fmt.Errorf("%w: %v", ErrJSONParsing, err)
	}
	return nil
}

// StatusCode returns the HTTP status code (e.g. 200, 404).
func (r *Response) StatusCode() int {
	if r == nil {
		return 0
	}
	return r.statusCode
}

// StatusLine returns the HTTP status line (e.g. "HTTP/1.1 200 OK").
func (r *Response) StatusLine() string {
	if r == nil {
		return ""
	}
	return r.statusLine
}

// Header returns the value of the specified header name, or empty string.
func (r *Response) Header(name string) string {
	if r == nil || r.headers == nil {
		return ""
	}
	return r.headers.Get(name)
}

// Headers returns all response headers.
func (r *Response) Headers() http.Header {
	if r == nil || r.headers == nil {
		return make(http.Header)
	}
	return r.headers.Clone()
}

// Cookie returns the value of the specified cookie name, or empty string.
func (r *Response) Cookie(name string) string {
	if r == nil {
		return ""
	}
	for _, c := range r.cookies {
		if c.Name == name {
			return c.Value
		}
	}
	return ""
}

// Cookies returns all response cookies.
func (r *Response) Cookies() []*http.Cookie {
	if r == nil {
		return []*http.Cookie{}
	}
	return append([]*http.Cookie{}, r.cookies...)
}

// ContentType returns the Content-Type header value.
func (r *Response) ContentType() string {
	if r == nil {
		return ""
	}
	return r.contentType
}

// Time returns the roundtrip elapsed time for the request.
func (r *Response) Time() time.Duration {
	if r == nil {
		return 0
	}
	return r.elapsed
}

// Err returns any execution error that occurred.
func (r *Response) Err() error {
	if r == nil {
		return ErrNilResponse
	}
	return r.err
}

// Config returns the configuration associated with this response, or DefaultConfig if unset.
func (r *Response) Config() *Config {
	if r == nil || r.config == nil {
		return DefaultConfig()
	}
	return r.config
}

// RawResponse returns the underlying *http.Response.
func (r *Response) RawResponse() *http.Response {
	if r == nil {
		return nil
	}
	return r.raw
}

// RawRequest returns the underlying *http.Request.
func (r *Response) RawRequest() *http.Request {
	if r == nil {
		return nil
	}
	return r.req
}

// Path evaluates a gjson path expression against the JSON body.
func (r *Response) Path(path string) gjson.Result {
	if r == nil || len(r.body) == 0 {
		return gjson.Result{}
	}
	return gjson.GetBytes(r.body, path)
}

// JsonPath returns a JsonPath queryable instance for the response body.
func (r *Response) JsonPath() *jsonpath.JsonPath {
	if r == nil {
		return jsonpath.From("")
	}
	return jsonpath.FromBytes(r.body)
}

// XmlPath returns an XmlPath queryable instance for the response body.
func (r *Response) XmlPath() (*xmlpath.XmlPath, error) {
	if r == nil {
		return nil, ErrNilResponse
	}
	return xmlpath.FromBytes(r.body)
}

// Then transitions to the ValidatableResponse assertion fluent interface.
func (r *Response) Then() *ValidatableResponse {
	rootPath := ""
	if r != nil {
		rootPath = r.rootPath
	}
	return &ValidatableResponse{
		response: r,
		rootPath: rootPath,
		failures: make([]string, 0),
	}
}

// Extract returns an ExtractableResponse for extracting values after inspection.
func (r *Response) Extract() *ExtractableResponse {
	return &ExtractableResponse{resp: r}
}

// AsObject is a generic helper to deserialize the response body into type T.
func AsObject[T any](r *Response) (T, error) {
	var target T
	if r == nil {
		return target, ErrNilResponse
	}
	if err := r.As(&target); err != nil {
		return target, err
	}
	return target, nil
}
