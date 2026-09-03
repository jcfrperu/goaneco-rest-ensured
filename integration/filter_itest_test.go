package integration_test

// Ported from FilterITest.java

import (
	"bytes"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func TestJavaITest_Filter_TimingAndLogging(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("TimingFilterMeasuresRequestDuration", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var measured time.Duration
		timingFilter := rest.NewTimingFilter(func(d time.Duration) {
			measured = d
		})

		resp := rest.Given().
			BaseURI(ts.URL).
			Filter(timingFilter).
			Get("/lotto")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.GreaterOrEqual(measured, time.Duration(0))
	})

	t.Run("RequestLoggingFilterLogsToWriter", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var buf bytes.Buffer
		logFilter := rest.NewRequestLoggingFilter(&buf, rest.LogDetailAll)

		resp := rest.Given().
			BaseURI(ts.URL).
			Filter(logFilter).
			Get("/lotto")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.NotEmpty(buf.String())
	})

	t.Run("ResponseLoggingFilterLogsToWriter", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var buf bytes.Buffer
		logFilter := rest.NewResponseLoggingFilter(&buf, rest.LogDetailAll)

		resp := rest.Given().
			BaseURI(ts.URL).
			Filter(logFilter).
			Get("/lotto")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.NotEmpty(buf.String())
	})

	t.Run("ErrorLoggingFilterDoesNotLogOnSuccess", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var buf bytes.Buffer
		errFilter := &rest.ErrorLoggingFilter{Output: &buf}

		resp := rest.Given().
			BaseURI(ts.URL).
			Filter(errFilter).
			Get("/lotto")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Empty(buf.String(), "error filter should not log on success")
	})

	t.Run("ErrorLoggingFilterLogsOnErrorStatus", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var buf bytes.Buffer
		errFilter := &rest.ErrorLoggingFilter{Output: &buf}

		resp := rest.Given().
			BaseURI(ts.URL).
			Filter(errFilter).
			Get("/409")

		must.NoError(resp.Err())
		is.Equal(http.StatusConflict, resp.StatusCode())
		is.NotEmpty(buf.String(), "error filter should log on error status")
	})

	t.Run("StatusCodeBasedLoggingFilterLogsWhenPredicateMatches", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var buf bytes.Buffer
		logFilter := &rest.StatusCodeBasedLoggingFilter{
			Predicate: func(code int) bool { return code == http.StatusOK },
			Output:    &buf,
		}

		resp := rest.Given().
			BaseURI(ts.URL).
			Filter(logFilter).
			Get("/lotto")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.NotEmpty(buf.String(), "should log when predicate returns true")
	})

	t.Run("StatusCodeBasedLoggingFilterDoesNotLogWhenPredicateDoesNotMatch", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var buf bytes.Buffer
		logFilter := &rest.StatusCodeBasedLoggingFilter{
			Predicate: func(code int) bool { return code == http.StatusInternalServerError },
			Output:    &buf,
		}

		resp := rest.Given().
			BaseURI(ts.URL).
			Filter(logFilter).
			Get("/lotto")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Empty(buf.String(), "should not log when predicate returns false")
	})

	t.Run("MultipleFiltersAreAppliedInOrder", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var order []string
		f1 := &testOrderFilter{name: "first", order: &order}
		f2 := &testOrderFilter{name: "second", order: &order}

		resp := rest.Given().
			BaseURI(ts.URL).
			Filter(f1, f2).
			Get("/lotto")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal([]string{"first", "second"}, order)
	})
}

// testOrderFilter records its name when executed.
type testOrderFilter struct {
	name  string
	order *[]string
}

func (f *testOrderFilter) Filter(req *rest.FilterableRequest, ctx *rest.FilterContext) (*rest.Response, error) {
	*f.order = append(*f.order, f.name)
	return ctx.Next(req)
}

