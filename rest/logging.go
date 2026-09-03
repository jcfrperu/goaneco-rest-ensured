package rest

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
)

// RequestLogSpec configures request logging.
type RequestLogSpec struct {
	req    *Request
	detail LogDetail
	writer io.Writer
}

// All logs all request details (method, URI, headers, cookies, params, body).
func (l *RequestLogSpec) All() *Request {
	l.detail = LogDetailAll
	l.apply()
	return l.req
}

// Headers logs request headers only.
func (l *RequestLogSpec) Headers() *Request {
	l.detail = LogDetailHeaders
	l.apply()
	return l.req
}

// Cookies logs request cookies only.
func (l *RequestLogSpec) Cookies() *Request {
	l.detail = LogDetailCookies
	l.apply()
	return l.req
}

// Body logs request body only.
func (l *RequestLogSpec) Body() *Request {
	l.detail = LogDetailBody
	l.apply()
	return l.req
}

// IfValidationFails enables full request/response logging whenever any assertion fails.
// It sets EnableLoggingIfValidationFails in the request Config so that recordFailure
// triggers the log, rather than registering a premature filter that runs before validation.
func (l *RequestLogSpec) IfValidationFails() *Request {
	cfg := l.req.config
	if cfg == nil {
		cfg = DefaultConfig()
	}
	lc := cfg.LogConfig()
	lc.EnableLoggingIfValidationFails = true
	if l.writer != nil {
		lc.Output = l.writer
	}
	l.req.config = cfg.WithLog(lc)
	return l.req
}

// Writer redirects log output to a custom io.Writer.
func (l *RequestLogSpec) Writer(w io.Writer) *RequestLogSpec {
	l.writer = w
	return l
}

func (l *RequestLogSpec) apply() {
	targetWriter := l.writer
	if targetWriter == nil {
		targetWriter = os.Stdout
	}
	l.req.filters = append(l.req.filters, &RequestLoggingFilter{
		Detail: l.detail,
		Output: targetWriter,
	})
}

// ResponseLogSpec configures response logging.
type ResponseLogSpec struct {
	resp   *Response
	valid  *ValidatableResponse
	detail LogDetail
	writer io.Writer
}

// Writer redirects response log output to a custom io.Writer.
func (l *ResponseLogSpec) Writer(w io.Writer) *ResponseLogSpec {
	l.writer = w
	return l
}

// All logs all response details (status, headers, cookies, body).
func (l *ResponseLogSpec) All() *ValidatableResponse {
	l.detail = LogDetailAll
	l.printNow()
	return l.valid
}

// Status logs response status code and line.
func (l *ResponseLogSpec) Status() *ValidatableResponse {
	l.detail = LogDetailStatus
	l.printNow()
	return l.valid
}

// Headers logs response headers only.
func (l *ResponseLogSpec) Headers() *ValidatableResponse {
	l.detail = LogDetailHeaders
	l.printNow()
	return l.valid
}

// Cookies logs response cookies only.
func (l *ResponseLogSpec) Cookies() *ValidatableResponse {
	l.detail = LogDetailCookies
	l.printNow()
	return l.valid
}

// Body logs response body.
func (l *ResponseLogSpec) Body() *ValidatableResponse {
	l.detail = LogDetailBody
	l.printNow()
	return l.valid
}

// IfValidationFails enables full request and response logging the first time any
// assertion on this ValidatableResponse fails, mirroring the request-level
// r.Log().IfValidationFails() behaviour in REST Assured Java.
func (l *ResponseLogSpec) IfValidationFails() *ValidatableResponse {
	if l.valid == nil || l.valid.response == nil {
		return l.valid
	}
	cfg := l.valid.response.Config()
	lc := cfg.LogConfig()
	lc.EnableLoggingIfValidationFails = true
	if l.writer != nil {
		lc.Output = l.writer
	}
	l.valid.response.config = cfg.WithLog(lc)
	return l.valid
}

// IfError logs response details if status code is >= 400.
func (l *ResponseLogSpec) IfError() *ValidatableResponse {
	if l.resp != nil && l.resp.StatusCode() >= 400 {
		l.detail = LogDetailAll
		l.printNow()
	}
	return l.valid
}

// IfStatusCodeIs logs response details if status matches given code.
func (l *ResponseLogSpec) IfStatusCodeIs(expectedCode int) *ValidatableResponse {
	if l.resp != nil && l.resp.StatusCode() == expectedCode {
		l.detail = LogDetailAll
		l.printNow()
	}
	return l.valid
}

func (l *ResponseLogSpec) printNow() {
	if l.resp == nil {
		return
	}
	targetWriter := l.writer
	if targetWriter == nil {
		targetWriter = os.Stdout
	}
	logResponse(targetWriter, l.resp, l.detail)
}

