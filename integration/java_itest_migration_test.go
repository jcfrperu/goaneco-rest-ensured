package integration_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/matcher"
	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

// TestJavaITest_JSONGet tests scenarios ported from Java's JSONGetITest.java
func TestJavaITest_JSONGet(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("LottoEndpoint", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			StatusCode(http.StatusOK).
			ContentType(rest.ContentTypeJSON).
			Body("lotto.lottoId", matcher.EqualTo(5)).
			Body("lotto.winning-numbers", matcher.HasItems(2, 45, 34)).
			Body("lotto.winners.winnerId", matcher.Contains(23, 54)).
			Body("lotto.winners.0.numbers", matcher.HasSize(6)).
			AssertAllNoFail(t)
	})

	t.Run("JsonStoreEndpoint", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			Get("/jsonStore").
			Then().
			StatusCode(http.StatusOK).
			Body("store.bicycle.color", matcher.EqualTo("red")).
			Body("store.bicycle.price", matcher.CloseToNum(19.95, 0.01)).
			Body("store.book.category", matcher.HasItemValue("reference")).
			Body("store.book.author", matcher.HasItems("Nigel Rees", "Evelyn Waugh", "J. R. R. Tolkien")).
			Body("store.book.price", matcher.ContainsInAnyOrder(8.95, 12.99, 8.99, 22.99)).
			AssertAllNoFail(t)
	})

	t.Run("NumbersAndAmounts", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			Get("/numbers").
			Then().
			StatusCode(http.StatusOK).
			Body("values.pi", matcher.CloseToNum(3.14, 0.001)).
			Body("values.answer", matcher.EqualTo(42)).
			AssertAllNoFail(t)

		rest.Given().
			BaseURI(ts.URL).
			Get("/amount").
			Then().
			StatusCode(http.StatusOK).
			Body("amount", matcher.EqualTo(250.0)).
			AssertAllNoFail(t)
	})

	t.Run("AnonymousListWithNumbers", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			Get("/anonymous_list_with_numbers").
			Then().
			StatusCode(http.StatusOK).
			Body("", matcher.HasSize(3)).
			Body("", matcher.HasItems(100, 50, 31)).
			AssertAllNoFail(t)
	})

	t.Run("JsonListPersons", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			Get("/jsonList").
			Then().
			StatusCode(http.StatusOK).
			Body("0.name", matcher.EqualTo("Anders")).
			Body("0.address.zipCode", matcher.EqualTo("12345")).
			Body("1.name", matcher.EqualTo("Sven")).
			Body("1.address.street", matcher.EqualTo("Svensgatan")).
			AssertAllNoFail(t)
	})

	t.Run("JsonWithAtSign", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			Get("/jsonWithAtSign").
			Then().
			StatusCode(http.StatusOK).
			Body("body.@id", matcher.EqualTo(10)).
			Body("body.content", matcher.ContainsString("content")).
			AssertAllNoFail(t)
	})

	t.Run("GreetingQueryParameters", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			QueryParam("firstName", "John").
			QueryParam("lastName", "Doe").
			Get("/greet").
			Then().
			StatusCode(http.StatusOK).
			Body("greeting", matcher.EqualTo("Greetings John Doe")).
			AssertAllNoFail(t)
	})

	t.Run("NonAsciiAndI18n", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			Get("/i18n").
			Then().
			StatusCode(http.StatusOK).
			Body("ön", matcher.EqualTo("Är ån")).
			AssertAllNoFail(t)

		rest.Given().
			BaseURI(ts.URL).
			Get("/utf8-body-json").
			Then().
			StatusCode(http.StatusOK).
			Body("value", matcher.ContainsString("啊")).
			AssertAllNoFail(t)
	})
}

