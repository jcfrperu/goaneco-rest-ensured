package integration_test

// Ported from LoggingITest.java and GivenWhenThenLoggingITest.java

import (
	"bytes"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func TestJavaITest_Logging_ResponseLogSpec(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("LogAllDetailsUsingResponseLogSpec", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var buf bytes.Buffer
		resp := rest.Given().
			BaseURI(ts.URL).
			Filter(rest.NewResponseLoggingFilter(&buf, rest.LogDetailAll)).
			Get("/lotto")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Contains(buf.String(), "--- Response Details ---")
		is.Contains(buf.String(), "200")
	})

	t.Run("LogStatusOnlyUsingResponseLogSpec", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var buf bytes.Buffer
		resp := rest.Given().
			BaseURI(ts.URL).
			Filter(rest.NewResponseLoggingFilter(&buf, rest.LogDetailStatus)).
			Get("/lotto")

		must.NoError(resp.Err())
		out := buf.String()
		is.Contains(out, "200")
		is.NotContains(out, "Body:")
	})

	t.Run("LogHeadersOnlyUsingResponseLogSpec", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var buf bytes.Buffer
		resp := rest.Given().
			BaseURI(ts.URL).
			Filter(rest.NewResponseLoggingFilter(&buf, rest.LogDetailHeaders)).
			Get("/lotto")

		must.NoError(resp.Err())
		out := buf.String()
		is.Contains(out, "Content-Type")
		is.NotContains(out, "Status code:")
	})

	t.Run("LogBodyOnlyUsingResponseLogSpec", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var buf bytes.Buffer
		resp := rest.Given().
			BaseURI(ts.URL).
			Filter(rest.NewResponseLoggingFilter(&buf, rest.LogDetailBody)).
			Get("/lotto")

		must.NoError(resp.Err())
		out := buf.String()
		is.Contains(out, "lotto")
	})

	t.Run("LogCookiesOnlyUsingResponseLogSpec", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var buf bytes.Buffer
		resp := rest.Given().
			BaseURI(ts.URL).
			Filter(rest.NewResponseLoggingFilter(&buf, rest.LogDetailCookies)).
			Get("/setCookies")

		must.NoError(resp.Err())
		out := buf.String()
		is.Contains(out, "--- Response Details ---")
	})

	t.Run("LogResponseWithCookiesLogDetailAll", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var buf bytes.Buffer
		resp := rest.Given().
			BaseURI(ts.URL).
			Filter(rest.NewResponseLoggingFilter(&buf, rest.LogDetailAll)).
			Get("/setCookies")

		must.NoError(resp.Err())
		out := buf.String()
		is.Contains(out, "--- Response Details ---")
		is.Contains(out, "200")
	})

	t.Run("LogResponseBodyWithPrettyPrintingWhenJson", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var buf bytes.Buffer
		resp := rest.Given().
			BaseURI(ts.URL).
			Filter(rest.NewResponseLoggingFilter(&buf, rest.LogDetailBody)).
			Get("/lotto")

		must.NoError(resp.Err())
		out := buf.String()
		// Pretty-printed JSON should have indentation
		is.Contains(out, "lotto")
		is.Contains(out, "lottoId")
	})

	t.Run("LogIfStatusCodeIsEqualTo", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var buf bytes.Buffer
		filter := &rest.StatusCodeBasedLoggingFilter{
			Predicate: func(code int) bool { return code == http.StatusConflict },
			Output:    &buf,
		}

		// 200 → does NOT trigger logging
		resp1 := rest.Given().BaseURI(ts.URL).Filter(filter).Get("/lotto")
		must.NoError(resp1.Err())
		is.Empty(buf.String())

		// 409 → DOES trigger logging
		resp2 := rest.Given().BaseURI(ts.URL).Filter(filter).Get("/409")
		must.NoError(resp2.Err())
		is.Contains(buf.String(), "409")
	})

	t.Run("LogIfStatusCodeMatchesHamcrestMatcher", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var buf bytes.Buffer
		filter := &rest.StatusCodeBasedLoggingFilter{
			Predicate: func(code int) bool { return code >= 500 },
			Output:    &buf,
		}

		// 200 → does NOT trigger
		resp1 := rest.Given().BaseURI(ts.URL).Filter(filter).Get("/lotto")
		must.NoError(resp1.Err())
		is.Empty(buf.String())

		// 500 → DOES trigger
		resp2 := rest.Given().BaseURI(ts.URL).Filter(filter).Get("/statusCode500")
		must.NoError(resp2.Err())
		is.Contains(buf.String(), "500")
	})

	t.Run("LogIfValidationFails_DoesNotLogWhenSucceeds", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var buf bytes.Buffer
		resp := rest.Given().
			BaseURI(ts.URL).
			Log().Writer(&buf).IfValidationFails().
			Get("/lotto")

		must.NoError(resp.Err())
		// Validation passes — nothing logged until AssertAll runs
		is.Equal(http.StatusOK, resp.StatusCode())
	})
}

