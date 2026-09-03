package integration_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func TestIntegration_Authentication(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	ts := integrationServer

	// 1. Basic Auth success
	respBasic := rest.Given().
		BaseURI(ts.URL).
		Auth().Basic("admin", "secret").
		Get("/auth/basic")

	must.NoError(respBasic.Err())
	is.Equal(http.StatusOK, respBasic.StatusCode())
	is.Equal("authenticated", respBasic.JsonPath().GetString("status"))

	// 2. Basic Auth failure
	respBasicFail := rest.Given().
		BaseURI(ts.URL).
		Auth().Basic("admin", "wrongpassword").
		Get("/auth/basic")

	is.Equal(http.StatusUnauthorized, respBasicFail.StatusCode())

	// 3. Bearer Token success
	respBearer := rest.Given().
		BaseURI(ts.URL).
		Auth().OAuth2("secret-token-123").
		Get("/auth/bearer")

	must.NoError(respBearer.Err())
	is.Equal(http.StatusOK, respBearer.StatusCode())

	// 4. Bearer Token failure
	respBearerFail := rest.Given().
		BaseURI(ts.URL).
		Auth().OAuth2("invalid-token").
		Get("/auth/bearer")

	is.Equal(http.StatusUnauthorized, respBearerFail.StatusCode())
}

func TestIntegration_FormAuthFilter_And_SessionPersistence(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	ts := integrationServer

	// Configure FormAuthFilter for /login with CSRF
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

	// Request to protected endpoint using the filter
	respSecured := rest.Given().
		BaseURI(ts.URL).
		Filter(formAuth).
		Get("/secured")

	must.NoError(respSecured.Err())
	is.Equal(http.StatusOK, respSecured.StatusCode())
	is.Equal("classified", respSecured.JsonPath().GetString("secret"))
	is.Equal("john", respSecured.JsonPath().GetString("user"))
}
