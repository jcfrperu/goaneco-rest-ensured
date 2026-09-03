package integration_test

// Ported from FailureConfigITest.java

import (
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func TestJavaITest_FailureConfig_Listener(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("ResponseValidationFailureListenerIsCalledOnFailure", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var listenerCalled int32
		listener := func(req *http.Request, resp *rest.Response, failures []string) {
			atomic.AddInt32(&listenerCalled, 1)
		}

		cfg := rest.DefaultConfig().WithFailure(rest.FailureConfig{
			Listeners: []func(req *http.Request, resp *rest.Response, failures []string){listener},
		})

		valid := rest.Given().
			Config(cfg).
			BaseURI(ts.URL).
			Get("/reflect").
			Then().
			StatusCode(http.StatusBadRequest) // fails: actual is 200

		must.True(valid.HasFailures())
		is.Equal(int32(1), atomic.LoadInt32(&listenerCalled), "listener should have been called once")
	})

	t.Run("ResponseValidationFailureListenerIsNotCalledOnSuccess", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		var listenerCalled int32
		listener := func(req *http.Request, resp *rest.Response, failures []string) {
			atomic.AddInt32(&listenerCalled, 1)
		}

		cfg := rest.DefaultConfig().WithFailure(rest.FailureConfig{
			Listeners: []func(req *http.Request, resp *rest.Response, failures []string){listener},
		})

		valid := rest.Given().
			Config(cfg).
			BaseURI(ts.URL).
			Get("/reflect").
			Then().
			StatusCode(http.StatusOK) // passes

		is.False(valid.HasFailures())
		is.Equal(int32(0), atomic.LoadInt32(&listenerCalled), "listener should not be called when validation passes")
	})

	t.Run("MultipleListenersAreAllCalled", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var count1, count2 int32
		listener1 := func(req *http.Request, resp *rest.Response, failures []string) {
			atomic.AddInt32(&count1, 1)
		}
		listener2 := func(req *http.Request, resp *rest.Response, failures []string) {
			atomic.AddInt32(&count2, 1)
		}

		cfg := rest.DefaultConfig().WithFailure(rest.FailureConfig{
			Listeners: []func(req *http.Request, resp *rest.Response, failures []string){listener1, listener2},
		})

		valid := rest.Given().
			Config(cfg).
			BaseURI(ts.URL).
			Get("/reflect").
			Then().
			StatusCode(http.StatusBadRequest) // fails

		must.True(valid.HasFailures())
		is.Equal(int32(1), atomic.LoadInt32(&count1))
		is.Equal(int32(1), atomic.LoadInt32(&count2))
	})
}
