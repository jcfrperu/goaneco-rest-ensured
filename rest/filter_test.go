package rest_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

// parseCookies parses a slice of raw Set-Cookie header strings using the standard library.
// Mirrors Java's CookieMatcher.getCookies(List<String>).
func parseCookies(lines []string) []*http.Cookie {
	header := http.Header{}
	for _, line := range lines {
		header.Add("Set-Cookie", line)
	}
	resp := &http.Response{Header: header}
	return resp.Cookies()
}

type CustomHeaderFilter struct {
	HeaderName  string
	HeaderValue string
}

func (f *CustomHeaderFilter) Filter(req *rest.FilterableRequest, ctx *rest.FilterContext) (*rest.Response, error) {
	req.Headers.Set(f.HeaderName, f.HeaderValue)
	return ctx.Next(req)
}

func TestFilterExecution(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	var receivedHeader string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeader = r.Header.Get("X-Intercepted-By")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"filtered"}`))
	}))
	t.Cleanup(ts.Close)

	customFilter := &CustomHeaderFilter{
		HeaderName:  "X-Intercepted-By",
		HeaderValue: "CustomFilterV1",
	}

	var (
		timingFilterCalled bool
		recordedDuration   time.Duration
	)
	timingFilter := &rest.TimingFilter{
		OnComplete: func(d time.Duration) {
			timingFilterCalled = true
			recordedDuration = d
		},
	}

	resp := rest.Given().
		BaseURI(ts.URL).
		Filter(customFilter, timingFilter).
		Get("/filter-check")

	must.NoError(resp.Err())
	is.Equal(http.StatusOK, resp.StatusCode())
	is.Equal("CustomFilterV1", receivedHeader)
	is.True(timingFilterCalled)
	is.GreaterOrEqual(recordedDuration, time.Duration(0))
}

func TestLoggingFilters(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"logged":true}`))
	}))
	t.Cleanup(ts.Close)

	var reqLogBuf bytes.Buffer
	reqLogFilter := &rest.RequestLoggingFilter{
		Detail: rest.LogDetailAll,
		Output: &reqLogBuf,
	}

	var respLogBuf bytes.Buffer
	respLogFilter := &rest.ResponseLoggingFilter{
		Detail: rest.LogDetailAll,
		Output: &respLogBuf,
	}

	resp := rest.Given().
		BaseURI(ts.URL).
		Filter(reqLogFilter, respLogFilter).
		Get("/logging-test")

	must.NoError(resp.Err())
	is.Equal(http.StatusOK, resp.StatusCode())

	is.Contains(reqLogBuf.String(), "--- Request Details ---")
	is.Contains(respLogBuf.String(), "--- Response Details ---")
}

type OrderTrackerFilter struct {
	order   int
	name    string
	tracker *[]string
}

func (f *OrderTrackerFilter) Order() int {
	return f.order
}

func (f *OrderTrackerFilter) Filter(req *rest.FilterableRequest, ctx *rest.FilterContext) (*rest.Response, error) {
	*f.tracker = append(*f.tracker, f.name)
	return ctx.Next(req)
}

func TestOrderedFilterExecution(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(ts.Close)

	var orderSequence []string
	filterFirst := &OrderTrackerFilter{order: 10, name: "SecondAdded_Order10", tracker: &orderSequence}
	filterSecond := &OrderTrackerFilter{order: 1, name: "FirstAdded_Order1", tracker: &orderSequence}
	filterThird := &OrderTrackerFilter{order: 50, name: "ThirdAdded_Order50", tracker: &orderSequence}

	// Added out of order: order 10, then order 1, then order 50
	resp := rest.Given().
		BaseURI(ts.URL).
		Filter(filterFirst, filterSecond, filterThird).
		Get("/ordered-check")

	must.NoError(resp.Err())
	is.Equal(http.StatusOK, resp.StatusCode())

	// Verified sorted execution: order 1 (FirstAdded_Order1) -> order 10 (SecondAdded_Order10) -> order 50 (ThirdAdded_Order50)
	expectedOrder := []string{"FirstAdded_Order1", "SecondAdded_Order10", "ThirdAdded_Order50"}
	is.Equal(expectedOrder, orderSequence)
}

