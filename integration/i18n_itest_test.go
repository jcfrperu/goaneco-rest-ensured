package integration_test

// Ported from GivenWhenTheni18nITest.java

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func TestJavaITest_I18n(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("BodyPathWithNonASCIIKeyAndValue", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/i18n")

		must.NoError(resp.Err())
		is.Equal("Är ån", resp.JsonPath().GetString("ön"))
	})
}