func TestJavaITest_Filter_RequestModification(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("FilterCanAddQueryParam", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		addParamFilter := &addQueryParamFilter{key: "firstName", value: "Jane"}

		resp := rest.Given().
			BaseURI(ts.URL).
			Filter(addParamFilter).
			QueryParam("lastName", "Doe").
			Get("/greet")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Contains(resp.AsString(), "Jane")
	})

	t.Run("FilterCanAddHeader", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		addHeaderFilter := &addHeaderFilterImpl{name: "X-Filter-Added", value: "yes"}

		resp := rest.Given().
			BaseURI(ts.URL).
			Filter(addHeaderFilter).
			Get("/headers")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Contains(resp.AsString(), "X-Filter-Added")
	})

	t.Run("FilterCanRemoveHeader", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		removeHeaderFilter := &removeHeaderFilterImpl{name: "X-Should-Remove"}

		resp := rest.Given().
			BaseURI(ts.URL).
			Header("X-Should-Remove", "remove-me").
			Filter(removeHeaderFilter).
			Get("/headers")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.NotContains(resp.AsString(), "X-Should-Remove")
	})

	t.Run("FilterCanReplaceHeaderValue", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		replaceHeaderFilter := &replaceHeaderFilterImpl{name: "X-Replace-Me", newValue: "replaced-value"}

		resp := rest.Given().
			BaseURI(ts.URL).
			Header("X-Replace-Me", "original-value").
			Filter(replaceHeaderFilter).
			Get("/headers")

		must.NoError(resp.Err())
		is.Contains(resp.AsString(), "replaced-value")
		is.NotContains(resp.AsString(), "original-value")
	})

	t.Run("FilterCanRemoveCookie", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		removeCookieFilter := &removeCookieFilterImpl{name: "secret_cookie"}

		resp := rest.Given().
			BaseURI(ts.URL).
			Cookie("secret_cookie", "should-be-removed").
			Cookie("visible_cookie", "keep-me").
			Filter(removeCookieFilter).
			Get("/cookies")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		body := resp.AsString()
		is.NotContains(body, "secret_cookie")
		is.Contains(body, "keep-me")
	})

	t.Run("FilterCanRemoveAllCookies", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		removeAllCookiesFilter := &removeAllCookiesFilterImpl{}

		resp := rest.Given().
			BaseURI(ts.URL).
			Cookie("cookie1", "v1").
			Cookie("cookie2", "v2").
			Filter(removeAllCookiesFilter).
			Get("/cookies")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		body := resp.AsString()
		is.NotContains(body, "cookie1")
		is.NotContains(body, "cookie2")
	})
}

type addQueryParamFilter struct{ key, value string }

func (f *addQueryParamFilter) Filter(req *rest.FilterableRequest, ctx *rest.FilterContext) (*rest.Response, error) {
	// URI is already built; modify it directly to add the query param.
	u, err := url.Parse(req.URI)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set(f.key, f.value)
	u.RawQuery = q.Encode()
	req.URI = u.String()
	return ctx.Next(req)
}

type addHeaderFilterImpl struct{ name, value string }

func (f *addHeaderFilterImpl) Filter(req *rest.FilterableRequest, ctx *rest.FilterContext) (*rest.Response, error) {
	if req.Headers == nil {
		req.Headers = make(http.Header)
	}
	req.Headers.Set(f.name, f.value)
	return ctx.Next(req)
}

type removeHeaderFilterImpl struct{ name string }

func (f *removeHeaderFilterImpl) Filter(req *rest.FilterableRequest, ctx *rest.FilterContext) (*rest.Response, error) {
	if req.Headers != nil {
		req.Headers.Del(f.name)
	}
	return ctx.Next(req)
}

type replaceHeaderFilterImpl struct{ name, newValue string }

func (f *replaceHeaderFilterImpl) Filter(req *rest.FilterableRequest, ctx *rest.FilterContext) (*rest.Response, error) {
	if req.Headers != nil {
		req.Headers.Set(f.name, f.newValue)
	}
	return ctx.Next(req)
}

type removeCookieFilterImpl struct{ name string }

func (f *removeCookieFilterImpl) Filter(req *rest.FilterableRequest, ctx *rest.FilterContext) (*rest.Response, error) {
	var filtered []*http.Cookie
	for _, c := range req.Cookies {
		if c.Name != f.name {
			filtered = append(filtered, c)
		}
	}
	req.Cookies = filtered
	return ctx.Next(req)
}

type removeAllCookiesFilterImpl struct{}

func (f *removeAllCookiesFilterImpl) Filter(req *rest.FilterableRequest, ctx *rest.FilterContext) (*rest.Response, error) {
	req.Cookies = nil
	return ctx.Next(req)
}

