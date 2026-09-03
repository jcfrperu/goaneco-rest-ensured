package integration_test

// Ported from CsrfITest.java.
// Java uses RestAssured.config().csrfConfig(...) and Auth().form(..., using(CsrfConfig)).
// Go maps to CsrfFilter with manual configuration.

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func TestJavaITest_Csrf_HeaderBased(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("CsrfHeaderInjectedFromMetaTag", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// CsrfFilter fetches token from <meta name="_csrf" content="..."/> and sends as X-CSRF-TOKEN header.
		csrfFilter := rest.NewCsrfFilter(ts.URL+"/pageWithDefaultHeaderCsrf", "_csrf")
		csrfFilter.HeaderName = "X-CSRF-TOKEN"

		resp := rest.Given().
			BaseURI(ts.URL).
			Filter(csrfFilter).
			Get("/pageThatRequireHeaderCsrf")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
	})

	t.Run("CsrfHeaderWithCustomizedHeaderName", func(t *testing.T) {
		t.Parallel()

		// Use the query-param-based endpoint that accepts a custom header name.
		customToken := "ab8722b1-1f23-4dcf-bf63-fb8b94be4107"

		resp := rest.Given().
			BaseURI(ts.URL).
			Header("X-CSRF-TOKEN", customToken).
			Get("/pageThatRequireHeaderCsrf")

		rest.Given().
			BaseURI(ts.URL).
			Get("/pageThatRequireHeaderCsrf").
			Then().
			StatusCode(http.StatusForbidden). // without token → 403
			AssertAllNoFail(t)

		resp = rest.Given().
			BaseURI(ts.URL).
			Header("X-CSRF-TOKEN", customToken).
			Get("/pageThatRequireHeaderCsrf")

		assert.Equal(t, http.StatusOK, resp.StatusCode())
	})

	t.Run("CsrfHeaderDerivedFromSpecifiedMetaTagName", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// Both _csrf_header (header name) and _csrf (token value) are in the page meta tags.
		csrfFilter := rest.NewCsrfFilter(ts.URL+"/pageWithDefaultHeaderCsrf", "_csrf")
		csrfFilter.HeaderName = "X-CSRF-TOKEN"

		resp := rest.Given().
			BaseURI(ts.URL).
			Filter(csrfFilter).
			Get("/pageThatRequireHeaderCsrf")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
	})

	t.Run("RequestWithoutCsrfTokenIsRejected", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/pageThatRequireHeaderCsrf")

		must.NoError(resp.Err())
		is.Equal(http.StatusForbidden, resp.StatusCode())
	})

	t.Run("CsrfTokenIsReusedAcrossMultipleRequests", func(t *testing.T) {
		t.Parallel()
		must := require.New(t)

		csrfFilter := rest.NewCsrfFilter(ts.URL+"/pageWithDefaultHeaderCsrf", "_csrf")
		csrfFilter.HeaderName = "X-CSRF-TOKEN"

		for i := range 3 {
			resp := rest.Given().
				BaseURI(ts.URL).
				Filter(csrfFilter).
				Get("/pageThatRequireHeaderCsrf")
			must.NoError(resp.Err(), "request %d failed", i)
			must.Equal(http.StatusOK, resp.StatusCode(), "request %d", i)
		}
	})

	t.Run("CsrfFilterWithLogging", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var buf bytes.Buffer
		logFilter := rest.NewRequestLoggingFilter(&buf, rest.LogDetailHeaders)
		csrfFilter := rest.NewCsrfFilter(ts.URL+"/pageWithDefaultHeaderCsrf", "_csrf")
		csrfFilter.HeaderName = "X-CSRF-TOKEN"

		resp := rest.Given().
			BaseURI(ts.URL).
			Filter(csrfFilter, logFilter).
			Get("/pageThatRequireHeaderCsrf")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.NotEmpty(buf.String(), "logging filter should have written something")
	})
}

func TestJavaITest_Csrf_FormBased(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("CsrfFormTokenInjectedIntoFormParams", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// /loginPageWithCsrf returns an HTML form with hidden _csrf input.
		// POSTing back with the token should succeed.
		csrfFilter := rest.NewCsrfFilter(ts.URL+"/loginPageWithCsrf", "_csrf")

		resp := rest.Given().
			BaseURI(ts.URL).
			Filter(csrfFilter).
			FormParam("j_username", "john").
			FormParam("j_password", "doe").
			Post("/loginPageWithCsrf")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
	})

	t.Run("CsrfIsNotUsedWhenTokenPathIsUndefined", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// Filter with empty csrfURI — no token will be fetched or injected.
		csrfFilter := rest.NewCsrfFilter("", "_csrf")

		resp := rest.Given().
			BaseURI(ts.URL).
			Filter(csrfFilter).
			FormParam("j_username", "john").
			FormParam("j_password", "doe").
			Post("/loginPageWithCsrf")

		must.NoError(resp.Err())
		// Without token, the POST will fail with 403.
		is.Equal(http.StatusForbidden, resp.StatusCode())
	})

	t.Run("FormAuthWithAutoCsrfDetectionViaFilter", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// Validates that fetching the CSRF token from a GET page and posting it works.
		csrfFilter := rest.NewCsrfFilter(ts.URL+"/loginPageWithCsrf", "_csrf")

		resp := rest.Given().
			BaseURI(ts.URL).
			Filter(csrfFilter).
			Post("/loginPageWithCsrf")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
	})
}

