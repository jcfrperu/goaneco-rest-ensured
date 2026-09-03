package integration_test

// Ported from JSONPostITest.java and additional JSONGetITest.java scenarios

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/matcher"
	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func TestJavaITest_JSONPost(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("SimpleJSONAndHamcrestMatcher", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			QueryParam("firstName", "John").
			QueryParam("lastName", "Doe").
			Post("/greet").
			Then().
			StatusCode(http.StatusOK).
			Body("greeting", matcher.EqualTo("Greetings John Doe")).
			AssertAllNoFail(t)
	})

	t.Run("FormParamsAcceptIntArguments", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			FormParam("list", "item1", "item2").
			Post("/multiValueParam").
			Then().
			StatusCode(http.StatusOK).
			Body("list", matcher.EqualTo("item1,item2")).
			AssertAllNoFail(t)
	})

	t.Run("RequestSpecificationAllowsSpecifyingStringBody", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Body(`{"action":"test","value":42}`).
			ContentType(rest.ContentTypeJSON).
			Post("/echo")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal("test", resp.JsonPath().GetString("action"))
		is.Equal(float64(42), resp.JsonPath().Get("value").Value())
	})

	t.Run("RequestSpecificationAllowsSpecifyingJsonBodyForPost", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			Body(`{"greeting":"Hello"}`).
			ContentType(rest.ContentTypeJSON).
			Post("/echo").
			Then().
			StatusCode(http.StatusOK).
			Body("greeting", matcher.EqualTo("Hello")).
			AssertAllNoFail(t)
	})

	t.Run("RequestSpecificationAllowsSpecifyingBinaryBodyForPost", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		binaryBody := []byte{1, 2, 3, 4, 5}
		resp := rest.Given().
			BaseURI(ts.URL).
			BodyBytes(binaryBody).
			ContentType("application/octet-stream").
			Post("/body")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal(binaryBody, resp.AsBytes())
	})

	t.Run("UriNotFoundWhenPost", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Post("/nonexistent-endpoint-xyz")

		is.Equal(http.StatusNotFound, resp.StatusCode())
	})

	t.Run("BodyWithSingleHamcrestMatching", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			Body(`{"hello":"Hello Scalatra"}`).
			ContentType(rest.ContentTypeJSON).
			Post("/echo").
			Then().
			StatusCode(http.StatusOK).
			Body("hello", matcher.EqualTo("Hello Scalatra")).
			AssertAllNoFail(t)
	})

	t.Run("RequestContentTypeIsSet", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			ContentType(rest.ContentTypeJSON).
			Body(`{"test":true}`).
			Post("/contentTypeAsBody")

		must.NoError(resp.Err())
		is.Contains(resp.AsString(), "application/json")
	})

	t.Run("MultiValueParametersSupportAppendingWhenPassingList", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			FormParam("list", "a", "b", "c").
			Post("/multiValueParam").
			Then().
			StatusCode(http.StatusOK).
			Body("list", matcher.EqualTo("a,b,c")).
			AssertAllNoFail(t)
	})

	t.Run("SupportsReturningPostBody", func(t *testing.T) {
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
		is.NotEmpty(resp.AsString())
	})

	t.Run("SupportsGettingResponseBodyWhenStatusCodeIs401", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// POST to bearer auth endpoint without token
		resp := rest.Given().
			BaseURI(ts.URL).
			Post("/auth/bearer")

		must.NoError(resp.Err())
		is.Equal(http.StatusUnauthorized, resp.StatusCode())
		is.NotEmpty(resp.AsString())
	})

	t.Run("QueryParametersInPostAreHandled", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			QueryParam("firstName", "Jane").
			QueryParam("lastName", "Doe").
			Post("/greet")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal("Greetings Jane Doe", resp.JsonPath().GetString("greeting"))
	})

	t.Run("ByteArrayBodyWithJsonContentType", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		jsonBytes := []byte(`{"byte":"array"}`)
		resp := rest.Given().
			BaseURI(ts.URL).
			BodyBytes(jsonBytes).
			ContentType(rest.ContentTypeJSON).
			Post("/echo")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal("array", resp.JsonPath().GetString("byte"))
	})

	t.Run("FormParamReflectEchosAllParams", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			FormParam("username", "alice").
			FormParam("role", "admin").
			Post("/form").
			Then().
			StatusCode(http.StatusOK).
			Body("username", matcher.EqualTo("alice")).
			Body("role", matcher.EqualTo("admin")).
			AssertAllNoFail(t)
	})
}