// TestJavaITest_XMLAndXPath tests scenarios ported from Java's XMLGetITest.java and XPathITest.java
func TestJavaITest_XMLAndXPath(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("GreetXML", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			QueryParam("firstName", "John").
			QueryParam("lastName", "Doe").
			Get("/greetXML")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Contains(resp.ContentType(), "application/xml")

		xp, err := resp.XmlPath()
		must.NoError(err)
		is.Equal("John", xp.GetString("//greeting/firstName"))
		is.Equal("Doe", xp.GetString("//greeting/lastName"))
	})

	t.Run("AnotherGreetXML_Nested", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			QueryParam("firstName", "Jane").
			QueryParam("lastName", "Doe").
			Get("/anotherGreetXML")

		must.NoError(resp.Err())
		xp, err := resp.XmlPath()
		must.NoError(err)
		is.Equal("Jane", xp.GetString("//greeting/name/firstName"))
		is.Equal("Doe", xp.GetString("//greeting/name/lastName"))
	})

	t.Run("GreetXMLAttribute", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			QueryParam("firstName", "Alice").
			QueryParam("lastName", "Smith").
			Get("/greetXMLAttribute")

		must.NoError(resp.Err())
		xp, err := resp.XmlPath()
		must.NoError(err)
		is.Equal("Alice", xp.GetString("//greeting/name/@firstName"))
		is.Equal("Smith", xp.GetString("//greeting/name/@lastName"))
	})

	t.Run("ShoppingCategories", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/shopping")

		must.NoError(resp.Err())
		xp, err := resp.XmlPath()
		must.NoError(err)

		items := xp.GetStringList("//shopping/category/item")
		is.Contains(items, "Chocolate")
		is.Contains(items, "Coffee")
		is.Contains(items, "Paper")
		is.Contains(items, "Pens")
		is.Contains(items, "Mango")
	})

	t.Run("MusicVideos", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/videos")

		must.NoError(resp.Err())
		xp, err := resp.XmlPath()
		must.NoError(err)
		is.Equal("Bohemian Rhapsody", xp.GetString("//music/title"))
		is.Equal("Queen", xp.GetString("//music/artist"))
	})

	t.Run("XmlWithMinusAndUnderscore", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp1 := rest.Given().
			BaseURI(ts.URL).
			QueryParam("firstName", "John").
			QueryParam("lastName", "Doe").
			Get("/xmlWithMinusInRoot")
		must.NoError(resp1.Err())
		xp1, err1 := resp1.XmlPath()
		must.NoError(err1)
		is.Equal("John", xp1.GetString("//a-greeting/firstName"))

		resp2 := rest.Given().
			BaseURI(ts.URL).
			QueryParam("firstName", "John").
			QueryParam("lastName", "Doe").
			Get("/xmlWithUnderscoreInChild")
		must.NoError(resp2.Err())
		xp2, err2 := resp2.XmlPath()
		must.NoError(err2)
		is.Equal("John", xp2.GetString("//greeting/your_firstName"))
	})
}

// TestJavaITest_MultiValueParams tests scenarios ported from Java's ParamITest.java
func TestJavaITest_MultiValueParams(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("MultiValueQueryParam", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			QueryParam("list", "item1", "item2", "item3").
			Get("/multiValueParam").
			Then().
			StatusCode(http.StatusOK).
			Body("list", matcher.EqualTo("item1,item2,item3")).
			AssertAllNoFail(t)
	})

	t.Run("ThreeMultiValueFormParams", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			FormParam("list", "a", "b").
			FormParam("list2", "c", "d").
			FormParam("list3", "e", "f").
			Post("/threeMultiValueParam").
			Then().
			StatusCode(http.StatusOK).
			Body("list", matcher.EqualTo("a,b")).
			Body("list2", matcher.EqualTo("c,d")).
			Body("list3", matcher.EqualTo("e,f")).
			AssertAllNoFail(t)
	})

	t.Run("NoValueParam", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			QueryParam("flag", "").
			Get("/noValueParam").
			Then().
			StatusCode(http.StatusOK).
			Body("params", matcher.ContainsString("flag=")).
			AssertAllNoFail(t)
	})
}

// TestJavaITest_HeadersAndContentTypes tests scenarios ported from Java's HeaderITest.java and ContentTypeITest.java
func TestJavaITest_HeadersAndContentTypes(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("MultiValueHeader", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			Get("/multiValueHeader").
			Then().
			StatusCode(http.StatusOK).
			HeaderExists("MultiHeader").
			AssertAllNoFail(t)
	})

	t.Run("HeadersWithValues", func(t *testing.T) {
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

	t.Run("CustomPlusJsonContentType", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			Get("/mimeTypeWithPlusJson").
			Then().
			StatusCode(http.StatusOK).
			HeaderMatching("Content-Type", matcher.ContainsString("application/something+json")).
			Body("message", matcher.EqualTo("It works")).
			AssertAllNoFail(t)
	})

	t.Run("HTMLContentType", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			Get("/textHTML").
			Then().
			StatusCode(http.StatusOK).
			ContentType(rest.ContentTypeHTML).
			BodyContains("paragraph 1").
			BodyContains("paragraph 2").
			AssertAllNoFail(t)
	})
}

