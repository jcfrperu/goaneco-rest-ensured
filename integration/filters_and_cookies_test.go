package integration_test

import (
	"bytes"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func TestIntegration_CookieFilter_Persistence(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	ts := integrationServer

	cookieFilter := rest.NewCookieFilter()

	// 1. First request receives Set-Cookie
	respSet := rest.Given().
		BaseURI(ts.URL).
		Filter(cookieFilter).
		QueryParam("name", "auth_token").
		QueryParam("value", "secret-token-abc").
		Post("/cookies")

	must.NoError(respSet.Err())
	is.Equal(http.StatusOK, respSet.StatusCode())

	// 2. Second request should automatically send the cookie stored in CookieFilter
	respGet := rest.Given().
		BaseURI(ts.URL).
		Filter(cookieFilter).
		Get("/cookies")

	must.NoError(respGet.Err())
	is.Equal(http.StatusOK, respGet.StatusCode())
	is.Equal("secret-token-abc", respGet.JsonPath().GetString("auth_token"))
}

func TestIntegration_TimingAndLoggingFilters(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	ts := integrationServer

	var logBuf bytes.Buffer
	reqLog := rest.NewRequestLoggingFilter(&logBuf, rest.LogDetailAll)
	respLog := rest.NewResponseLoggingFilter(&logBuf, rest.LogDetailAll)

	var durationCaptured time.Duration
	timing := rest.NewTimingFilter(func(d time.Duration) {
		durationCaptured = d
	})

	resp := rest.Given().
		BaseURI(ts.URL).
		Filter(reqLog, respLog, timing).
		Get("/hello")

	must.NoError(resp.Err())
	is.Equal(http.StatusOK, resp.StatusCode())
	is.GreaterOrEqual(durationCaptured, time.Duration(0))
	is.Contains(logBuf.String(), "Request method:\tGET")
	is.Contains(logBuf.String(), "HTTP/1.1 200 OK")
}

type testOrderedTrackerFilter struct {
	id      int
	order   int
	tracker *[]int
	mu      *sync.Mutex
}

func (f *testOrderedTrackerFilter) Order() int {
	return f.order
}

func (f *testOrderedTrackerFilter) Filter(req *rest.FilterableRequest, ctx *rest.FilterContext) (*rest.Response, error) {
	f.mu.Lock()
	*f.tracker = append(*f.tracker, f.id)
	f.mu.Unlock()
	return ctx.Next(req)
}

func TestIntegration_OrderedFilterExecution(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	ts := integrationServer

	var (
		mu      sync.Mutex
		tracker []int
	)

	fHigh := &testOrderedTrackerFilter{id: 300, order: 300, tracker: &tracker, mu: &mu}
	fLow := &testOrderedTrackerFilter{id: 100, order: 100, tracker: &tracker, mu: &mu}
	fMid := &testOrderedTrackerFilter{id: 200, order: 200, tracker: &tracker, mu: &mu}

	// Attach in random order: 300, 100, 200
	resp := rest.Given().
		BaseURI(ts.URL).
		Filter(fHigh, fLow, fMid).
		Get("/hello")

	must.NoError(resp.Err())
	is.Equal(http.StatusOK, resp.StatusCode())

	mu.Lock()
	orderResult := append([]int{}, tracker...)
	mu.Unlock()

	// Must be sorted and executed in ascending order: 100, 200, 300
	is.Equal([]int{100, 200, 300}, orderResult)
}
