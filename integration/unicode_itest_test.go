package integration_test

// Ported from UnicodeITest.java

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func TestJavaITest_Unicode_JSONBody(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("PureBodyContainsUnicodeCharacters", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/utf8-body-json")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Contains(resp.AsString(), "啊 ☆")
	})

	t.Run("JsonPathExtractsUnicodeValue", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/utf8-body-json")

		must.NoError(resp.Err())
		is.Equal("啊 ☆", resp.JsonPath().GetString("value"))
	})
}

func TestJavaITest_Unicode_XMLBody(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("XmlBodyContainsUnicodeCharacters", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/utf8-body-xml")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Contains(resp.AsString(), "啊 ☆")
	})

	t.Run("XmlPathExtractsUnicodeValue", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/utf8-body-xml")

		must.NoError(resp.Err())
		xp, err := resp.XmlPath()
		must.NoError(err)
		is.Equal("啊 ☆", xp.GetString("//value"))
	})
}

func TestJavaITest_Unicode_RequestBody(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("UnicodeInJSONRequestBodyIsPreserved", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			ContentType(rest.ContentTypeJSON).
			Body(`{"title":"äöüß€'"}`).
			Post("/reflect")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal("äöüß€'", resp.JsonPath().GetString("title"))
	})

	t.Run("ChineseCharactersRoundTripInJSON", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			ContentType(rest.ContentTypeJSON).
			Body(`{"value":"啊 ☆"}`).
			Post("/reflect")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal("啊 ☆", resp.JsonPath().GetString("value"))
	})
}