// TestJavaITest_CookiesAndSession tests scenarios ported from Java's CookieITest.java and SessionIdITest.java
func TestJavaITest_CookiesAndSession(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("SetMultipleCookies", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			Get("/setCookies").
			Then().
			StatusCode(http.StatusOK).
			Cookie("key1", "value1").
			Cookie("key2", "value2").
			Cookie("key3", "value3").
			AssertAllNoFail(t)
	})

	t.Run("SessionIdValidation", func(t *testing.T) {
		t.Parallel()
		// First request gets new session
		resp1 := rest.Given().
			BaseURI(ts.URL).
			Get("/sessionId")

		resp1.Then().
			StatusCode(http.StatusOK).
			CookieExists("jsessionid").
			Body("sessionId", matcher.EqualTo("1234")).
			AssertAllNoFail(t)

		// Second request sends session back
		cookie := resp1.Cookie("jsessionid")
		rest.Given().
			BaseURI(ts.URL).
			Cookie("jsessionid", cookie).
			Get("/sessionId").
			Then().
			StatusCode(http.StatusOK).
			Body("status", matcher.EqualTo("Success")).
			AssertAllNoFail(t)
	})
}

// TestJavaITest_CsrfAndFormAuth tests scenarios ported from Java's CsrfITest.java
func TestJavaITest_CsrfAndFormAuth(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("SpringSecurityCheck", func(t *testing.T) {
		t.Parallel()
		resp := rest.Given().
			BaseURI(ts.URL).
			FormParam("j_username", "John").
			FormParam("j_password", "Doe").
			Post("/j_spring_security_check")

		resp.Then().
			StatusCode(http.StatusOK).
			CookieExists("jsessionid").
			BodyEquals("OK").
			AssertAllNoFail(t)

		// Access secured resource with session cookie
		sessionCookie := resp.Cookie("jsessionid")
		rest.Given().
			BaseURI(ts.URL).
			Cookie("jsessionid", sessionCookie).
			Get("/session-required").
			Then().
			StatusCode(http.StatusOK).
			Body("message", matcher.EqualTo("session valid")).
			AssertAllNoFail(t)
	})

	t.Run("HeaderCsrfValidation", func(t *testing.T) {
		t.Parallel()
		// Failed without header
		rest.Given().
			BaseURI(ts.URL).
			Post("/pageThatRequireHeaderCsrf").
			Then().
			StatusCode(http.StatusForbidden).
			AssertAllNoFail(t)

		// Succeed with valid CSRF header
		rest.Given().
			BaseURI(ts.URL).
			Header("X-CSRF-TOKEN", "ab8722b1-1f23-4dcf-bf63-fb8b94be4107").
			Post("/pageThatRequireHeaderCsrf").
			Then().
			StatusCode(http.StatusOK).
			BodyEquals("OK").
			AssertAllNoFail(t)
	})
}

