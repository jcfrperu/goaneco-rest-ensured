package rest

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/tidwall/gjson"

	"github.com/jcfrperu/goaneco-rest-ensured/jsonschema"
)

// ValidatableResponse provides fluent assertion methods for validating HTTP responses.
type ValidatableResponse struct {
	response      *Response
	t             testing.TB
	rootPath      string
	failures      []string
	failureLogged bool // guards against logging the same request/response on every failure
}

// AssertWith attaches a *testing.T or testing.TB to report assertion failures immediately.
func (v *ValidatableResponse) AssertWith(t testing.TB) *ValidatableResponse {
	v.t = t
	return v
}

// AssertThat is syntactic sugar returning the ValidatableResponse.
func (v *ValidatableResponse) AssertThat() *ValidatableResponse {
	return v
}

// And is syntactic sugar returning the ValidatableResponse.
func (v *ValidatableResponse) And() *ValidatableResponse {
	return v
}

// Extract transitions to the ExtractableResponse for extracting values.
func (v *ValidatableResponse) Extract() *ExtractableResponse {
	return &ExtractableResponse{resp: v.response}
}

// IsValid returns true if no assertion failures have been recorded.
func (v *ValidatableResponse) IsValid() bool {
	if v == nil {
		return false
	}
	return len(v.failures) == 0
}

// Spec validates the response against a reusable ResponseSpec.
func (v *ValidatableResponse) Spec(spec *ResponseSpec) *ValidatableResponse {
	if spec != nil {
		return spec.Validate(v)
	}
	return v
}

// Log returns a ResponseLogSpec for logging response data.
func (v *ValidatableResponse) Log() *ResponseLogSpec {
	return &ResponseLogSpec{
		resp:  v.response,
		valid: v,
	}
}

// checkResponse returns false and records a failure when the response is nil or carries a
// transport error, preventing nil-pointer panics in subsequent assertion methods.
func (v *ValidatableResponse) checkResponse() bool {
	if v.response == nil {
		v.recordFailure("response is nil")
		return false
	}
	if v.response.Err() != nil {
		v.recordFailure(fmt.Sprintf("request failed with error: %v", v.response.Err()))
		return false
	}
	return true
}

// StatusCode asserts that the response status code equals expected.
func (v *ValidatableResponse) StatusCode(expected int) *ValidatableResponse {
	if !v.checkResponse() {
		return v
	}
	actual := v.response.StatusCode()
	if actual != expected {
		v.recordFailure(fmt.Sprintf("expected status code %d, but got %d", expected, actual))
	}
	return v
}

// StatusCodeBetween asserts that status code is within [minCode, maxCode] inclusive.
func (v *ValidatableResponse) StatusCodeBetween(minCode, maxCode int) *ValidatableResponse {
	if !v.checkResponse() {
		return v
	}
	actual := v.response.StatusCode()
	if actual < minCode || actual > maxCode {
		v.recordFailure(fmt.Sprintf("expected status code between %d and %d, but got %d", minCode, maxCode, actual))
	}
	return v
}

// StatusLine asserts that the status line equals expected.
func (v *ValidatableResponse) StatusLine(expected string) *ValidatableResponse {
	if !v.checkResponse() {
		return v
	}
	actual := v.response.StatusLine()
	if actual != expected {
		v.recordFailure(fmt.Sprintf("expected status line %q, but got %q", expected, actual))
	}
	return v
}

// StatusLineContains asserts that status line contains the specified substring.
func (v *ValidatableResponse) StatusLineContains(substring string) *ValidatableResponse {
	if !v.checkResponse() {
		return v
	}
	actual := v.response.StatusLine()
	if !strings.Contains(actual, substring) {
		v.recordFailure(fmt.Sprintf("expected status line %q to contain %q", actual, substring))
	}
	return v
}

// Header asserts that the header equals expected value.
func (v *ValidatableResponse) Header(name, expected string) *ValidatableResponse {
	if !v.checkResponse() {
		return v
	}
	actual := v.response.Header(name)
	if actual != expected {
		v.recordFailure(fmt.Sprintf("expected header %q = %q, but got %q", name, expected, actual))
	}
	return v
}

