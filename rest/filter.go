package rest

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

// FilterableRequest exposes request details for filter inspection and modification.
// Ctx carries the request context so filters can respect cancellation and deadlines.
type FilterableRequest struct {
	Ctx         context.Context
	Method      string
	URI         string
	Headers     http.Header
	Cookies     []*http.Cookie
	QueryParams map[string][]string
	FormParams  map[string][]string
	PathParams  map[string]string
	Body        []byte
	ContentType string
}

// Filter defines the middleware interface for request/response interception.
type Filter interface {
	Filter(req *FilterableRequest, ctx *FilterContext) (*Response, error)
}

// OrderedFilter allows filters to specify an explicit execution order (lower executed first).
type OrderedFilter interface {
	Filter
	Order() int
}

// FilterContext coordinates filter chain execution.
type FilterContext struct {
	filters  []Filter
	index    int
	executor func(req *FilterableRequest) (*Response, error)
}

// Next executes the next filter in the chain or the final HTTP request executor.
// Each call creates an immutable sub-context starting one position ahead, so a filter
// can call Next more than once (e.g. for retry logic) without corrupting the shared index.
func (fc *FilterContext) Next(req *FilterableRequest) (*Response, error) {
	if fc == nil {
		return nil, nil
	}
	if fc.index < len(fc.filters) {
		currentFilter := fc.filters[fc.index]
		subCtx := &FilterContext{
			filters:  fc.filters,
			index:    fc.index + 1,
			executor: fc.executor,
		}
		return currentFilter.Filter(req, subCtx)
	}
	if fc.executor != nil {
		return fc.executor(req)
	}
	return nil, nil
}

// TimingFilter measures request execution time.
type TimingFilter struct {
	OnComplete func(d time.Duration)
}

// NewTimingFilter creates a new TimingFilter with the given callback.
func NewTimingFilter(onComplete func(d time.Duration)) *TimingFilter {
	return &TimingFilter{OnComplete: onComplete}
}

// Filter records execution duration and calls the callback.
func (tf *TimingFilter) Filter(req *FilterableRequest, ctx *FilterContext) (*Response, error) {
	if tf == nil {
		return ctx.Next(req)
	}
	start := time.Now()
	resp, err := ctx.Next(req)
	duration := time.Since(start)
	if tf.OnComplete != nil {
		tf.OnComplete(duration)
	}
	return resp, err
}

// RequestLoggingFilter logs request details according to the LogDetail setting.
type RequestLoggingFilter struct {
	Detail LogDetail
	Output io.Writer
}

// NewRequestLoggingFilter creates a new RequestLoggingFilter.
func NewRequestLoggingFilter(out io.Writer, detail LogDetail) *RequestLoggingFilter {
	return &RequestLoggingFilter{Output: out, Detail: detail}
}

// Filter logs the request and proceeds to the next filter.
func (rlf *RequestLoggingFilter) Filter(req *FilterableRequest, ctx *FilterContext) (*Response, error) {
	if rlf != nil && rlf.Output != nil && rlf.Detail != LogDetailNone {
		logRequest(rlf.Output, req, rlf.Detail, nil)
	}
	return ctx.Next(req)
}

// ResponseLoggingFilter logs response details.
type ResponseLoggingFilter struct {
	Detail LogDetail
	Output io.Writer
}

// NewResponseLoggingFilter creates a new ResponseLoggingFilter.
func NewResponseLoggingFilter(out io.Writer, detail LogDetail) *ResponseLoggingFilter {
	return &ResponseLoggingFilter{Output: out, Detail: detail}
}

// Filter processes the request and logs the resulting response.
func (rlf *ResponseLoggingFilter) Filter(req *FilterableRequest, ctx *FilterContext) (*Response, error) {
	resp, err := ctx.Next(req)
	if err == nil && resp != nil && rlf != nil && rlf.Output != nil && rlf.Detail != LogDetailNone {
		logResponse(rlf.Output, resp, rlf.Detail)
	}
	return resp, err
}

// CookieFilter maintains an RFC6265 compliant cookie jar across HTTP requests.
type CookieFilter struct {
	mu  sync.RWMutex
	jar http.CookieJar
}

// NewCookieFilter creates a new CookieFilter with a standard in-memory cookie jar.
func NewCookieFilter() *CookieFilter {
	jar, err := cookiejar.New(nil)
	if err != nil {
		// cookiejar.New(nil) never returns an error in the standard library;
		// panic here surfaces any future stdlib change immediately rather than
		// producing a silently broken filter with a nil jar.
		panic(fmt.Sprintf("restassured: failed to create cookie jar: %v", err))
	}
	return &CookieFilter{jar: jar}
}

// Jar returns the underlying http.CookieJar.
func (cf *CookieFilter) Jar() http.CookieJar {
	if cf == nil {
		return nil
	}
	cf.mu.RLock()
	defer cf.mu.RUnlock()
	return cf.jar
}