func TestCookieFilter(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/set" {
			http.SetCookie(w, &http.Cookie{Name: "session_id", Value: "cookie_val_123", Path: "/"})
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`ok`))
			return
		}
		if r.URL.Path == "/check" {
			c, err := r.Cookie("session_id")
			if err == nil && c.Value == "cookie_val_123" {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`cookie_matched`))
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(ts.Close)

	cookieFilter := rest.NewCookieFilter()
	is.NotNil(cookieFilter.Jar())

	// First request: sets cookie
	resp1 := rest.Given().
		BaseURI(ts.URL).
		Filter(cookieFilter).
		Get("/set")
	must.NoError(resp1.Err())
	is.Equal(http.StatusOK, resp1.StatusCode())

	// Second request: automatically carries cookie from jar
	resp2 := rest.Given().
		BaseURI(ts.URL).
		Filter(cookieFilter).
		Get("/check")
	must.NoError(resp2.Err())
	is.Equal(http.StatusOK, resp2.StatusCode())
	is.Equal("cookie_matched", resp2.AsString())
}

func TestSessionFilter(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login" {
			http.SetCookie(w, &http.Cookie{Name: "JSESSIONID", Value: "SESS_ABC_999", Path: "/"})
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.URL.Path == "/profile" {
			c, err := r.Cookie("JSESSIONID")
			if err == nil && c.Value == "SESS_ABC_999" {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"user":"authenticated"}`))
				return
			}
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}))
	t.Cleanup(ts.Close)

	sessionFilter := rest.NewSessionFilter()
	is.Equal("", sessionFilter.SessionID())

	// Login and capture session ID
	resp1 := rest.Given().
		BaseURI(ts.URL).
		Filter(sessionFilter).
		Post("/login")
	must.NoError(resp1.Err())
	is.Equal("SESS_ABC_999", sessionFilter.SessionID())

	// Subsequent request uses tracked session ID
	resp2 := rest.Given().
		BaseURI(ts.URL).
		Filter(sessionFilter).
		Get("/profile")
	must.NoError(resp2.Err())
	is.Equal(http.StatusOK, resp2.StatusCode())

	// Manual override
	sessionFilter.SetSessionID("MANUAL_SESSION")
	is.Equal("MANUAL_SESSION", sessionFilter.SessionID())
}

func TestErrorLoggingFilter(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/err" {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`server failure`))
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(ts.Close)

	var errBuf bytes.Buffer
	errFilter := &rest.ErrorLoggingFilter{Output: &errBuf}

	// Successful request does not log to errBuf
	resp1 := rest.Given().
		BaseURI(ts.URL).
		Filter(errFilter).
		Get("/ok")
	must.NoError(resp1.Err())
	is.Empty(errBuf.String())

	// Failed request logs to errBuf
	resp2 := rest.Given().
		BaseURI(ts.URL).
		Filter(errFilter).
		Get("/err")
	must.NoError(resp2.Err())
	is.Contains(errBuf.String(), "--- Request Details ---")
	is.Contains(errBuf.String(), "--- Response Details ---")
	is.Contains(errBuf.String(), "500")
}

func TestStatusCodeBasedLoggingFilter(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/target" {
			w.WriteHeader(http.StatusTeapot)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(ts.Close)

	var logBuf bytes.Buffer
	filter := &rest.StatusCodeBasedLoggingFilter{
		Predicate: func(code int) bool { return code == http.StatusTeapot },
		Output:    &logBuf,
	}

	// 200 does not log
	resp1 := rest.Given().BaseURI(ts.URL).Filter(filter).Get("/ok")
	must.NoError(resp1.Err())
	is.Empty(logBuf.String())

	// 418 Teapot matches predicate and logs
	resp2 := rest.Given().BaseURI(ts.URL).Filter(filter).Get("/target")
	must.NoError(resp2.Err())
	is.Contains(logBuf.String(), "418")
}

func TestCsrfFilter_And_ExtractCsrfFromHTML(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	tests := []struct {
		name      string
		html      string
		fieldName string
		expected  string
	}{
		{
			name:      "input tag with name and value",
			html:      `<html><body><form><input type="hidden" name="_csrf" value="token_val_input_123" /></form></body></html>`,
			fieldName: "_csrf",
			expected:  "token_val_input_123",
		},
		{
			name:      "meta tag with name and content",
			html:      `<html><head><meta name="_csrf" content="token_val_meta_456" /></head></html>`,
			fieldName: "_csrf",
			expected:  "token_val_meta_456",
		},
		{
			name:      "nonexistent field returns empty string",
			html:      `<html><body><form><input type="hidden" name="_csrf" value="token_val_input_123" /></form></body></html>`,
			fieldName: "nonexistent",
			expected:  "",
		},
		{
			name:      "empty html returns empty string",
			html:      "",
			fieldName: "_csrf",
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)
			got := rest.ExtractCsrfFromHTML(tt.html, tt.fieldName)
			is.Equal(tt.expected, got)
		})
	}

	// Filter token injection
	csrfFilter := rest.NewCsrfFilter("/login", "_csrf")
	csrfFilter.HeaderName = "X-CSRF-TOKEN"
	csrfFilter.SetToken("test_csrf_token")
	is.Equal("test_csrf_token", csrfFilter.Token())

	fReq := &rest.FilterableRequest{
		FormParams: make(map[string][]string),
		Headers:    make(http.Header),
	}
	fCtx := &rest.FilterContext{}
	_, _ = csrfFilter.Filter(fReq, fCtx)

	// Header-mode CSRF: token is set in the header only, not in FormParams.
	is.Nil(fReq.FormParams["_csrf"])
	is.Equal("test_csrf_token", fReq.Headers.Get("X-CSRF-TOKEN"))
}

func TestFormAuthFilter(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	formAuthFilter := rest.NewFormAuthFilter("admin", "secret", rest.SpringSecurityFormAuth())
	is.Equal("admin", formAuthFilter.Username)
	is.Equal("secret", formAuthFilter.Password)
	is.Equal("/login", formAuthFilter.Config.FormAction)
	is.Equal("_csrf", formAuthFilter.Config.CsrfField)
}

func TestNilFilterSafety(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	var (
		nilCookieFilter  *rest.CookieFilter
		nilSessionFilter *rest.SessionFilter
		nilErrorFilter   *rest.ErrorLoggingFilter
		nilStatusFilter  *rest.StatusCodeBasedLoggingFilter
		nilCsrfFilter    *rest.CsrfFilter
		nilFormFilter    *rest.FormAuthFilter
		nilTimingFilter  *rest.TimingFilter
	)

	is.Nil(nilCookieFilter.Jar())
	is.Equal("", nilSessionFilter.SessionID())
	nilSessionFilter.SetSessionID("abc") // should not panic
	is.Equal("", nilCsrfFilter.Token())
	nilCsrfFilter.SetToken("xyz") // should not panic

	req := &rest.FilterableRequest{}
	ctx := &rest.FilterContext{}

	resp, err := nilCookieFilter.Filter(req, ctx)
	is.Nil(resp)
	is.NoError(err)

	resp, err = nilSessionFilter.Filter(req, ctx)
	is.Nil(resp)
	is.NoError(err)

	resp, err = nilErrorFilter.Filter(req, ctx)
	is.Nil(resp)
	is.NoError(err)

	resp, err = nilStatusFilter.Filter(req, ctx)
	is.Nil(resp)
	is.NoError(err)

	resp, err = nilCsrfFilter.Filter(req, ctx)
	is.Nil(resp)
	is.NoError(err)

	resp, err = nilFormFilter.Filter(req, ctx)
	is.Nil(resp)
	is.NoError(err)

	resp, err = nilTimingFilter.Filter(req, ctx)
	is.Nil(resp)
	is.NoError(err)
}

func TestCookieFilter_DomainScoping(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	// Mirrors Java CookieFilterTest.doesntAddCookiesToNonMatchingUrlRequest.
	// Go's cookiejar scopes by host (not port), so we use path-scoping instead:
	// a cookie with Path=/restricted must not be forwarded to /other.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/set":
			http.SetCookie(w, &http.Cookie{Name: "path_cookie", Value: "secret", Path: "/restricted"})
			w.WriteHeader(http.StatusOK)
		case "/other":
			if c, err := r.Cookie("path_cookie"); err == nil {
				_, _ = w.Write([]byte("got:" + c.Value))
			} else {
				_, _ = w.Write([]byte("no-cookie"))
			}
			w.WriteHeader(http.StatusOK)
		case "/restricted":
			if c, err := r.Cookie("path_cookie"); err == nil {
				_, _ = w.Write([]byte("got:" + c.Value))
			} else {
				_, _ = w.Write([]byte("no-cookie"))
			}
			w.WriteHeader(http.StatusOK)
		}
	}))
	t.Cleanup(ts.Close)

	cookieFilter := rest.NewCookieFilter()

	// First request sets the cookie in the jar (Path=/restricted)
	resp1 := rest.Given().BaseURI(ts.URL).Filter(cookieFilter).Get("/set")
	must.NoError(resp1.Err())

	// Request to /other must NOT carry the path-scoped cookie
	resp2 := rest.Given().BaseURI(ts.URL).Filter(cookieFilter).Get("/other")
	must.NoError(resp2.Err())
	is.Equal("no-cookie", resp2.AsString())

	// Request to /restricted MUST carry the path-scoped cookie
	resp3 := rest.Given().BaseURI(ts.URL).Filter(cookieFilter).Get("/restricted")
	must.NoError(resp3.Err())
	is.Equal("got:secret", resp3.AsString())
}

func TestCookieFilter_PreserveExistingCookies(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	// Mirrors Java CookieFilterTest.preserveCookies.
	// A cookie manually set on the request must not be overwritten by the jar.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/set" {
			// Response sets cookieName2; request already had cookieName1
			http.SetCookie(w, &http.Cookie{Name: "cookieName2", Value: "cookieValue2", Path: "/"})
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.URL.Path == "/check" {
			c1, err1 := r.Cookie("cookieName1")
			c2, err2 := r.Cookie("cookieName2")
			if err1 == nil && err2 == nil {
				_, _ = w.Write([]byte(c1.Value + "+" + c2.Value))
			}
			w.WriteHeader(http.StatusOK)
			return
		}
	}))
	t.Cleanup(ts.Close)

	cookieFilter := rest.NewCookieFilter()

	// First request: manually set cookieName1; response sets cookieName2
	resp1 := rest.Given().
		BaseURI(ts.URL).
		Cookie("cookieName1", "cookieInitialValue").
		Filter(cookieFilter).
		Get("/set")
	must.NoError(resp1.Err())

	// Second request: both cookies must be present; cookieName1 keeps its original value
	resp2 := rest.Given().
		BaseURI(ts.URL).
		Cookie("cookieName1", "cookieInitialValue").
		Filter(cookieFilter).
		Get("/check")
	must.NoError(resp2.Err())
	is.Equal("cookieInitialValue+cookieValue2", resp2.AsString())
}

func TestCookieParsing(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	// Mirrors Java CookieMatcherTest.testSetVersion.
	// Verifies that Set-Cookie header strings are parsed into structured cookies
	// with the correct name, value, domain, path, MaxAge, Secure, and HttpOnly flags.
	lines := []string{
		"DEVICE_ID=123; Domain=.test.com; Expires=Thu, 12 Oct 2023 09:34:31 GMT; Path=/; Secure; HttpOnly; SameSite=Lax",
		"SPRING_SECURITY_REMEMBER_ME_COOKIE=12345; Domain=.test.com; Path=/; Max-Age=1209600",
		"COOKIE_WITH_ZERO_MAX_AGE=1234; Domain=.test.com; Path=/; Max-Age=0",
		"COOKIE_WITH_NEGATIVE_MAX_AGE=123456; Domain=.test.com; Path=/; Max-Age=-1",
	}
	cookies := parseCookies(lines)
	is.Len(cookies, 4)

	// Index the cookies by name for assertion
	byName := make(map[string]*http.Cookie, len(cookies))
	for _, c := range cookies {
		byName[c.Name] = c
	}

	spring := byName["SPRING_SECURITY_REMEMBER_ME_COOKIE"]
	is.Equal("12345", spring.Value)
	is.Equal(".test.com", spring.Domain)
	is.Equal("/", spring.Path)
	is.Equal(1209600, spring.MaxAge)
	is.False(spring.Secure)
	is.False(spring.HttpOnly)

	// Go's net/http maps both Max-Age=0 and Max-Age<0 to MaxAge=-1 ("delete immediately").
	// Java preserves the literal value; Go normalises both to -1.
	zeroAge := byName["COOKIE_WITH_ZERO_MAX_AGE"]
	is.Equal("1234", zeroAge.Value)
	is.Equal(".test.com", zeroAge.Domain)
	is.Equal("/", zeroAge.Path)
	is.Equal(-1, zeroAge.MaxAge) // Go normalises Max-Age=0 → -1

	negAge := byName["COOKIE_WITH_NEGATIVE_MAX_AGE"]
	is.Equal("123456", negAge.Value)
	is.Equal(".test.com", negAge.Domain)
	is.Equal("/", negAge.Path)
	is.Equal(-1, negAge.MaxAge)

	device := byName["DEVICE_ID"]
	is.Equal("123", device.Value)
	is.Equal(".test.com", device.Domain)
	is.Equal("/", device.Path)
	is.True(device.Secure)
	is.True(device.HttpOnly)
}

func TestCookieParsing_EmptyCookieValue(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	// Mirrors Java CookieMatcherTest.deals_with_empty_cookie_values.
	// An empty string in the list results in no cookie being parsed for that entry;
	// the non-empty entries are still parsed correctly.
	lines := []string{
		"un=bob; domain=bob.com; path=/",
		"",
		"_session_id=asdfwerwersdfwere; domain=bob.com; path=/; HttpOnly",
	}
	cookies := parseCookies(lines)
	// The two valid lines produce two cookies; empty line produces nothing.
	is.Len(cookies, 2)
	byName := make(map[string]*http.Cookie, len(cookies))
	for _, c := range cookies {
		byName[c.Name] = c
	}
	is.Equal("bob", byName["un"].Value)
	is.True(byName["_session_id"].HttpOnly)
}

// TestCookieFilter_DuplicateNameDifferentPath mirrors Java's
// CookieFilterTest.addDuplicateNameCookiesLikeInBrowser.
// Java's CookieFilter(true) allows sending duplicate-named cookies from different paths.
// Go's CookieFilter deduplicates by name (preserves the first/most-specific match),
// so the jar stores both entries but only the most specific one is forwarded per request.
func TestCookieFilter_DuplicateNameDifferentPath(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/set":
			// Two cookies with same name but different paths — stored as separate jar entries
			http.SetCookie(w, &http.Cookie{Name: "cookieName", Value: "xxx", Path: "/bar"})
			http.SetCookie(w, &http.Cookie{Name: "cookieName", Value: "yyy", Path: "/"})
			w.WriteHeader(http.StatusOK)
		case "/bar/check":
			var vals []string
			for _, c := range r.Cookies() {
				if c.Name == "cookieName" {
					vals = append(vals, c.Value)
				}
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(strings.Join(vals, ",")))
		case "/root":
			var vals []string
			for _, c := range r.Cookies() {
				if c.Name == "cookieName" {
					vals = append(vals, c.Value)
				}
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(strings.Join(vals, ",")))
		}
	}))
	t.Cleanup(ts.Close)

	cookieFilter := rest.NewCookieFilter()

	resp1 := rest.Given().BaseURI(ts.URL).Filter(cookieFilter).Get("/set")
	must.NoError(resp1.Err())

	// /bar/check matches both Path=/bar (more specific, sent first by jar) and Path=/.
	// Go's CookieFilter deduplicates by name — only the most-specific path cookie is forwarded.
	resp2 := rest.Given().BaseURI(ts.URL).Filter(cookieFilter).Get("/bar/check")
	must.NoError(resp2.Err())
	is.Contains(resp2.AsString(), "xxx") // /bar path is more specific

	// /root only matches Path=/ — the root-scoped cookie is forwarded
	resp3 := rest.Given().BaseURI(ts.URL).Filter(cookieFilter).Get("/root")
	must.NoError(resp3.Err())
	is.Contains(resp3.AsString(), "yyy") // / path matches /root
}
