package integration_test

// Ported from OAuthITest.java and SSLITest.java.
// Java uses SignedOAuthRequest / OAuth1 libraries; Go maps to Bearer token (OAuth2).
// SSL tests use httptest.NewTLSServer() with InsecureSkipVerify.

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func TestJavaITest_OAuth_BearerToken(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("OAuth2BearerTokenSuccess", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			Auth().OAuth2("secret-token-123").
			Post("/auth/bearer").
			Then().
			StatusCode(http.StatusOK).
			Body("status", "authenticated").
			AssertAllNoFail(t)
	})

	t.Run("OAuth2BearerTokenFailure", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Auth().OAuth2("wrong-token").
			Post("/auth/bearer")

		is.Equal(http.StatusUnauthorized, resp.StatusCode())
	})

	t.Run("OAuth2BearerTokenViaSentInAuthorizationHeader", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// Manually setting Bearer token via header is equivalent to Auth().OAuth2()
		resp := rest.Given().
			BaseURI(ts.URL).
			Header("Authorization", "Bearer secret-token-123").
			Post("/auth/bearer")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal("authenticated", resp.JsonPath().GetString("status"))
	})

	t.Run("OAuth2BearerTokenViaRequestSpec", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		spec := rest.NewRequestSpecBuilder().
			SetBaseURI(ts.URL).
			SetAuth(&rest.OAuth2Scheme{AccessToken: "secret-token-123"}).
			Build()

		resp := rest.Given().
			Spec(spec).
			Post("/auth/bearer")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
	})

	t.Run("OAuth2WithEmptyTokenReturns401", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Auth().OAuth2("").
			Post("/auth/bearer")

		must.NoError(resp.Err())
		is.Equal(http.StatusUnauthorized, resp.StatusCode())
	})

	t.Run("OAuth2TokenReusedAcrossRequests", func(t *testing.T) {
		t.Parallel()
		must := require.New(t)

		spec := rest.NewRequestSpecBuilder().
			SetBaseURI(ts.URL).
			SetAuth(&rest.OAuth2Scheme{AccessToken: "secret-token-123"}).
			Build()

		for i := range 3 {
			resp := rest.Given().
				Spec(spec).
				Post("/auth/bearer")
			must.NoError(resp.Err(), "iteration %d", i)
			must.Equal(http.StatusOK, resp.StatusCode(), "iteration %d", i)
		}
	})
}

func TestJavaITest_SSL_TLSServer(t *testing.T) {
	t.Parallel()

	// httptest.NewTLSServer uses a self-signed cert; InsecureSkipVerify is required.
	tlsServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"tls":true,"secure":true}`))
	}))
	t.Cleanup(tlsServer.Close)

	t.Run("SSLWithInsecureSkipVerify", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(tlsServer.URL).
			RelaxedHTTPSValidation().
			Get("/")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.True(resp.JsonPath().GetBool("tls"))
		is.True(resp.JsonPath().GetBool("secure"))
	})

	t.Run("SSLWithInsecureSkipVerifyViaSpec", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		spec := rest.NewRequestSpecBuilder().
			SetBaseURI(tlsServer.URL).
			Build()
		spec.InsecureSkipVerify = true

		resp := rest.Given().
			Spec(spec).
			Get("/")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
	})

	t.Run("TLSServerWithAuthAndInsecureSkipVerify", func(t *testing.T) {
		t.Parallel()

		tlsAuthServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader != "Bearer tls-token" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"tls_auth":true}`))
		}))
		t.Cleanup(tlsAuthServer.Close)

		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(tlsAuthServer.URL).
			RelaxedHTTPSValidation().
			Auth().OAuth2("tls-token").
			Get("/")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.True(resp.JsonPath().GetBool("tls_auth"))
	})

	t.Run("TLSServerWithoutInsecureSkipVerifyReturnsError", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		// Self-signed cert without RelaxedHTTPSValidation will fail certificate validation
		resp := rest.Given().
			BaseURI(tlsServer.URL).
			Get("/")

		// Should return a connection/TLS error
		is.Error(resp.Err())
	})
}

func TestJavaITest_SSL_AdditionalScenarios(t *testing.T) {
	t.Parallel()

	tlsServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ssl":true,"path":"` + r.URL.Path + `"}`))
	}))
	t.Cleanup(tlsServer.Close)

	t.Run("SSLWithPostMethod", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(tlsServer.URL).
			RelaxedHTTPSValidation().
			Body(`{"test":true}`).
			ContentType(rest.ContentTypeJSON).
			Post("/secure")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.True(resp.JsonPath().GetBool("ssl"))
	})

	t.Run("SSLWithCustomHeaders", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(tlsServer.URL).
			RelaxedHTTPSValidation().
			Header("X-Custom", "tls-value").
			Get("/")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.True(resp.JsonPath().GetBool("ssl"))
	})

	t.Run("SSLConcurrentRequests", func(t *testing.T) {
		t.Parallel()
		must := require.New(t)

		for i := range 3 {
			resp := rest.Given().
				BaseURI(tlsServer.URL).
				RelaxedHTTPSValidation().
				Get("/")
			must.NoError(resp.Err(), "concurrent TLS request %d failed", i)
			must.Equal(http.StatusOK, resp.StatusCode(), "concurrent TLS request %d", i)
		}
	})

	t.Run("SSLSpecCanBeReused", func(t *testing.T) {
		t.Parallel()
		must := require.New(t)

		spec := rest.NewRequestSpecBuilder().
			SetBaseURI(tlsServer.URL).
			Build()
		spec.InsecureSkipVerify = true

		for i := range 2 {
			resp := rest.Given().
				Spec(spec).
				Get("/")
			must.NoError(resp.Err(), "TLS spec request %d", i)
			must.Equal(http.StatusOK, resp.StatusCode(), "TLS spec request %d", i)
		}
	})
}

func TestJavaITest_OAuth_InlineCustomServer(t *testing.T) {
	t.Parallel()

	// A custom server that validates Bearer tokens and returns user info
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		switch authHeader {
		case "Bearer token-alice":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"user":"alice","role":"admin"}`))
		case "Bearer token-bob":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"user":"bob","role":"viewer"}`))
		default:
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		}
	}))
	t.Cleanup(tokenServer.Close)

	t.Run("DifferentBearerTokensReturnDifferentUsers", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		respAlice := rest.Given().
			BaseURI(tokenServer.URL).
			Auth().OAuth2("token-alice").
			Get("/")

		must.NoError(respAlice.Err())
		is.Equal(http.StatusOK, respAlice.StatusCode())
		is.Equal("alice", respAlice.JsonPath().GetString("user"))
		is.Equal("admin", respAlice.JsonPath().GetString("role"))

		respBob := rest.Given().
			BaseURI(tokenServer.URL).
			Auth().OAuth2("token-bob").
			Get("/")

		must.NoError(respBob.Err())
		is.Equal(http.StatusOK, respBob.StatusCode())
		is.Equal("bob", respBob.JsonPath().GetString("user"))
		is.Equal("viewer", respBob.JsonPath().GetString("role"))
	})

	t.Run("InvalidTokenReturns401WithBody", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(tokenServer.URL).
			Auth().OAuth2("invalid-token-xyz").
			Get("/")

		must.NoError(resp.Err())
		is.Equal(http.StatusUnauthorized, resp.StatusCode())
		is.NotEmpty(resp.AsString())
	})
}