// HeaderContains asserts that the header value contains substring.
func (v *ValidatableResponse) HeaderContains(name, substring string) *ValidatableResponse {
	if !v.checkResponse() {
		return v
	}
	actual := v.response.Header(name)
	if !strings.Contains(actual, substring) {
		v.recordFailure(fmt.Sprintf("expected header %q (%q) to contain %q", name, actual, substring))
	}
	return v
}

// HeaderExists asserts that the header is present in the response (regardless of its value).
func (v *ValidatableResponse) HeaderExists(name string) *ValidatableResponse {
	if !v.checkResponse() {
		return v
	}
	// Use the canonical key so "content-type" and "Content-Type" match equally.
	if _, ok := v.response.headers[http.CanonicalHeaderKey(name)]; !ok {
		v.recordFailure(fmt.Sprintf("expected header %q to exist", name))
	}
	return v
}

// Cookie asserts that cookie with name has the expected value.
func (v *ValidatableResponse) Cookie(name, expected string) *ValidatableResponse {
	if !v.checkResponse() {
		return v
	}
	actual := v.response.Cookie(name)
	if actual != expected {
		v.recordFailure(fmt.Sprintf("expected cookie %q = %q, but got %q", name, expected, actual))
	}
	return v
}

// CookieExists asserts that a cookie with the given name exists (regardless of its value).
func (v *ValidatableResponse) CookieExists(name string) *ValidatableResponse {
	if !v.checkResponse() {
		return v
	}
	for _, c := range v.response.cookies {
		if c.Name == name {
			return v
		}
	}
	v.recordFailure(fmt.Sprintf("expected cookie %q to exist", name))
	return v
}

// ContentType asserts that the base media type of the Content-Type header equals expected.
// Parameters such as "; charset=utf-8" are stripped before comparison.
// ContentTypeAny ("*/*") matches any non-empty Content-Type value.
func (v *ValidatableResponse) ContentType(expected ContentType) *ValidatableResponse {
	if !v.checkResponse() {
		return v
	}
	actual := v.response.ContentType()
	if expected == ContentTypeAny {
		if actual == "" {
			v.recordFailure(fmt.Sprintf("expected content-type %q (any), but response had no Content-Type header", expected))
		}
		return v
	}
	actualBase := strings.ToLower(strings.TrimSpace(strings.SplitN(actual, ";", 2)[0]))
	expectedBase := strings.ToLower(strings.TrimSpace(strings.SplitN(string(expected), ";", 2)[0]))
	if actualBase != expectedBase {
		v.recordFailure(fmt.Sprintf("expected content-type %q, but got %q", expected, actual))
	}
	return v
}

// ContentTypeContains asserts that the Content-Type header contains substring.
func (v *ValidatableResponse) ContentTypeContains(substring string) *ValidatableResponse {
	if !v.checkResponse() {
		return v
	}
	actual := v.response.ContentType()
	if !strings.Contains(actual, substring) {
		v.recordFailure(fmt.Sprintf("expected content-type %q to contain %q", actual, substring))
	}
	return v
}

// TimeLessThan asserts that the response roundtrip time was strictly less than maxDuration.
func (v *ValidatableResponse) TimeLessThan(maxDuration time.Duration) *ValidatableResponse {
	if !v.checkResponse() {
		return v
	}
	actual := v.response.Time()
	if actual >= maxDuration {
		v.recordFailure(fmt.Sprintf("expected response time < %v, but took %v", maxDuration, actual))
	}
	return v
}

// RootPath sets the base root path for subsequent Body assertions.
func (v *ValidatableResponse) RootPath(path string) *ValidatableResponse {
	v.rootPath = path
	return v
}

// AppendRootPath appends a sub-path to the current root path.
func (v *ValidatableResponse) AppendRootPath(subPath string) *ValidatableResponse {
	if v.rootPath == "" {
		v.rootPath = subPath
	} else {
		v.rootPath = v.rootPath + "." + strings.TrimPrefix(subPath, ".")
	}
	return v
}

// NoRootPath resets the root path to empty string.
func (v *ValidatableResponse) NoRootPath() *ValidatableResponse {
	v.rootPath = ""
	return v
}