// PrettyPrintXML returns a 2-space indented XML string from the input.
// Returns ("", false) if input is empty or not valid XML.
func PrettyPrintXML(input string) (string, bool) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", false
	}
	var buf bytes.Buffer
	dec := xml.NewDecoder(strings.NewReader(input))
	enc := xml.NewEncoder(&buf)
	enc.Indent("", "  ")
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", false
		}
		if err := enc.EncodeToken(tok); err != nil {
			return "", false
		}
	}
	if err := enc.Flush(); err != nil {
		return "", false
	}
	return buf.String(), true
}

// shouldLog reports whether the given section should be included for the current detail level.
func shouldLog(detail, section LogDetail) bool {
	return detail == LogDetailAll || detail == section
}

// isBlacklisted reports whether name appears in the blacklist (case-insensitive).
func isBlacklisted(name string, blacklist []string) bool {
	for _, b := range blacklist {
		if strings.EqualFold(name, b) {
			return true
		}
	}
	return false
}

// headerValue returns the value to log: "[REDACTED]" when the header is blacklisted.
func headerValue(name, value string, blacklist []string) string {
	if isBlacklisted(name, blacklist) {
		return "[REDACTED]"
	}
	return value
}

// logRequest logs request details. blacklist is a list of header names whose values
// are redacted; pass nil to disable redaction.
func logRequest(w io.Writer, req *FilterableRequest, detail LogDetail, blacklist []string) {
	var sb strings.Builder
	sb.WriteString("\n--- Request Details ---\n")
	if shouldLog(detail, LogDetailStatus) {
		sb.WriteString(fmt.Sprintf("Request method:\t%s\n", req.Method))
		sb.WriteString(fmt.Sprintf("Request URI:\t%s\n", req.URI))
	}

	if shouldLog(detail, LogDetailHeaders) {
		sb.WriteString("Headers:\n")
		for name, values := range req.Headers {
			for _, val := range values {
				sb.WriteString(fmt.Sprintf("\t%s: %s\n", name, headerValue(name, val, blacklist)))
			}
		}
	}

	if shouldLog(detail, LogDetailCookies) {
		sb.WriteString("Cookies:\n")
		for _, cookie := range req.Cookies {
			sb.WriteString(fmt.Sprintf("\t%s=%s\n", cookie.Name, headerValue("Cookie", cookie.Value, blacklist)))
		}
	}

	if shouldLog(detail, LogDetailBody) {
		sb.WriteString("Body:\n")
		if len(req.Body) > 0 {
			sb.WriteString(prettyBody(req.Body) + "\n")
		}
	}
	sb.WriteString("-----------------------\n")
	_, _ = fmt.Fprint(w, sb.String())
}

// logResponse writes response details to w according to detail. Header values are
// redacted for any name that appears in the response Config's BlacklistHeaders list.
func logResponse(w io.Writer, resp *Response, detail LogDetail) {
	var blacklist []string
	if resp != nil {
		blacklist = resp.Config().LogConfig().BlacklistHeaders
	}

	var sb strings.Builder
	sb.WriteString("\n--- Response Details ---\n")
	if shouldLog(detail, LogDetailStatus) {
		sb.WriteString(fmt.Sprintf("Status line:\t%s\n", resp.StatusLine()))
		sb.WriteString(fmt.Sprintf("Status code:\t%d\n", resp.StatusCode()))
	}

	if shouldLog(detail, LogDetailHeaders) {
		sb.WriteString("Headers:\n")
		for name, values := range resp.Headers() {
			for _, val := range values {
				sb.WriteString(fmt.Sprintf("\t%s: %s\n", name, headerValue(name, val, blacklist)))
			}
		}
	}

	if shouldLog(detail, LogDetailCookies) {
		sb.WriteString("Cookies:\n")
		for _, cookie := range resp.Cookies() {
			sb.WriteString(fmt.Sprintf("\t%s=%s\n", cookie.Name, headerValue("Cookie", cookie.Value, blacklist)))
		}
	}

	if shouldLog(detail, LogDetailBody) {
		sb.WriteString("Body:\n")
		body := resp.AsBytes()
		if len(body) > 0 {
			sb.WriteString(prettyBody(body) + "\n")
		}
	}
	sb.WriteString("------------------------\n")
	_, _ = fmt.Fprint(w, sb.String())
}

// prettyBody returns a human-readable string for body bytes: indented JSON, indented XML, or raw text.
func prettyBody(body []byte) string {
	var prettyJSON bytes.Buffer
	if json.Indent(&prettyJSON, body, "", "  ") == nil {
		return prettyJSON.String()
	}
	if pretty, ok := PrettyPrintXML(string(body)); ok {
		return pretty
	}
	return string(body)
}
