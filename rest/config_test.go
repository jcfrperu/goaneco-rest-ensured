package rest_test

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func TestConfig_DeepCopyAndImmutability(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	cfg1 := rest.DefaultConfig()
	must.NotNil(cfg1)

	// 1. Encoder map deep copy
	cfg2 := cfg1.WithEncoder(rest.EncoderConfig{
		DefaultCharset: "UTF-8",
		ContentTypeCharsets: map[string]string{
			"application/json": "UTF-8",
		},
	})
	is.Equal("UTF-8", cfg2.EncoderConfig().ContentTypeCharsets["application/json"])
	is.Empty(cfg1.EncoderConfig().ContentTypeCharsets["application/json"])

	// 2. Log blacklist slice deep copy
	cfg3 := cfg1.WithLog(rest.LogConfig{
		BlacklistHeaders: []string{"X-Secret-1"},
	})
	is.Equal([]string{"X-Secret-1"}, cfg3.LogConfig().BlacklistHeaders)
	is.Equal([]string{"Authorization", "Cookie"}, cfg1.LogConfig().BlacklistHeaders)

	// 3. Decoder slice deep copy
	cfg4 := cfg1.WithDecoder(rest.DecoderConfig{
		ContentDecoders: []string{"custom-codec"},
	})
	is.Equal([]string{"custom-codec"}, cfg4.DecoderConfig().ContentDecoders)
	is.Equal([]string{"gzip", "deflate"}, cfg1.DecoderConfig().ContentDecoders)

	// 4. SSL, Redirect, Param, Csrf, Session configs
	cfg5 := cfg1.
		WithSSL(rest.SSLConfig{InsecureSkipVerify: true}).
		WithRedirect(rest.RedirectConfig{Follow: false, MaxCount: 3}).
		WithParam(rest.ParamConfig{EmptyParamsBehavior: "omit"}).
		WithCsrf(rest.CsrfConfig{TokenField: "custom_csrf"}).
		WithSession(rest.SessionConfig{SessionName: "MY_SESS"})

	is.True(cfg5.SSLConfig().InsecureSkipVerify)
	is.False(cfg5.RedirectConfig().Follow)
	is.Equal(3, cfg5.RedirectConfig().MaxCount)
	is.Equal(rest.EmptyParamsOmit, cfg5.ParamConfig().EmptyParamsBehavior)
	is.Equal("custom_csrf", cfg5.CsrfConfig().TokenField)
	is.Equal("MY_SESS", cfg5.SessionConfig().SessionName)

	// Original cfg1 remained unmodified
	is.False(cfg1.SSLConfig().InsecureSkipVerify)
	is.True(cfg1.RedirectConfig().Follow)
}

func TestConfig_GzipResponseDecompression(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var buf bytes.Buffer
		gw := gzip.NewWriter(&buf)
		_, _ = gw.Write([]byte(`{"decompressed":true,"engine":"gzip"}`))
		_ = gw.Close()

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Encoding", "gzip")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(buf.Bytes())
	}))
	t.Cleanup(ts.Close)

	resp := rest.Given().
		BaseURI(ts.URL).
		Get("/gzip-data")

	must.NoError(resp.Err())
	is.Equal(http.StatusOK, resp.StatusCode())
	is.Equal(`{"decompressed":true,"engine":"gzip"}`, resp.AsString())
	is.True(resp.JsonPath().GetBool("decompressed"))
	is.Equal("gzip", resp.JsonPath().GetString("engine"))
}

func TestConfig_DeflateResponseDecompression(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var buf bytes.Buffer
		fw, _ := flate.NewWriter(&buf, flate.DefaultCompression)
		_, _ = fw.Write([]byte(`{"compressed":true,"engine":"deflate"}`))
		_ = fw.Close()

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Encoding", "deflate")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(buf.Bytes())
	}))
	t.Cleanup(ts.Close)

	resp := rest.Given().
		BaseURI(ts.URL).
		Get("/deflate-data")

	must.NoError(resp.Err())
	is.Equal(http.StatusOK, resp.StatusCode())
	is.Equal(`{"compressed":true,"engine":"deflate"}`, resp.AsString())
	is.True(resp.JsonPath().GetBool("compressed"))
	is.Equal("deflate", resp.JsonPath().GetString("engine"))
}

