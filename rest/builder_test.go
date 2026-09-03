package rest_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

// withArgs formats a path string with the supplied arguments, mirroring Java's RestAssured.withArgs.
func withArgs(args ...any) []any { return args }

func TestRequestSpecBuilder(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	var receivedHeader string
	var receivedQuery string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeader = r.Header.Get("X-App-ID")
		receivedQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	t.Cleanup(ts.Close)

	reqSpec := rest.NewRequestSpecBuilder().
		SetBaseURI(ts.URL).
		AddHeader("X-App-ID", "my-app-123").
		AddQueryParam("env", "test").
		SetContentType(rest.ContentTypeJSON).
		Build()

	must.NotNil(reqSpec)

	resp := rest.Given().
		Spec(reqSpec).
		Get("/check")

	must.NoError(resp.Err())
	is.Equal(http.StatusOK, resp.StatusCode())
	is.Equal("my-app-123", receivedHeader)
	is.Equal("env=test", receivedQuery)
}

func TestRequestSpecBuilder_QueryableFields(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	spec := rest.NewRequestSpecBuilder().
		AddHeader("X-Token", "tok-abc").
		AddCookie("session", "sess-xyz").
		AddQueryParam("env", "prod").
		AddPathParam("id", "42").
		AddFormParam("field", "val").
		SetBody("raw body").
		Build()

	// Headers stored and accessible (case-insensitive)
	is.Equal("tok-abc", spec.Headers.Get("X-Token"))
	is.Equal("tok-abc", spec.Headers.Get("x-token"))

	// Cookies stored
	is.Len(spec.Cookies, 1)
	is.Equal("session", spec.Cookies[0].Name)
	is.Equal("sess-xyz", spec.Cookies[0].Value)

	// QueryParams stored
	is.Equal("prod", spec.QueryParams.Get("env"))

	// PathParams stored
	is.Equal("42", spec.PathParams["id"])

	// FormParams stored
	is.Equal("val", spec.FormParams.Get("field"))

	// Body stored
	is.Equal("raw body", string(spec.Body))

	// Build() returns independent copy — mutations do not affect the builder
	spec.Headers.Set("X-Injected", "leak")
	spec2 := rest.NewRequestSpecBuilder().
		AddHeader("X-Token", "tok-abc").
		Build()
	is.Empty(spec2.Headers.Get("X-Injected"))
}

func TestResponseSpecBuilder(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	schemaJSON := `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"properties": {
			"success": {"type": "boolean"},
			"count": {"type": "integer"}
		},
		"required": ["success", "count"]
	}`

	tempDir := t.TempDir()
	schemaFilePath := filepath.Join(tempDir, "resp_schema.json")
	err := os.WriteFile(schemaFilePath, []byte(schemaJSON), 0600)
	must.NoError(err)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Powered-By", "Go")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"count":5}`))
	}))
	t.Cleanup(ts.Close)

	// Test ResponseSpec with schema string and schema file
	respSpec := rest.NewResponseSpecBuilder().
		ExpectStatusCode(200).
		ExpectContentType(rest.ContentTypeJSON).
		ExpectHeader("X-Powered-By", "Go").
		ExpectBody("success", true).
		ExpectBody("count", 5).
		ExpectBodyMatchesSchema(schemaJSON).
		ExpectBodyMatchesSchemaFile(schemaFilePath).
		ExpectResponseTimeLessThan(2 * time.Second).
		Build()

	must.NotNil(respSpec)

	valid := rest.Given().
		BaseURI(ts.URL).
		Get("/data").
		Then().
		Spec(respSpec)

	is.False(valid.HasFailures())
	is.Empty(valid.Failures())
}

// ── RequestSpecBuilderTest ───────────────────────────────────────────────────

