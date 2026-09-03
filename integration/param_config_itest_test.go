package integration_test

// Ported from ParamConfigITest.java
// Note: Java's REPLACE strategy is not implemented in Go; only merge tests are ported.

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func TestJavaITest_ParamConfig_MergeByDefault(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("MergesQueryParamsByDefault", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			QueryParam("list", "value1").
			QueryParam("list", "value2").
			Get("/multiValueParam")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal("value1,value2", resp.JsonPath().GetString("list"))
	})

	t.Run("MergesFormParamsByDefault", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			FormParam("list", "value1").
			FormParam("list", "value2").
			Post("/multiValueParam")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal("value1,value2", resp.JsonPath().GetString("list"))
	})

	t.Run("MultipleFormParamValuesAreMerged", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			FormParam("list", "value1").
			FormParam("list", "value2").
			FormParam("list2", "value3").
			FormParam("list2", "value4").
			FormParam("list3", "value5").
			FormParam("list3", "value6").
			Post("/threeMultiValueParam")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal("value1,value2", resp.JsonPath().GetString("list"))
		is.Equal("value3,value4", resp.JsonPath().GetString("list2"))
		is.Equal("value5,value6", resp.JsonPath().GetString("list3"))
	})
}