func TestConfig_FailureListeners(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	var (
		listenerCalled   bool
		capturedFailures []string
	)

	cfg := rest.DefaultConfig().WithFailure(rest.FailureConfig{
		Listeners: []func(req *http.Request, resp *rest.Response, failures []string){
			func(req *http.Request, resp *rest.Response, failures []string) {
				listenerCalled = true
				capturedFailures = append([]string{}, failures...)
			},
		},
	})

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	t.Cleanup(ts.Close)

	resp := rest.Given().
		Config(cfg).
		BaseURI(ts.URL).
		Get("/test")

	// Trigger a failed assertion
	valResp := resp.Then().StatusCode(http.StatusBadRequest)
	is.False(valResp.IsValid())
	is.True(listenerCalled)
	is.NotEmpty(capturedFailures)
}

func TestConfig_ParamOmitBehavior(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	var capturedQuery string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(ts.Close)

	cfg := rest.DefaultConfig().WithParam(rest.ParamConfig{
		EmptyParamsBehavior: "omit",
	})

	resp := rest.Given().
		Config(cfg).
		BaseURI(ts.URL).
		QueryParam("present", "value123").
		QueryParam("empty", "").
		Get("/params")

	must.NoError(resp.Err())
	is.Equal(http.StatusOK, resp.StatusCode())
	is.Equal("present=value123", capturedQuery)

	// Test "error" behavior
	errCfg := rest.DefaultConfig().WithParam(rest.ParamConfig{
		EmptyParamsBehavior: "error",
	})
	respErr := rest.Given().
		Config(errCfg).
		BaseURI(ts.URL).
		QueryParam("empty_param", "").
		Get("/params")

	is.Error(respErr.Err())
	is.Contains(respErr.Err().Error(), "empty parameter \"empty_param\" is not allowed")
}

func TestConfig_RedirectPolicy(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "/destination", http.StatusFound)
			return
		}
		if r.URL.Path == "/destination" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`arrived`))
			return
		}
	}))
	t.Cleanup(ts.Close)

	// 1. FollowRedirects = false stops at 302
	noRedirectCfg := rest.DefaultConfig().WithRedirect(rest.RedirectConfig{
		Follow: false,
	})

	respNoFollow := rest.Given().
		Config(noRedirectCfg).
		BaseURI(ts.URL).
		Get("/redirect")

	must.NoError(respNoFollow.Err())
	is.Equal(http.StatusFound, respNoFollow.StatusCode())

	// 2. Default config follows redirect to 200 OK
	respFollow := rest.Given().
		BaseURI(ts.URL).
		Get("/redirect")

	must.NoError(respFollow.Err())
	is.Equal(http.StatusOK, respFollow.StatusCode())
	is.Equal("arrived", respFollow.AsString())
}

// TestConfig_SameHostRedirectKeepsSensitiveHeaders mirrors Java's
// RedirectSensitiveHeaderStrippingTest.keepsSensitiveHeadersOnSameHostRedirect.
// When a redirect stays on the same host, Authorization, Cookie, and Proxy-Authorization
// headers must be forwarded to the redirect target.
func TestConfig_SameHostRedirectKeepsSensitiveHeaders(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	var receivedAuth, receivedCookie, receivedProxy string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/start":
			http.Redirect(w, r, "/target", http.StatusFound)
		case "/target":
			receivedAuth = r.Header.Get("Authorization")
			receivedCookie = r.Header.Get("Cookie")
			receivedProxy = r.Header.Get("Proxy-Authorization")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		}
	}))
	t.Cleanup(ts.Close)

	resp := rest.Given().
		BaseURI(ts.URL).
		Header("Authorization", "Bearer secret").
		Header("Cookie", "session=cookie").
		Header("Proxy-Authorization", "Basic proxy").
		Get("/start")

	must.NoError(resp.Err())
	is.Equal(http.StatusOK, resp.StatusCode())
	is.Equal("Bearer secret", receivedAuth)
	is.Equal("session=cookie", receivedCookie)
	is.Equal("Basic proxy", receivedProxy)
}

