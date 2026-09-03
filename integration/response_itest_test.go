package integration_test

// Ported from ResponseITest.java

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func TestJavaITest_Response_BodyAccessors(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("ResponseBodyAsString", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto")

		must.NoError(resp.Err())
		body := resp.AsString()
		is.NotEmpty(body)
		is.Contains(body, "lotto")
		is.Contains(body, "lottoId")
	})

	t.Run("ResponseBodyAsBytes", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto")

		must.NoError(resp.Err())
		body := resp.AsBytes()
		is.NotEmpty(body)
		is.Greater(len(body), 0)
	})

	t.Run("ResponseBodyAsStringForHTML", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/textHTML")

		must.NoError(resp.Err())
		body := resp.AsString()
		is.Contains(body, "<!DOCTYPE html>")
		is.Contains(body, "paragraph 1")
		is.Contains(body, "paragraph 2")
	})

	t.Run("EmptyBodyIsReturnable", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/emptyBody")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal("", resp.AsString())
		is.Empty(resp.AsBytes())
	})

	t.Run("ResponseBodyAsStringForXML", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			QueryParam("firstName", "John").
			QueryParam("lastName", "Doe").
			Get("/greetXML")

		must.NoError(resp.Err())
		body := resp.AsString()
		is.Contains(body, "<?xml")
		is.Contains(body, "John")
		is.Contains(body, "Doe")
	})
}

func TestJavaITest_Response_StatusCodeAndLine(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("StatusCodeIsAccessible", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
	})

	t.Run("StatusLineContainsCodeAndText", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto")

		must.NoError(resp.Err())
		statusLine := resp.StatusLine()
		is.Contains(statusLine, "200")
		is.Contains(statusLine, "OK")
	})

	t.Run("StatusLineForErrorResponse", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/409")

		must.NoError(resp.Err())
		is.Equal(http.StatusConflict, resp.StatusCode())
		is.Contains(resp.StatusLine(), "409")
	})

	t.Run("BodyIsReadableEvenFor4xxResponses", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/409")

		must.NoError(resp.Err())
		is.Equal(http.StatusConflict, resp.StatusCode())
		is.Equal("ERROR", resp.AsString())
	})

	t.Run("StatusCode500WithBodyContent", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/statusCode500")

		must.NoError(resp.Err())
		is.Equal(http.StatusInternalServerError, resp.StatusCode())
		is.Contains(resp.AsString(), "expected error occurred")
	})
}

func TestJavaITest_Response_HeadersAccess(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("HeadersAreAccessible", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto")

		must.NoError(resp.Err())
		headers := resp.Headers()
		is.NotEmpty(headers)
		is.NotEmpty(headers.Get("Content-Type"))
	})

	t.Run("SingleHeaderIsAccessible", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto")

		must.NoError(resp.Err())
		ct := resp.Header("Content-Type")
		is.Contains(ct, "application/json")
	})

	t.Run("MissingHeaderReturnsEmptyString", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto")

		must.NoError(resp.Err())
		is.Empty(resp.Header("X-Nonexistent-Header-XYZ"))
	})

	t.Run("EchoedHeaderIsReadable", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Header("X-Echo-Me", "echo-value").
			Body(`{}`).
			ContentType(rest.ContentTypeJSON).
			Post("/echo")

		must.NoError(resp.Err())
		is.Equal("echo-value", resp.Header("X-Echo-Me"))
	})
}

func TestJavaITest_Response_CookiesAccess(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("CookiesAreAccessible", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/setCookies")

		must.NoError(resp.Err())
		cookies := resp.Cookies()
		is.Len(cookies, 3)
	})

	t.Run("SingleCookieValueIsAccessible", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/setCookies")

		must.NoError(resp.Err())
		is.Equal("value1", resp.Cookie("key1"))
		is.Equal("value2", resp.Cookie("key2"))
		is.Equal("value3", resp.Cookie("key3"))
	})

	t.Run("MissingCookieReturnsEmptyString", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto")

		must.NoError(resp.Err())
		is.Empty(resp.Cookie("nonexistent_cookie"))
	})
}

func TestJavaITest_Response_ContentType(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("ContentTypeIsAccessibleAsString", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto")

		must.NoError(resp.Err())
		ct := resp.ContentType()
		is.Contains(ct, "application/json")
	})

	t.Run("XMLContentTypeIsAccessible", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			QueryParam("firstName", "John").
			QueryParam("lastName", "Doe").
			Get("/greetXML")

		must.NoError(resp.Err())
		ct := resp.ContentType()
		is.Contains(ct, "application/xml")
	})

	t.Run("HTMLContentTypeIsAccessible", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/textHTML")

		must.NoError(resp.Err())
		ct := resp.ContentType()
		is.Contains(ct, "text/html")
	})
}

