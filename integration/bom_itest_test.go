package integration_test

// Ported from BomITest.java

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func TestJavaITest_Bom(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("CanParseXMLWithBOM", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/xmlWithBom")

		must.NoError(resp.Err())
		xp, err := resp.XmlPath()
		must.NoError(err)
		is.Equal("build", xp.GetString("//project/target/@name"))
	})
}