// StatusCodeMatching asserts that the status code matches the given matcher.
func (v *ValidatableResponse) StatusCodeMatching(m any) *ValidatableResponse {
	if !v.checkResponse() {
		return v
	}
	actual := v.response.StatusCode()
	switch matcher := m.(type) {
	case interface {
		Matches(int) bool
		Describe() string
	}:
		if !matcher.Matches(actual) {
			v.recordFailure(fmt.Sprintf("expected status code to match %s, but got %d", matcher.Describe(), actual))
		}
	case interface {
		Matches(any) bool
		Describe() string
	}:
		if !matcher.Matches(actual) {
			v.recordFailure(fmt.Sprintf("expected status code to match %s, but got %d", matcher.Describe(), actual))
		}
	default:
		v.recordFailure(fmt.Sprintf("invalid matcher type %T for status code", m))
	}
	return v
}

// HeaderMatching asserts that the header value matches the given matcher.
func (v *ValidatableResponse) HeaderMatching(name string, m any) *ValidatableResponse {
	if !v.checkResponse() {
		return v
	}
	actual := v.response.Header(name)
	switch matcher := m.(type) {
	case interface {
		Matches(string) bool
		Describe() string
	}:
		if !matcher.Matches(actual) {
			v.recordFailure(fmt.Sprintf("expected header %q to match %s, but got %q", name, matcher.Describe(), actual))
		}
	case interface {
		Matches(any) bool
		Describe() string
	}:
		if !matcher.Matches(actual) {
			v.recordFailure(fmt.Sprintf("expected header %q to match %s, but got %q", name, matcher.Describe(), actual))
		}
	default:
		v.recordFailure(fmt.Sprintf("invalid matcher type %T for header %q", m, name))
	}
	return v
}

// CookieMatching asserts that the cookie value matches the given matcher.
func (v *ValidatableResponse) CookieMatching(name string, m any) *ValidatableResponse {
	if !v.checkResponse() {
		return v
	}
	actual := v.response.Cookie(name)
	switch matcher := m.(type) {
	case interface {
		Matches(string) bool
		Describe() string
	}:
		if !matcher.Matches(actual) {
			v.recordFailure(fmt.Sprintf("expected cookie %q to match %s, but got %q", name, matcher.Describe(), actual))
		}
	case interface {
		Matches(any) bool
		Describe() string
	}:
		if !matcher.Matches(actual) {
			v.recordFailure(fmt.Sprintf("expected cookie %q to match %s, but got %q", name, matcher.Describe(), actual))
		}
	default:
		v.recordFailure(fmt.Sprintf("invalid matcher type %T for cookie %q", m, name))
	}
	return v
}

// evalGjsonPath evaluates path against body using gjson, with two fallback strategies
// when the direct lookup finds nothing:
//  1. Escape "@" characters, which gjson treats as modifier prefixes.
//  2. Insert "#." at each dot boundary to try GPath-style array projections
//     (e.g. "store.book.price" → "store.#.price" → "store.book.#.price").
//     This fallback is skipped when the path contains gjson query predicates "#(…)"
//     to avoid splitting on operators inside them.
func evalGjsonPath(body []byte, path string) gjson.Result {
	res := gjson.GetBytes(body, path)
	if res.Exists() {
		return res
	}

	// Try escaping @ symbols
	if strings.Contains(path, "@") {
		escaped := strings.ReplaceAll(path, "@", "\\@")
		res = gjson.GetBytes(body, escaped)
		if res.Exists() {
			return res
		}
	}

	// Try converting dot segments to .#. for GPath-style array projections (e.g. store.book.price -> store.book.#.price).
	// Skip when the path contains gjson query predicates (#(...)) to avoid splitting on decimal digits inside them.
	// Limit iterations to avoid quadratic cost on very deep paths.
	if strings.Contains(path, "#(") {
		return gjson.Result{}
	}
	parts := strings.Split(path, ".")
	const maxProjectionDepth = 10
	limit := len(parts)
	if limit > maxProjectionDepth+1 {
		limit = maxProjectionDepth + 1
	}
	if len(parts) > 1 {
		for i := 1; i < limit; i++ {
			altPath := strings.Join(parts[:i], ".") + ".#." + strings.Join(parts[i:], ".")
			if strings.Contains(altPath, "@") {
				altPath = strings.ReplaceAll(altPath, "@", "\\@")
			}
			res = gjson.GetBytes(body, altPath)
			if res.Exists() {
				return res
			}
		}
	}
	return res
}