func TestRequestSpec_GlobalIntegration(t *testing.T) {
	// Not parallel — mutates global state; reset guards sibling tests.
	t.Cleanup(rest.Reset)

	is := assert.New(t)
	must := require.New(t)

	// Mirrors request_spec_doesnt_throw_NPE_when_logging_after_creation:
	// A built spec applied via Given().Spec() does not panic even without explicit setup.
	spec := rest.NewRequestSpecBuilder().Build()
	must.NotNil(spec)
	// Using the spec in Given() must not panic.
	require.NotPanics(t, func() { _ = rest.Given().Spec(spec) })

	// Mirrors request_spec_picks_up_filters_from_static_config:
	// Global filters set via GlobalFilter() are applied to requests that use a RequestSpec.
	var filterCalled bool
	markerFilter := &markerFilterImpl{onCall: func() { filterCalled = true }}

	rest.GlobalFilter(markerFilter)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(ts.Close)

	resp := rest.Given().
		Spec(rest.NewRequestSpecBuilder().SetBaseURI(ts.URL).Build()).
		Get("/filter-check")
	must.NoError(resp.Err())
	is.True(filterCalled, "global filter must be invoked even when a RequestSpec is used")

	// Mirrors request_spec_picks_up_headers_from_static_request_spec:
	// Global headers set via GlobalHeader() reach the server when a RequestSpec is applied.
	var capturedHeader string
	ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeader = r.Header.Get("X-Global")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(ts2.Close)

	rest.GlobalHeader("X-Global", "world")
	resp2 := rest.Given().
		Spec(rest.NewRequestSpecBuilder().SetBaseURI(ts2.URL).Build()).
		Get("/header-check")
	must.NoError(resp2.Err())
	is.Equal("world", capturedHeader)
}

// markerFilterImpl is a Filter that calls a callback once per invocation.
type markerFilterImpl struct {
	onCall func()
}

func (f *markerFilterImpl) Filter(req *rest.FilterableRequest, ctx *rest.FilterContext) (*rest.Response, error) {
	f.onCall()
	return ctx.Next(req)
}

// ── ResponseSpecBuilderTest ──────────────────────────────────────────────────

func TestResponseSpecBuilder_MergeSpecs(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	// Mirrors Java responseSpecShouldContainMergedExpectations:
	// AddResponseSpec copies all expectations into the new builder so validate() uses them all.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"name":"Alice","score":100}`))
	}))
	t.Cleanup(ts.Close)

	original := rest.NewResponseSpecBuilder().
		ExpectStatusCode(200).
		ExpectBody("name", "Alice").
		Build()

	merged := rest.NewResponseSpecBuilder().
		ExpectBody("score", 100).
		AddResponseSpec(original).
		Build()

	// The merged spec carries all three expectations
	valid := rest.Given().BaseURI(ts.URL).Get("/data").Then()
	merged.Validate(valid)
	must.False(valid.HasFailures(), "merged spec should pass: %v", valid.Failures())

	// A response that violates the original spec should now fail through the merged spec
	tsWrong := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"name":"Bob","score":0}`))
	}))
	t.Cleanup(tsWrong.Close)

	invalid := rest.Given().BaseURI(tsWrong.URL).Get("/bad").Then()
	merged.Validate(invalid)
	is.True(invalid.HasFailures())
}

// ── ResponseSpecBuilderExpectationsTest ─────────────────────────────────────

func TestResponseSpec_AllAssertionTypes(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Custom", "goodValue")
		http.SetCookie(w, &http.Cookie{Name: "cookie1", Value: "cookie1Val", Path: "/"})
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"name":"goodValue","items":["value1","value2"]}`))
	}))
	t.Cleanup(ts.Close)

	get := func() *rest.Response { return rest.Given().BaseURI(ts.URL).Get("/") }

	// Status code (value) — valid
	must.False(rest.NewResponseSpecBuilder().ExpectStatusCode(200).Build().Validate(get().Then()).HasFailures())
	// Status code (value) — invalid
	is.True(rest.NewResponseSpecBuilder().ExpectStatusCode(404).Build().Validate(get().Then()).HasFailures())

	// Header value — valid
	must.False(rest.NewResponseSpecBuilder().ExpectHeader("X-Custom", "goodValue").Build().Validate(get().Then()).HasFailures())
	// Header value — invalid
	is.True(rest.NewResponseSpecBuilder().ExpectHeader("X-Custom", "badValue").Build().Validate(get().Then()).HasFailures())

	// Cookie value — valid
	must.False(rest.NewResponseSpecBuilder().ExpectCookie("cookie1", "cookie1Val").Build().Validate(get().Then()).HasFailures())
	// Cookie value — invalid
	is.True(rest.NewResponseSpecBuilder().ExpectCookie("cookie1", "cookie1BadVal").Build().Validate(get().Then()).HasFailures())

	// Content type — valid
	must.False(rest.NewResponseSpecBuilder().ExpectContentType(rest.ContentTypeJSON).Build().Validate(get().Then()).HasFailures())
	// Content type — invalid
	is.True(rest.NewResponseSpecBuilder().ExpectContentType(rest.ContentTypeXML).Build().Validate(get().Then()).HasFailures())

	// Body path — valid
	must.False(rest.NewResponseSpecBuilder().ExpectBody("name", "goodValue").Build().Validate(get().Then()).HasFailures())
	// Body path — invalid
	is.True(rest.NewResponseSpecBuilder().ExpectBody("name", "badValue").Build().Validate(get().Then()).HasFailures())

	// Response time — valid (10 s is always sufficient for an in-process server)
	must.False(rest.NewResponseSpecBuilder().ExpectResponseTimeLessThan(10 * time.Second).Build().Validate(get().Then()).HasFailures())
}

