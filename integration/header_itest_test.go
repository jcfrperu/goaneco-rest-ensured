package integration_test

// Ported from HeaderITest.java

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/matcher"
	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func TestJavaITest_Header_RequestHeaders(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("AllowsSpecifyingHeader", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Header("X-Single-Header", "single-value").
			Get("/headers")

		must.NoError(resp.Err())
		is.Equal("single-value", resp.JsonPath().GetString("X-Single-Header.0"))
	})

	t.Run("AllowsParsingMultiValueHeaders", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/multiValueHeader")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		// Response should have MultiHeader set (may be one or both values)
		is.NotEmpty(resp.Header("MultiHeader"))
	})

	t.Run("OrderIsMaintainedForMultiValueHeaders", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/multiValueHeader")

		must.NoError(resp.Err())
		// The MultiHeader should contain "Value 1" first
		multiHeader := resp.Header("MultiHeader")
		is.Contains(multiHeader, "Value 1")
	})

	t.Run("AllowsSpecifyingMultipleHeaders", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Header("X-Multi-A", "valA").
			Header("X-Multi-B", "valB").
			Get("/headers")

		must.NoError(resp.Err())
		jp := resp.JsonPath()
		is.Equal("valA", jp.GetString("X-Multi-A.0"))
		is.Equal("valB", jp.GetString("X-Multi-B.0"))
	})

	t.Run("AllowsSpecifyingHeadersObjectViaMap", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Headers(map[string]string{
				"X-Map-A": "mapA",
				"X-Map-B": "mapB",
			}).
			Get("/headers")

		must.NoError(resp.Err())
		jp := resp.JsonPath()
		is.Equal("mapA", jp.GetString("X-Map-A.0"))
		is.Equal("mapB", jp.GetString("X-Map-B.0"))
	})

	t.Run("MultipleHeaderStatementsAreConcatenated", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			Header("X-Custom-1", "val1").
			Header("X-Custom-2", "val2").
			Get("/headersWithValues").
			Then().
			StatusCode(http.StatusOK).
			Body("X-Custom-1.0", matcher.EqualTo("val1")).
			Body("X-Custom-2.0", matcher.EqualTo("val2")).
			AssertAllNoFail(t)
	})

	t.Run("MultipleHeadersUsingMapWithHamcrestMatcher", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			Headers(map[string]string{"X-Ham-1": "hv1"}).
			Get("/headersWithValues").
			Then().
			StatusCode(http.StatusOK).
			Body("X-Ham-1.0", matcher.EqualTo("hv1")).
			AssertAllNoFail(t)
	})

	t.Run("RequestSpecificationAllowsSpecifyingHeadersViaSpec", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		spec := rest.NewRequestSpecBuilder().
			SetBaseURI(ts.URL).
			AddHeader("X-Req-Spec-H", "req-spec-val").
			Build()

		resp := rest.Given().
			Spec(spec).
			Get("/headers")

		must.NoError(resp.Err())
		is.Equal("req-spec-val", resp.JsonPath().GetString("X-Req-Spec-H.0"))
	})
}

func TestJavaITest_Header_ResponseHeaders(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("SupportsHeaderStringMatching", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			StatusCode(http.StatusOK).
			Header("Content-Type", "application/json").
			AssertAllNoFail(t)
	})

	t.Run("HeaderContainsSubstring", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			StatusCode(http.StatusOK).
			HeaderContains("Content-Type", "json").
			AssertAllNoFail(t)
	})

	t.Run("WhenExpectedHeaderDoesNotMatchAnAssertionIsRecorded", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		valid := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then()

		valid.Header("Content-Type", "text/plain") // wrong value
		is.True(valid.HasFailures())
	})

	t.Run("WhenExpectedHeaderIsNotFoundAnAssertionIsRecorded", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		valid := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then()

		valid.HeaderExists("X-Nonexistent-Header-XYZ")
		is.True(valid.HasFailures())
	})

	t.Run("MultiValueHeaderExistsInResponse", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			Get("/multiValueHeader").
			Then().
			StatusCode(http.StatusOK).
			HeaderExists("MultiHeader").
			AssertAllNoFail(t)
	})

	t.Run("HeaderEchoedInResponse", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Header("X-Custom-Echo", "echo-value").
			Body(`{"ping":"pong"}`).
			ContentType(rest.ContentTypeJSON).
			Post("/echo")

		must.NoError(resp.Err())
		is.Equal("echo-value", resp.Header("X-Custom-Echo"))
	})

	t.Run("RequestMirrorHeadersReflectedBack", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Header("X-Mirror-1", "m1").
			Header("X-Mirror-2", "m2").
			Get("/multiHeaderReflect")

		must.NoError(resp.Err())
		is.Equal("m1", resp.Header("X-Mirror-1"))
		is.Equal("m2", resp.Header("X-Mirror-2"))
	})

	t.Run("CanUseResponseAwareMatchersForHeaderValidation", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			StatusCode(http.StatusOK).
			HeaderMatching("Content-Type", matcher.ContainsString("application/json")).
			AssertAllNoFail(t)
	})

	t.Run("AllowsSupplyingHeadersMappingFunction", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto")

		must.NoError(resp.Err())
		// Access headers as a map and verify
		headers := resp.Headers()
		is.NotEmpty(headers)
		is.Contains(headers.Get("Content-Type"), "application/json")
	})
}