func TestJavaITest_Filter_ResponseInspection(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("FilterCanCaptureResponseStatusCode", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var capturedStatus int
		captureFilter := &captureStatusFilter{captured: &capturedStatus}

		resp := rest.Given().
			BaseURI(ts.URL).
			Filter(captureFilter).
			Get("/lotto")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal(http.StatusOK, capturedStatus)
	})

	t.Run("FilterCanInspectResponseBody", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var capturedBody string
		bodyFilter := &captureBodyFilter{body: &capturedBody}

		resp := rest.Given().
			BaseURI(ts.URL).
			Filter(bodyFilter).
			Get("/lotto")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Contains(capturedBody, "lotto")
	})

	t.Run("FilterIsAppliedWhenAddedViaRequestSpec", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var executed bool
		execFilter := &executionRecorderFilter{executed: &executed}

		spec := rest.NewRequestSpecBuilder().
			SetBaseURI(ts.URL).
			AddFilter(execFilter).
			Build()

		resp := rest.Given().
			Spec(spec).
			Get("/lotto")

		must.NoError(resp.Err())
		is.True(executed, "filter from request spec must execute")
	})

	t.Run("CookieFilterPersistsCookiesAcrossRequests", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		cookieFilter := rest.NewCookieFilter()

		resp1 := rest.Given().
			BaseURI(ts.URL).
			Filter(cookieFilter).
			Get("/setCookies")

		must.NoError(resp1.Err())
		is.Equal(http.StatusOK, resp1.StatusCode())
		is.NotNil(cookieFilter.Jar())
	})

	t.Run("SessionFilterPersistsSessionAcrossRequests", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// /sessionId sets a "jsessionid" cookie (lowercase); create filter to match.
		sessionFilter := rest.NewSessionFilter("jsessionid")

		resp1 := rest.Given().
			BaseURI(ts.URL).
			Filter(sessionFilter).
			Get("/sessionId")

		must.NoError(resp1.Err())
		sid := sessionFilter.SessionID()
		is.NotEmpty(sid, "session ID should be captured")

		// Second request: filter auto-injects the captured cookie; /sessionId validates it.
		resp2 := rest.Given().
			BaseURI(ts.URL).
			Filter(sessionFilter).
			Get("/sessionId")

		must.NoError(resp2.Err())
		is.Equal(http.StatusOK, resp2.StatusCode())
		is.Equal("Success", resp2.JsonPath().GetString("status"))
	})
}

type captureStatusFilter struct{ captured *int }

func (f *captureStatusFilter) Filter(req *rest.FilterableRequest, ctx *rest.FilterContext) (*rest.Response, error) {
	resp, err := ctx.Next(req)
	if resp != nil {
		*f.captured = resp.StatusCode()
	}
	return resp, err
}

type captureBodyFilter struct{ body *string }

func (f *captureBodyFilter) Filter(req *rest.FilterableRequest, ctx *rest.FilterContext) (*rest.Response, error) {
	resp, err := ctx.Next(req)
	if resp != nil {
		*f.body = resp.AsString()
	}
	return resp, err
}

type executionRecorderFilter struct{ executed *bool }

func (f *executionRecorderFilter) Filter(req *rest.FilterableRequest, ctx *rest.FilterContext) (*rest.Response, error) {
	*f.executed = true
	return ctx.Next(req)
}

func TestJavaITest_Filter_LoggingCombinations(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("RequestAndResponseLoggingFiltersWorkTogether", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var reqBuf, respBuf bytes.Buffer
		reqFilter := rest.NewRequestLoggingFilter(&reqBuf, rest.LogDetailAll)
		respFilter := rest.NewResponseLoggingFilter(&respBuf, rest.LogDetailAll)

		resp := rest.Given().
			BaseURI(ts.URL).
			Filter(reqFilter, respFilter).
			QueryParam("firstName", "Test").
			QueryParam("lastName", "User").
			Get("/greet")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.NotEmpty(reqBuf.String())
		is.NotEmpty(respBuf.String())
	})

	t.Run("RequestLoggingHeadersOnlyDoesNotLogBody", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var buf bytes.Buffer
		logFilter := rest.NewRequestLoggingFilter(&buf, rest.LogDetailHeaders)

		resp := rest.Given().
			BaseURI(ts.URL).
			Header("X-Test-Header", "test-value").
			Filter(logFilter).
			Get("/lotto")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.NotEmpty(buf.String())
	})

	t.Run("ResponseLoggingBodyOnlyWritesBody", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var buf bytes.Buffer
		logFilter := rest.NewResponseLoggingFilter(&buf, rest.LogDetailBody)

		resp := rest.Given().
			BaseURI(ts.URL).
			Filter(logFilter).
			Get("/lotto")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.NotEmpty(buf.String())
	})

	t.Run("ErrorLoggingFilterLogsOn500", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var buf bytes.Buffer
		errFilter := &rest.ErrorLoggingFilter{Output: &buf}

		resp := rest.Given().
			BaseURI(ts.URL).
			Filter(errFilter).
			Get("/statusCode500")

		must.NoError(resp.Err())
		is.Equal(http.StatusInternalServerError, resp.StatusCode())
		is.NotEmpty(buf.String())
	})
}