// Filter applies stored cookies to outgoing requests and collects Set-Cookie headers from responses.
func (cf *CookieFilter) Filter(req *FilterableRequest, ctx *FilterContext) (*Response, error) {
	if cf == nil {
		return ctx.Next(req)
	}

	parsedURL, parseErr := url.Parse(req.URI)
	if parseErr == nil && cf.jar != nil {
		cf.mu.RLock()
		jarCookies := cf.jar.Cookies(parsedURL)
		cf.mu.RUnlock()

		for _, jc := range jarCookies {
			exists := false
			for _, rc := range req.Cookies {
				if rc.Name == jc.Name {
					exists = true
					break
				}
			}
			if !exists {
				req.Cookies = append(req.Cookies, jc)
			}
		}
	}

	resp, err := ctx.Next(req)
	if err == nil && resp != nil && cf.jar != nil && parseErr == nil {
		cf.mu.Lock()
		cf.jar.SetCookies(parsedURL, resp.Cookies())
		cf.mu.Unlock()
	}
	return resp, err
}

// SessionFilter tracks and injects session identifiers (e.g. JSESSIONID, session_id) across requests.
type SessionFilter struct {
	mu          sync.RWMutex
	sessionID   string
	sessionName string
}

// NewSessionFilter creates a new SessionFilter for the given cookie/header session identifier (default "JSESSIONID").
func NewSessionFilter(sessionName ...string) *SessionFilter {
	name := "JSESSIONID"
	if len(sessionName) > 0 && sessionName[0] != "" {
		name = sessionName[0]
	}
	return &SessionFilter{
		sessionName: name,
	}
}

// SessionID returns the currently tracked session ID.
func (sf *SessionFilter) SessionID() string {
	if sf == nil {
		return ""
	}
	sf.mu.RLock()
	defer sf.mu.RUnlock()
	return sf.sessionID
}

// SetSessionID sets the tracked session ID.
func (sf *SessionFilter) SetSessionID(id string) {
	if sf == nil {
		return
	}
	sf.mu.Lock()
	defer sf.mu.Unlock()
	sf.sessionID = id
}

// Filter injects the session cookie if tracked, and captures new session cookies from responses.
func (sf *SessionFilter) Filter(req *FilterableRequest, ctx *FilterContext) (*Response, error) {
	if sf == nil {
		return ctx.Next(req)
	}

	sf.mu.RLock()
	sid := sf.sessionID
	sname := sf.sessionName
	sf.mu.RUnlock()

	if sid != "" {
		hasCookie := false
		for _, c := range req.Cookies {
			if c.Name == sname {
				hasCookie = true
				break
			}
		}
		if !hasCookie {
			req.Cookies = append(req.Cookies, &http.Cookie{
				Name:  sname,
				Value: sid,
			})
		}
	}

	resp, err := ctx.Next(req)
	if err == nil && resp != nil {
		for _, c := range resp.Cookies() {
			if c.Name == sname && c.Value != "" {
				sf.SetSessionID(c.Value)
				break
			}
		}
	}
	return resp, err
}

// ErrorLoggingFilter logs request and response details only when response status is >= 400 or an error occurs.
type ErrorLoggingFilter struct {
	Output io.Writer
}

// Filter executes the request and logs if status code is >= 400, a transport error
// occurred, or the response itself carries an error (e.g. network timeout).
func (elf *ErrorLoggingFilter) Filter(req *FilterableRequest, ctx *FilterContext) (*Response, error) {
	if elf == nil {
		return ctx.Next(req)
	}
	resp, err := ctx.Next(req)
	out := elf.Output
	if out == nil {
		out = os.Stderr
	}

	if err != nil || (resp != nil && (resp.Err() != nil || resp.StatusCode() >= 400)) {
		var blacklist []string
		if resp != nil {
			blacklist = resp.Config().LogConfig().BlacklistHeaders
		}
		logRequest(out, req, LogDetailAll, blacklist)
		if resp != nil {
			logResponse(out, resp, LogDetailAll)
		}
	}
	return resp, err
}

// StatusCodeBasedLoggingFilter logs request and response details if a predicate returns true for the status code.
type StatusCodeBasedLoggingFilter struct {
	Predicate func(statusCode int) bool
	Output    io.Writer
}

// Filter executes the request and logs if Predicate matches the response status code.
func (sblf *StatusCodeBasedLoggingFilter) Filter(req *FilterableRequest, ctx *FilterContext) (*Response, error) {
	if sblf == nil {
		return ctx.Next(req)
	}
	resp, err := ctx.Next(req)
	out := sblf.Output
	if out == nil {
		out = os.Stderr
	}

	if resp != nil && sblf.Predicate != nil && sblf.Predicate(resp.StatusCode()) {
		blacklist := resp.Config().LogConfig().BlacklistHeaders
		logRequest(out, req, LogDetailAll, blacklist)
		logResponse(out, resp, LogDetailAll)
	}
	return resp, err
}

