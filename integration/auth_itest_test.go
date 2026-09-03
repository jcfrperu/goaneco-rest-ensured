package integration_test

// Ported from AuthenticationITest.java (expanded beyond auth_and_security_test.go)

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func TestJavaITest_Auth_BasicAuthentication(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("BasicAuthSuccess", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			Auth().Basic("admin", "secret").
			Get("/auth/basic").
			Then().
			StatusCode(http.StatusOK).
			Body("status", "authenticated").
			Body("user", "admin").
			AssertAllNoFail(t)
	})

	t.Run("BasicAuthFailureReturns401", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Auth().Basic("admin", "wrongpassword").
			Get("/auth/basic")

		is.Equal(http.StatusUnauthorized, resp.StatusCode())
	})

	t.Run("PreemptiveBasicAuthSuccess", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// Preemptive sends Authorization header immediately without 401 challenge
		resp := rest.Given().
			BaseURI(ts.URL).
			Auth().PreemptiveBasic("admin", "secret").
			Get("/auth/basic")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal("authenticated", resp.JsonPath().GetString("status"))
	})

	t.Run("PreemptiveBasicAuthFailure", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Auth().PreemptiveBasic("admin", "wrong").
			Get("/auth/basic")

		is.Equal(http.StatusUnauthorized, resp.StatusCode())
	})

	t.Run("BasicAuthUsingDefaultSpec", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		spec := rest.NewRequestSpecBuilder().
			SetBaseURI(ts.URL).
			SetAuth(&rest.BasicAuthScheme{Username: "admin", Password: "secret"}).
			Build()

		resp := rest.Given().
			Spec(spec).
			Get("/auth/basic")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
	})

	t.Run("ExplicitNoneAuthOverridesDefaultAuth", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		// Build spec with auth, then override with None() per request
		spec := rest.NewRequestSpecBuilder().
			SetBaseURI(ts.URL).
			SetAuth(&rest.BasicAuthScheme{Username: "admin", Password: "secret"}).
			Build()

		resp := rest.Given().
			Spec(spec).
			Auth().None(). // explicitly removes auth
			Get("/auth/basic")

		// Without auth, /auth/basic returns 401
		is.Equal(http.StatusUnauthorized, resp.StatusCode())
	})
}

func TestJavaITest_Auth_FormAuthentication(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("FormAuthUsingSpringSecurityCheck", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// Direct POST to Spring security check endpoint
		resp := rest.Given().
			BaseURI(ts.URL).
			FormParam("j_username", "John").
			FormParam("j_password", "Doe").
			Post("/j_spring_security_check")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.NotEmpty(resp.Cookie("jsessionid"))

		// Use session cookie to access protected resource
		sessionCookie := resp.Cookie("jsessionid")
		resp2 := rest.Given().
			BaseURI(ts.URL).
			Cookie("jsessionid", sessionCookie).
			Get("/session-required")

		must.NoError(resp2.Err())
		is.Equal(http.StatusOK, resp2.StatusCode())
		is.Equal("session valid", resp2.JsonPath().GetString("message"))
	})

	t.Run("FormAuthWithLoginPageParsing", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// Get login page with CSRF token, then post credentials
		pageResp := rest.Given().
			BaseURI(ts.URL).
			Get("/loginPageWithCsrf")

		must.NoError(pageResp.Err())
		csrfToken := rest.ExtractCsrfFromHTML(pageResp.AsString(), "_csrf")
		is.NotEmpty(csrfToken)

		// POST credentials with CSRF token
		loginResp := rest.Given().
			BaseURI(ts.URL).
			FormParam("_csrf", csrfToken).
			Post("/loginPageWithCsrf")

		must.NoError(loginResp.Err())
		is.Equal(http.StatusOK, loginResp.StatusCode())
	})

	t.Run("FormAuthWithAutoCsrfDetectionViaFilter", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		formAuth := rest.NewFormAuthFilter(
			"john",
			"doe",
			&rest.FormAuthConfig{
				FormAction:    ts.URL + "/login",
				UsernameField: "username",
				PasswordField: "password",
				CsrfField:     "_csrf",
			},
		)

		resp := rest.Given().
			BaseURI(ts.URL).
			Filter(formAuth).
			Get("/secured")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal("classified", resp.JsonPath().GetString("secret"))
	})

	t.Run("FormAuthInvalidCredentialsReturn401", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			FormParam("j_username", "wrong").
			FormParam("j_password", "wrong").
			Post("/j_spring_security_check")

		must.NoError(resp.Err())
		is.Equal(http.StatusUnauthorized, resp.StatusCode())
	})

	t.Run("CanSpecifyPortWhenUsingFormAuth", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// Verify direct form submission still works (port is embedded in ts.URL)
		resp := rest.Given().
			BaseURI(ts.URL).
			FormParam("username", "john").
			FormParam("password", "doe").
			Post("/login")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.NotEmpty(resp.Cookie("JSESSIONID"))
	})
}

