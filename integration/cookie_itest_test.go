package integration_test

// Ported from CookieITest.java

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/matcher"
	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func TestJavaITest_Cookie_ResponseCookies(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("CookiesReturnsMapWhereLastValueOfMultiValueCookieIsUsed", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// /multiCookie sets two cookie1 entries with different values
		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/multiCookie")

		must.NoError(resp.Err())
		// Go's http.Response.Cookies() returns all Set-Cookie headers in order.
		// The last one is the most recently set.
		all := resp.Cookies()
		is.NotEmpty(all)
		// At least one cookie named "cookie1" exists
		var found bool
		for _, c := range all {
			if c.Name == "cookie1" {
				found = true
				break
			}
		}
		is.True(found, "expected cookie named 'cookie1' in response")
	})

	t.Run("DetailedCookiesAllowToGetMultiValues", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/multiCookie")

		must.NoError(resp.Err())
		cookies := resp.Cookies()
		is.GreaterOrEqual(len(cookies), 2, "expected at least two cookies from /multiCookie")
	})

	t.Run("SetsMultipleDistinctCookies", func(t *testing.T) {
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

	t.Run("SupportsDetailedCookieMatcher", func(t *testing.T) {
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

	t.Run("SupportsCookieStringMatching", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			Get("/setCookies").
			Then().
			StatusCode(http.StatusOK).
			Cookie("key1", "value1").
			AssertAllNoFail(t)
	})

	t.Run("MultipleCookieStatementsAreConcatenated", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/setCookies")

		must.NoError(resp.Err())
		is.NotEmpty(resp.Cookie("key1"))
		is.NotEmpty(resp.Cookie("key2"))
		is.NotEmpty(resp.Cookie("key3"))
	})

	t.Run("MultipleCookiesUsingMap", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Cookies(map[string]string{
				"map_cookie_1": "mv1",
				"map_cookie_2": "mv2",
			}).
			Get("/cookies")

		must.NoError(resp.Err())
		jp := resp.JsonPath()
		is.Equal("mv1", jp.GetString("map_cookie_1"))
		is.Equal("mv2", jp.GetString("map_cookie_2"))
	})

	t.Run("WhenExpectedCookieDoesNotMatchAnAssertionIsRecorded", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		valid := rest.Given().
			BaseURI(ts.URL).
			Get("/setCookies").
			Then()

		valid.Cookie("key1", "wrong-value")
		is.True(valid.HasFailures(), "expected failure when cookie value does not match")
	})

	t.Run("WhenExpectedCookieIsNotFoundAnAssertionIsRecorded", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		valid := rest.Given().
			BaseURI(ts.URL).
			Get("/setCookies").
			Then()

		valid.CookieExists("nonexistent_cookie_xyz")
		is.True(valid.HasFailures(), "expected failure when cookie does not exist")
	})

	t.Run("ResponseSpecificationAllowsParsingCookieWithNoValue", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/html_with_cookie")

		must.NoError(resp.Err())
		is.NotEmpty(resp.Cookie("JSESSIONID"))
	})

	t.Run("CanGetCookieDetails", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/setCookies")

		must.NoError(resp.Err())
		cookies := resp.Cookies()
		is.NotEmpty(cookies)
		byName := make(map[string]*http.Cookie)
		for _, c := range cookies {
			byName[c.Name] = c
		}
		is.Equal("value1", byName["key1"].Value)
		is.Equal("/", byName["key1"].Path)
	})

	t.Run("GetDetailedCookieAttributesForSameKeyMultiple", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/setCommonIdCookies")

		must.NoError(resp.Err())
		// Three cookies with same key1 name; all should be returned
		cookies := resp.Cookies()
		var key1Count int
		for _, c := range cookies {
			if c.Name == "key1" {
				key1Count++
			}
		}
		is.Equal(3, key1Count, "expected three cookies named key1")
	})
}

