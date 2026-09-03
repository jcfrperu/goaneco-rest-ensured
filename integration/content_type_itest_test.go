package integration_test

// Ported from ContentTypeITest.java.
// Java tests EncoderConfig (charset appending), response content-type validation, and accept headers.
// Go maps charset tests to Content-Type header inspection via /returnContentTypeAsBody.

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/matcher"
	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func TestJavaITest_ContentType_ResponseValidation(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("CanValidateResponseContentType", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		// ContentType() assertion against a wrong value must record a failure.
		valid := rest.Given().
			BaseURI(ts.URL).
			Get("/hello").
			Then().
			ContentType(rest.ContentType("something/wrong"))

		is.True(valid.HasFailures(), "wrong content type should record a failure")
	})

	t.Run("CanValidateResponseContentTypeWithHamcrestMatcher", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/hello")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())

		// ContentType contains "application/json".
		valid := rest.Given().
			BaseURI(ts.URL).
			Get("/hello").
			Then().
			ContentTypeContains("application/json")

		is.False(valid.HasFailures())
	})

	t.Run("JSONResponseHasCorrectContentType", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto")

		must.NoError(resp.Err())
		is.Contains(resp.ContentType(), "application/json")
	})

	t.Run("XMLResponseHasCorrectContentType", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			QueryParam("firstName", "John").
			QueryParam("lastName", "Doe").
			Get("/greetXML")

		must.NoError(resp.Err())
		is.Contains(resp.ContentType(), "application/xml")
	})

	t.Run("TextXMLResponseHasCorrectContentType", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/textXML")

		must.NoError(resp.Err())
		is.Contains(resp.ContentType(), "text/xml")
	})

	t.Run("ContentTypeMatchingWithContainsStringMatcher", func(t *testing.T) {
		t.Parallel()

		rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			HeaderMatching("Content-Type", matcher.ContainsString("application/json")).
			AssertAllNoFail(t)
	})

	t.Run("ContentTypeSpecificationMatchesResponseType", func(t *testing.T) {
		t.Parallel()

		rest.Given().
			BaseURI(ts.URL).
			Get("/hello").
			Then().
			ContentTypeContains("application/json").
			AssertAllNoFail(t)
	})

	t.Run("PlusJsonContentTypeIsRecognized", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/mimeTypeWithPlusJson")

		must.NoError(resp.Err())
		is.Contains(resp.ContentType(), "+json")
	})

	t.Run("PlusXmlContentTypeIsRecognized", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/mimeTypeWithPlusXml")

		must.NoError(resp.Err())
		is.Contains(resp.ContentType(), "+xml")
	})
}

func TestJavaITest_ContentType_RequestContentType(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("RequestContentTypeIsEchoedBack", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// /returnContentTypeAsBody echoes the request Content-Type as the response body.
		resp := rest.Given().
			BaseURI(ts.URL).
			ContentType(rest.ContentTypeJSON).
			Body(`{"test":true}`).
			Post("/returnContentTypeAsBody")

		must.NoError(resp.Err())
		is.Contains(resp.AsString(), "application/json")
	})

	t.Run("OctetStreamContentTypeIsPreserved", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			ContentType(rest.ContentType("application/octet-stream")).
			BodyBytes([]byte{42}).
			Post("/returnContentTypeAsBody")

		must.NoError(resp.Err())
		is.Equal("application/octet-stream", resp.AsString())
	})

	t.Run("CustomVendorContentTypeIsPreserved", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			ContentType(rest.ContentType("application/vnd.example-v1+json")).
			Body(`{"vendor":true}`).
			Post("/returnContentTypeAsBody")

		must.NoError(resp.Err())
		is.Equal("application/vnd.example-v1+json", resp.AsString())
	})

	t.Run("ZipContentTypeIsPreserved", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			ContentType(rest.ContentType("application/zip")).
			BodyBytes([]byte{42}).
			Post("/returnContentTypeAsBody")

		must.NoError(resp.Err())
		is.Equal("application/zip", resp.AsString())
	})

	t.Run("URLEncodedContentTypeIsSetForFormParams", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			FormParam("username", "alice").
			Post("/returnContentTypeAsBody")

		must.NoError(resp.Err())
		is.Contains(resp.AsString(), "application/x-www-form-urlencoded")
	})

	t.Run("JSONContentTypeCanBeSetExplicitlyForPost", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			ContentType(rest.ContentTypeJSON).
			BodyObject(map[string]any{"key": "value"}).
			Post("/returnContentTypeAsBody")

		must.NoError(resp.Err())
		is.Contains(resp.AsString(), "application/json")
	})

	t.Run("XMLContentTypeIsPreserved", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			ContentType(rest.ContentTypeXML).
			Body("<root><value>test</value></root>").
			Post("/returnContentTypeAsBody")

		must.NoError(resp.Err())
		is.Contains(resp.AsString(), "application/xml")
	})

	t.Run("TextPlainContentTypeIsPreserved", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			ContentType(rest.ContentTypeText).
			Body("plain text body").
			Post("/returnContentTypeAsBody")

		must.NoError(resp.Err())
		is.Contains(resp.AsString(), "text/plain")
	})
}

