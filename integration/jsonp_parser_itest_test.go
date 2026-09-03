package integration_test

// Ported from JSONPGetITest.java and DefaultParserITest.java

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func TestJavaITest_JSONP(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("ReturnsResponseOfJSONP", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			QueryParam("callback", "methodCall").
			QueryParam("firstName", "John").
			QueryParam("lastName", "Doe").
			Get("/jsonp")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		body := resp.AsString()
		is.Equal(`methodCall({"greeting":"Greetings John Doe"});`, body)
	})

	t.Run("JSONPBodyContainsCallback", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			QueryParam("callback", "myFunc").
			QueryParam("firstName", "Alice").
			QueryParam("lastName", "Bob").
			Get("/jsonp")

		must.NoError(resp.Err())
		body := resp.AsString()
		is.Contains(body, "myFunc(")
		is.Contains(body, "Alice Bob")
		is.Contains(body, ");")
	})
}

func TestJavaITest_DefaultParser(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("CustomMimeTypeWithJSONCompatibleBodyIsParseable", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// /customMimeTypeJsonCompatible returns application/vnd.uoml+json with JSON body
		// Go parses JSON regardless of content-type when the body is JSON-shaped
		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/customMimeTypeJsonCompatible")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Contains(resp.ContentType(), "json")

		var result map[string]any
		must.NoError(resp.As(&result))
		is.Equal("It works", result["message"])
	})

	t.Run("NoContentTypeJsonCompatibleBodyIsParseable", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// /noContentTypeJsonCompatible returns no Content-Type header with JSON body
		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/noContentTypeJsonCompatible")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())

		var result map[string]any
		must.NoError(resp.As(&result))
		is.Equal("It works", result["message"])
	})

	t.Run("PlusJsonMimeTypeBodyIsParseable", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/mimeTypeWithPlusJson")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Contains(resp.ContentType(), "+json")

		var result map[string]any
		must.NoError(resp.As(&result))
		is.Equal("It works", result["message"])
	})

	t.Run("JsonBodyAccessibleViaJsonPathEvenWithCustomMimeType", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/customMimeTypeJsonCompatible")

		must.NoError(resp.Err())
		is.Equal("It works", resp.JsonPath().GetString("message"))
	})
}
