package integration_test

// Ported from GivenWhenThenExtractITest.java

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func TestJavaITest_GivenWhenThenExtract(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("ExtractResponseAsStringWorks", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/hello")

		must.NoError(resp.Err())
		body := resp.Then().Extract().AsString()
		is.Contains(body, "message")
		is.Contains(body, "Hello World")
	})

	t.Run("ExtractSinglePathWorks", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/hello")

		must.NoError(resp.Err())
		msg := resp.Then().Extract().PathString("message")
		is.Equal("Hello World", msg)
	})

	t.Run("ExtractEntireResponseWorks", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/hello")

		must.NoError(resp.Err())
		extracted := resp.Then().Extract().Response()
		is.NotEmpty(extracted.Header("Content-Type"))
	})

	t.Run("ExtractSinglePathWorksAfterBodyValidation", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/hello")

		must.NoError(resp.Err())
		valid := resp.Then().Body("message", "Hello World")
		is.False(valid.HasFailures(), valid.Failures())
		msg := valid.Extract().PathString("message")
		is.Equal("Hello World", msg)
	})

	t.Run("ExtractSinglePathWorksAfterStatusCodeAndBodyValidation", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/hello")

		must.NoError(resp.Err())
		valid := resp.Then().
			StatusCode(http.StatusOK).
			Body("message", "Hello World")
		is.False(valid.HasFailures(), valid.Failures())
		msg := valid.Extract().PathString("message")
		is.Equal("Hello World", msg)
	})

	t.Run("ExtractSinglePathWorksAfterMultipleBodyValidations", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto")

		must.NoError(resp.Err())
		valid := resp.Then().
			StatusCode(http.StatusOK).
			Body("lotto.lottoId", float64(5))
		is.False(valid.HasFailures(), valid.Failures())
		lottoId := valid.Extract().Path("lotto.lottoId").Int()
		is.Equal(int64(5), lottoId)
	})

	t.Run("ExtractProductsListWorks", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/products")

		must.NoError(resp.Err())
		is.Equal(float64(2), resp.JsonPath().Get("0.id").Float())
		is.Equal("An ice sculpture", resp.JsonPath().GetString("0.name"))
		is.Equal(float64(12.5), resp.JsonPath().Get("0.price").Float())
		is.Equal(float64(3), resp.JsonPath().Get("1.id").Float())
		is.Equal("A blue mouse", resp.JsonPath().GetString("1.name"))
		is.Equal(float64(25.5), resp.JsonPath().Get("1.price").Float())
	})

	t.Run("ExtractProductDimensionsWorks", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/products")

		must.NoError(resp.Err())
		is.InDelta(7.0, resp.JsonPath().Get("0.dimensions.length").Float(), 0.001)
		is.InDelta(12.0, resp.JsonPath().Get("0.dimensions.width").Float(), 0.001)
		is.InDelta(9.5, resp.JsonPath().Get("0.dimensions.height").Float(), 0.001)
		is.InDelta(3.1, resp.JsonPath().Get("1.dimensions.length").Float(), 0.001)
		is.InDelta(1.0, resp.JsonPath().Get("1.dimensions.width").Float(), 0.001)
		is.InDelta(1.0, resp.JsonPath().Get("1.dimensions.height").Float(), 0.001)
	})
}