func TestConfig_HeaderBehavior(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	var capturedHeaders http.Header
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeaders = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(ts.Close)

	// Request.Header() uses http.Header.Set() → same key overwrites (case-insensitive)
	resp := rest.Given().
		BaseURI(ts.URL).
		Header("X-Version", "first").
		Header("x-version", "second").
		Get("/")
	must.NoError(resp.Err())
	is.Equal([]string{"second"}, capturedHeaders.Values("X-Version"))

	// ContentType() is overwritable — last call wins
	resp2 := rest.Given().
		BaseURI(ts.URL).
		ContentType(rest.ContentTypeJSON).
		ContentType(rest.ContentTypeXML).
		Get("/")
	must.NoError(resp2.Err())
	is.Equal(string(rest.ContentTypeXML), capturedHeaders.Get("Content-Type"))

	// RequestSpecBuilder.AddHeader() uses http.Header.Add() → accumulates values for the same key
	spec := rest.NewRequestSpecBuilder().
		AddHeader("Accept-Language", "en").
		AddHeader("Accept-Language", "fr").
		Build()
	langs := spec.Headers.Values("Accept-Language")
	is.Len(langs, 2)
	is.Contains(langs, "en")
	is.Contains(langs, "fr")

	// http.Header is always case-insensitive — Get normalises the key
	spec2 := rest.NewRequestSpecBuilder().
		AddHeader("X-Api-Key", "key-123").
		Build()
	is.Equal("key-123", spec2.Headers.Get("x-api-key"))
	is.Equal("key-123", spec2.Headers.Get("X-API-KEY"))

	// BlacklistHeaders defaults contain Authorization and Cookie
	defaultBlacklist := rest.DefaultConfig().LogConfig().BlacklistHeaders
	is.Contains(defaultBlacklist, "Authorization")
	is.Contains(defaultBlacklist, "Cookie")

	// Custom blacklist replaces the default list
	cfg := rest.DefaultConfig().WithLog(rest.LogConfig{
		BlacklistHeaders: []string{"X-Api-Key", "X-Secret"},
	})
	is.Contains(cfg.LogConfig().BlacklistHeaders, "X-Api-Key")
	is.Contains(cfg.LogConfig().BlacklistHeaders, "X-Secret")
	is.NotContains(cfg.LogConfig().BlacklistHeaders, "Authorization")
}

func TestConfig_HTTPClientDefaults(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	httpCfg := rest.DefaultConfig().HTTPClientConfig()
	is.Equal(30*time.Second, httpCfg.Timeout)
	is.True(httpCfg.FollowRedirects)
	is.True(httpCfg.ReuseClient)
	is.False(httpCfg.DisableKeepAlive)
	is.Nil(httpCfg.TLSConfig)
	is.Nil(httpCfg.CustomClient)
	is.False(httpCfg.InsecureSkipVerify)
}

func TestConfig_HTTPClientCustomConfig(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	// Setting custom values is preserved
	cfg := rest.DefaultConfig().WithHTTPClient(rest.HTTPClientConfig{
		Timeout:          5 * time.Second,
		FollowRedirects:  false,
		DisableKeepAlive: true,
		MaxIdleConns:     10,
		ReuseClient:      false,
	})
	httpCfg := cfg.HTTPClientConfig()
	is.Equal(5*time.Second, httpCfg.Timeout)
	is.False(httpCfg.FollowRedirects)
	is.True(httpCfg.DisableKeepAlive)
	is.Equal(10, httpCfg.MaxIdleConns)
	is.False(httpCfg.ReuseClient)

	// CustomClient injection
	customClient := &http.Client{Timeout: 15 * time.Second}
	cfg2 := rest.DefaultConfig().WithHTTPClient(rest.HTTPClientConfig{
		CustomClient: customClient,
	})
	is.Equal(customClient, cfg2.HTTPClientConfig().CustomClient)

	// Timeout applied: short deadline causes request failure
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(ts.Close)

	timeoutCfg := rest.DefaultConfig().WithHTTPClient(rest.HTTPClientConfig{
		Timeout: 20 * time.Millisecond,
	})
	resp := rest.Given().Config(timeoutCfg).BaseURI(ts.URL).Get("/slow")
	must.Error(resp.Err())
}