func TestJavaITest_Response_JsonAndXmlPath(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("JsonPathExtractsString", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			QueryParam("firstName", "Alice").
			QueryParam("lastName", "Wonder").
			Get("/greet")

		must.NoError(resp.Err())
		is.Equal("Greetings Alice Wonder", resp.JsonPath().GetString("greeting"))
	})

	t.Run("JsonPathExtractsNumber", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto")

		must.NoError(resp.Err())
		is.Equal(float64(5), resp.JsonPath().Get("lotto.lottoId").Value())
	})

	t.Run("JsonPathExtractsList", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto")

		must.NoError(resp.Err())
		winning := resp.JsonPath().GetIntList("lotto.winning-numbers")
		is.Contains(winning, 2)
		is.Contains(winning, 45)
	})

	t.Run("XmlPathExtractsString", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			QueryParam("firstName", "Bob").
			QueryParam("lastName", "Smith").
			Get("/greetXML")

		must.NoError(resp.Err())
		xp, err := resp.XmlPath()
		must.NoError(err)
		is.Equal("Bob", xp.GetString("//greeting/firstName"))
		is.Equal("Smith", xp.GetString("//greeting/lastName"))
	})
}

func TestJavaITest_Response_TimeMeasurement(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("ResponseTimeIsGreaterThanDelay", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/delay/50")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.GreaterOrEqual(resp.Time(), 50*time.Millisecond)
	})

	t.Run("ResponseTimeIsMeasuredForFastRequests", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto")

		must.NoError(resp.Err())
		// Any fast request should still have a positive elapsed time
		is.GreaterOrEqual(resp.Time(), time.Duration(0))
		is.Less(resp.Time(), 2*time.Second)
	})
}

func TestJavaITest_Response_HTTPMethods(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("PostCanReturnBodyAsString", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Body(`{"returned":true}`).
			ContentType(rest.ContentTypeJSON).
			Post("/echo")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		body := resp.AsString()
		is.NotEmpty(body)
		is.Contains(body, "returned")
	})

	t.Run("PutCanReturnBodyAsString", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Body(`{"updated":true}`).
			ContentType(rest.ContentTypeJSON).
			Put("/echo")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.NotEmpty(resp.AsString())
	})

	t.Run("DeleteCanReturnBodyAsString", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Body(`{"deleted":true}`).
			ContentType(rest.ContentTypeJSON).
			Delete("/echo")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.NotEmpty(resp.AsString())
	})

	t.Run("PatchCanReturnBodyAsString", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Body(`{"patched":true}`).
			ContentType(rest.ContentTypeJSON).
			Patch("/echo")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.NotEmpty(resp.AsString())
	})

	t.Run("CanGetAsStringMultipleTimes", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto")

		must.NoError(resp.Err())
		body1 := resp.AsString()
		body2 := resp.AsString()
		body3 := resp.AsString()
		is.Equal(body1, body2)
		is.Equal(body2, body3)
		is.Contains(body1, "lotto")
	})

	t.Run("CanGetAsByteArrayMultipleTimes", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto")

		must.NoError(resp.Err())
		bytes1 := resp.AsBytes()
		bytes2 := resp.AsBytes()
		is.Equal(bytes1, bytes2)
		is.NotEmpty(bytes1)
	})

	t.Run("ResponseCharsetIsAccessibleViaContentType", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto")

		must.NoError(resp.Err())
		ct := resp.ContentType()
		is.NotEmpty(ct)
	})
}

func TestJavaITest_Response_ExtractableResponse(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("ExtractStatusCodeFromValidatable", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		statusCode := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			AssertWith(t).
			StatusCode(http.StatusOK).
			Extract().
			StatusCode()

		is.Equal(http.StatusOK, statusCode)
	})

	t.Run("ExtractHeaderFromValidatable", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		ct := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			AssertWith(t).
			StatusCode(http.StatusOK).
			Extract().
			Header("Content-Type")

		is.Contains(ct, "application/json")
	})

	t.Run("ExtractBodyAsStringFromValidatable", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		body := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			AssertWith(t).
			StatusCode(http.StatusOK).
			Extract().
			AsString()

		is.Contains(body, "lotto")
		is.Contains(body, "lottoId")
	})

	t.Run("ExtractResponseObjectAndInspectDirectly", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			AssertWith(t).
			StatusCode(http.StatusOK).
			Extract().
			Response()

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Contains(resp.ContentType(), "application/json")
		is.NotEmpty(resp.AsString())
	})

	t.Run("ExtractBodyPath", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		lottoID := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			AssertWith(t).
			StatusCode(http.StatusOK).
			Extract().
			JsonPath().Get("lotto.lottoId").Value()

		is.Equal(float64(5), lottoID)
	})
}
