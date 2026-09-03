package integration_test

// Ported from HttpClientConfigITest.java

import (
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

type countingFilter struct {
	count *int32
}

func (f *countingFilter) Filter(req *rest.FilterableRequest, ctx *rest.FilterContext) (*rest.Response, error) {
	atomic.AddInt32(f.count, 1)
	return ctx.Next(req)
}

func TestJavaITest_HttpClientConfig_DefaultHeaders(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("CustomHeadersViaNonStaticConfigAreVisible", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Header("header1", "value1").
			Header("header2", "value2").
			Get("/multiHeaderReflect")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal("value1", resp.Header("Header1"))
		is.Equal("value2", resp.Header("Header2"))
	})
}

func TestJavaITest_HttpClientConfig_ClientReuse(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("ReuseHTTPClientSucceedsBothRequests", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		cfg := rest.DefaultConfig().WithHTTPClient(
			rest.DefaultConfig().HTTPClientConfig().ReuseHTTPClient(),
		)

		resp1 := rest.Given().
			Config(cfg).
			BaseURI(ts.URL).
			Get("/lotto")

		must.NoError(resp1.Err())
		is.Equal(http.StatusOK, resp1.StatusCode())

		resp2 := rest.Given().
			Config(cfg).
			BaseURI(ts.URL).
			Header("name", "value").
			Post("/reflect")

		must.NoError(resp2.Err())
		is.Equal(http.StatusOK, resp2.StatusCode())
	})

	t.Run("NonReuseClientSucceedsBothRequests", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// Default config: no reuse — each request gets a fresh client
		resp1 := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto")

		must.NoError(resp1.Err())
		is.Equal(http.StatusOK, resp1.StatusCode())

		resp2 := rest.Given().
			BaseURI(ts.URL).
			QueryParam("firstName", "Jane").
			QueryParam("lastName", "Doe").
			Get("/greet")

		must.NoError(resp2.Err())
		is.Equal(http.StatusOK, resp2.StatusCode())
		is.Equal("Greetings Jane Doe", resp2.JsonPath().GetString("greeting"))
	})
}

func TestJavaITest_HttpClientConfig_FilterCanObserveConfig(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("FilterCanInterceptAndModifyRequest", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var filteredCount int32
		captureFilter := &countingFilter{count: &filteredCount}

		resp := rest.Given().
			BaseURI(ts.URL).
			Filter(captureFilter).
			Get("/lotto")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal(int32(1), atomic.LoadInt32(&filteredCount))
	})
}