// ── HeaderConfigTest ────────────────────────────────────────────────────────

func TestHeaderConfig_OverwriteAndMerge(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	// shouldOverwriteHeaderWithName returns true for explicitly overwritable header
	cfg := rest.NewHeaderConfig().OverwriteHeadersWithName("header")
	is.True(cfg.ShouldOverwriteHeaderWithName("header"))

	// returns false when a different header is overwritable
	cfg2 := rest.NewHeaderConfig().OverwriteHeadersWithName("header2")
	is.False(cfg2.ShouldOverwriteHeaderWithName("header"))

	// case-insensitive: "HeadEr" and "header" resolve to the same entry
	cfg3 := rest.NewHeaderConfig().OverwriteHeadersWithName("HeadEr")
	is.True(cfg3.ShouldOverwriteHeaderWithName("header"))

	// multiple names in one call all become overwritable
	cfg4 := rest.NewHeaderConfig().OverwriteHeadersWithName("Header2", "header2", "heaDer1")
	is.True(cfg4.ShouldOverwriteHeaderWithName("header2"))
	is.True(cfg4.ShouldOverwriteHeaderWithName("Header1"))

	// content-type is overwritable by default
	is.True(rest.NewHeaderConfig().ShouldOverwriteHeaderWithName("content-type"))

	// accept is overwritable by default
	is.True(rest.NewHeaderConfig().ShouldOverwriteHeaderWithName("accept"))

	// accept becomes mergeable when explicitly configured
	cfg5 := rest.NewHeaderConfig().MergeHeadersWithName("Accept")
	is.False(cfg5.ShouldOverwriteHeaderWithName("accept"))

	// content-type becomes mergeable when explicitly configured
	cfg6 := rest.NewHeaderConfig().MergeHeadersWithName("Content-type")
	is.False(cfg6.ShouldOverwriteHeaderWithName("content-type"))

	// mergeHeadersWithName works for multiple names
	cfg7 := rest.NewHeaderConfig().MergeHeadersWithName("Content-type", "accept")
	is.False(cfg7.ShouldOverwriteHeaderWithName("content-type"))
	is.False(cfg7.ShouldOverwriteHeaderWithName("Accept"))
}

func TestHeaderConfig_ImmutabilityAndConfigIntegration(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	// OverwriteHeadersWithName does not mutate the original
	base := rest.NewHeaderConfig()
	derived := base.OverwriteHeadersWithName("x-custom")
	is.True(derived.ShouldOverwriteHeaderWithName("x-custom"))
	is.False(base.ShouldOverwriteHeaderWithName("x-custom"))

	// HeaderConfig can be stored in Config and retrieved
	cfg := rest.DefaultConfig().WithHeaderConfig(
		rest.NewHeaderConfig().MergeHeadersWithName("Accept"),
	)
	is.False(cfg.HeaderConfig().ShouldOverwriteHeaderWithName("accept"))
	is.True(cfg.HeaderConfig().ShouldOverwriteHeaderWithName("content-type"))

	// nil Config returns default HeaderConfig
	var nilCfg *rest.Config
	is.NotNil(nilCfg.HeaderConfig())
	is.True(nilCfg.HeaderConfig().ShouldOverwriteHeaderWithName("content-type"))
}

// ── HttpClientConfigTest ─────────────────────────────────────────────────────

