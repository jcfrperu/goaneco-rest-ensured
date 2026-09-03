package rest_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/jsonschema"
	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func TestValidatableResponseAssertions(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Server", "GoRestEnsured")
		http.SetCookie(w, &http.Cookie{Name: "session", Value: "sess-99"})
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"status": "success",
			"code": 200,
			"user": {
				"id": 10,
				"name": "Eve",
				"roles": ["admin", "editor"]
			}
		}`))
	}))
	t.Cleanup(ts.Close)

	// Fluent assertions matching Java Given().When().Then()
	rest.Given().
		BaseURI(ts.URL).
		When().
		Get("/user").
		Then().
		AssertWith(t).
		AssertThat().
		StatusCode(200).
		StatusCodeBetween(200, 299).
		Header("X-Server", "GoRestEnsured").
		HeaderExists("X-Server").
		Cookie("session", "sess-99").
		CookieExists("session").
		ContentType(rest.ContentTypeJSON).
		TimeLessThan(2*time.Second).
		Body("status", "success").
		Body("code", 200).
		Body("user.id", 10).
		Body("user.name", "Eve").
		Body("user.roles.0", "admin").
		BodyContainsElement("user.roles", "editor")

	// Root path scoping matching Java rootPath()
	rest.Given().
		BaseURI(ts.URL).
		When().
		Get("/user").
		Then().
		AssertWith(t).
		RootPath("user").
		Body("id", 10).
		Body("name", "Eve").
		AppendRootPath("roles").
		Body("0", "admin").
		NoRootPath().
		Body("status", "success")

	// Verify AssertAll on successful assertions
	valid := rest.Given().
		BaseURI(ts.URL).
		Get("/user").
		Then().
		StatusCode(200)

	is.NoError(valid.AssertAll())
}

func TestValidatableResponseFailures(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"count": 42, "flag": true, "score": 9.5}`))
	}))
	t.Cleanup(ts.Close)

	val := rest.Given().
		BaseURI(ts.URL).
		Get("/fail-check").
		Then().
		StatusCode(404).
		StatusCodeBetween(400, 499).
		StatusLine("HTTP/1.1 500 Internal Server Error").
		StatusLineContains("500").
		Header("Missing-Header", "Val").
		HeaderContains("Missing-Header", "Val").
		HeaderExists("Missing-Header").
		Cookie("Missing-Cookie", "Val").
		CookieExists("Missing-Cookie").
		ContentType(rest.ContentTypeXML).
		ContentTypeContains("xml").
		TimeLessThan(1*time.Nanosecond).
		Body("count", 100).
		Body("flag", false).
		Body("score", 1.23).
		Body("nonexistent", "val").
		BodyContains("missing_string").
		BodyEquals(`{"other":true}`).
		BodyJSONEquals(`{"other":true}`).
		BodyContainsElement("count", 1).
		BodyLength("count", 10).
		BodyLength("missing", 5)

	is.True(val.HasFailures())
	is.NotEmpty(val.Failures())
	is.Error(val.AssertAll())

	// Edge case: nil response
	var nilResp *rest.Response
	nilVal := nilResp.Then().StatusCode(200)
	is.True(nilVal.HasFailures())

	// Edge case: response with error
	errResp := &rest.Response{}
	is.Error(errResp.As(nil))
	errVal := rest.Given().BaseURI("http://invalid.localhost:99999").Get("/err").Then().StatusCode(200)
	is.True(errVal.HasFailures())
}