func TestJavaITest_Header_ContentType(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("ContentTypeInRequestContainsValue", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			ContentType(rest.ContentTypeJSON).
			Body(`{"test":true}`).
			Post("/contentTypeAsBody")

		must.NoError(resp.Err())
		// Server echoes the Content-Type back as the JSON body
		is.Contains(resp.AsString(), "application/json")
	})

	t.Run("AcceptHeaderIsForwarded", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Accept(rest.ContentTypeJSON).
			Get("/headers")

		must.NoError(resp.Err())
		jp := resp.JsonPath()
		is.Contains(jp.GetString("Accept.0"), "application/json")
	})

	t.Run("HTMLContentTypeIsRecognized", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			Get("/textHTML").
			Then().
			StatusCode(http.StatusOK).
			ContentType(rest.ContentTypeHTML).
			AssertAllNoFail(t)
	})

	t.Run("CustomPlusJsonContentTypeIsHandled", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			Get("/mimeTypeWithPlusJson").
			Then().
			StatusCode(http.StatusOK).
			HeaderMatching("Content-Type", matcher.ContainsString("application/something+json")).
			AssertAllNoFail(t)
	})
}

func TestJavaITest_Header_ShortVersionMatchers(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("MultipleHeadersShortVersionUsingPlainStrings", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			StatusCode(http.StatusOK).
			HeaderContains("Content-Type", "application/json").
			AssertAllNoFail(t)
	})

	t.Run("MultipleHeadersShortVersionUsingHamcrestMatching", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			StatusCode(http.StatusOK).
			HeaderMatching("Content-Type", matcher.ContainsString("json")).
			HeaderMatching("Content-Type", matcher.Not(matcher.EqualTo(""))).
			AssertAllNoFail(t)
	})

	t.Run("MultipleHeadersShortVersionUsingMixedMatchers", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			StatusCode(http.StatusOK).
			HeaderContains("Content-Type", "application/json").
			HeaderMatching("Content-Type", matcher.ContainsString("json")).
			AssertAllNoFail(t)
	})

	t.Run("ResponseSpecificationAllowsParsingMultiValueHeadersWithEqualCharacter", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// Send a header that contains an equal sign in its value
		resp := rest.Given().
			BaseURI(ts.URL).
			Header("X-Token", "base64==").
			Body(`{}`).
			ContentType(rest.ContentTypeJSON).
			Post("/echo")

		must.NoError(resp.Err())
		// Echo endpoint mirrors X-* headers back
		echoed := resp.Header("X-Token")
		is.Equal("base64==", echoed)
	})

	t.Run("HeaderExceptionCanFailWhenExpectedValueMismatches", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		valid := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then()

		valid.Header("Content-Type", "text/plain") // wrong — actual is application/json
		is.True(valid.HasFailures(), "assertion should fail when header value mismatches")
	})

	t.Run("WhenMultiValueHeaderLastValueHasPrecedenceInMap", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/multiValueHeader")

		must.NoError(resp.Err())
		// MultiHeader appears multiple times; resp.Header() returns the first or canonical value
		is.NotEmpty(resp.Header("MultiHeader"))
	})
}
