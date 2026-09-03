package integration_test

// Ported from OptionsITest.java

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func TestJavaITest_Options(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("OptionsSupportsStringBody", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Body("a body").
			Options("/returnBodyAsBody")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal("a body", resp.AsString())
	})
}