func TestJavaITest_Filter_FilterableRequestAccess(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("FilterableRequestExposesMethod", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var capturedMethod string
		methodFilter := &captureMethodFilter{method: &capturedMethod}

		resp := rest.Given().
			BaseURI(ts.URL).
			Filter(methodFilter).
			Get("/lotto")

		must.NoError(resp.Err())
		is.Equal("GET", strings.ToUpper(capturedMethod))
	})

	t.Run("FilterableRequestExposesURI", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var capturedURI string
		uriFilter := &captureURIFilter{uri: &capturedURI}

		resp := rest.Given().
			BaseURI(ts.URL).
			Filter(uriFilter).
			Get("/lotto")

		must.NoError(resp.Err())
		is.Contains(capturedURI, "/lotto")
	})

	t.Run("FilterableRequestExposesHeaders", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var capturedHeaders http.Header
		headerFilter := &captureHeadersFilter{headers: &capturedHeaders}

		resp := rest.Given().
			BaseURI(ts.URL).
			Header("X-Custom-Header", "custom-value").
			Filter(headerFilter).
			Get("/lotto")

		must.NoError(resp.Err())
		is.Equal("custom-value", capturedHeaders.Get("X-Custom-Header"))
	})

	t.Run("FilterableRequestExposesCookies", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var capturedCookies []*http.Cookie
		cookieFilter := &captureRequestCookiesFilter{cookies: &capturedCookies}

		resp := rest.Given().
			BaseURI(ts.URL).
			Cookie("test_cookie", "test_value").
			Filter(cookieFilter).
			Get("/cookies")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		var found bool
		for _, c := range capturedCookies {
			if c.Name == "test_cookie" && c.Value == "test_value" {
				found = true
			}
		}
		is.True(found, "request cookies should be accessible from filter")
	})

	t.Run("FilterableRequestExposesQueryParams", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var capturedParams map[string][]string
		paramFilter := &captureParamsFilter{params: &capturedParams}

		resp := rest.Given().
			BaseURI(ts.URL).
			QueryParam("count", "42").
			Filter(paramFilter).
			Get("/lotto")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal([]string{"42"}, capturedParams["count"])
	})

	t.Run("FilterableRequestExposesContentType", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var capturedCT string
		ctFilter := &captureContentTypeFilter{ct: &capturedCT}

		resp := rest.Given().
			BaseURI(ts.URL).
			ContentType(rest.ContentTypeJSON).
			Body(`{"key":"value"}`).
			Filter(ctFilter).
			Post("/echo")

		must.NoError(resp.Err())
		is.Contains(capturedCT, "application/json")
	})
}

type captureMethodFilter struct{ method *string }

func (f *captureMethodFilter) Filter(req *rest.FilterableRequest, ctx *rest.FilterContext) (*rest.Response, error) {
	*f.method = req.Method
	return ctx.Next(req)
}

type captureURIFilter struct{ uri *string }

func (f *captureURIFilter) Filter(req *rest.FilterableRequest, ctx *rest.FilterContext) (*rest.Response, error) {
	*f.uri = req.URI
	return ctx.Next(req)
}

type captureHeadersFilter struct{ headers *http.Header }

func (f *captureHeadersFilter) Filter(req *rest.FilterableRequest, ctx *rest.FilterContext) (*rest.Response, error) {
	*f.headers = req.Headers.Clone()
	return ctx.Next(req)
}

type captureRequestCookiesFilter struct{ cookies *[]*http.Cookie }

func (f *captureRequestCookiesFilter) Filter(req *rest.FilterableRequest, ctx *rest.FilterContext) (*rest.Response, error) {
	*f.cookies = append([]*http.Cookie{}, req.Cookies...)
	return ctx.Next(req)
}

type captureParamsFilter struct{ params *map[string][]string }

func (f *captureParamsFilter) Filter(req *rest.FilterableRequest, ctx *rest.FilterContext) (*rest.Response, error) {
	copied := make(map[string][]string)
	for k, v := range req.QueryParams {
		copied[k] = v
	}
	*f.params = copied
	return ctx.Next(req)
}

type captureContentTypeFilter struct{ ct *string }

func (f *captureContentTypeFilter) Filter(req *rest.FilterableRequest, ctx *rest.FilterContext) (*rest.Response, error) {
	*f.ct = req.ContentType
	return ctx.Next(req)
}
