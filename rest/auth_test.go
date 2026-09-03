package rest_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func TestAuthenticationSchemes(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		switch r.URL.Path {
		case "/basic":
			u, p, ok := r.BasicAuth()
			if ok && u == "user1" && p == "pass1" {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"auth":"basic"}`))
				return
			}
			w.WriteHeader(http.StatusUnauthorized)
		case "/bearer":
			if strings.HasPrefix(authHeader, "Bearer ") && strings.TrimPrefix(authHeader, "Bearer ") == "token-xyz" {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"auth":"bearer"}`))
				return
			}
			w.WriteHeader(http.StatusUnauthorized)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(ts.Close)

	// 1. Basic Auth
	respBasic := rest.Given().
		BaseURI(ts.URL).
		Auth().Basic("user1", "pass1").
		Get("/basic")
	must.NoError(respBasic.Err())
	is.Equal(http.StatusOK, respBasic.StatusCode())

	// 2. Preemptive Basic Auth
	respPreemptive := rest.Given().
		BaseURI(ts.URL).
		Auth().PreemptiveBasic("user1", "pass1").
		Get("/basic")
	must.NoError(respPreemptive.Err())
	is.Equal(http.StatusOK, respPreemptive.StatusCode())

	// 3. OAuth 2.0 Bearer Token
	respOAuth := rest.Given().
		BaseURI(ts.URL).
		Auth().OAuth2("token-xyz").
		Get("/bearer")
	must.NoError(respOAuth.Err())
	is.Equal(http.StatusOK, respOAuth.StatusCode())

	// 4. Invalid credentials returns 401
	respInvalid := rest.Given().
		BaseURI(ts.URL).
		Auth().Basic("wrong", "credentials").
		Get("/basic")
	must.NoError(respInvalid.Err())
	is.Equal(http.StatusUnauthorized, respInvalid.StatusCode())
}

func TestNilAuthSchemes(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	var (
		nilBasic      *rest.BasicAuthScheme
		nilPreemptive *rest.PreemptiveBasicAuthScheme
		nilOAuth      *rest.OAuth2Scheme
		nilNoAuth     *rest.NoAuthScheme
	)

	req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	must.NoError(err)

	is.NoError(nilBasic.Authenticate(req))
	is.NoError(nilPreemptive.Authenticate(req))
	is.NoError(nilOAuth.Authenticate(req))
	is.NoError(nilNoAuth.Authenticate(req))

	var nilForm *rest.FormAuthScheme
	is.NoError(nilForm.Authenticate(req))
}

func TestFormAuth_BlankCredentials(t *testing.T) {
	t.Parallel()
	// Mirrors Java RestAssuredTest: Form() panics when username is blank or password is empty.

	// blank username (whitespace-only) → panic
	require.Panics(t, func() {
		rest.Form("    ", "somepass", nil)
	}, "expected panic for blank username")

	// empty username (equivalent of null in Java) → panic
	require.Panics(t, func() {
		rest.Form("", "somepass", nil)
	}, "expected panic for empty username")

	// empty password → panic
	require.Panics(t, func() {
		rest.Form("validuser", "", nil)
	}, "expected panic for empty password")

	// valid credentials → no panic
	require.NotPanics(t, func() {
		scheme := rest.Form("validuser", "validpass", nil)
		assert.NotNil(t, scheme)
	})
}

func TestFormAuthScheme(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	// Default config
	req := rest.Given().Auth().Form("myuser", "mypass")
	is.NotNil(req)

	// Spring security config
	req2 := rest.Given().Auth().Form("myuser", "mypass", rest.SpringSecurityFormAuth())
	is.NotNil(req2)

	// Direct FormAuthConfig
	customCfg := &rest.FormAuthConfig{
		FormAction:    "/custom-login",
		UsernameField: "u",
		PasswordField: "p",
		CsrfField:     "csrf_token",
		LogRequest:    true,
	}
	is.Equal("/custom-login", customCfg.FormAction)
	is.Equal("u", customCfg.UsernameField)
	is.Equal("p", customCfg.PasswordField)
	is.Equal("csrf_token", customCfg.CsrfField)
	is.True(customCfg.LogRequest)
}