func TestExtractableResponse(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	type UserData struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Custom", "Val123")
		http.SetCookie(w, &http.Cookie{Name: "c1", Value: "cookie_val"})
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"name": "Grace",
			"age": 30,
			"scores": [90, 85, 95]
		}`))
	}))
	t.Cleanup(ts.Close)

	extract := rest.Given().
		BaseURI(ts.URL).
		When().
		Get("/extract").
		Then().
		StatusCode(200).
		Extract()

	// Extract methods
	is.Equal(http.StatusOK, extract.StatusCode())
	is.Equal("Val123", extract.Header("X-Custom"))
	is.Equal("cookie_val", extract.Cookie("c1"))
	is.Equal("application/json", extract.ContentType())
	is.Contains(extract.AsString(), "Grace")
	is.NotEmpty(extract.AsBytes())

	// Path and JsonPath extraction
	is.Equal("Grace", extract.Path("name").String())
	is.Equal(30, int(extract.Path("age").Int()))
	is.Equal("Grace", extract.JsonPath().GetString("name"))
	is.Equal(30, extract.JsonPath().GetInt("age"))

	// Object deserialization via As
	var user UserData
	err := extract.As(&user)
	must.NoError(err)
	is.Equal("Grace", user.Name)
	is.Equal(30, user.Age)

	// Generic deserialization via AsObject
	userTyped, err := rest.AsObject[UserData](extract.Response())
	must.NoError(err)
	is.Equal("Grace", userTyped.Name)
	is.Equal(30, userTyped.Age)

	// Response accessor
	must.NotNil(extract.Response())
}

func TestBodyJSONEqualsAndStringAssertions(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"alpha": 1, "beta": "two"}`))
	}))
	t.Cleanup(ts.Close)

	valid := rest.Given().
		BaseURI(ts.URL).
		Get("/json-eq").
		Then().
		AssertWith(t).
		BodyJSONEquals(`{"beta": "two", "alpha": 1}`).
		BodyContains("alpha").
		BodyEquals(`{"alpha": 1, "beta": "two"}`)

	must.False(valid.HasFailures())
	is.Empty(valid.Failures())
}