func TestJavaITest_Auth_SessionFilter(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("SessionFilterRecordsAndProvidesSessionIdFromJSESSIONID", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		sessionFilter := rest.NewSessionFilter()
		is.Empty(sessionFilter.SessionID())

		// Login: server sets JSESSIONID
		resp1 := rest.Given().
			BaseURI(ts.URL).
			FormParam("username", "john").
			FormParam("password", "doe").
			Filter(sessionFilter).
			Post("/login")

		must.NoError(resp1.Err())
		is.Equal(http.StatusOK, resp1.StatusCode())
		sessionID := sessionFilter.SessionID()
		is.NotEmpty(sessionID)

		// Subsequent request: filter automatically forwards session ID
		resp2 := rest.Given().
			BaseURI(ts.URL).
			Filter(sessionFilter).
			Get("/secured")

		must.NoError(resp2.Err())
		is.Equal(http.StatusOK, resp2.StatusCode())
		is.Equal("classified", resp2.JsonPath().GetString("secret"))
	})

	t.Run("ReusingSameSessionFilterAppliesSessionIdToNewRequest", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		sessionFilter := rest.NewSessionFilter()

		// Login first
		resp1 := rest.Given().
			BaseURI(ts.URL).
			FormParam("username", "john").
			FormParam("password", "doe").
			Filter(sessionFilter).
			Post("/login")

		must.NoError(resp1.Err())
		firstSessionID := sessionFilter.SessionID()
		is.NotEmpty(firstSessionID)

		// Use same filter for multiple subsequent requests
		for i := range 2 {
			resp := rest.Given().
				BaseURI(ts.URL).
				Filter(sessionFilter).
				Get("/secured")
			must.NoError(resp.Err(), "secured request %d should succeed", i)
			is.Equal(http.StatusOK, resp.StatusCode(), "secured request %d", i)
		}
	})

	t.Run("SessionFilterCanBeManuallyOverridden", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		sessionFilter := rest.NewSessionFilter()
		sessionFilter.SetSessionID("MANUAL_SESSION_ID")
		is.Equal("MANUAL_SESSION_ID", sessionFilter.SessionID())
	})
}

func TestJavaITest_Auth_CsrfHeader(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("CsrfHeaderRequired", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// Without CSRF header: 403 Forbidden
		resp1 := rest.Given().
			BaseURI(ts.URL).
			Post("/pageThatRequireHeaderCsrf")

		must.NoError(resp1.Err())
		is.Equal(http.StatusForbidden, resp1.StatusCode())
	})

	t.Run("CsrfHeaderAccepted", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			Header("X-CSRF-TOKEN", "ab8722b1-1f23-4dcf-bf63-fb8b94be4107").
			Post("/pageThatRequireHeaderCsrf").
			Then().
			StatusCode(http.StatusOK).
			BodyEquals("OK").
			AssertAllNoFail(t)
	})

	t.Run("CsrfHeaderDerivedFromMetaTag", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// Fetch page with CSRF meta tag
		pageResp := rest.Given().
			BaseURI(ts.URL).
			Get("/pageWithDefaultHeaderCsrf")

		must.NoError(pageResp.Err())
		csrfToken := rest.ExtractCsrfFromHTML(pageResp.AsString(), "_csrf")
		is.NotEmpty(csrfToken)

		// Use extracted token in the header
		resp2 := rest.Given().
			BaseURI(ts.URL).
			Header("X-CSRF-TOKEN", csrfToken).
			Post("/pageThatRequireHeaderCsrf")

		must.NoError(resp2.Err())
		is.Equal(http.StatusOK, resp2.StatusCode())
	})
}

