package integration_test

// Ported from DecoderConfigITest.java

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func TestJavaITest_DecoderConfig_ContentDecoders(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("ByDefaultGzipResponsesAreDecompressed", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// /gzip/data returns gzip-compressed JSON; default config auto-decompresses
		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/gzip/data")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.True(resp.JsonPath().GetBool("compressed"))
		is.Equal("gzip", resp.JsonPath().GetString("codec"))
	})

	t.Run("DeflateResponsesAreDecompressedByDefault", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// /deflate/data returns deflate-compressed JSON
		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/deflate/data")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.True(resp.JsonPath().GetBool("compressed"))
		is.Equal("deflate", resp.JsonPath().GetString("codec"))
	})

	t.Run("OnlyGzipDecoderCanBeConfigured", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		cfg := rest.DefaultConfig().WithDecoder(rest.DecoderConfig{
			ContentDecoders: []string{"gzip"},
		})

		resp := rest.Given().
			Config(cfg).
			BaseURI(ts.URL).
			Get("/gzip/data")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.True(resp.JsonPath().GetBool("compressed"))
	})

	t.Run("EmptyContentDecodersStillAllowsTransportLevelGzip", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// Go's http.Transport decompresses gzip transparently at the transport level.
		// Setting ContentDecoders to empty only disables our manual decompression
		// pass in executeHTTP; the transport-level decompression still happens.
		cfg := rest.DefaultConfig().WithDecoder(rest.DecoderConfig{
			ContentDecoders: []string{},
		})

		resp := rest.Given().
			Config(cfg).
			BaseURI(ts.URL).
			Get("/gzip-json")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		// Transport-level decompression means body is still readable
		is.NotEmpty(resp.AsBytes())
	})

	t.Run("JsonBodyWithUTF16DefaultCharsetIsAccessible", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// Default charset is UTF-8 for JSON; body should be accessible
		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal(float64(5), resp.JsonPath().Get("lotto.lottoId").Float())
	})
}