// TestJavaITest_PathParams tests scenarios ported from Java's PathParamITest.java
func TestJavaITest_PathParams(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("TwoPartPathParams", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			PathParam("firstName", "John").
			PathParam("lastName", "Doe").
			Get("/greeting/{firstName}/{lastName}").
			Then().
			StatusCode(http.StatusOK).
			Body("firstName", matcher.EqualTo("John")).
			Body("lastName", matcher.EqualTo("Doe")).
			Body("fullName", matcher.EqualTo("John Doe")).
			AssertAllNoFail(t)
	})

	t.Run("ThreePartPathParams", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			PathParam("firstName", "John").
			PathParam("middleName", "Fitzgerald").
			PathParam("lastName", "Kennedy").
			Get("/greeting/{firstName}/{middleName}/{lastName}").
			Then().
			StatusCode(http.StatusOK).
			Body("firstName", matcher.EqualTo("John")).
			Body("middleName", matcher.EqualTo("Fitzgerald")).
			Body("lastName", matcher.EqualTo("Kennedy")).
			AssertAllNoFail(t)
	})

	t.Run("SupportsPassingPathParamsAsMap", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			PathParams(map[string]any{
				"firstName": "Alice",
				"lastName":  "Wonderland",
			}).
			Get("/greeting/{firstName}/{lastName}")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Contains(resp.AsString(), "Alice")
		is.Contains(resp.AsString(), "Wonderland")
	})

	t.Run("PathParamWithIntegerValue", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// Path param values are always strings in the URL
		resp := rest.Given().
			BaseURI(ts.URL).
			PathParam("firstName", "42").
			PathParam("lastName", "99").
			Get("/greeting/{firstName}/{lastName}")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
	})

	t.Run("PathParamWithSpecialCharacters", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			PathParam("firstName", "John").
			PathParam("lastName", "O'Brien").
			Get("/greeting/{firstName}/{lastName}")

		must.NoError(resp.Err())
		// Server should handle URL-encoded values
		is.Equal(http.StatusOK, resp.StatusCode())
	})

	t.Run("PathParamWithUnicodeValue", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			PathParam("firstName", "Ångström").
			PathParam("lastName", "Müller").
			Get("/greeting/{firstName}/{lastName}")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
	})

	t.Run("NamedPathParamDefinedMultipleTimesWorks", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// Both firstName and lastName are provided in a map
		resp := rest.Given().
			BaseURI(ts.URL).
			PathParams(map[string]any{
				"firstName": "Bob",
				"lastName":  "Bob",
			}).
			Get("/greeting/{firstName}/{lastName}")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Contains(resp.AsString(), "Bob")
	})

	t.Run("PathParamsMergeWithPerRequestParam", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		spec := rest.NewRequestSpecBuilder().
			SetBaseURI(ts.URL).
			AddPathParam("firstName", "SpecFirst").
			Build()

		resp := rest.Given().
			Spec(spec).
			PathParam("lastName", "RequestLast").
			Get("/greeting/{firstName}/{lastName}")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Contains(resp.AsString(), "SpecFirst")
		is.Contains(resp.AsString(), "RequestLast")
	})

	t.Run("PathParamWithStatusCode", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			PathParam("code", "409").
			Get("/status/{code}")

		must.NoError(resp.Err())
		is.Equal(http.StatusConflict, resp.StatusCode())
	})

	t.Run("PathParamMixedWithQueryParam", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			PathParam("firstName", "Mix").
			PathParam("lastName", "Param").
			Get("/greeting/{firstName}/{lastName}")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Contains(resp.AsString(), "Mix")
	})

	t.Run("ThreePathParamsAsMap", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			PathParams(map[string]any{
				"firstName":  "James",
				"middleName": "Earl",
				"lastName":   "Jones",
			}).
			Get("/greeting/{firstName}/{middleName}/{lastName}")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		body := resp.AsString()
		is.Contains(body, "James")
		is.Contains(body, "Earl")
		is.Contains(body, "Jones")
	})
}

// TestJavaITest_Gzip tests scenarios ported from Java's GzipITest.java
func TestJavaITest_Gzip(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("GzipJsonPayload", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			Get("/gzip-json").
			Then().
			StatusCode(http.StatusOK).
			Body("hello", matcher.EqualTo("Hello Scalatra")).
			AssertAllNoFail(t)
	})
}

// TestJavaITest_StatusCodesAndErrors tests scenarios ported from Java's ErrorMessageITest.java
func TestJavaITest_StatusCodesAndErrors(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("Conflict409", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			Get("/409").
			Then().
			StatusCode(http.StatusConflict).
			BodyEquals("ERROR").
			AssertAllNoFail(t)
	})

	t.Run("StatusCode500", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			Get("/statusCode500").
			Then().
			StatusCode(http.StatusInternalServerError).
			BodyContains("expected error occurred").
			AssertAllNoFail(t)
	})

	t.Run("EmptyBody", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			Get("/emptyBody").
			Then().
			StatusCode(http.StatusOK).
			BodyEquals("").
			AssertAllNoFail(t)
	})
}