// matchesExpected dispatches val against expected using type-switch on well-known matcher
// interfaces (Matches(int), Matches(string), Matches(any), etc.) to avoid reflection on
// the common path. It falls back to reflect-based dispatch for custom matchers with
// non-standard Matches signatures. Returns (matched, description, isMatcher); when
// isMatcher=false, expected is not a matcher and the caller should fall back to equality.
func matchesExpected(val any, expected any) (bool, string, bool) {
	if expected == nil {
		return val == nil, "nil", true
	}

	// Fast path: type assertions for common matcher interfaces avoid reflection on the hot path.
	// Each Matches(T) signature is a distinct interface; matchers implement exactly one.
	switch m := expected.(type) {
	case interface {
		Matches(any) bool
		Describe() string
	}:
		return m.Matches(val), m.Describe(), true
	case interface {
		Matches(int) bool
		Describe() string
	}:
		desc := m.Describe()
		switch v := val.(type) {
		case int:
			return m.Matches(v), desc, true
		case float64:
			if v >= math.MinInt && v <= math.MaxInt {
				return m.Matches(int(v)), desc, true
			}
			return false, desc, true
		case int64:
			if v >= math.MinInt && v <= math.MaxInt {
				return m.Matches(int(v)), desc, true
			}
			return false, desc, true
		case int32:
			return m.Matches(int(v)), desc, true
		}
		return false, desc, true
	case interface {
		Matches(int64) bool
		Describe() string
	}:
		desc := m.Describe()
		switch v := val.(type) {
		case int64:
			return m.Matches(v), desc, true
		case int:
			return m.Matches(int64(v)), desc, true
		case float64:
			return m.Matches(int64(v)), desc, true
		}
		return false, desc, true
	case interface {
		Matches(float64) bool
		Describe() string
	}:
		desc := m.Describe()
		switch v := val.(type) {
		case float64:
			return m.Matches(v), desc, true
		case int:
			return m.Matches(float64(v)), desc, true
		case int64:
			return m.Matches(float64(v)), desc, true
		}
		return false, desc, true
	case interface {
		Matches(bool) bool
		Describe() string
	}:
		desc := m.Describe()
		if v, ok := val.(bool); ok {
			return m.Matches(v), desc, true
		}
		return false, desc, true
	case interface {
		Matches(string) bool
		Describe() string
	}:
		desc := m.Describe()
		if s, ok := val.(string); ok {
			return m.Matches(s), desc, true
		}
		return false, desc, true
	}

	// Reflection fallback for custom Matches signatures not covered above.
	expVal := reflect.ValueOf(expected)
	matchesMethod := expVal.MethodByName("Matches")
	describeMethod := expVal.MethodByName("Describe")
	if !matchesMethod.IsValid() || !describeMethod.IsValid() {
		return false, "", false
	}

	desc := "matcher"
	if dRes := describeMethod.Call(nil); len(dRes) > 0 {
		desc = dRes[0].String()
	}

	mType := matchesMethod.Type()
	if mType.NumIn() != 1 || mType.NumOut() != 1 || mType.Out(0).Kind() != reflect.Bool {
		return false, "", false
	}
	targetType := mType.In(0)
	argVal := reflect.ValueOf(val)
	if !argVal.IsValid() {
		if targetType.Kind() == reflect.Interface || targetType.Kind() == reflect.Pointer {
			res := matchesMethod.Call([]reflect.Value{reflect.Zero(targetType)})
			return res[0].Bool(), desc, true
		}
		return false, desc, true
	}
	if argVal.Type().AssignableTo(targetType) {
		res := matchesMethod.Call([]reflect.Value{argVal})
		return res[0].Bool(), desc, true
	}
	if argVal.Type().ConvertibleTo(targetType) {
		res := matchesMethod.Call([]reflect.Value{argVal.Convert(targetType)})
		return res[0].Bool(), desc, true
	}
	if targetType.Kind() == reflect.Int {
		if f, ok := val.(float64); ok {
			res := matchesMethod.Call([]reflect.Value{reflect.ValueOf(int(f))})
			return res[0].Bool(), desc, true
		}
	}
	if targetType.Kind() == reflect.Float64 {
		if i, ok := val.(int); ok {
			res := matchesMethod.Call([]reflect.Value{reflect.ValueOf(float64(i))})
			return res[0].Bool(), desc, true
		}
	}
	if targetType.Kind() == reflect.String {
		res := matchesMethod.Call([]reflect.Value{reflect.ValueOf(fmt.Sprintf("%v", val))})
		return res[0].Bool(), desc, true
	}
	return false, desc, true
}

