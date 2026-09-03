package integration_test

// Ported from BigDecimalITest.java and DoubleITest.java

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func TestJavaITest_BigDecimal_Amount(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("AmountIsReturnedAs250Point00", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/amount")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		// Java uses BigDecimal(250.00); Go returns float64
		is.InDelta(250.0, resp.JsonPath().Get("amount").Float(), 0.001)
	})

	t.Run("AmountMatchesExact250", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/amount")

		must.NoError(resp.Err())
		is.Equal(float64(250), resp.JsonPath().Get("amount").Float())
	})
}

func TestJavaITest_Double_Amount(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("DoubleWorks", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/amount")

		must.NoError(resp.Err())
		is.Equal(float64(250), resp.JsonPath().Get("amount").Float())
	})

	t.Run("CanUseCloseToMatcherForDoubles", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/amount")

		must.NoError(resp.Err())
		amount := resp.JsonPath().Get("amount").Float()
		is.InDelta(250.0, amount, 0.001)
	})
}

func TestJavaITest_NumberTypes_AnonymousList(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("AnonymousListContainsExpectedNumbers", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/anonymous_list_with_numbers")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())

		var numbers []float64
		must.NoError(resp.As(&numbers))
		is.Len(numbers, 3)
		is.Contains(numbers, float64(100))
		is.Contains(numbers, float64(50))
		is.Contains(numbers, float64(31))
	})

	t.Run("AnonymousListWithFloatCanUseCloseTo", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/anonymous_list_with_numbers")

		must.NoError(resp.Err())

		var numbers []float64
		must.NoError(resp.As(&numbers))
		is.InDelta(31.0, numbers[2], 0.001)
	})
}
