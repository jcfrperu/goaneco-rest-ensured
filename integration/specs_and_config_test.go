package integration_test

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func TestIntegration_SpecsAndRedirects(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	ts := integrationServer

	reqSpec := rest.NewRequestSpecBuilder().
		SetBaseURI(ts.URL).
		AddHeader("X-Integration", "SpecsTest").
		Build()

	respSpec := rest.NewResponseSpecBuilder().
		ExpectStatusCode(http.StatusOK).
		ExpectContentType(rest.ContentTypeJSON).
		ExpectResponseTimeLessThan(3 * time.Second).
		Build()

	// 1. Multi-hop redirect following to final 200 OK
	valid := rest.Given().
		Spec(reqSpec).
		Get("/redirect/step1").
		Then().
		Spec(respSpec).
		Body("status", "redirect_complete").
		Body("step", "final")

	must.False(valid.HasFailures())
	is.Empty(valid.Failures())

	// 2. Disabling redirect stops at 302
	noFollowCfg := rest.DefaultConfig().WithRedirect(rest.RedirectConfig{
		Follow: false,
	})

	respStop := rest.Given().
		Config(noFollowCfg).
		BaseURI(ts.URL).
		Get("/redirect/step1")

	must.NoError(respStop.Err())
	is.Equal(http.StatusFound, respStop.StatusCode())
}

func TestIntegration_Decompression_GzipAndDeflate(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	ts := integrationServer

	// 1. Transparent GZIP decoding
	respGzip := rest.Given().
		BaseURI(ts.URL).
		Get("/gzip/data")

	must.NoError(respGzip.Err())
	is.Equal(http.StatusOK, respGzip.StatusCode())
	is.True(respGzip.JsonPath().GetBool("compressed"))
	is.Equal("gzip", respGzip.JsonPath().GetString("codec"))
	is.Equal("integration-test", respGzip.JsonPath().GetString("payload"))

	// 2. Transparent Deflate decoding
	respDeflate := rest.Given().
		BaseURI(ts.URL).
		Get("/deflate/data")

	must.NoError(respDeflate.Err())
	is.Equal(http.StatusOK, respDeflate.StatusCode())
	is.True(respDeflate.JsonPath().GetBool("compressed"))
	is.Equal("deflate", respDeflate.JsonPath().GetString("codec"))
	is.Equal("integration-test", respDeflate.JsonPath().GetString("payload"))
}

func TestIntegration_RedirectSensitiveHeaderStripping(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	// Go's http.Client strips Authorization/Cookie headers when the redirect
	// target has a different hostname (not just different port on same IP).
	// We simulate cross-hostname by redirecting from 127.0.0.1 to "localhost".

	var authOnSink atomic.Value
	var cookieOnSink atomic.Value

	// Sink listens on 127.0.0.1; we'll redirect to it via "localhost" alias
	sink := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authOnSink.Store(r.Header.Get("Authorization"))
		cookieOnSink.Store(r.Header.Get("Cookie"))
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(sink.Close)

	// Extract port to build "localhost:PORT" cross-hostname redirect URL
	sinkPort := sink.Listener.Addr().(*net.TCPAddr).Port
	crossHostTarget := fmt.Sprintf("http://localhost:%d/leak", sinkPort)

	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, crossHostTarget, http.StatusFound)
	}))
	t.Cleanup(origin.Close)

	// Cross-hostname redirect: Go stdlib strips Authorization and Cookie
	resp := rest.Given().
		BaseURI(origin.URL).
		Header("Authorization", "Bearer secret").
		Cookie("session", "cookie-val").
		Get("/start")
	must.NoError(resp.Err())
	is.Equal(http.StatusOK, resp.StatusCode())
	is.Empty(authOnSink.Load())
	is.Empty(cookieOnSink.Load())

	// Same-host redirect (path-only Location): headers are preserved
	var authOnTarget atomic.Value
	sameHost := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/start" {
			http.Redirect(w, r, "/target", http.StatusFound)
			return
		}
		authOnTarget.Store(r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(sameHost.Close)

	resp2 := rest.Given().
		BaseURI(sameHost.URL).
		Header("Authorization", "Bearer secret").
		Get("/start")
	must.NoError(resp2.Err())
	is.Equal(http.StatusOK, resp2.StatusCode())
	is.Equal("Bearer secret", authOnTarget.Load())
}