// Body asserts that the JSON path expression evaluates to expected value or satisfies the matcher.
func (v *ValidatableResponse) Body(path string, expected any) *ValidatableResponse {
	if !v.checkResponse() {
		return v
	}
	fullPath := v.resolvePath(path)
	if fullPath == "" {
		var val any
		if gjson.ValidBytes(v.response.body) {
			val = gjson.ParseBytes(v.response.body).Value()
		} else {
			val = v.response.AsString()
		}
		if matched, desc, isMatcher := matchesExpected(val, expected); isMatcher {
			if !matched {
				v.recordFailure(fmt.Sprintf("response body mismatch:\nExpected: %s\n     but: was <%v>",
					desc, val))
			}
			return v
		}
		if !reflect.DeepEqual(expected, val) && v.response.AsString() != fmt.Sprintf("%v", expected) {
			v.recordFailure(fmt.Sprintf("expected body %v (%T), but got %v", expected, expected, val))
		}
		return v
	}

	result := evalGjsonPath(v.response.body, fullPath)
	if !result.Exists() {
		v.recordFailure(fmt.Sprintf("path %q does not exist in response body", fullPath))
		return v
	}

	val := result.Value()
	if matched, desc, isMatcher := matchesExpected(val, expected); isMatcher {
		if !matched {
			v.recordFailure(fmt.Sprintf("path %q mismatch:\nExpected: %s\n     but: was <%v>",
				fullPath, desc, val))
		}
		return v
	}

	if !matchValue(result, expected) {
		v.recordFailure(fmt.Sprintf("path %q: expected %v (%T), but got %v (%s)",
			fullPath, expected, expected, result.Value(), result.Type.String()))
	}
	return v
}

// BodyContains asserts that the raw response body contains substring.
func (v *ValidatableResponse) BodyContains(substring string) *ValidatableResponse {
	if !v.checkResponse() {
		return v
	}
	actual := v.response.AsString()
	if !strings.Contains(actual, substring) {
		v.recordFailure(fmt.Sprintf("expected response body to contain %q", substring))
	}
	return v
}

// BodyEquals asserts that the raw response body equals expected string.
func (v *ValidatableResponse) BodyEquals(expected string) *ValidatableResponse {
	if !v.checkResponse() {
		return v
	}
	actual := v.response.AsString()
	if actual != expected {
		v.recordFailure(fmt.Sprintf("expected response body %q, but got %q", expected, actual))
	}
	return v
}

// BodyJSONEquals asserts that the response body is semantically equal to expected JSON.
func (v *ValidatableResponse) BodyJSONEquals(expectedJSON string) *ValidatableResponse {
	if !v.checkResponse() {
		return v
	}
	var exp, act any
	if err := json.Unmarshal([]byte(expectedJSON), &exp); err != nil {
		v.recordFailure(fmt.Sprintf("invalid expected JSON: %v", err))
		return v
	}
	if err := json.Unmarshal(v.response.body, &act); err != nil {
		v.recordFailure(fmt.Sprintf("invalid actual JSON: %v", err))
		return v
	}
	if !reflect.DeepEqual(exp, act) {
		v.recordFailure(fmt.Sprintf("expected JSON %s, but got %s", expectedJSON, string(v.response.body)))
	}
	return v
}