// CsrfFilter extracts CSRF tokens from HTML pages and automatically injects them into outgoing requests.
type CsrfFilter struct {
	mu         sync.RWMutex
	CsrfURI    string
	TokenField string
	HeaderName string
	token      string
}

// NewCsrfFilter creates a new CsrfFilter.
func NewCsrfFilter(csrfURI string, tokenField ...string) *CsrfFilter {
	field := "_csrf"
	if len(tokenField) > 0 && tokenField[0] != "" {
		field = tokenField[0]
	}
	return &CsrfFilter{
		CsrfURI:    csrfURI,
		TokenField: field,
	}
}

// Token returns the extracted CSRF token.
func (cf *CsrfFilter) Token() string {
	if cf == nil {
		return ""
	}
	cf.mu.RLock()
	defer cf.mu.RUnlock()
	return cf.token
}

// SetToken manually updates the CSRF token.
func (cf *CsrfFilter) SetToken(token string) {
	if cf == nil {
		return
	}
	cf.mu.Lock()
	defer cf.mu.Unlock()
	cf.token = token
}

// Filter fetches the CSRF token from CsrfURI on first use (if not already set), then injects it.
func (cf *CsrfFilter) Filter(req *FilterableRequest, ctx *FilterContext) (*Response, error) {
	if cf == nil {
		return ctx.Next(req)
	}

	cf.mu.RLock()
	token := cf.token
	csrfURI := cf.CsrfURI
	field := cf.TokenField
	headerName := cf.HeaderName
	cf.mu.RUnlock()

	// Auto-fetch token from CsrfURI when not yet set.
	// Double-checked locking: the RLock fast-path avoids contention once the token
	// is cached; the WLock re-check prevents duplicate fetches under concurrency.
	if token == "" && csrfURI != "" {
		fetched := cf.fetchCSRFToken(req, csrfURI, field)
		cf.mu.Lock()
		if cf.token == "" {
			cf.token = fetched
		}
		token = cf.token
		cf.mu.Unlock()
	}

	if token != "" {
		if headerName != "" && req.Headers != nil {
			req.Headers.Set(headerName, token)
		} else if headerName == "" {
			// Form-based CSRF: inject into the URL-encoded body only when the method
			// supports a body (POST/PUT/PATCH) and the content type is form-encoded
			// (or the body is empty, in which case we initialise a form body).
			// Skipping this block prevents corrupting JSON/XML payloads and avoids
			// injecting a body into GET/HEAD/OPTIONS requests.
			method := strings.ToUpper(req.Method)
			isBodyMethod := method == "POST" || method == "PUT" || method == "PATCH"
			isFormEncoded := strings.HasPrefix(strings.ToLower(req.ContentType), "application/x-www-form-urlencoded")
			if isBodyMethod && (isFormEncoded || len(req.Body) == 0) {
				encoded := url.QueryEscape(field) + "=" + url.QueryEscape(token)
				if len(req.Body) > 0 {
					req.Body = append(req.Body, '&')
					req.Body = append(req.Body, []byte(encoded)...)
				} else {
					req.Body = []byte(encoded)
					if req.Headers != nil {
						req.Headers.Set("Content-Type", string(ContentTypeURLEncoded))
					}
				}
			}
			if req.FormParams != nil {
				req.FormParams[field] = []string{token}
			}
		}
	}

	return ctx.Next(req)
}

// fetchCSRFToken performs a GET to csrfURI and extracts the token from the HTML response.
// Must be called while cf.mu write-lock is held by the caller.
func (cf *CsrfFilter) fetchCSRFToken(req *FilterableRequest, csrfURI, field string) string {
	fetchURL := csrfURI
	if !strings.HasPrefix(fetchURL, "http://") && !strings.HasPrefix(fetchURL, "https://") {
		if u, err := url.Parse(req.URI); err == nil {
			fetchURL = fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, fetchURL)
		}
	}
	getReq, err := http.NewRequestWithContext(req.Ctx, http.MethodGet, fetchURL, nil)
	if err != nil {
		return ""
	}
	for _, c := range req.Cookies {
		getReq.AddCookie(c)
	}
	if auth := req.Headers.Get("Authorization"); auth != "" {
		getReq.Header.Set("Authorization", auth)
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(getReq)
	if err != nil {
		return ""
	}
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	return ExtractCsrfFromHTML(string(body), field)
}