func TestHTTPClientConfig_ParamsAndReuse(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	// Default HTTPClientConfig has no Params (Go has no Apache-style cookie policy by default)
	defaultCfg := rest.HTTPClientConfig{}
	is.Empty(defaultCfg.Params)

	// SetParam accumulates; multiple calls with different keys preserve all entries
	cfg := rest.HTTPClientConfig{}.
		SetParam("max_redirects", "100").
		ReuseHTTPClient().
		SetParam("cookie_policy", "browser")
	is.Equal("100", cfg.Params["max_redirects"])
	is.Equal("browser", cfg.Params["cookie_policy"])
	is.True(cfg.ReuseClient)

	// SetParam with same key: later value wins
	cfg2 := rest.HTTPClientConfig{}.
		SetParam("max_redirects", "50").
		SetParam("max_redirects", "100")
	is.Equal("100", cfg2.Params["max_redirects"])

	// AddParams (bulk) preserves other configuration
	cfg3 := rest.HTTPClientConfig{}.
		AddParams(map[string]any{"max_redirects": "100"}).
		ReuseHTTPClient().
		AddParams(map[string]any{"cookie_policy": "browser"})
	is.Equal("100", cfg3.Params["max_redirects"])
	is.Equal("browser", cfg3.Params["cookie_policy"])
	is.True(cfg3.ReuseClient)

	// AddParams: later map overwrites same key from earlier map
	cfg4 := rest.HTTPClientConfig{}.
		AddParams(map[string]any{"max_redirects": "50"}).
		ReuseHTTPClient().
		AddParams(map[string]any{"max_redirects": "100"})
	is.Equal("100", cfg4.Params["max_redirects"])
}

// ── GZIPDecompressingEntityTest ──────────────────────────────────────────────

func TestGzip_DirectDecompression(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	// Mirrors Java GZIPDecompressingEntityTest.returns_gzipped_decompressed_content_when_content_length_is_minus_one.
	// Go's compress/gzip.Reader handles unknown-length streams the same way.
	input := `{"userId":"e047379","firstName":"Ninju","lastName":"BohraB","cookieNotice":true}`

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, err := gw.Write([]byte(input))
	must.NoError(err)
	must.NoError(gw.Close())

	gr, err := gzip.NewReader(&buf)
	must.NoError(err)
	decompressed, err := io.ReadAll(gr)
	must.NoError(err)
	is.Equal(input, string(decompressed))

	// Mirrors Java GZIPDecompressingEntityTest.should_not_fail_on_empty_response_with_gzip.
	// An empty gzip stream decompresses to empty bytes without error.
	var emptyBuf bytes.Buffer
	gw2 := gzip.NewWriter(&emptyBuf)
	must.NoError(gw2.Close())

	gr2, err := gzip.NewReader(&emptyBuf)
	must.NoError(err)
	emptyOut, err := io.ReadAll(gr2)
	must.NoError(err)
	is.Empty(emptyOut)
}

func TestConfig_NilConfigSafety(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	var nilCfg *rest.Config

	is.False(nilCfg.SSLConfig().InsecureSkipVerify)
	is.Equal(30*time.Second, nilCfg.HTTPClientConfig().Timeout)
	is.True(nilCfg.LogConfig().EnablePrettyPrinting)
	is.Equal("UTF-8", nilCfg.EncoderConfig().DefaultCharset)
	is.Equal("UTF-8", nilCfg.DecoderConfig().DefaultCharset)
	is.True(nilCfg.RedirectConfig().Follow)
	is.Empty(nilCfg.FailureConfig().Listeners)
	is.Equal(rest.EmptyParamsInclude, nilCfg.ParamConfig().EmptyParamsBehavior)
	is.Equal("_csrf", nilCfg.CsrfConfig().TokenField)
	is.Equal("JSESSIONID", nilCfg.SessionConfig().SessionName)

	clone := nilCfg.Clone()
	is.NotNil(clone)
}

// TestConfig_CrossHostRedirectStripsSensitiveHeadersByDefault mirrors Java's
// RedirectSensitiveHeaderStrippingTest.stripsSensitiveHeadersOnCrossHostRedirectByDefault.
// Go's net/http client strips Authorization and Cookie headers on cross-host redirects
// (different hostname). Proxy-Authorization is NOT stripped by Go stdlib (unlike Java).
func TestConfig_CrossHostRedirectStripsSensitiveHeadersByDefault(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	var authOnSink, cookieOnSink string

	// Sink server receives the redirected request
	sink := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authOnSink = r.Header.Get("Authorization")
		cookieOnSink = r.Header.Get("Cookie")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	t.Cleanup(sink.Close)

	// Origin server redirects to a different hostname (127.0.0.1 → localhost).
	// Go's CheckRedirect strips Authorization/Cookie for different-host targets.
	sinkURL := "http://localhost" + sink.Listener.Addr().String()[len("127.0.0.1"):]
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, sinkURL+"/leak", http.StatusFound)
	}))
	t.Cleanup(origin.Close)

	resp := rest.Given().
		BaseURI(origin.URL).
		Header("Authorization", "Bearer secret").
		Header("Cookie", "session=cookie").
		Get("/start")
	must.NoError(resp.Err())
	is.Equal(http.StatusOK, resp.StatusCode())

	// Authorization and Cookie must be stripped on cross-host redirect
	is.Empty(authOnSink)
	is.Empty(cookieOnSink)
}