// TestJavaITest_PathParams_Extended adds more PathParamITest.java scenarios.
func TestJavaITest_PathParams_Extended(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("SupportsPassingPathParamsToRequestSpec", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		spec := rest.NewRequestSpecBuilder().
			SetBaseURI(ts.URL).
			AddPathParam("firstName", "John").
			AddPathParam("lastName", "Doe").
			Build()

		resp := rest.Given().
			Spec(spec).
			Get("/greeting/{firstName}/{lastName}")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Contains(resp.AsString(), "John")
	})

	t.Run("SupportsPassingPathParamsAsMapToRequestSpec", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		spec := rest.NewRequestSpecBuilder().
			SetBaseURI(ts.URL).
			AddPathParam("firstName", "Jane").
			AddPathParam("lastName", "Smith").
			Build()

		resp := rest.Given().
			Spec(spec).
			Get("/greeting/{firstName}/{lastName}")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Contains(resp.AsString(), "Jane")
	})

	t.Run("SupportsPassingIntPathParamsToRequestSpec", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		spec := rest.NewRequestSpecBuilder().
			SetBaseURI(ts.URL).
			AddPathParam("code", 200).
			Build()

		resp := rest.Given().
			Spec(spec).
			Get("/status/{code}")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
	})

	t.Run("SupportsPassingPathParamsToGet", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			PathParam("firstName", "Bob").
			PathParam("lastName", "Marley").
			Get("/greeting/{firstName}/{lastName}")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Contains(resp.AsString(), "Bob")
	})

	t.Run("SupportsPassingPathParamsAsMapToGet", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			PathParams(map[string]any{"firstName": "Carl", "lastName": "Sagan"}).
			Get("/greeting/{firstName}/{lastName}")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Contains(resp.AsString(), "Carl")
	})

	t.Run("URLEncodesPathParams", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// Space should be percent-encoded in the URL.
		resp := rest.Given().
			BaseURI(ts.URL).
			PathParam("firstName", "John Paul").
			PathParam("lastName", "Jones").
			Get("/greeting/{firstName}/{lastName}")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
	})

	t.Run("URLEncodesPathParamsInMap", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			PathParams(map[string]any{"firstName": "Mary Jane", "lastName": "Watson"}).
			Get("/greeting/{firstName}/{lastName}")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
	})

	t.Run("DisablingURLEncodingPathParamsWorks", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			URLEncodingEnabled(false).
			PathParam("firstName", "John").
			PathParam("lastName", "Doe").
			Get("/greeting/{firstName}/{lastName}")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
	})

	t.Run("UnnamedQueryParametersWorks", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// Positional path params (unnamed) via Get() varargs.
		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/status/{0}", 201)

		must.NoError(resp.Err())
		is.Equal(http.StatusCreated, resp.StatusCode())
	})

	t.Run("MixingUnnamedPathParametersAndQueryParametersWorks", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			QueryParam("firstName", "Query").
			QueryParam("lastName", "Param").
			Get("/greet")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Contains(resp.AsString(), "Query")
	})

	t.Run("NamedPathParametersWorksWithUnicodeParameterValues", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			PathParam("firstName", "Ångström").
			PathParam("lastName", "Müller").
			Get("/greeting/{firstName}/{lastName}")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
	})

	t.Run("PassingInSinglePathParamDefinedMultipleTimesWorks", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			PathParam("firstName", "Echo").
			PathParam("lastName", "Echo").
			Get("/greeting/{firstName}/{lastName}")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		body := resp.AsString()
		is.Contains(body, "Echo")
	})

	t.Run("MergesPathParamsMapWithNonMapWhenGiven", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// Spec provides firstName, per-request provides lastName.
		spec := rest.NewRequestSpecBuilder().
			SetBaseURI(ts.URL).
			AddPathParam("firstName", "Merged").
			Build()

		resp := rest.Given().
			Spec(spec).
			PathParam("lastName", "Params").
			Get("/greeting/{firstName}/{lastName}")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		body := resp.AsString()
		is.Contains(body, "Merged")
		is.Contains(body, "Params")
	})

	t.Run("CanUsePathParamsWithNonStandardChars", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			PathParam("firstName", "O'Brien").
			PathParam("lastName", "Smith-Jones").
			Get("/greeting/{firstName}/{lastName}")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
	})

	t.Run("PositionalPathParamsWithStatusCode", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/status/{0}", 404)

		must.NoError(resp.Err())
		is.Equal(http.StatusNotFound, resp.StatusCode())
	})

	t.Run("NamedPathParametersCanBeAppendedBeforeSubPath", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// greeting/{firstName}/{lastName} where params are supplied individually
		resp := rest.Given().
			BaseURI(ts.URL).
			PathParam("firstName", "Pre").
			PathParam("lastName", "Sub").
			Get("/greeting/{firstName}/{lastName}")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Contains(resp.AsString(), "Pre")
	})

	t.Run("UnnamedPathParametersWorksWhenThereAreMultipleTemplates", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// Two positional path params.
		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/greeting/{0}/{1}", "Pos", "Params")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		body := resp.AsString()
		is.Contains(body, "Pos")
	})

	t.Run("NamedPathParamsSpecViaRequestSpecBuilderAddPathParams", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		spec := rest.NewRequestSpecBuilder().
			SetBaseURI(ts.URL).
			AddPathParam("firstName", "Spec").
			AddPathParam("lastName", "Builder").
			Build()

		resp := rest.Given().
			Spec(spec).
			Get("/greeting/{firstName}/{lastName}")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		body := resp.AsString()
		is.Contains(body, "Spec")
	})
}

