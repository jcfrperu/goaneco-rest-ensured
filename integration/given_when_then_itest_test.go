package integration_test

// Ported from GivenWhenThenITest.java

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func TestJavaITest_GivenWhenThen(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("SimpleGivenWhenThenWorks", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			QueryParam("firstName", "John").
			QueryParam("lastName", "Doe").
			Get("/greet")

		must.NoError(resp.Err())
		valid := resp.Then().
			StatusCode(http.StatusOK).
			Body("greeting", "Greetings John Doe")

		is.False(valid.HasFailures(), valid.Failures())
	})

	t.Run("GivenWhenThenWorksWithXPathAssertions", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			QueryParam("firstName", "John").
			QueryParam("lastName", "Doe").
			Get("/greetXML")

		must.NoError(resp.Err())
		xp, err := resp.XmlPath()
		must.NoError(err)
		is.Equal("John", xp.GetString("//greeting/firstName"))
	})

	t.Run("GivenWhenThenWorksWithMultipleBodyAssertions", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			QueryParam("firstName", "John").
			QueryParam("lastName", "Doe").
			Get("/greetXML")

		must.NoError(resp.Err())
		body := resp.AsString()
		is.Contains(body, "greeting")
		is.Contains(body, "John")
	})
}
