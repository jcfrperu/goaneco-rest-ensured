package integration_test

// Ported from PatchITest.java

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func TestJavaITest_Patch(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("RequestSpecificationAllowsSpecifyingCookieForPatch", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Cookie("username", "John").
			Cookie("token", "1234").
			Patch("/cookie")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		body := resp.AsString()
		is.Contains(body, "username")
		is.Contains(body, "token")
	})

	t.Run("BodyMatcherWithoutKeyForPatch", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			QueryParam("firstName", "John").
			QueryParam("lastName", "Doe").
			Patch("/greetPatch")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal("Greetings John Doe", resp.JsonPath().GetString("greeting"))
	})

	t.Run("PatchSupportsBinaryBody", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		body := []byte("a body")
		resp := rest.Given().
			BaseURI(ts.URL).
			BodyBytes(body).
			Patch("/binaryBody")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal("97, 32, 98, 111, 100, 121", resp.AsString())
	})

	t.Run("PatchSupportsStringBody", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Body("a body").
			Patch("/body")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal("a body", resp.AsString())
	})

	t.Run("PatchWithFormParams", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			FormParam("firstName", "John").
			FormParam("lastName", "Doe").
			Patch("/greetPatch")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal("Greetings John Doe", resp.JsonPath().GetString("greeting"))
	})

	t.Run("PatchSupportsMultiValueFormParameters", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			FormParam("list", "a", "b", "c").
			Patch("/multiValueParam")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal("a,b,c", resp.JsonPath().GetString("list"))
	})

	t.Run("CanUseMapAsBodyToPatch", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		greeting := map[string]string{
			"firstName": "John",
			"lastName":  "Doe",
		}
		bodyBytes, err := json.Marshal(greeting)
		must.NoError(err)

		resp := rest.Given().
			BaseURI(ts.URL).
			ContentType(rest.ContentTypeJSON).
			BodyBytes(bodyBytes).
			Patch("/jsonGreet")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal("John Doe", resp.JsonPath().GetString("fullName"))
	})
}