// BodyContainsElement asserts that the array at path contains expected element.
func (v *ValidatableResponse) BodyContainsElement(path string, element any) *ValidatableResponse {
	if !v.checkResponse() {
		return v
	}
	fullPath := v.resolvePath(path)
	result := evalGjsonPath(v.response.body, fullPath)

	if !result.Exists() {
		v.recordFailure(fmt.Sprintf("path %q does not exist in response body", fullPath))
		return v
	}
	if !result.IsArray() {
		v.recordFailure(fmt.Sprintf("path %q is not an array", fullPath))
		return v
	}

	found := false
	for _, item := range result.Array() {
		if matched, _, isMatcher := matchesExpected(item.Value(), element); isMatcher {
			if matched {
				found = true
				break
			}
		} else if matchValue(item, element) {
			found = true
			break
		}
	}

	if !found {
		v.recordFailure(fmt.Sprintf("path %q: array does not contain element %v", fullPath, element))
	}
	return v
}

// BodyLength asserts that the array or string at path has expected length.
func (v *ValidatableResponse) BodyLength(path string, expectedLength int) *ValidatableResponse {
	if !v.checkResponse() {
		return v
	}
	fullPath := v.resolvePath(path)
	result := evalGjsonPath(v.response.body, fullPath)

	if result.IsArray() {
		actualLen := len(result.Array())
		if actualLen != expectedLength {
			v.recordFailure(fmt.Sprintf("path %q: expected array length %d, but got %d", fullPath, expectedLength, actualLen))
		}
		return v
	}

	if result.Type == gjson.String {
		actualLen := utf8.RuneCountInString(result.String())
		if actualLen != expectedLength {
			v.recordFailure(fmt.Sprintf("path %q: expected string length %d, but got %d", fullPath, expectedLength, actualLen))
		}
		return v
	}

	v.recordFailure(fmt.Sprintf("path %q is neither an array nor a string", fullPath))
	return v
}

// BodyMatchesSchema asserts that the response JSON matches the given JSON Schema string.
func (v *ValidatableResponse) BodyMatchesSchema(schemaJSON string) *ValidatableResponse {
	if !v.checkResponse() {
		return v
	}
	validator := jsonschema.MatchesJsonSchema(schemaJSON)
	if err := validator.ValidateBytes(v.response.body); err != nil {
		v.recordFailure(fmt.Sprintf("body schema validation failed: %v", err))
	}
	return v
}

// BodyMatchesSchemaFile asserts that the response JSON matches the JSON Schema loaded from disk.
func (v *ValidatableResponse) BodyMatchesSchemaFile(schemaFilePath string) *ValidatableResponse {
	if !v.checkResponse() {
		return v
	}
	validator := jsonschema.MatchesJsonSchemaFile(schemaFilePath)
	if err := validator.ValidateBytes(v.response.body); err != nil {
		v.recordFailure(fmt.Sprintf("body schema validation from file %q failed: %v", schemaFilePath, err))
	}
	return v
}

// BodyMatchesSchemaValidator asserts that the response JSON matches a custom jsonschema.Validator instance.
func (v *ValidatableResponse) BodyMatchesSchemaValidator(validator *jsonschema.Validator) *ValidatableResponse {
	if !v.checkResponse() {
		return v
	}
	if validator == nil {
		v.recordFailure("provided schema validator is nil")
		return v
	}
	if err := validator.ValidateBytes(v.response.body); err != nil {
		v.recordFailure(fmt.Sprintf("body schema validation failed: %v", err))
	}
	return v
}

// Failures returns all recorded assertion failure messages.
func (v *ValidatableResponse) Failures() []string {
	return append([]string{}, v.failures...)
}

// HasFailures returns true if any assertion failed.
func (v *ValidatableResponse) HasFailures() bool {
	return len(v.failures) > 0
}

// AssertAll returns an error aggregating all failures if any occurred.
func (v *ValidatableResponse) AssertAll() error {
	if len(v.failures) == 0 {
		return nil
	}
	return fmt.Errorf("%w:\n- %s", ErrAssertionFailed, strings.Join(v.failures, "\n- "))
}

// Assert checks that all assertions passed using t.Errorf on failure.
func (v *ValidatableResponse) Assert(t testing.TB) {
	t.Helper()
	if len(v.failures) > 0 {
		t.Errorf("%v:\n- %s", ErrAssertionFailed, strings.Join(v.failures, "\n- "))
	}
}