func TestJavaITest_Logging_RequestLogSpec(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("LogRequestAllDetails", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var buf bytes.Buffer
		resp := rest.Given().
			BaseURI(ts.URL).
			Header("X-Trace", "trace-001").
			Cookie("req_cookie", "rc_val").
			QueryParam("param1", "value1").
			Filter(rest.NewRequestLoggingFilter(&buf, rest.LogDetailAll)).
			Get("/greet")

		must.NoError(resp.Err())
		out := buf.String()
		is.Contains(out, "--- Request Details ---")
		is.Contains(out, "GET")
		is.Contains(out, "X-Trace")
	})

	t.Run("LogRequestHeadersOnly", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var buf bytes.Buffer
		resp := rest.Given().
			BaseURI(ts.URL).
			Header("X-Log-Header", "hdr-val").
			Filter(rest.NewRequestLoggingFilter(&buf, rest.LogDetailHeaders)).
			Get("/greet")

		must.NoError(resp.Err())
		out := buf.String()
		is.Contains(out, "X-Log-Header")
	})

	t.Run("LogRequestBodyOnly", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var buf bytes.Buffer
		resp := rest.Given().
			BaseURI(ts.URL).
			Body(`{"action":"test"}`).
			ContentType(rest.ContentTypeJSON).
			Filter(rest.NewRequestLoggingFilter(&buf, rest.LogDetailBody)).
			Post("/echo")

		must.NoError(resp.Err())
		out := buf.String()
		is.Contains(out, "action")
	})

	t.Run("LogRequestCookiesOnly", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var buf bytes.Buffer
		resp := rest.Given().
			BaseURI(ts.URL).
			Cookie("session_log_cookie", "slc_val").
			Filter(rest.NewRequestLoggingFilter(&buf, rest.LogDetailCookies)).
			Get("/greet")

		must.NoError(resp.Err())
		out := buf.String()
		is.Contains(out, "session_log_cookie")
	})
}