// TestJavaITest_JSONGet_Extended2 adds missing JSONGetITest.java scenarios.
func TestJavaITest_JSONGet_Extended2(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("GpathJSONAndHamcrestMatcher", func(t *testing.T) {
		t.Parallel()

		// books with price > 10 — expect at least one.
		rest.Given().
			BaseURI(ts.URL).
			Get("/jsonStore").
			Then().
			StatusCode(http.StatusOK).
			Body("store.book.price", matcher.HasItem(matcher.GreaterThanNum(10.0))).
			AssertAllNoFail(t)
	})

	t.Run("GpathAssertionWithHamcrestMatcherAndJSONReturnsArray", func(t *testing.T) {
		t.Parallel()

		rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			StatusCode(http.StatusOK).
			Body("lotto.winning-numbers", matcher.HasItemValue(45)).
			Body("lotto.winning-numbers", matcher.HasSizeMatcher(matcher.GreaterThan(0))).
			AssertAllNoFail(t)
	})

	t.Run("MultipleSingleParametersAreConcatenated", func(t *testing.T) {
		t.Parallel()

		rest.Given().
			BaseURI(ts.URL).
			QueryParam("firstName", "John").
			QueryParam("lastName", "Doe").
			Get("/greet").
			Then().
			Body("greeting", matcher.ContainsString("John")).
			Body("greeting", matcher.ContainsString("Doe")).
			AssertAllNoFail(t)
	})

	t.Run("MultipleParametersAreConcatenated", func(t *testing.T) {
		t.Parallel()

		rest.Given().
			BaseURI(ts.URL).
			QueryParam("firstName", "Jane").
			QueryParam("lastName", "Smith").
			Get("/greet").
			Then().
			Body("greeting", matcher.ContainsString("Jane")).
			Body("greeting", matcher.ContainsString("Smith")).
			AssertAllNoFail(t)
	})

	t.Run("NewSyntaxWithParameters", func(t *testing.T) {
		t.Parallel()

		rest.Given().
			BaseURI(ts.URL).
			QueryParam("firstName", "Syntax").
			QueryParam("lastName", "Test").
			Get("/greet").
			Then().
			StatusCode(http.StatusOK).
			Body("greeting", matcher.ContainsString("Syntax")).
			AssertAllNoFail(t)
	})

	t.Run("NewSyntaxWithCorrectStatusCodeUsingHamcrestMatcher", func(t *testing.T) {
		t.Parallel()

		rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			StatusCodeMatching(matcher.EqualTo(http.StatusOK)).
			AssertAllNoFail(t)
	})

	t.Run("NewSyntaxWithCorrectStatusLineUsingHamcrestMatcher", func(t *testing.T) {
		t.Parallel()

		rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			StatusLineContains("200").
			AssertAllNoFail(t)
	})

	t.Run("MultipleBodyHamcrestMatchersLongVersion", func(t *testing.T) {
		t.Parallel()

		rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			StatusCode(http.StatusOK).
			Body("lotto.lottoId", matcher.EqualTo(5)).
			Body("lotto.winning-numbers", matcher.HasItemValue(2)).
			Body("lotto.winning-numbers", matcher.HasItemValue(45)).
			AssertAllNoFail(t)
	})

	t.Run("GetFirstTwoElementsInList", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto")

		must.NoError(resp.Err())
		jp := resp.JsonPath()
		// winning-numbers[0] and [1] both exist.
		is.NotNil(jp.Get("lotto.winning-numbers.0"))
		is.NotNil(jp.Get("lotto.winning-numbers.1"))
	})

	t.Run("GetFirstAndLastElementsInList", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto")

		must.NoError(resp.Err())
		jp := resp.JsonPath()
		first := jp.Get("lotto.winning-numbers.0")
		is.NotNil(first, "first element should exist")
	})

	t.Run("GetRangeInList", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto")

		must.NoError(resp.Err())
		jp := resp.JsonPath()
		// Get all winning numbers as slice.
		nums := jp.Get("lotto.winning-numbers").Value()
		is.NotNil(nums)
	})

	t.Run("SpecificationSyntax", func(t *testing.T) {
		t.Parallel()

		reqSpec := rest.NewRequestSpecBuilder().
			SetBaseURI(ts.URL).
			Build()

		rest.Given().
			Spec(reqSpec).
			QueryParam("firstName", "Spec").
			QueryParam("lastName", "Syntax").
			Get("/greet").
			Then().
			StatusCode(http.StatusOK).
			Body("greeting", matcher.ContainsString("Spec")).
			AssertAllNoFail(t)
	})

	t.Run("UuidIsTreatedAsString", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/uuid")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		id := resp.JsonPath().GetString("id")
		is.NotEmpty(id)
		// UUID format: 8-4-4-4-12 hex chars separated by hyphens.
		parts := strings.Split(id, "-")
		is.Len(parts, 5)
	})

	t.Run("ContentTypeButNoBodyWhenError", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/409")

		must.NoError(resp.Err())
		is.Equal(http.StatusConflict, resp.StatusCode())
		// Response still has a body even on error status.
	})

	t.Run("ThrowsExceptionWhenJsonPathIsInvalid", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto")

		must.NoError(resp.Err())
		// An invalid/nonexistent path returns empty/nil rather than panicking.
		result := resp.JsonPath().Get("lotto.nonexistentKey")
		is.Empty(result.Value())
	})

	t.Run("MultipleBodyJsonStringMatchersShortVersion", func(t *testing.T) {
		t.Parallel()

		rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			Body("lotto.lottoId", 5).
			Body("lotto.winners.0.winnerId", matcher.Anything()).
			AssertAllNoFail(t)
	})

	t.Run("HasItemHamcrestMatchingWorkForArray", func(t *testing.T) {
		t.Parallel()

		rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			Body("lotto.winning-numbers", matcher.HasItemValue(2)).
			Body("lotto.winning-numbers", matcher.HasItemValue(45)).
			AssertAllNoFail(t)
	})

	t.Run("StatusCodeHasPriorityOverJsonParsingWhenErrorOccurs", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// malformedJson returns invalid JSON with status 200.
		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/malformedJson")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		// Body should not be empty even if JSON parsing fails.
		is.NotEmpty(resp.AsString())
	})

	t.Run("SupportsGettingListItemInNonArrayStyle", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/jsonList")

		must.NoError(resp.Err())
		// Access element by dot-notation index.
		jp := resp.JsonPath()
		name := jp.GetString("0.name")
		is.Equal("Anders", name)
	})

	t.Run("NewSyntax", func(t *testing.T) {
		t.Parallel()

		rest.Given().
			BaseURI(ts.URL).
			Get("/hello").
			Then().
			StatusCode(http.StatusOK).
			AssertAllNoFail(t)
	})

	t.Run("RestAssuredSupportsPrintingTheResponse", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto")

		must.NoError(resp.Err())
		// AsString() is the Go equivalent of print()/prettyPrint() on the response body.
		body := resp.AsString()
		is.Contains(body, "lotto")
	})

	t.Run("SupportsGettingResponseBodyWhenStatusCodeIs401InGet", func(t *testing.T) {
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

	t.Run("SupportsValidatingCookiesWithNoValue", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/html_with_cookie")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		// JSESSIONID should be present even with no explicit value check.
		is.NotEmpty(resp.Cookie("JSESSIONID"))
	})
}
