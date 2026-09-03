package integration_test

// Ported from PutITest.java

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func TestJavaITest_Put(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("RequestSpecificationAllowsSpecifyingCookieForPut", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Cookie("username", "John").
			Cookie("token", "1234").
			Put("/cookie")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		body := resp.AsString()
		is.Contains(body, "username")
		is.Contains(body, "token")
	})

	t.Run("BodyMatcherWithoutKeyForPut", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			QueryParam("firstName", "John").
			QueryParam("lastName", "Doe").
			Put("/greetPut")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal("Greetings John Doe", resp.JsonPath().GetString("greeting"))
	})

	t.Run("PutSupportsBinaryBody", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		body := []byte("a body")
		resp := rest.Given().
			BaseURI(ts.URL).
			BodyBytes(body).
			Put("/binaryBody")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal("97, 32, 98, 111, 100, 121", resp.AsString())
	})

	t.Run("PutSupportsStringBody", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Body("a body").
			Put("/body")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal("a body", resp.AsString())
	})

	t.Run("PutWithFormParams", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			FormParam("firstName", "John").
			FormParam("lastName", "Doe").
			Put("/greetPut")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal("Greetings John Doe", resp.JsonPath().GetString("greeting"))
	})

	t.Run("PutSupportsMultiValueFormParameters", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			FormParam("list", "1", "2", "3").
			Put("/multiValueParam")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal("1,2,3", resp.JsonPath().GetString("list"))
	})
}