// ExtractCsrfFromHTML parses an HTML document to find a CSRF input field or meta tag value.
func ExtractCsrfFromHTML(htmlContent string, fieldName string) string {
	if htmlContent == "" || fieldName == "" {
		return ""
	}
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return ""
	}

	var foundToken string
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if foundToken != "" {
			return
		}
		if n.Type == html.ElementNode {
			if n.Data == "input" {
				var nameVal, valueVal string
				for _, attr := range n.Attr {
					if attr.Key == "name" {
						nameVal = attr.Val
					} else if attr.Key == "value" {
						valueVal = attr.Val
					}
				}
				if nameVal == fieldName {
					foundToken = valueVal
					return
				}
			} else if n.Data == "meta" {
				var nameVal, contentVal string
				for _, attr := range n.Attr {
					if attr.Key == "name" {
						nameVal = attr.Val
					} else if attr.Key == "content" {
						contentVal = attr.Val
					}
				}
				if nameVal == fieldName {
					foundToken = contentVal
					return
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)
	return foundToken
}

// FormAuthFilter coordinates automated form login before request execution.
type FormAuthFilter struct {
	mu            sync.RWMutex
	Username      string
	Password      string
	Config        *FormAuthConfig
	SessionFilter *SessionFilter
	httpClient    *http.Client
}

// NewFormAuthFilter creates a new FormAuthFilter.
func NewFormAuthFilter(username, password string, cfg ...*FormAuthConfig) *FormAuthFilter {
	var formCfg *FormAuthConfig
	if len(cfg) > 0 && cfg[0] != nil {
		formCfg = cfg[0]
	} else {
		formCfg = DefaultFormAuthConfig()
	}
	timeout := formCfg.LoginTimeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &FormAuthFilter{
		Username:      username,
		Password:      password,
		Config:        formCfg,
		SessionFilter: NewSessionFilter(),
		httpClient:    &http.Client{Timeout: timeout},
	}
}

// Filter delegates request processing and handles session cookies from form authentication.
func (faf *FormAuthFilter) Filter(req *FilterableRequest, ctx *FilterContext) (*Response, error) {
	if faf == nil {
		return ctx.Next(req)
	}

	faf.mu.Lock()
	if faf.SessionFilter == nil {
		faf.SessionFilter = NewSessionFilter()
	}
	hasSession := faf.SessionFilter.SessionID() != ""
	faf.mu.Unlock()

	if !hasSession && faf.Config != nil && faf.Config.FormAction != "" {
		loginURL := faf.Config.FormAction
		if !strings.HasPrefix(loginURL, "http://") && !strings.HasPrefix(loginURL, "https://") {
			if u, err := url.Parse(req.URI); err == nil {
				loginURL = fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, loginURL)
			}
		}

		client := faf.httpClient
		if client == nil {
			client = &http.Client{Timeout: 10 * time.Second}
		}

		// Optional: Fetch login page for CSRF token
		csrfToken := ""
		if faf.Config.CsrfField != "" {
			getReq, err := http.NewRequestWithContext(req.Ctx, http.MethodGet, loginURL, nil)
			if err == nil {
				// Propagate cookies and auth from the current request context.
				for _, c := range req.Cookies {
					getReq.AddCookie(c)
				}
				if auth := req.Headers.Get("Authorization"); auth != "" {
					getReq.Header.Set("Authorization", auth)
				}
				if getResp, err := client.Do(getReq); err == nil {
					body, _ := io.ReadAll(getResp.Body)
					_ = getResp.Body.Close()
					csrfToken = ExtractCsrfFromHTML(string(body), faf.Config.CsrfField)
				}
			}
		}

		// Submit credentials via POST
		formValues := url.Values{}
		userField := faf.Config.UsernameField
		if userField == "" {
			userField = "username"
		}
		passField := faf.Config.PasswordField
		if passField == "" {
			passField = "password"
		}
		formValues.Set(userField, faf.Username)
		formValues.Set(passField, faf.Password)
		if faf.Config.CsrfField != "" && csrfToken != "" {
			formValues.Set(faf.Config.CsrfField, csrfToken)
		}

		postReq, err := http.NewRequestWithContext(req.Ctx, http.MethodPost, loginURL, strings.NewReader(formValues.Encode()))
		if err == nil {
			postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			// Do not follow redirects: many frameworks (Spring Security, Laravel, etc.)
			// respond with 302 + Set-Cookie. If the redirect is followed automatically,
			// the final response has no Set-Cookie and the session cookie is lost.
			noRedirectClient := &http.Client{
				Timeout: client.Timeout,
				CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
					return http.ErrUseLastResponse
				},
			}
			if postResp, err := noRedirectClient.Do(postReq); err == nil {
				sessionName := faf.SessionFilter.sessionName
				for _, c := range postResp.Cookies() {
					if c.Name == sessionName {
						faf.SessionFilter.SetSessionID(c.Value)
						break
					}
				}
				_ = postResp.Body.Close()
			}
		}
	}

	return faf.SessionFilter.Filter(req, ctx)
}
