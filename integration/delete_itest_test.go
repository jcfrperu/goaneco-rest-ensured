package integration_test

// Ported from DeleteITest.java

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func TestJavaITest_Delete(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("RequestSpecificationAllowsSpecifyingCookieForDelete", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Cookie("username", "John").
			Cookie("token", "1234").
			Delete("/cookie")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		body := resp.AsString()
		is.Contains(body, "username")
		is.Contains(body, "token")
	})

	t.Run("BodyMatcherWithoutKeyForDelete", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			QueryParam("firstName", "John").
			QueryParam("lastName", "Doe").
			Delete("/greet")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal("Greetings John Doe", resp.JsonPath().GetString("greeting"))
	})

	t.Run("DeleteSupportsStringBody", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Body("a body").
			Delete("/body")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal("a body", resp.AsString())
	})
}