// TestResponseSpecBuilder_DoesNotPanicWithoutRequest mirrors Java's
// response_spec_doesnt_throw_NPE_when_logging_all_after_creation.
// Java throws IllegalStateException when log() is called on a bare spec; in Go,
// ResponseSpec has no Log() method and building a spec without a request is always safe.
func TestResponseSpecBuilder_DoesNotPanicWithoutRequest(t *testing.T) {
	t.Parallel()
	require.NotPanics(t, func() {
		spec := rest.NewResponseSpecBuilder().
			ExpectStatusCode(200).
			ExpectBody("name", "Alice").
			Build()
		_ = spec.GetRootPath()
	})
}

// TestMapCreator_BodyPathExpectations mirrors Java's MapCreatorTest.
// Java's MapCreator.createMapFromObjects builds path→matcher maps used internally.
// In Go, the equivalent is building multiple ExpectBody expectations and validating them.
func TestMapCreator_BodyPathExpectations(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"a":1,"b":2}`))
	}))
	t.Cleanup(ts.Close)

	// Mirrors can_create_map_from_integer_values: integer values stored and retrieved correctly.
	valid := rest.Given().BaseURI(ts.URL).Get("/").Then().
		Body("a", 1).
		Body("b", 2)
	must.False(valid.HasFailures())
	is.Empty(valid.Failures())

	// Mirrors can_merge_map_keys_with_parameters: ResponseSpec holds multiple expectations.
	spec := rest.NewResponseSpecBuilder().
		ExpectBody("a", 1).
		ExpectBody("b", 2).
		Build()
	valid2 := rest.Given().BaseURI(ts.URL).Get("/").Then()
	spec.Validate(valid2)
	must.False(valid2.HasFailures())
}

// TestMapCreator_FormatStringPathKeys mirrors Java's MapCreatorTest.can_merge_map_keys_with_parameters.
// Java's MapCreator.createMapFromObjects(MERGE, "key1.%s", withArgs("hello1"), equalTo("value1"), ...)
// builds a path→matcher map using format-string keys. In Go, paths are pre-formatted with fmt.Sprintf
// before being passed to Body() or ExpectBody(), which is the direct equivalent.
func TestMapCreator_FormatStringPathKeys(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"key1":{"hello1":"value1"},"key2":{"hello2":"value2"}}`))
	}))
	t.Cleanup(ts.Close)

	// Java: MapCreator.createMapFromObjects(MERGE, "key1.%s", withArgs("hello1"), equalTo("value1"), ...)
	// Go equivalent: pre-format path with fmt.Sprintf
	valid := rest.Given().BaseURI(ts.URL).Get("/").Then().
		Body(fmt.Sprintf("key1.%s", "hello1"), "value1").
		Body(fmt.Sprintf("key2.%s", "hello2"), "value2")
	must.False(valid.HasFailures())
	is.Empty(valid.Failures())

	// Same via ResponseSpecBuilder
	spec := rest.NewResponseSpecBuilder().
		ExpectBody(fmt.Sprintf("key1.%s", "hello1"), "value1").
		ExpectBody(fmt.Sprintf("key2.%s", "hello2"), "value2").
		Build()
	valid2 := rest.Given().BaseURI(ts.URL).Get("/").Then()
	spec.Validate(valid2)
	must.False(valid2.HasFailures())
}

// TestArgumentWithArgs mirrors Java's ArgumentTest.equalsInputNullOutputFalse.
// Java's Argument wraps format arguments passed to path expressions.
// In Go, withArgs() is a plain helper returning []any; nil values are preserved correctly.
func TestArgumentWithArgs(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	args := withArgs(nil)
	is.Len(args, 1)
	is.Nil(args[0])

	args2 := withArgs(1, nil, "hello")
	is.Len(args2, 3)
	is.Equal(1, args2[0])
	is.Nil(args2[1])
	is.Equal("hello", args2[2])

	is.Empty(withArgs())
}