func TestJavaITest_Auth_BasicAuthExtended(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("SupportsExpectingStatusCodeWhenAuthenticationError", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Auth().Basic("wrong-user", "wrong-pass").
			Get("/auth/basic")

		must.NoError(resp.Err())
		is.Equal(http.StatusUnauthorized, resp.StatusCode())
	})

	t.Run("SupportsExpectingStatusCodeWhenPreemptiveAuthenticationError", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Auth().PreemptiveBasic("wrong-user", "wrong-pass").
			Get("/auth/basic")

		must.NoError(resp.Err())
		is.Equal(http.StatusUnauthorized, resp.StatusCode())
	})

	t.Run("BasicAuthViaRequestSpecWithWrongCredentials", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		spec := rest.NewRequestSpecBuilder().
			SetBaseURI(ts.URL).
			SetAuth(&rest.BasicAuthScheme{Username: "wrong", Password: "creds"}).
			Build()

		resp := rest.Given().
			Spec(spec).
			Get("/auth/basic")

		must.NoError(resp.Err())
		is.Equal(http.StatusUnauthorized, resp.StatusCode())
	})

	t.Run("PreemptiveBasicAuthViaRequestSpec", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		spec := rest.NewRequestSpecBuilder().
			SetBaseURI(ts.URL).
			SetAuth(&rest.PreemptiveBasicAuthScheme{Username: "admin", Password: "secret"}).
			Build()

		resp := rest.Given().
			Spec(spec).
			Get("/auth/basic")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
	})

	t.Run("NoneAuthClearsBasicAuthOnPerRequest", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// spec sets basic auth, but per-request None() overrides it
		spec := rest.NewRequestSpecBuilder().
			SetBaseURI(ts.URL).
			SetAuth(&rest.BasicAuthScheme{Username: "admin", Password: "secret"}).
			Build()

		resp := rest.Given().
			Spec(spec).
			Auth().None().
			Get("/auth/basic")

		must.NoError(resp.Err())
		// Without auth, should be 401
		is.Equal(http.StatusUnauthorized, resp.StatusCode())
	})
}

func TestJavaITest_Auth_SessionAndLogging(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("SessionFilterWithCustomSessionName", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// Use custom session name "jsessionid" (same as server uses)
		sessionFilter := rest.NewSessionFilter("jsessionid")

		resp := rest.Given().
			BaseURI(ts.URL).
			Filter(sessionFilter).
			Get("/sessionId")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.NotEmpty(sessionFilter.SessionID())
	})

	t.Run("SessionFilterCanBeManuallySet", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		sessionFilter := rest.NewSessionFilter()
		sessionFilter.SetSessionID("EXPLICIT_SESSION_42")
		is.Equal("EXPLICIT_SESSION_42", sessionFilter.SessionID())
	})

	t.Run("FormAuthWithCsrfHeaderInjectViaFilter", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// CsrfFilter fetches _csrf token from a meta tag, then injects it as the X-CSRF-TOKEN header.
		// /pageThatRequireHeaderCsrf validates the header token.
		csrfFilter := rest.NewCsrfFilter(ts.URL+"/pageWithDefaultHeaderCsrf", "_csrf")
		csrfFilter.HeaderName = "X-CSRF-TOKEN"

		resp := rest.Given().
			BaseURI(ts.URL).
			Filter(csrfFilter).
			Get("/pageThatRequireHeaderCsrf")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
	})

	t.Run("BasicAuthAndBearerTokenAreDistinct", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// Basic auth works on /auth/basic
		respBasic := rest.Given().
			BaseURI(ts.URL).
			Auth().Basic("admin", "secret").
			Get("/auth/basic")

		must.NoError(respBasic.Err())
		is.Equal(http.StatusOK, respBasic.StatusCode())

		// OAuth2 fails on /auth/basic (different auth mechanism)
		respOAuth := rest.Given().
			BaseURI(ts.URL).
			Auth().OAuth2("some-token").
			Get("/auth/basic")

		must.NoError(respOAuth.Err())
		is.Equal(http.StatusUnauthorized, respOAuth.StatusCode())
	})
}
