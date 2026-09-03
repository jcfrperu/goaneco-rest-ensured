package rest_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestGlobalConfiguration(t *testing.T) {
	is := assert.New(t)

	// Ensure clean state before and after
	rest.Reset()
	defer rest.Reset()

	// Initial defaults
	is.Equal("http://localhost", rest.GetBaseURI())
	is.Equal(0, rest.GetPort())
	is.Equal("", rest.GetBasePath())
	is.Equal("", rest.GetRootPath())
	is.True(rest.IsURLEncodingEnabled())

	// Mutate global state
	rest.BaseURI("https://api.github.com")
	rest.Port(443)
	rest.BasePath("/v3")
	rest.RootPath("data.items")
	rest.URLEncodingEnabled(false)

	is.Equal("https://api.github.com", rest.GetBaseURI())
	is.Equal(443, rest.GetPort())
	is.Equal("/v3", rest.GetBasePath())
	is.Equal("data.items", rest.GetRootPath())
	is.False(rest.IsURLEncodingEnabled())

	// Reset restores defaults
	rest.Reset()
	is.Equal("http://localhost", rest.GetBaseURI())
	is.Equal(0, rest.GetPort())
	is.Equal("", rest.GetBasePath())
	is.Equal("", rest.GetRootPath())
	is.True(rest.IsURLEncodingEnabled())
}

func TestGlobalStateConcurrency(t *testing.T) {
	rest.Reset()
	defer rest.Reset()

	const (
		numWorkers = 10
		iterations = 100
	)
	var wg sync.WaitGroup
	for i := range numWorkers {
		wg.Go(func() {
			for j := range iterations {
				if i%2 == 0 {
					rest.BaseURI("http://localhost")
					rest.Port(8080 + (j % 10))
					rest.BasePath("/api")
				} else {
					_ = rest.GetBaseURI()
					_ = rest.GetPort()
					_ = rest.GetBasePath()
					_ = rest.Given()
				}
			}
		})
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(5 * time.Second):
		t.Fatal("concurrent global state access timed out (possible deadlock)")
	}
}

func TestEntryPoints(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	req1 := rest.Given()
	is.NotNil(req1)

	req2 := rest.When()
	is.NotNil(req2)

	req3 := rest.With()
	is.NotNil(req3)
}

func TestGlobalSettersAndHTTPShortcuts(t *testing.T) {
	is := assert.New(t)
	must := require.New(t)

	rest.Reset()
	defer rest.Reset()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Received-Header", r.Header.Get("X-Global-Header"))
		w.Header().Set("X-Received-Method", r.Method)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"method":"` + r.Method + `"}`))
	}))
	t.Cleanup(ts.Close)

	rest.BaseURI(ts.URL)
	rest.GlobalHeader("X-Global-Header", "GlobalValue123")
	rest.GlobalHeaders(map[string]string{"X-Another": "Val"})
	rest.GlobalCookie("global_cookie", "cookie_val")
	rest.GlobalAuth(&rest.NoAuthScheme{})
	rest.GlobalConfig(rest.DefaultConfig())

	// Test all HTTP shortcut functions
	respGet := rest.Get("/test-get")
	must.NoError(respGet.Err())
	is.Equal(http.StatusOK, respGet.StatusCode())
	is.Equal("GlobalValue123", respGet.Header("X-Received-Header"))
	is.Equal("GET", respGet.Header("X-Received-Method"))

	respPost := rest.Post("/test-post")
	must.NoError(respPost.Err())
	is.Equal("POST", respPost.Header("X-Received-Method"))

	respPut := rest.Put("/test-put")
	must.NoError(respPut.Err())
	is.Equal("PUT", respPut.Header("X-Received-Method"))

	respDelete := rest.Delete("/test-delete")
	must.NoError(respDelete.Err())
	is.Equal("DELETE", respDelete.Header("X-Received-Method"))

	respHead := rest.Head("/test-head")
	must.NoError(respHead.Err())
	is.Equal("HEAD", respHead.Header("X-Received-Method"))

	respPatch := rest.Patch("/test-patch")
	must.NoError(respPatch.Err())
	is.Equal("PATCH", respPatch.Header("X-Received-Method"))

	respOptions := rest.Options("/test-options")
	must.NoError(respOptions.Err())
	is.Equal("OPTIONS", respOptions.Header("X-Received-Method"))
}

func TestGlobalFilter(t *testing.T) {
	is := assert.New(t)
	must := require.New(t)

	rest.Reset()
	defer rest.Reset()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Received-Echo", r.Header.Get("X-Filter-Injected"))
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(ts.Close)

	var filterExecuted bool
	timingFilter := rest.NewTimingFilter(func(d time.Duration) {
		filterExecuted = true
	})

	rest.BaseURI(ts.URL)
	rest.GlobalFilter(timingFilter)

	resp := rest.Get("/filter-test")
	must.NoError(resp.Err())
	is.Equal(http.StatusOK, resp.StatusCode())
	is.True(filterExecuted)
}