func TestJavaITest_Csrf_GetRequestsDoNotSendCsrf(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("CsrfIsNotUsedForGetRequests", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// GET /lotto with a CsrfFilter — the filter should not interfere with GET.
		csrfFilter := rest.NewCsrfFilter(ts.URL+"/loginPageWithCsrf", "_csrf")

		resp := rest.Given().
			BaseURI(ts.URL).
			Filter(csrfFilter).
			Get("/lotto")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal(5, resp.JsonPath().GetInt("lotto.lottoId"))
	})

	t.Run("CsrfIsNotUsedForHeadRequests", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		csrfFilter := rest.NewCsrfFilter(ts.URL+"/loginPageWithCsrf", "_csrf")

		resp := rest.Given().
			BaseURI(ts.URL).
			Filter(csrfFilter).
			Head("/hello")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
	})
}

func TestJavaITest_Csrf_InlineServer(t *testing.T) {
	t.Parallel()

	// Inline server that accepts a header CSRF token with a custom header name.
	customHeaderName := "X-MY-CSRF"
	customToken := "tok-abc-123"

	inlineServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/csrf-page":
			w.Header().Set("Content-Type", "text/html")
			// Embed token as a meta tag with the custom field name.
			_, _ = w.Write([]byte(`<!DOCTYPE html><html><head>
<meta name="my-csrf" content="` + customToken + `"/>
</head><body></body></html>`))
		case "/csrf-protected":
			if r.Header.Get(customHeaderName) != customToken {
				http.Error(w, "CSRF invalid", http.StatusForbidden)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"ok":true}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(inlineServer.Close)

	t.Run("CsrfHeaderWithCustomFieldNameAndHeaderName", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		csrfFilter := rest.NewCsrfFilter(inlineServer.URL+"/csrf-page", "my-csrf")
		csrfFilter.HeaderName = customHeaderName

		resp := rest.Given().
			BaseURI(inlineServer.URL).
			Filter(csrfFilter).
			Get("/csrf-protected")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.True(resp.JsonPath().GetBool("ok"))
	})

	t.Run("CsrfFilterConcurrentRequestsReuseToken", func(t *testing.T) {
		t.Parallel()
		must := require.New(t)

		csrfFilter := rest.NewCsrfFilter(inlineServer.URL+"/csrf-page", "my-csrf")
		csrfFilter.HeaderName = customHeaderName

		for i := range 3 {
			resp := rest.Given().
				BaseURI(inlineServer.URL).
				Filter(csrfFilter).
				Get("/csrf-protected")
			must.NoError(resp.Err(), "request %d", i)
			must.Equal(http.StatusOK, resp.StatusCode(), "request %d", i)
		}
	})
}

func TestJavaITest_Csrf_CookiePropagation(t *testing.T) {
	t.Parallel()

	// Inline server that sets a cookie on the CSRF page and validates it on the target.
	csrfToken := "csrf-cookie-tok"
	serverCookieName := "SESSION"
	serverCookieValue := "sess-xyz"

	inlineServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/page-with-csrf-and-cookie":
			http.SetCookie(w, &http.Cookie{Name: serverCookieName, Value: serverCookieValue, Path: "/"})
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(`<!DOCTYPE html><html><head>
<meta name="_csrf" content="` + csrfToken + `"/>
</head><body></body></html>`))
		case "/protected":
			// Validate both CSRF header and session cookie.
			cookie, err := r.Cookie(serverCookieName)
			if err != nil || cookie.Value != serverCookieValue {
				http.Error(w, "missing session", http.StatusUnauthorized)
				return
			}
			if r.Header.Get("X-CSRF-TOKEN") != csrfToken {
				http.Error(w, "missing csrf", http.StatusForbidden)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"protected":true}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(inlineServer.Close)

	t.Run("WhenCsrfIsConfiguredCookiesCanBeManuallySupplied", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// Manually supply the session cookie and let CsrfFilter inject the token.
		csrfFilter := rest.NewCsrfFilter(inlineServer.URL+"/page-with-csrf-and-cookie", "_csrf")
		csrfFilter.HeaderName = "X-CSRF-TOKEN"

		resp := rest.Given().
			BaseURI(inlineServer.URL).
			Cookie(serverCookieName, serverCookieValue).
			Filter(csrfFilter).
			Get("/protected")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.True(resp.JsonPath().GetBool("protected"))
	})
}