func TestJavaITest_Logging_ErrorFilter(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("ErrorLoggingFilterOnlyLogsErrors", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var buf bytes.Buffer
		errFilter := &rest.ErrorLoggingFilter{Output: &buf}

		// Successful request: nothing logged
		resp1 := rest.Given().BaseURI(ts.URL).Filter(errFilter).Get("/lotto")
		must.NoError(resp1.Err())
		is.Empty(buf.String())

		// Error response: logs request and response
		resp2 := rest.Given().BaseURI(ts.URL).Filter(errFilter).Get("/409")
		must.NoError(resp2.Err())
		out := buf.String()
		is.Contains(out, "--- Request Details ---")
		is.Contains(out, "--- Response Details ---")
		is.Contains(out, "409")
	})

	t.Run("ResponseLoggingFilterLogsNonErrors", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var buf bytes.Buffer
		filter := rest.NewResponseLoggingFilter(&buf, rest.LogDetailAll)

		resp := rest.Given().BaseURI(ts.URL).Filter(filter).Get("/greet")
		must.NoError(resp.Err())
		is.Contains(buf.String(), "--- Response Details ---")
		is.Contains(buf.String(), "200")
	})

	t.Run("LoggingRequestAndResponseTogether", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var reqBuf, respBuf bytes.Buffer
		reqFilter := rest.NewRequestLoggingFilter(&reqBuf, rest.LogDetailAll)
		respFilter := rest.NewResponseLoggingFilter(&respBuf, rest.LogDetailAll)

		resp := rest.Given().
			BaseURI(ts.URL).
			Header("X-Dual-Log", "dual").
			Filter(reqFilter, respFilter).
			Get("/lotto")

		must.NoError(resp.Err())
		is.Contains(reqBuf.String(), "--- Request Details ---")
		is.Contains(respBuf.String(), "--- Response Details ---")
		is.Contains(respBuf.String(), "200")
	})

	t.Run("LoggingDoesNotDisruptResponseChain", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var buf bytes.Buffer
		resp := rest.Given().
			BaseURI(ts.URL).
			Filter(rest.NewResponseLoggingFilter(&buf, rest.LogDetailAll)).
			Get("/lotto")

		must.NoError(resp.Err())
		// Full response body is still accessible after logging
		is.Equal(http.StatusOK, resp.StatusCode())
		jp := resp.JsonPath()
		is.Equal(float64(5), jp.Get("lotto.lottoId").Value())
	})
}

func TestJavaITest_Logging_GivenWhenThenSyntax(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("LogsEverythingResponseUsingGivenWhenThenSyntax", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var buf bytes.Buffer
		valid := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then()

		valid.Log().Writer(&buf).All()

		must.NotEmpty(buf.String())
		is.Contains(buf.String(), "--- Response Details ---")
		is.Equal(http.StatusOK, valid.Extract().StatusCode())
	})

	t.Run("LogOnlyHeadersUsingGivenWhenThenSyntax", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var buf bytes.Buffer
		valid := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then()

		valid.Log().Writer(&buf).Headers()

		must.NotEmpty(buf.String())
		is.Contains(buf.String(), "Content-Type")
	})

	t.Run("LogResponseThatHasCookiesWithLogDetailCookies", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		var buf bytes.Buffer
		valid := rest.Given().
			BaseURI(ts.URL).
			Get("/setCookies").
			Then()

		valid.Log().Writer(&buf).Cookies()

		// Cookies section is logged
		is.Contains(buf.String(), "--- Response Details ---")
	})

	t.Run("LogBodyWithPrettyPrintingWhenJSON", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var buf bytes.Buffer
		resp := rest.Given().
			BaseURI(ts.URL).
			Filter(rest.NewResponseLoggingFilter(&buf, rest.LogDetailBody)).
			Get("/jsonList")

		must.NoError(resp.Err())
		out := buf.String()
		is.Contains(out, "Anders")
		is.Contains(out, "Sven")
	})

	t.Run("LogBodyWithPrettyPrintingWhenXML", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var buf bytes.Buffer
		resp := rest.Given().
			BaseURI(ts.URL).
			Filter(rest.NewResponseLoggingFilter(&buf, rest.LogDetailBody)).
			Get("/videos")

		must.NoError(resp.Err())
		out := buf.String()
		is.Contains(out, "Bohemian Rhapsody")
	})
}

func TestJavaITest_Logging_ResponseTime(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("ResponseTimeIsMeasured", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/delay/50")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.GreaterOrEqual(resp.Time(), 50*time.Millisecond)
	})

	t.Run("ResponseTimeAssertionViaSpec", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		valid := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then()

		valid.TimeLessThan(5 * time.Second)

		is.False(valid.HasFailures(), "response time should be under 5s: %v", valid.Failures())
	})
}