func TestJavaITest_JSONPost_Extended(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("BodyHamcrestMatcherWithoutKey", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			Get("/anonymous_list_with_numbers").
			Then().
			StatusCode(http.StatusOK).
			Body("", matcher.HasSize(3)).
			AssertAllNoFail(t)
	})

	t.Run("QueryParametersInPostAreUrlEncoded", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			QueryParam("firstName", "John & Jane").
			QueryParam("lastName", "Doe/Smith").
			Post("/greet")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
	})

	t.Run("CustomJsonCompatibleContentTypeWithBody", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			ContentType("application/vnd.test+json").
			BodyBytes([]byte(`{"vendor":true}`)).
			Put("/echo")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
	})

	t.Run("RequestSpecificationAllowsSpecifyingIntForPost", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			ContentType("text/plain").
			Body("42").
			Post("/body")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal("42", resp.AsString())
	})

	t.Run("RequestSpecificationAllowsSpecifyingBooleanForPost", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			ContentType("text/plain").
			Body("true").
			Post("/body")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal("true", resp.AsString())
	})
}

func TestJavaITest_JSONGet_Extended(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("NewSyntaxStatusCodeInt", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			StatusCode(200).
			AssertAllNoFail(t)
	})

	t.Run("NewSyntaxWithWrongStatusCode", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		valid := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then()

		valid.StatusCode(999) // intentionally wrong
		is.True(valid.HasFailures())
	})

	t.Run("NewSyntaxWithCorrectStatusLineContains", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			StatusLineContains("200").
			AssertAllNoFail(t)
	})

	t.Run("JsonHamcrestEqualBody", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			Get("/amount").
			Then().
			StatusCode(http.StatusOK).
			Body("amount", matcher.EqualTo(250.0)).
			AssertAllNoFail(t)
	})

	t.Run("ContentTypeSpecification", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			StatusCode(http.StatusOK).
			ContentType(rest.ContentTypeJSON).
			AssertAllNoFail(t)
	})

	t.Run("ContentTypeSpecificationWithHamcrestMatcher", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			StatusCode(http.StatusOK).
			HeaderMatching("Content-Type", matcher.ContainsString("application/json")).
			AssertAllNoFail(t)
	})

	t.Run("SupportsGettingListSize", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			Get("/anonymous_list_with_numbers").
			Then().
			StatusCode(http.StatusOK).
			Body("", matcher.HasSize(3)).
			AssertAllNoFail(t)
	})

	t.Run("FindAllBooksWithPriceGreaterThanTen", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/jsonStore")

		must.NoError(resp.Err())

		// jsonStore has 4 books: 8.95, 12.99, 8.99, 22.99
		// Books with price > 10: 12.99, 22.99
		var store StoreResponse
		must.NoError(resp.As(&store))
		var expensiveCount int
		for _, b := range store.Store.Book {
			if b.Price > 10.0 {
				expensiveCount++
			}
		}
		is.Equal(2, expensiveCount)
	})

	t.Run("RestAssuredSupportsFullyQualifiedURI", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			Get(ts.URL + "/lotto")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
	})

	t.Run("MultipleBodyHamcrestMatchersShortVersion", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			Body("lotto.lottoId", matcher.EqualTo(5)).
			Body("lotto.winning-numbers", matcher.HasItems(2, 45)).
			AssertAllNoFail(t)
	})

	t.Run("RequestSpecificationAllowsSpecifyingCookie", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Cookie("req_cookie", "cookie_value").
			Get("/cookies")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal("cookie_value", resp.JsonPath().GetString("req_cookie"))
	})

	t.Run("JsonPathWithAtSignKey", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			Get("/jsonWithAtSign").
			Then().
			StatusCode(http.StatusOK).
			Body("body.@id", matcher.EqualTo(10)).
			AssertAllNoFail(t)
	})

	t.Run("ParameterSupportWithStandardHashMap", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			QueryParam("firstName", "John").
			QueryParam("lastName", "Doe").
			Get("/greet")

		must.NoError(resp.Err())
		is.Equal("Greetings John Doe", resp.JsonPath().GetString("greeting"))
	})

	t.Run("GetLastElementInList", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/jsonList")

		must.NoError(resp.Err())
		jp := resp.JsonPath()
		// jsonList is an array of objects; access last element by index
		lastName := jp.GetString("1.name")
		is.Equal("Sven", lastName)
	})

	t.Run("SupportsGettingMap", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/numbers")

		must.NoError(resp.Err())
		jp := resp.JsonPath()
		is.NotEmpty(jp.GetString("values.pi"))
	})

	t.Run("MixingSingleAndMultipleParametersConcatenatesThem", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			QueryParam("firstName", "John").
			QueryParam("lastName", "Doe").
			Get("/greet").
			Then().
			StatusCode(http.StatusOK).
			Body("greeting", matcher.ContainsString("John")).
			Body("greeting", matcher.ContainsString("Doe")).
			AssertAllNoFail(t)
	})

	t.Run("JSONP_ReturnsCallbackWrappedResponse", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			QueryParam("callback", "myCallback").
			QueryParam("firstName", "John").
			QueryParam("lastName", "Doe").
			Get("/jsonp")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Contains(resp.AsString(), "myCallback")
		is.Contains(resp.AsString(), "Greetings John Doe")
	})

	t.Run("QueryParamWithBooleanWorks", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			QueryParam("firstName", "true").
			QueryParam("lastName", "false").
			Get("/greet")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Contains(resp.AsString(), "true")
	})

	t.Run("FormParamTreatedAsQueryParamForGetRequest", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// GET with form params: server reads from query string
		resp := rest.Given().
			BaseURI(ts.URL).
			QueryParam("firstName", "GetParam").
			QueryParam("lastName", "Test").
			Get("/greet")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Contains(resp.AsString(), "GetParam")
	})

	t.Run("SupportsGettingListItemByIndex", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/jsonList")

		must.NoError(resp.Err())
		jp := resp.JsonPath()
		// Access first element by index
		is.Equal("Anders", jp.GetString("0.name"))
	})

	t.Run("SupportsGettingSingleFloat", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/amount")

		must.NoError(resp.Err())
		val := resp.JsonPath().Get("amount").Value()
		f, ok := val.(float64)
		is.True(ok, "amount should be float64")
		is.InDelta(250.0, f, 0.01)
	})

	t.Run("SupportsParsingJsonWhenContentTypeEndsWithPlusJson", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/mimeTypeWithPlusJson")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Contains(resp.ContentType(), "+json")
		is.Equal("It works", resp.JsonPath().GetString("message"))
	})

	t.Run("ContentTypeButNoBody", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/contentTypeButNoBody")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Contains(resp.ContentType(), "application/json")
		is.Empty(resp.AsString())
	})

	t.Run("NewSyntaxWithWrongStatusLine", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		valid := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then()

		valid.StatusLineContains("999 Nonexistent Status") // intentionally wrong
		is.True(valid.HasFailures(), "wrong status line should produce failure")
	})

	t.Run("SupportsGettingResponseBodyWhenStatusCodeIs401", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/auth/basic")

		must.NoError(resp.Err())
		is.Equal(http.StatusUnauthorized, resp.StatusCode())
		is.NotEmpty(resp.AsString())
	})

	t.Run("CanParseJsonPathWithAtSignKeyInGet", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			Get("/jsonWithAtSign").
			Then().
			StatusCode(http.StatusOK).
			Body("body.@id", matcher.EqualTo(10)).
			AssertAllNoFail(t)
	})

	t.Run("WinnersArrayContainsExpectedIds", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			StatusCode(http.StatusOK).
			Body("lotto.winners.winnerId", matcher.HasItems(23, 54)).
			AssertAllNoFail(t)
	})
}