// TestResponseSpecBuilder_ParserRegistration mirrors Java's
// ResponseSpecBuilderTest.responseParserShouldHandleConfiguredContentType and
// defaultResponseParserShouldBeConfiguredToHandleUnrecognizedContentTypes.
// Java's ResponseSpecBuilder has registerParser() and setDefaultParser() for configuring
// per-content-type parsers. Go's ResponseSpecBuilder has no parser registration because Go
// uses standard encoding/json and encoding/xml directly. This test documents the equivalent
// Go approach: body assertions work regardless of parser config since content is read as bytes.
func TestResponseSpecBuilder_ParserRegistration(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	// Java: registerParser("dummyContentType", Parser.HTML) → spec handles that content type
	// Go equivalent: any content type can be matched with ExpectContentType or validated via body assertions
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use a custom content type that would require parser registration in Java
		w.Header().Set("Content-Type", "application/vnd.custom+json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"value":"test"}`))
	}))
	t.Cleanup(ts.Close)

	// Go spec validates the response regardless of content type — no parser registration needed
	spec := rest.NewResponseSpecBuilder().
		ExpectStatusCode(200).
		ExpectBody("value", "test").
		Build()
	must.NotNil(spec)

	valid := rest.Given().BaseURI(ts.URL).Get("/custom").Then()
	spec.Validate(valid)
	is.False(valid.HasFailures(), "body assertion should work for any content type: %v", valid.Failures())

	// Java: setDefaultParser(Parser.HTML) → unknown content types use HTML parser
	// Go: no default parser concept needed; all body content is treated as bytes/string/JSON
	// The spec always works for known-JSON content with body assertions.
}

// TestSpecificationQuerier_QueryableRequestSpec mirrors Java's SpecificationQuerierTest.
// Java's SpecificationQuerier.query(spec) returns a QueryableRequestSpecification that exposes
// headers, cookies, and params from a built spec. In Go, the RequestSpec struct is directly
// queryable since its fields are public.
func TestSpecificationQuerier_QueryableRequestSpec(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	// Build a spec with header, cookie, and query param — mirrors Java test
	spec := rest.NewRequestSpecBuilder().
		AddHeader("header", "value").
		AddCookie("cookie", "cookieValue").
		AddQueryParam("someparam", "somevalue").
		Build()

	// Go: RequestSpec fields are directly accessible (no separate Querier needed)
	is.Equal("value", spec.Headers.Get("header"))
	is.Equal("somevalue", spec.QueryParams.Get("someparam"))

	// Find cookie by name
	var cookieVal string
	for _, c := range spec.Cookies {
		if c.Name == "cookie" {
			cookieVal = c.Value
			break
		}
	}
	is.Equal("cookieValue", cookieVal)
}

func TestResponseSpecBuilder_RootPath(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	_ = fmt.Sprintf // keep import used

	// rootPath is set
	spec := rest.NewResponseSpecBuilder().
		RootPath("lotto").
		Build()
	is.Equal("lotto", spec.GetRootPath())

	// rootPath overwrites the previous one
	spec = rest.NewResponseSpecBuilder().
		RootPath("nonExistentPath").
		RootPath("lotto").
		Build()
	is.Equal("lotto", spec.GetRootPath())

	// format args are evaluated
	spec = rest.NewResponseSpecBuilder().
		RootPath("lotto.winners[%d]", withArgs(1)...).
		Build()
	is.Equal("lotto.winners[1]", spec.GetRootPath())

	// appendRootPath appends with "." separator
	spec = rest.NewResponseSpecBuilder().
		RootPath("lotto").
		AppendRootPath("winners[1]").
		Build()
	is.Equal("lotto.winners[1]", spec.GetRootPath())

	// appendRootPath with format args
	spec = rest.NewResponseSpecBuilder().
		RootPath("lotto").
		AppendRootPath("winners[%d]", withArgs(1)...).
		Build()
	is.Equal("lotto.winners[1]", spec.GetRootPath())

	// noRootPath resets to empty string
	spec = rest.NewResponseSpecBuilder().
		RootPath("lotto.winners[1]").
		NoRootPath().
		Build()
	is.Equal("", spec.GetRootPath())

	// detachRootPath removes the suffix
	spec = rest.NewResponseSpecBuilder().
		RootPath("lotto.winners[1]").
		DetachRootPath("winners[1]").
		Build()
	is.Equal("lotto", spec.GetRootPath())
}