// TestConfig_CrossHostRedirectForwardsHeadersWhenStrippingDisabled mirrors Java's
// RedirectSensitiveHeaderStrippingTest.forwardsSensitiveHeadersOnCrossHostRedirectWhenStrippingIsDisabled.
// Go does not have a built-in "disable sensitive header stripping" config. The equivalent
// is to use a custom http.Client with a CheckRedirect function that re-injects the headers.
// This test documents that by using a custom client the caller can preserve headers.
func TestConfig_CrossHostRedirectForwardsHeadersWhenStrippingDisabled(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	var authOnSink, cookieOnSink string

	sink := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authOnSink = r.Header.Get("Authorization")
		cookieOnSink = r.Header.Get("Cookie")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	t.Cleanup(sink.Close)

	sinkURL := "http://localhost" + sink.Listener.Addr().String()[len("127.0.0.1"):]
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, sinkURL+"/leak", http.StatusFound)
	}))
	t.Cleanup(origin.Close)

	// Custom client with a CheckRedirect that copies sensitive headers from prior requests
	customClient := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) > 0 {
				for _, hdr := range []string{"Authorization", "Cookie", "Proxy-Authorization"} {
					if val := via[0].Header.Get(hdr); val != "" {
						req.Header.Set(hdr, val)
					}
				}
			}
			return nil
		},
	}

	cfg := rest.DefaultConfig().WithHTTPClient(rest.HTTPClientConfig{
		CustomClient: customClient,
	})

	resp := rest.Given().
		Config(cfg).
		BaseURI(origin.URL).
		Header("Authorization", "Bearer secret").
		Header("Cookie", "session=cookie").
		Get("/start")
	must.NoError(resp.Err())
	is.Equal(http.StatusOK, resp.StatusCode())

	// With stripping disabled via custom client, sensitive headers must be forwarded
	is.Equal("Bearer secret", authOnSink)
	is.Equal("session=cookie", cookieOnSink)
}

// TestConfig_CrossHostRedirectStripsOAuth2Header mirrors Java's
// RedirectSensitiveHeaderStrippingTest.stripsAuthInterceptorInjectedHeaderOnCrossHostRedirect.
// Java's OAuthSigner interceptor injects Authorization on every hop, so the stripper must run
// after it. In Go, Auth().OAuth2() injects the header before the first request.
// Go's net/http strips Authorization on cross-host redirects (different hostname). We trigger
// this by redirecting from 127.0.0.1 to localhost — same approach as the other redirect tests.
func TestConfig_CrossHostRedirectStripsOAuth2Header(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	var authOnSink string

	sink := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authOnSink = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	t.Cleanup(sink.Close)

	// Rewrite the sink URL from "127.0.0.1:PORT" to "localhost:PORT" so Go's http.Client
	// considers the redirect a cross-host hop and strips Authorization/Cookie.
	sinkURL := "http://localhost" + sink.Listener.Addr().String()[len("127.0.0.1"):]

	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, sinkURL+"/leak", http.StatusFound)
	}))
	t.Cleanup(origin.Close)

	resp := rest.Given().
		BaseURI(origin.URL).
		Auth().OAuth2("interceptor-injected-token").
		Get("/start")

	must.NoError(resp.Err())
	is.Equal(http.StatusOK, resp.StatusCode())

	// Authorization injected by OAuth2 must be stripped on cross-host redirect
	is.Empty(authOnSink, "OAuth2-injected Authorization header must not leak to cross-host redirect target")
}