func TestAdditionalValidations(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"pi": 3.14159,
			"active": true,
			"greeting": "hello world",
			"nil_field": null,
			"numbers": [1, 2, 3]
		}`))
	}))
	t.Cleanup(ts.Close)

	resp := rest.Given().
		BaseURI(ts.URL).
		Get("/types")

	valid := resp.Then().
		AssertWith(t).
		StatusLine(fmt.Sprintf("%s %s", "HTTP/1.1", "200 OK")).
		StatusLineContains("200 OK").
		HeaderContains("Content-Type", "json").
		ContentTypeContains("json").
		Body("pi", 3.14159).
		Body("pi", float32(3.14159)).
		Body("active", true).
		Body("greeting", "hello world").
		Body("nil_field", nil).
		BodyLength("greeting", 11).
		BodyLength("numbers", 3).
		BodyContainsElement("numbers", 2)

	must.False(valid.HasFailures())
	is.Empty(valid.Failures())

	// Response getters
	must.NotNil(resp.RawResponse())
	must.NotNil(resp.RawRequest())
	is.Equal(ts.URL+"/types", resp.RawRequest().URL.String())
}

// TestCookieMatcherFailureMessages mirrors Java's CookieMatcherMessagesTest.
// Java's CookieMatcher produces Hamcrest-formatted error messages; Go produces
// simpler format strings that still include the cookie name and mismatched values.
func TestCookieMatcherFailureMessages(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "DEVICE_ID", Value: "123", Path: "/"})
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(ts.Close)

	// shouldPrintValidErrorMessageForStandardMatchers equivalent:
	// DEVICE_ID = "123", assert it equals "X" → failure message names the cookie and value
	valid := rest.Given().
		BaseURI(ts.URL).
		Get("/").
		Then().
		Cookie("DEVICE_ID", "X")

	is.True(valid.HasFailures())
	failures := valid.Failures()
	is.NotEmpty(failures)
	is.Contains(failures[0], "DEVICE_ID")
	is.Contains(failures[0], "123")

	// CookieExists failure: absent cookie names the expected cookie
	valid2 := rest.Given().
		BaseURI(ts.URL).
		Get("/").
		Then().
		CookieExists("MISSING_COOKIE")

	is.True(valid2.HasFailures())
	is.Contains(valid2.Failures()[0], "MISSING_COOKIE")
}

// TestMatcherErrorMessages mirrors Java's MatcherErrorMessageBuilderTests.
// Java's MatcherErrorMessageBuilder formats "Expected X to equal <Y> but was <Z>.\n".
// Go's recordFailure uses simpler fmt.Sprintf patterns; the tests verify the key
// values (expected and actual) appear in the failure message.
func TestMatcherErrorMessages(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(ts.Close)

	// shouldPrintNicelyWithDefaultEqualMatcher equivalent:
	// status code expected 500 but got 200 → both values appear in the message
	valid := rest.Given().
		BaseURI(ts.URL).
		Get("/").
		Then().
		StatusCode(500)

	is.True(valid.HasFailures())
	is.NotEmpty(valid.Failures())
	is.Contains(valid.Failures()[0], "200")
	is.Contains(valid.Failures()[0], "500")

	// shouldIncludeMismatchedDescription equivalent:
	// header expected "expected-value" but not present → message names header and expected value
	valid2 := rest.Given().
		BaseURI(ts.URL).
		Get("/").
		Then().
		Header("X-Missing", "expected-value")

	is.True(valid2.HasFailures())
	is.Contains(valid2.Failures()[0], "X-Missing")
	is.Contains(valid2.Failures()[0], "expected-value")

	// shouldPrintNicelyWithoutMismatchDescription equivalent:
	// body path absent → message references the missing path
	valid3 := rest.Given().
		BaseURI(ts.URL).
		Get("/").
		Then().
		Body("nonexistent.path", "value")

	is.True(valid3.HasFailures())
	is.Contains(valid3.Failures()[0], "nonexistent.path")
}

// TestSerializationCandidateBehavior mirrors Java's SerializationSupportTest.
// Java checks SerializationSupport.isSerializableCandidate() per type.
// In Go, serialization is governed by json.Marshal semantics via BodyObject().
func TestSerializationCandidateBehavior(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	var capturedBody string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		capturedBody = string(b)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(ts.Close)

	type Pojo struct {
		Name string `json:"name"`
	}

	// plainObjectIsASerializationCandidate: struct → serialized to JSON
	must.NoError(rest.Given().BaseURI(ts.URL).BodyObject(Pojo{Name: "test"}).Post("/").Err())
	is.JSONEq(`{"name":"test"}`, capturedBody)

	// Map → serializable
	must.NoError(rest.Given().BaseURI(ts.URL).BodyObject(map[string]any{"a": 1}).Post("/").Err())
	is.JSONEq(`{"a":1}`, capturedBody)

	// nullIsNotASerializationCandidate in Java; Go serializes nil to "null"
	must.NoError(rest.Given().BaseURI(ts.URL).BodyObject(nil).Post("/").Err())
	is.Equal("null", capturedBody)

	// Non-marshalable type (channel) → error (plainEnumConstantIsNotASerializationCandidate analogue)
	is.Error(rest.Given().BaseURI(ts.URL).BodyObject(make(chan int)).Post("/").Err())
}

// TestSerializationCandidates_EnumAndNil mirrors Java's SerializationSupportTest additional cases.
// Java's SerializationSupport.isSerializableCandidate() returns false for enum constants and null.
// In Go there are no enums; the closest analogues are typed integer constants and types with
// custom String() implementations. json.Marshal succeeds for both (returns the underlying value).
// This test documents the Go behavior for completeness.
func TestSerializationCandidates_EnumAndNil(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	var capturedBody string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		capturedBody = string(b)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(ts.Close)

	// plainEnumConstantIsNotASerializationCandidate (Java): Go's nearest analogue is a typed int
	// constant. Unlike Java enums, json.Marshal serializes it as its underlying integer value.
	type Status int
	const StatusActive Status = 1
	must.NoError(rest.Given().BaseURI(ts.URL).BodyObject(StatusActive).Post("/").Err())
	is.Equal("1", capturedBody) // serialized as integer, not skipped

	// enumConstantWithBodyIsNotASerializationCandidate (Java): Go's nearest analogue is a struct
	// that implements Stringer. json.Marshal serializes its fields, not the String() output.
	type StatusWithString struct {
		Code int `json:"code"`
	}
	must.NoError(rest.Given().BaseURI(ts.URL).BodyObject(StatusWithString{Code: 2}).Post("/").Err())
	is.JSONEq(`{"code":2}`, capturedBody)

	// nullIsNotASerializationCandidate (Java): In Go, json.Marshal(nil) produces "null".
	// This is a Go-specific behavior difference — nil is serialized, not skipped.
	must.NoError(rest.Given().BaseURI(ts.URL).BodyObject(nil).Post("/").Err())
	is.Equal("null", capturedBody)
}

// TestCookieMatcherFailureMessages_CustomMatcher mirrors Java's
// CookieMatcherMessagesTest.shouldPrintValidErrorMessageForCustomMatcher.
// Java uses a TypeSafeDiagnosingMatcher that appends a custom mismatch description.
// Go's CookieMatching uses the CookieMatching interface (Matches(string)bool + Describe()string).
func TestCookieMatcherFailureMessages_CustomMatcher(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "DEVICE_ID", Value: "123", Path: "/"})
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(ts.Close)

	// Custom matcher that checks the cookie value contains "X"
	containsX := &containsXCookieMatcher{}

	valid := rest.Given().
		BaseURI(ts.URL).
		Get("/").
		Then().
		CookieMatching("DEVICE_ID", containsX)

	is.True(valid.HasFailures())
	failures := valid.Failures()
	is.NotEmpty(failures)
	// Failure message should reference the cookie name and the custom description
	is.Contains(failures[0], "DEVICE_ID")
	is.Contains(failures[0], "containing X")
}

// containsXCookieMatcher implements the CookieMatching interface used by ValidatableResponse.CookieMatching.
type containsXCookieMatcher struct{}

func (m *containsXCookieMatcher) Matches(val string) bool {
	return strings.Contains(val, "X")
}

func (m *containsXCookieMatcher) Describe() string {
	return "containing X"
}

func TestValidatableResponse_JsonSchemaAssertions(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	schemaJSON := `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"properties": {
			"id": {"type": "integer"},
			"name": {"type": "string"}
		},
		"required": ["id", "name"]
	}`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id": 1, "name": "Valid Item"}`))
	}))
	t.Cleanup(ts.Close)

	// 1. Matches schema string
	valid := rest.Given().
		BaseURI(ts.URL).
		Get("/item").
		Then().
		StatusCode(http.StatusOK).
		BodyMatchesSchema(schemaJSON)

	is.True(valid.IsValid())
	is.Empty(valid.Failures())

	// 2. Matches schema file
	tempDir := t.TempDir()
	schemaFilePath := filepath.Join(tempDir, "schema.json")
	err := os.WriteFile(schemaFilePath, []byte(schemaJSON), 0600)
	must.NoError(err)

	validFile := rest.Given().
		BaseURI(ts.URL).
		Get("/item").
		Then().
		BodyMatchesSchemaFile(schemaFilePath)

	is.True(validFile.IsValid())
	is.Empty(validFile.Failures())

	// 3. Matches custom validator instance
	customValidator := jsonschema.MatchesJsonSchema(schemaJSON)
	validCustom := rest.Given().
		BaseURI(ts.URL).
		Get("/item").
		Then().
		BodyMatchesSchemaValidator(customValidator)

	is.True(validCustom.IsValid())

	// 4. Fails when schema doesn't match
	invalidSchema := `{
		"type": "object",
		"properties": {
			"id": {"type": "string"}
		},
		"required": ["id"]
	}`
	invalid := rest.Given().
		BaseURI(ts.URL).
		Get("/item").
		Then().
		BodyMatchesSchema(invalidSchema)

	is.False(invalid.IsValid())
	is.NotEmpty(invalid.Failures())

	// 5. Nil validator records failure
	nilVal := rest.Given().
		BaseURI(ts.URL).
		Get("/item").
		Then().
		BodyMatchesSchemaValidator(nil)

	is.False(nilVal.IsValid())
}