// AssertAllNoFail fails immediately with t.Fatalf if any assertion failed.
func (v *ValidatableResponse) AssertAllNoFail(t testing.TB) {
	t.Helper()
	if len(v.failures) > 0 {
		t.Fatalf("%v:\n- %s", ErrAssertionFailed, strings.Join(v.failures, "\n- "))
	}
}

// resolvePath prefixes path with v.rootPath when one has been set, enabling all Body
// assertions to operate relative to a common JSON sub-tree without repeating the prefix.
func (v *ValidatableResponse) resolvePath(path string) string {
	if v.rootPath == "" {
		return path
	}
	if path == "" {
		return v.rootPath
	}
	return v.rootPath + "." + strings.TrimPrefix(path, ".")
}

// recordFailure appends msg to the failure list, reports it to t if set, notifies
// registered FailureConfig listeners (with a snapshot copy of failures so they cannot
// observe future appends), and logs the full request/response when
// EnableLoggingIfValidationFails is set.
func (v *ValidatableResponse) recordFailure(msg string) {
	v.failures = append(v.failures, msg)
	if v.t != nil {
		v.t.Helper()
		v.t.Error(msg)
	}

	if v.response != nil {
		cfg := v.response.Config()
		// Trigger failure listeners
		listeners := cfg.FailureConfig().Listeners
		if len(listeners) > 0 {
			req := v.response.RawRequest()
			// Pass a snapshot copy so listeners cannot observe future appends
			// and so the backing array cannot be mutated behind their backs.
			snapshot := append([]string{}, v.failures...)
			for _, listener := range listeners {
				if listener != nil {
					listener(req, v.response, snapshot)
				}
			}
		}

		// Trigger logging on validation failure if enabled — log only once across all failures.
		if cfg.LogConfig().EnableLoggingIfValidationFails && !v.failureLogged {
			v.failureLogged = true
			out := cfg.LogConfig().Output
			if out == nil {
				out = os.Stderr
			}
			if v.response.req != nil {
				fReq := &FilterableRequest{
					Method:      v.response.req.Method,
					URI:         v.response.req.URL.String(),
					Headers:     v.response.req.Header,
					Cookies:     v.response.req.Cookies(),
					Body:        v.response.reqBody,
					ContentType: v.response.req.Header.Get("Content-Type"),
				}
				logRequest(out, fReq, LogDetailAll, cfg.LogConfig().BlacklistHeaders)
			}
			logResponse(out, v.response, LogDetailAll)
		}
	}
}

// matchValue compares a gjson.Result against a typed Go value using the most appropriate
// comparison: exact string/integer/float/boolean equality for scalar types, and JSON
// marshal-then-compare for composite or unknown types.
func matchValue(result gjson.Result, expected any) bool {
	switch exp := expected.(type) {
	case string:
		return result.String() == exp
	case int:
		return result.Int() == int64(exp)
	case int8:
		return result.Int() == int64(exp)
	case int16:
		return result.Int() == int64(exp)
	case int32:
		return result.Int() == int64(exp)
	case int64:
		return result.Int() == exp
	case uint:
		return result.Uint() == uint64(exp)
	case uint8:
		return result.Uint() == uint64(exp)
	case uint16:
		return result.Uint() == uint64(exp)
	case uint32:
		return result.Uint() == uint64(exp)
	case uint64:
		return result.Uint() == exp
	case float64:
		return math.Abs(result.Float()-exp) < 1e-9
	case float32:
		return math.Abs(result.Float()-float64(exp)) < 1e-6
	case bool:
		return result.Bool() == exp
	case nil:
		return !result.Exists() || result.Type == gjson.Null
	default:
		// Fallback to json equality check
		expJSON, err := json.Marshal(expected)
		if err != nil {
			return false
		}
		var expVal, actVal any
		if json.Unmarshal(expJSON, &expVal) != nil || json.Unmarshal([]byte(result.Raw), &actVal) != nil {
			return false
		}
		return reflect.DeepEqual(expVal, actVal)
	}
}
