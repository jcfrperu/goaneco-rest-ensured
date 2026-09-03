package rest

import (
	"net/http"
	"time"

	"github.com/tidwall/gjson"

	"github.com/jcfrperu/goaneco-rest-ensured/jsonpath"
	"github.com/jcfrperu/goaneco-rest-ensured/xmlpath"
)

// ExtractableResponse provides convenient methods to extract data from a validated response.
type ExtractableResponse struct {
	resp *Response
}

// Response returns the underlying Response object.
func (e *ExtractableResponse) Response() *Response {
	if e == nil {
		return nil
	}
	return e.resp
}

// AsString returns the response body as string.
func (e *ExtractableResponse) AsString() string {
	if e == nil || e.resp == nil {
		return ""
	}
	return e.resp.AsString()
}

// AsBytes returns the response body as bytes.
func (e *ExtractableResponse) AsBytes() []byte {
	if e == nil || e.resp == nil {
		return []byte{}
	}
	return e.resp.AsBytes()
}

// As unmarshals the response body into target pointer.
func (e *ExtractableResponse) As(v any) error {
	if e == nil || e.resp == nil {
		return ErrNilResponse
	}
	return e.resp.As(v)
}

// StatusCode returns the HTTP status code.
func (e *ExtractableResponse) StatusCode() int {
	if e == nil || e.resp == nil {
		return 0
	}
	return e.resp.StatusCode()
}

// StatusLine returns the HTTP status line.
func (e *ExtractableResponse) StatusLine() string {
	if e == nil || e.resp == nil {
		return ""
	}
	return e.resp.StatusLine()
}

// Header returns a header value by name.
func (e *ExtractableResponse) Header(name string) string {
	if e == nil || e.resp == nil {
		return ""
	}
	return e.resp.Header(name)
}

// Headers returns all response headers.
func (e *ExtractableResponse) Headers() http.Header {
	if e == nil || e.resp == nil {
		return make(http.Header)
	}
	return e.resp.Headers()
}

// Cookie returns a cookie value by name.
func (e *ExtractableResponse) Cookie(name string) string {
	if e == nil || e.resp == nil {
		return ""
	}
	return e.resp.Cookie(name)
}

// Cookies returns all response cookies.
func (e *ExtractableResponse) Cookies() []*http.Cookie {
	if e == nil || e.resp == nil {
		return []*http.Cookie{}
	}
	return e.resp.Cookies()
}

// ContentType returns the Content-Type header.
func (e *ExtractableResponse) ContentType() string {
	if e == nil || e.resp == nil {
		return ""
	}
	return e.resp.ContentType()
}

// Time returns the response duration.
func (e *ExtractableResponse) Time() time.Duration {
	if e == nil || e.resp == nil {
		return 0
	}
	return e.resp.Time()
}

// Path evaluates a gjson path expression.
func (e *ExtractableResponse) Path(path string) gjson.Result {
	if e == nil || e.resp == nil {
		return gjson.Result{}
	}
	return e.resp.Path(path)
}

// PathString extracts a string value at the specified gjson path.
func (e *ExtractableResponse) PathString(path string) string {
	return e.Path(path).String()
}

// PathInt extracts an int64 value at the specified gjson path.
func (e *ExtractableResponse) PathInt(path string) int64 {
	return e.Path(path).Int()
}

// PathFloat extracts a float64 value at the specified gjson path.
func (e *ExtractableResponse) PathFloat(path string) float64 {
	return e.Path(path).Float()
}

// PathBool extracts a boolean value at the specified gjson path.
func (e *ExtractableResponse) PathBool(path string) bool {
	return e.Path(path).Bool()
}

// PathArray extracts an array of gjson.Result at the specified path.
func (e *ExtractableResponse) PathArray(path string) []gjson.Result {
	return e.Path(path).Array()
}

// JsonPath returns a JsonPath queryable instance for the response body.
func (e *ExtractableResponse) JsonPath() *jsonpath.JsonPath {
	if e == nil || e.resp == nil {
		return jsonpath.From("")
	}
	return e.resp.JsonPath()
}

// XmlPath returns an XmlPath queryable instance for the response body.
func (e *ExtractableResponse) XmlPath() (*xmlpath.XmlPath, error) {
	if e == nil || e.resp == nil {
		return nil, ErrNilResponse
	}
	return e.resp.XmlPath()
}

// Err returns any execution error that occurred during the request.
func (e *ExtractableResponse) Err() error {
	if e == nil || e.resp == nil {
		return ErrNilResponse
	}
	return e.resp.Err()
}

// Config returns the configuration associated with the response.
func (e *ExtractableResponse) Config() *Config {
	if e == nil || e.resp == nil {
		return DefaultConfig()
	}
	return e.resp.Config()
}

// RawRequest returns the underlying *http.Request.
func (e *ExtractableResponse) RawRequest() *http.Request {
	if e == nil || e.resp == nil {
		return nil
	}
	return e.resp.RawRequest()
}

// RawResponse returns the underlying *http.Response.
func (e *ExtractableResponse) RawResponse() *http.Response {
	if e == nil || e.resp == nil {
		return nil
	}
	return e.resp.RawResponse()
}