func TestJavaITest_ContentType_Charset(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("UTF8CharsetInResponseIsRecognized", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/utf8-body-json")

		must.NoError(resp.Err())
		ct := resp.ContentType()
		is.Contains(strings.ToLower(ct), "charset")
	})

	t.Run("ContentTypeWithCharsetIsReturnedCorrectly", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/i18n")

		must.NoError(resp.Err())
		ct := resp.ContentType()
		is.Contains(ct, "application/json")
		is.Contains(ct, "charset")
	})

	t.Run("ContentTypeWithoutCharsetForJSONIsFine", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto")

		must.NoError(resp.Err())
		is.Contains(resp.ContentType(), "application/json")
		is.Equal(http.StatusOK, resp.StatusCode())
	})

	t.Run("ContentTypeWithCharsetCanBeValidatedWithContains", func(t *testing.T) {
		t.Parallel()

		rest.Given().
			BaseURI(ts.URL).
			Get("/utf8-body-json").
			Then().
			ContentTypeContains("application/json").
			AssertAllNoFail(t)
	})
}

func TestJavaITest_ContentType_AcceptHeader(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("AcceptHeaderIsForwardedCorrectly", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Accept(rest.ContentTypeJSON).
			Get("/headersWithValues")

		must.NoError(resp.Err())
		is.Contains(resp.AsString(), "application/json")
	})

	t.Run("AcceptAnyContentType", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Accept(rest.ContentTypeAny).
			Get("/hello")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
	})

	t.Run("RequestWithJSONAcceptReceivesJSONResponse", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Accept(rest.ContentTypeJSON).
			Get("/lotto")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Contains(resp.ContentType(), "application/json")
	})

	t.Run("MultipleAcceptableContentTypesCanBeSpecified", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Header("Accept", "application/json, text/plain").
			Get("/lotto")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
	})
}

func TestJavaITest_ContentType_SpecBuilder(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("ContentTypeCanBeSetInRequestSpec", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		spec := rest.NewRequestSpecBuilder().
			SetBaseURI(ts.URL).
			SetContentType(rest.ContentTypeJSON).
			Build()

		resp := rest.Given().
			Spec(spec).
			Body(`{"spec":true}`).
			Post("/returnContentTypeAsBody")

		must.NoError(resp.Err())
		is.Contains(resp.AsString(), "application/json")
	})

	t.Run("ContentTypeFromSpecCanBeOverriddenPerRequest", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		spec := rest.NewRequestSpecBuilder().
			SetBaseURI(ts.URL).
			SetContentType(rest.ContentTypeJSON).
			Build()

		resp := rest.Given().
			Spec(spec).
			ContentType(rest.ContentType("application/zip")).
			BodyBytes([]byte{1, 2, 3}).
			Post("/returnContentTypeAsBody")

		must.NoError(resp.Err())
		is.Equal("application/zip", resp.AsString())
	})
}