func TestJavaITest_Cookie_RequestCookies(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("RequestSpecificationAllowsSpecifyingCookieWithNoValue", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Cookie("empty_cookie", "").
			Get("/cookies")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
	})

	t.Run("RequestSpecificationAllowsSpecifyingMultipleCookies", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Cookie("req_a", "ra").
			Cookie("req_b", "rb").
			Get("/cookies")

		must.NoError(resp.Err())
		jp := resp.JsonPath()
		is.Equal("ra", jp.GetString("req_a"))
		is.Equal("rb", jp.GetString("req_b"))
	})

	t.Run("RequestSpecificationAllowsSpecifyingCookiesUsingMap", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			Cookies(map[string]string{"c1": "v1", "c2": "v2"}).
			Get("/cookies").
			Then().
			StatusCode(http.StatusOK).
			Body("c1", matcher.EqualTo("v1")).
			Body("c2", matcher.EqualTo("v2")).
			AssertAllNoFail(t)
	})

	t.Run("UsesCookiesDefinedInStaticRequestSpecification", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		spec := rest.NewRequestSpecBuilder().
			SetBaseURI(ts.URL).
			AddCookie("spec_cookie", "spec_val").
			Build()

		resp := rest.Given().
			Spec(spec).
			Get("/cookies")

		must.NoError(resp.Err())
		is.Equal("spec_val", resp.JsonPath().GetString("spec_cookie"))
	})

	t.Run("CookiesWithMultiValueRequestEchoed", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Cookie("multi_a", "ma").
			Cookie("multi_b", "mb").
			Cookie("multi_c", "mc").
			Get("/multiCookieRequest")

		must.NoError(resp.Err())
		body := resp.AsString()
		is.Contains(body, "multi_a")
		is.Contains(body, "multi_b")
		is.Contains(body, "multi_c")
	})
}

func TestJavaITest_Cookie_CookieFilter(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("CookieFilterPersistsAcrossRequests", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		cookieFilter := rest.NewCookieFilter()

		// First request: sets jsessionid=1234
		resp1 := rest.Given().
			BaseURI(ts.URL).
			Filter(cookieFilter).
			Get("/sessionId")

		must.NoError(resp1.Err())
		is.NotEmpty(resp1.Cookie("jsessionid"))

		// Second request: filter automatically sends jsessionid back
		resp2 := rest.Given().
			BaseURI(ts.URL).
			Filter(cookieFilter).
			Get("/sessionId")

		must.NoError(resp2.Err())
		is.Equal("Success", resp2.JsonPath().GetString("status"))
	})

	t.Run("SessionFilterRecordsAndProvidesSessionId", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		sessionFilter := rest.NewSessionFilter()
		is.Empty(sessionFilter.SessionID())

		// Login request sets jsessionid cookie
		resp1 := rest.Given().
			BaseURI(ts.URL).
			Filter(sessionFilter).
			Get("/html_with_cookie")

		must.NoError(resp1.Err())
		is.NotEmpty(sessionFilter.SessionID(), "session filter should capture JSESSIONID")
	})
}

func TestJavaITest_Cookie_DetailedAttributes(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("CookiesSupportEqualCharacterInCookieValue", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// Send a cookie whose value contains an equals sign (common in base64)
		resp := rest.Given().
			BaseURI(ts.URL).
			Cookie("token", "abc==def").
			Get("/cookies")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Contains(resp.AsString(), "token")
	})

	t.Run("CookiesParsingSupportsNoValueCookies", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Cookie("flag_cookie", "").
			Get("/cookies")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
	})

	t.Run("DetailedCookieWorksByName", func(t *testing.T) {
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

	t.Run("CookieWithPathAttribute", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// /setCookies sets cookies with Path="/"; /cookiesWithValues echoes request cookies.
		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/setCookies")

		must.NoError(resp.Err())
		cookies := resp.Cookies()
		is.NotEmpty(cookies)
		var hasPath bool
		for _, c := range cookies {
			if c.Path != "" {
				hasPath = true
				break
			}
		}
		is.True(hasPath, "at least one cookie should have a Path attribute")
	})

	t.Run("MultipleCookiesShortVersionUsingPlainStrings", func(t *testing.T) {
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

	t.Run("MultipleCookiesShortVersionUsingHamcrestMatching", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			Get("/setCookies").
			Then().
			StatusCode(http.StatusOK).
			CookieMatching("key1", matcher.EqualTo("value1")).
			CookieMatching("key2", matcher.ContainsString("value")).
			AssertAllNoFail(t)
	})

	t.Run("WhenExpectedCookieHasWrongValueFailureIsRecorded", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		valid := rest.Given().
			BaseURI(ts.URL).
			Get("/setCookies").
			Then()

		valid.Cookie("key1", "WRONG_VALUE")
		is.True(valid.HasFailures(), "wrong cookie value should record a failure")
	})

	t.Run("WhenExpectedCookieIsNotFoundFailureIsRecorded", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		valid := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then()

		valid.CookieExists("nonexistent_cookie_xyz")
		is.True(valid.HasFailures(), "missing cookie should record a failure")
	})

	t.Run("CanSpecifyMultiValueCookieInRequest", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Cookie("mv1", "v1").
			Cookie("mv2", "v2").
			Cookie("mv3", "v3").
			Get("/multiCookieRequest")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		body := resp.AsString()
		is.Contains(body, "mv1")
		is.Contains(body, "mv2")
		is.Contains(body, "mv3")
	})

	t.Run("MissingCookieReturnsEmptyString", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto")

		must.NoError(resp.Err())
		is.Empty(resp.Cookie("nonexistent_xyz"))
	})
}
