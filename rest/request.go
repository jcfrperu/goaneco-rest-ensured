package rest

import (
	"bytes"
	"cmp"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Request is the builder for configuring and executing an HTTP request.
type Request struct {
	ctx                context.Context
	ctxCancel          context.CancelFunc // non-nil when Timeout() created a child context
	baseURI            string
	basePath           string
	port               int
	rootPath           string
	urlEncodingEnabled bool
	headers            http.Header
	cookies            []*http.Cookie
	params             url.Values // method-agnostic params resolved to query or form at execution time
	queryParams        url.Values
	formParams         url.Values
	pathParams         map[string]string
	body               []byte
	contentType        string
	accept             string
	auth               AuthScheme
	filters            []Filter
	config             *Config
	insecureSkipVerify bool
	multiparts         []MultiPart
	buildErr           error
}

// NewRequest creates a new Request initialized with current global defaults.
func NewRequest() *Request {
	var (
		gBaseURI      string
		gBasePath     string
		gPort         int
		gInsecureSkip bool
		gURLEncoding  bool
		gRootPath     string
		gConfig       *Config
		gFilters      []Filter
		gHeaders      http.Header
		gCookies      []*http.Cookie
		gQueryParams  url.Values
		gAuth         AuthScheme
	)

	globalMu.RLock()
	gBaseURI = globalBaseURI
	gBasePath = globalBasePath
	gPort = globalPort
	gInsecureSkip = globalInsecureSkipVerify
	gURLEncoding = globalURLEncodingEnabled
	gRootPath = globalRootPath
	if globalConfig != nil {
		gConfig = globalConfig.Clone()
	} else {
		gConfig = DefaultConfig()
	}
	if len(globalFilters) > 0 {
		gFilters = append([]Filter{}, globalFilters...)
	}
	if len(globalHeaders) > 0 {
		gHeaders = make(http.Header, len(globalHeaders))
		for k, vs := range globalHeaders {
			gHeaders[k] = append([]string{}, vs...)
		}
	} else {
		gHeaders = make(http.Header)
	}
	if len(globalCookies) > 0 {
		gCookies = make([]*http.Cookie, len(globalCookies))
		for i, c := range globalCookies {
			clone := *c
			gCookies[i] = &clone
		}
	} else {
		gCookies = make([]*http.Cookie, 0)
	}
	if len(globalQueryParams) > 0 {
		gQueryParams = make(url.Values, len(globalQueryParams))
		for k, vs := range globalQueryParams {
			gQueryParams[k] = append([]string{}, vs...)
		}
	} else {
		gQueryParams = make(url.Values)
	}
	gAuth = globalAuth
	globalMu.RUnlock()

	return &Request{
		ctx:                context.Background(),
		baseURI:            gBaseURI,
		basePath:           gBasePath,
		port:               gPort,
		rootPath:           gRootPath,
		urlEncodingEnabled: gURLEncoding,
		headers:            gHeaders,
		cookies:            gCookies,
		params:             make(url.Values),
		queryParams:        gQueryParams,
		formParams:         make(url.Values),
		pathParams:         make(map[string]string),
		auth:               gAuth,
		filters:            gFilters,
		config:             gConfig,
		insecureSkipVerify: gInsecureSkip,
		multiparts:         make([]MultiPart, 0),
	}
}

// Context attaches a context.Context to the request for cancellation and timeouts.
func (r *Request) Context(ctx context.Context) *Request {
	if ctx != nil {
		r.ctx = ctx
	}
	return r
}

// Timeout creates a child context with the given deadline and attaches it to the request.
// The cancel function is called automatically when Execute() returns, so no goroutine or
// timer leak occurs. Use Context() directly when you need external cancellation control.
func (r *Request) Timeout(d time.Duration) *Request {
	if r.ctxCancel != nil {
		r.ctxCancel() // release any previous timeout context
	}
	ctx, cancel := context.WithTimeout(r.ctx, d)
	r.ctx = ctx
	r.ctxCancel = cancel
	return r
}

// BaseURI sets the base URL for the request (e.g. "https://api.example.com").
func (r *Request) BaseURI(uri string) *Request {
	r.baseURI = uri
	return r
}

// BasePath sets the base path appended to the base URI (e.g. "/v1").
func (r *Request) BasePath(path string) *Request {
	r.basePath = path
	return r
}

// Port sets an explicit TCP port for the request.
func (r *Request) Port(port int) *Request {
	r.port = port
	return r
}

// URLEncodingEnabled configures whether URL query parameters are automatically percent-encoded.
func (r *Request) URLEncodingEnabled(enabled bool) *Request {
	r.urlEncodingEnabled = enabled
	return r
}

// Header sets a single request header, overwriting any existing value for that name.
// To accumulate multiple values for the same header, use a RequestSpecBuilder instead.
func (r *Request) Header(name, value string) *Request {
	r.headers.Set(name, value)
	return r
}

// Headers sets multiple headers on the request, overwriting any existing values.
func (r *Request) Headers(headers map[string]string) *Request {
	for k, v := range headers {
		r.headers.Set(k, v)
	}
	return r
}

// Cookie adds a cookie by name and value.
func (r *Request) Cookie(name, value string) *Request {
	r.cookies = append(r.cookies, &http.Cookie{Name: name, Value: value})
	return r
}

// Cookies adds multiple cookies from a map.
func (r *Request) Cookies(cookies map[string]string) *Request {
	for k, v := range cookies {
		r.cookies = append(r.cookies, &http.Cookie{Name: k, Value: v})
	}
	return r
}

// Param sets a method-agnostic parameter (query param for GET/DELETE, form param for POST/PUT/PATCH).
func (r *Request) Param(name string, values ...any) *Request {
	for _, v := range values {
		r.params.Add(name, fmt.Sprintf("%v", v))
	}
	return r
}

// Params sets multiple method-agnostic parameters.
func (r *Request) Params(params map[string]any) *Request {
	for k, v := range params {
		r.Param(k, v)
	}
	return r
}

// QueryParam explicitly sets a query parameter.
func (r *Request) QueryParam(name string, values ...any) *Request {
	for _, v := range values {
		r.queryParams.Add(name, fmt.Sprintf("%v", v))
	}
	return r
}

// QueryParams sets multiple query parameters from a map.
func (r *Request) QueryParams(params map[string]any) *Request {
	for k, v := range params {
		r.QueryParam(k, v)
	}
	return r
}

// FormParam explicitly sets a form URL-encoded parameter.
func (r *Request) FormParam(name string, values ...any) *Request {
	for _, v := range values {
		r.formParams.Add(name, fmt.Sprintf("%v", v))
	}
	return r
}

// FormParams sets multiple form parameters from a map.
func (r *Request) FormParams(params map[string]any) *Request {
	for k, v := range params {
		r.FormParam(k, v)
	}
	return r
}

// PathParam replaces named placeholders in the URL path (e.g. {id} or :id).
func (r *Request) PathParam(name string, value any) *Request {
	r.pathParams[name] = fmt.Sprintf("%v", value)
	return r
}

// PathParams sets multiple named path parameters.
func (r *Request) PathParams(params map[string]any) *Request {
	for k, v := range params {
		r.PathParam(k, v)
	}
	return r
}

// Body sets the raw request body as string.
func (r *Request) Body(body string) *Request {
	r.body = []byte(body)
	return r
}

// BodyBytes sets the raw byte slice body.
func (r *Request) BodyBytes(body []byte) *Request {
	r.body = append([]byte{}, body...)
	return r
}

// BodyObject serializes the given object to JSON and sets it as request body.
// If marshaling fails the error is stored and surfaced when Execute() is called.
func (r *Request) BodyObject(obj any) *Request {
	data, err := json.Marshal(obj)
	if err != nil {
		r.buildErr = fmt.Errorf("BodyObject: %w", err)
		return r
	}
	r.body = data
	if r.contentType == "" {
		r.contentType = string(ContentTypeJSON)
	}
	return r
}

// ContentType sets the Content-Type header.
func (r *Request) ContentType(ct ContentType) *Request {
	r.contentType = string(ct)
	return r
}

// Accept sets the Accept header.
func (r *Request) Accept(ct ContentType) *Request {
	r.accept = string(ct)
	return r
}

// Auth returns the AuthSpec helper for configuring authentication.
func (r *Request) Auth() *AuthSpec {
	return &AuthSpec{req: r}
}

// Filter attaches one or more filters (middleware) to the request.
func (r *Request) Filter(filters ...Filter) *Request {
	r.filters = append(r.filters, filters...)
	return r
}

// Config sets custom configuration for this request.
func (r *Request) Config(cfg *Config) *Request {
	r.config = cfg
	return r
}

// Spec applies a reusable RequestSpec configuration to this request.
func (r *Request) Spec(spec *RequestSpec) *Request {
	if spec == nil {
		return r
	}
	if spec.BaseURI != "" {
		r.baseURI = spec.BaseURI
	}
	if spec.BasePath != "" {
		r.basePath = spec.BasePath
	}
	if spec.PortSet {
		r.port = spec.Port
	}
	if spec.ContentType != "" {
		r.contentType = spec.ContentType
	}
	if spec.Accept != "" {
		r.accept = spec.Accept
	}
	if spec.Auth != nil {
		r.auth = spec.Auth
	}
	if spec.Config != nil {
		r.config = spec.Config
	}
	if spec.InsecureSkipVerify {
		r.insecureSkipVerify = true
	}
	if len(spec.Body) > 0 {
		r.body = append([]byte{}, spec.Body...)
	}
	for k, values := range spec.Headers {
		for i, v := range values {
			// Use Set only on the first value to replace any previously inherited value;
			// subsequent values for the same header are always accumulated with Add so that
			// the full multi-value slice from the spec is preserved.
			if i == 0 && r.config != nil && r.config.HeaderConfig().ShouldOverwriteHeaderWithName(k) {
				r.headers.Set(k, v)
			} else {
				r.headers.Add(k, v)
			}
		}
	}
	for k, values := range spec.QueryParams {
		for _, v := range values {
			r.queryParams.Add(k, v)
		}
	}
	for k, values := range spec.FormParams {
		for _, v := range values {
			r.formParams.Add(k, v)
		}
	}
	for k, v := range spec.PathParams {
		r.pathParams[k] = v
	}
	r.cookies = append(r.cookies, spec.Cookies...)
	r.filters = append(r.filters, spec.Filters...)
	return r
}

// Log returns RequestLogSpec for logging configuration.
func (r *Request) Log() *RequestLogSpec {
	return &RequestLogSpec{req: r}
}

// MultiPart adds a multipart file or field without a filename.
func (r *Request) MultiPart(control string, data []byte, contentType ...string) *Request {
	ct := ""
	if len(contentType) > 0 {
		ct = contentType[0]
	}
	r.multiparts = append(r.multiparts, MultiPart{
		ControlName: control,
		Content:     data,
		ContentType: ct,
	})
	return r
}

// MultiPartNamed adds a multipart data upload with an explicit filename.
func (r *Request) MultiPartNamed(control, filename string, data []byte, contentType ...string) *Request {
	ct := ""
	if len(contentType) > 0 {
		ct = contentType[0]
	}
	r.multiparts = append(r.multiparts, MultiPart{
		ControlName: control,
		FileName:    filename,
		Content:     data,
		ContentType: ct,
	})
	return r
}

// MultiPartFile adds a local file as a multipart upload.
func (r *Request) MultiPartFile(control, filePath string, contentType ...string) *Request {
	ct := ""
	if len(contentType) > 0 {
		ct = contentType[0]
	}
	r.multiparts = append(r.multiparts, MultiPart{
		ControlName: control,
		FilePath:    filePath,
		ContentType: ct,
	})
	return r
}

// RelaxedHTTPSValidation disables SSL/TLS certificate verification.
func (r *Request) RelaxedHTTPSValidation() *Request {
	r.insecureSkipVerify = true
	return r
}

// Given returns the request builder for fluent chaining.
func (r *Request) Given() *Request { return r }

// When returns the request builder for fluent chaining.
func (r *Request) When() *Request { return r }

// Get executes a GET request.
func (r *Request) Get(path string, pathParams ...any) *Response {
	return r.Execute(MethodGet, path, pathParams...)
}

// Post executes a POST request.
func (r *Request) Post(path string, pathParams ...any) *Response {
	return r.Execute(MethodPost, path, pathParams...)
}

// Put executes a PUT request.
func (r *Request) Put(path string, pathParams ...any) *Response {
	return r.Execute(MethodPut, path, pathParams...)
}

// Delete executes a DELETE request.
func (r *Request) Delete(path string, pathParams ...any) *Response {
	return r.Execute(MethodDelete, path, pathParams...)
}

// Head executes a HEAD request.
func (r *Request) Head(path string, pathParams ...any) *Response {
	return r.Execute(MethodHead, path, pathParams...)
}

// Patch executes a PATCH request.
func (r *Request) Patch(path string, pathParams ...any) *Response {
	return r.Execute(MethodPatch, path, pathParams...)
}

// Options executes an OPTIONS request.
func (r *Request) Options(path string, pathParams ...any) *Response {
	return r.Execute(MethodOptions, path, pathParams...)
}

// Execute builds the final URL, serializes the body, runs the filter chain in order,
// dispatches the HTTP request, and returns the *Response. It is the central entry point
// called by all method shortcuts (Get, Post, Put, etc.).
// Note: if Timeout() was used, the context is cancelled when Execute returns and the
// Request should not be reused for subsequent calls without calling Timeout() again.
func (r *Request) Execute(method Method, path string, positionalParams ...any) *Response {
	cancel := r.ctxCancel
	r.ctxCancel = nil
	if cancel != nil {
		defer cancel()
	}
	if r.buildErr != nil {
		return &Response{err: fmt.Errorf("%w: %v", ErrRequestFailed, r.buildErr)}
	}

	effectiveQuery, effectiveForm := r.effectiveParams(string(method))

	resolvedURL, err := r.buildURL(path, effectiveQuery, positionalParams...)
	if err != nil {
		return &Response{err: fmt.Errorf("%w: %v", ErrInvalidURL, err)}
	}

	bodyBytes, finalContentType, err := r.prepareBody(string(method), effectiveForm)
	if err != nil {
		return &Response{err: fmt.Errorf("%w: %v", ErrRequestFailed, err)}
	}

	headers := r.headers.Clone()
	if finalContentType != "" {
		headers.Set("Content-Type", finalContentType)
	}
	if r.accept != "" {
		headers.Set("Accept", r.accept)
	}

	filterableReq := &FilterableRequest{
		Ctx:         r.ctx,
		Method:      string(method),
		URI:         resolvedURL,
		Headers:     headers,
		Cookies:     append([]*http.Cookie{}, r.cookies...),
		QueryParams: cloneValues(r.queryParams),
		FormParams:  cloneValues(r.formParams),
		PathParams:  cloneStringMap(r.pathParams),
		Body:        bodyBytes,
		ContentType: finalContentType,
	}

	// OrderedFilter implementations define explicit priority; stable sort preserves
	// registration order among filters with equal Order() values.
	activeFilters := append([]Filter{}, r.filters...)

	// Auto-register FormAuthFilter when using form-based auth so that the DSL
	// Given().Auth().Form(...) triggers the login without requiring manual filter setup.
	// The filter is cached on the FormAuthScheme so that session state (cookie, session ID)
	// persists across multiple Execute() calls on the same Request.
	if fa, ok := r.auth.(*FormAuthScheme); ok {
		activeFilters = append([]Filter{fa.cachedFilter()}, activeFilters...)
	}
	if len(activeFilters) > 1 {
		slices.SortStableFunc(activeFilters, func(a, b Filter) int {
			oa, oka := a.(OrderedFilter)
			ob, okb := b.(OrderedFilter)
			orderA, orderB := 0, 0
			if oka {
				orderA = oa.Order()
			}
			if okb {
				orderB = ob.Order()
			}
			return cmp.Compare(orderA, orderB)
		})
	}

	filterCtx := &FilterContext{
		filters: activeFilters,
		index:   0,
		executor: func(fReq *FilterableRequest) (*Response, error) {
			return r.executeHTTP(fReq)
		},
	}

	resp, err := filterCtx.Next(filterableReq)
	if err != nil {
		return &Response{err: fmt.Errorf("%w: %v", ErrRequestFailed, err)}
	}
	return resp
}

// effectiveParams computes the final query and form parameter sets for the request
// without mutating the request's own fields (so Execute remains idempotent). Generic
// r.params are routed to query params for GET/DELETE/HEAD/OPTIONS and to form params
// for POST/PUT/PATCH. Empty values are dropped when EmptyParamsBehavior is "omit".
func (r *Request) effectiveParams(method string) (query url.Values, form url.Values) {
	paramBehavior := EmptyParamsInclude
	if r.config != nil {
		paramBehavior = r.config.ParamConfig().EmptyParamsBehavior
	}
	isFormMethod := method == "POST" || method == "PUT" || method == "PATCH"

	query = cloneValues(r.queryParams)
	form = cloneValues(r.formParams)

	for k, vs := range r.params {
		for _, v := range vs {
			if v == "" && paramBehavior == "omit" {
				continue
			}
			if isFormMethod {
				form.Add(k, v)
			} else {
				query.Add(k, v)
			}
		}
	}
	return query, form
}

// executeHTTP creates and dispatches the *http.Request, reads the full response body,
// decompresses it when indicated by Content-Encoding or magic bytes, and returns a
// populated *Response. It is the terminal executor at the end of the filter chain.
func (r *Request) executeHTTP(fReq *FilterableRequest) (*Response, error) {
	cfg := r.config
	if cfg == nil {
		cfg = DefaultConfig()
	}

	httpReq, err := http.NewRequestWithContext(fReq.Ctx, fReq.Method, fReq.URI, bytes.NewReader(fReq.Body))
	if err != nil {
		return nil, fmt.Errorf("creating http request: %w", err)
	}

	for k, v := range fReq.Headers {
		httpReq.Header[k] = v
	}
	for _, cookie := range fReq.Cookies {
		httpReq.AddCookie(cookie)
	}

	if r.auth != nil {
		if err := r.auth.Authenticate(httpReq); err != nil {
			return nil, fmt.Errorf("authenticating request: %w", err)
		}
	}

	client, err := r.buildHTTPClient()
	if err != nil {
		return nil, fmt.Errorf("building http client: %w", err)
	}
	start := time.Now()
	httpResp, err := client.Do(httpReq)
	elapsed := time.Since(start)

	if err != nil {
		return &Response{
			req:      httpReq,
			config:   cfg,
			elapsed:  elapsed,
			err:      err,
			rootPath: r.rootPath,
		}, nil
	}
	defer httpResp.Body.Close()

	bodyData, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	contentEncoding := strings.ToLower(httpResp.Header.Get("Content-Encoding"))
	bodyData, err = decompressResponse(bodyData, contentEncoding, cfg.DecoderConfig())
	if err != nil {
		return nil, err
	}

	resp := &Response{
		raw:         httpResp,
		req:         httpReq,
		config:      cfg,
		body:        bodyData,
		reqBody:     fReq.Body,
		statusCode:  httpResp.StatusCode,
		statusLine:  httpResp.Proto + " " + httpResp.Status,
		headers:     httpResp.Header,
		cookies:     httpResp.Cookies(),
		contentType: httpResp.Header.Get("Content-Type"),
		elapsed:     elapsed,
		rootPath:    r.rootPath,
	}

	return resp, nil
}

// buildHTTPClient returns the http.Client to use for this request.
// When HTTPClientConfig.ReuseClient=true and no per-request TLS override is active,
// the shared transport is initialized exactly once (via sync.Once on the Config's
// transportCache) and reused across concurrent calls. A CustomClient bypasses all
// transport construction and is returned as-is.
func (r *Request) buildHTTPClient() (*http.Client, error) {
	cfg := r.config
	if cfg == nil {
		cfg = DefaultConfig()
	}
	httpCfg := cfg.HTTPClientConfig()

	if httpCfg.CustomClient != nil {
		return httpCfg.CustomClient, nil
	}

	// Reuse the cached transport when ReuseClient=true, unless the request has a
	// per-request insecureSkipVerify override that would change the TLS config.
	if httpCfg.ReuseClient && !r.insecureSkipVerify && cfg.tCache != nil {
		cfg.tCache.once.Do(func() {
			cfg.tCache.transport, cfg.tCache.err = r.buildTransport()
		})
		if cfg.tCache.err != nil {
			return nil, cfg.tCache.err
		}
		return r.buildClientWithTransport(cfg.tCache.transport), nil
	}

	transport, err := r.buildTransport()
	if err != nil {
		return nil, err
	}
	return r.buildClientWithTransport(transport), nil
}

// buildTransport clones http.DefaultTransport and applies project-specific overrides:
// keep-alive settings, max idle connections, a custom TLS config, an optional client
// certificate pair (CertFile + KeyFile), and the insecureSkipVerify flag. All TLS
// fields are merged rather than replaced so that a custom TLSConfig and a cert file
// can coexist.
func (r *Request) buildTransport() (*http.Transport, error) {
	cfg := r.config
	if cfg == nil {
		cfg = DefaultConfig()
	}
	httpCfg := cfg.HTTPClientConfig()
	sslCfg := cfg.SSLConfig()

	var transport *http.Transport
	if t, ok := http.DefaultTransport.(*http.Transport); ok {
		transport = t.Clone()
	} else {
		transport = &http.Transport{}
	}

	if httpCfg.DisableKeepAlive {
		transport.DisableKeepAlives = true
	}
	if httpCfg.MaxIdleConns > 0 {
		transport.MaxIdleConns = httpCfg.MaxIdleConns
	}

	if httpCfg.TLSConfig != nil {
		transport.TLSClientConfig = httpCfg.TLSConfig.Clone()
	}

	if sslCfg.CertFile != "" && sslCfg.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(sslCfg.CertFile, sslCfg.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("loading TLS certificate from %q and %q: %w", sslCfg.CertFile, sslCfg.KeyFile, err)
		}
		if transport.TLSClientConfig == nil {
			transport.TLSClientConfig = &tls.Config{Certificates: []tls.Certificate{cert}}
		} else {
			transport.TLSClientConfig.Certificates = append(transport.TLSClientConfig.Certificates, cert)
		}
	}

	if r.insecureSkipVerify || httpCfg.InsecureSkipVerify || sslCfg.InsecureSkipVerify {
		if transport.TLSClientConfig == nil {
			transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec
		} else {
			transport.TLSClientConfig.InsecureSkipVerify = true //nolint:gosec
		}
	}

	return transport, nil
}

// buildClientWithTransport wraps transport in an *http.Client, applying the request
// timeout and redirect policy from the active Config. When redirects are disabled,
// http.ErrUseLastResponse is returned to halt following. When a MaxCount is set,
// the redirect hook enforces it.
func (r *Request) buildClientWithTransport(transport *http.Transport) *http.Client {
	cfg := r.config
	if cfg == nil {
		cfg = DefaultConfig()
	}
	httpCfg := cfg.HTTPClientConfig()
	redirectCfg := cfg.RedirectConfig()

	client := &http.Client{
		Transport: transport,
		Timeout:   httpCfg.Timeout,
	}

	if !httpCfg.FollowRedirects || !redirectCfg.Follow {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	} else if redirectCfg.MaxCount > 0 {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			if len(via) >= redirectCfg.MaxCount {
				return fmt.Errorf("stopped after %d redirects", redirectCfg.MaxCount)
			}
			return nil
		}
	}

	return client
}

// prepareBody selects the correct body encoding for the request:
//  1. multipart/form-data when MultiPart parts are present (effectiveForm fields are included as text parts).
//  2. The raw body bytes when Body or BodyBytes was called explicitly.
//  3. application/x-www-form-urlencoded from effectiveForm for POST, PUT, or PATCH.
//  4. An empty body for all other cases (GET, DELETE, HEAD, etc.).
func (r *Request) prepareBody(method string, effectiveForm url.Values) ([]byte, string, error) {
	if len(r.multiparts) > 0 {
		return buildMultipartBody(r.multiparts, effectiveForm)
	}

	if len(r.body) > 0 {
		return r.body, r.contentType, nil
	}

	isFormMethod := method == "POST" || method == "PUT" || method == "PATCH"
	if len(effectiveForm) > 0 && isFormMethod {
		if r.config != nil && r.config.ParamConfig().EmptyParamsBehavior == EmptyParamsError {
			for k, vs := range effectiveForm {
				for _, v := range vs {
					if v == "" {
						return nil, "", fmt.Errorf("empty form parameter %q is not allowed by ParamConfig", k)
					}
				}
			}
		}
		return []byte(effectiveForm.Encode()), string(ContentTypeURLEncoded), nil
	}

	return nil, r.contentType, nil
}

// buildURL assembles the full request URL from baseURI, port, basePath, the given path,
// path parameter substitutions, and query parameters. An absolute path (starting with
// "http://" or "https://") skips baseURI assembly and is used directly. The port is
// omitted when it is the default for the scheme (80/http, 443/https) or already encoded
// in the host.
func (r *Request) buildURL(path string, effectiveQuery url.Values, positionalParams ...any) (string, error) {
	// First substitute path parameters in the input path
	substitutedPath := r.substitutePathParams(path, positionalParams...)

	// If path is already an absolute URL, use it directly
	if strings.HasPrefix(substitutedPath, "http://") || strings.HasPrefix(substitutedPath, "https://") {
		return r.appendQueryParams(substitutedPath, effectiveQuery)
	}

	rawBaseURI := r.substitutePathParams(r.baseURI)
	if rawBaseURI == "" {
		rawBaseURI = "http://localhost"
	}
	if !strings.HasPrefix(rawBaseURI, "http://") && !strings.HasPrefix(rawBaseURI, "https://") {
		rawBaseURI = "http://" + rawBaseURI
	}

	u, err := url.Parse(rawBaseURI)
	if err != nil {
		return "", fmt.Errorf("parsing baseURI %q: %w", rawBaseURI, err)
	}

	// Apply port only if not already encoded in the host and it is not the default for the scheme.
	if u.Port() == "" && r.port > 0 {
		isDefault := (u.Scheme == "http" && r.port == 80) || (u.Scheme == "https" && r.port == 443)
		if !isDefault {
			u.Host = u.Hostname() + ":" + strconv.Itoa(r.port)
		}
	}

	// Combine basePath and path
	basePath := strings.Trim(r.substitutePathParams(r.basePath), "/")
	cleanPath := strings.TrimPrefix(substitutedPath, "/")

	var fullPathParts []string
	if u.Path != "" && u.Path != "/" {
		fullPathParts = append(fullPathParts, strings.Trim(u.Path, "/"))
	}
	if basePath != "" {
		fullPathParts = append(fullPathParts, basePath)
	}
	if cleanPath != "" {
		fullPathParts = append(fullPathParts, cleanPath)
	}

	if len(fullPathParts) > 0 {
		u.Path = "/" + strings.Join(fullPathParts, "/")
	} else {
		u.Path = "/"
	}

	return r.appendQueryParams(u.String(), effectiveQuery)
}

// substitutePathParams replaces path parameter placeholders in raw with URL-encoded values.
// Named parameters use {name} or /:name syntax; positional parameters use {0}, {1}, …
// Keys are processed longest-first to prevent a short key (e.g. "id") from matching inside
// a longer one (e.g. "id_type"). The /:name form matches only complete path segments.
// Positional params are replaced highest-index-first so {1} never clobbers {10}.
func (r *Request) substitutePathParams(raw string, positionalParams ...any) string {
	// Sort keys descending by length to prevent prefix collisions.
	keys := make([]string, 0, len(r.pathParams))
	for k := range r.pathParams {
		keys = append(keys, k)
	}
	slices.SortFunc(keys, func(a, b string) int {
		return cmp.Compare(len(b), len(a)) // descending
	})

	for _, k := range keys {
		encoded := url.PathEscape(r.pathParams[k])
		raw = strings.ReplaceAll(raw, "{"+k+"}", encoded)
		// Replace /:name only when it is a complete segment token
		// (followed by '/', '?', '#', or end of string).
		needle := "/:" + k
		start := 0
		for {
			idx := strings.Index(raw[start:], needle)
			if idx < 0 {
				break
			}
			abs := start + idx
			end := abs + len(needle)
			if end == len(raw) || raw[end] == '/' || raw[end] == '?' || raw[end] == '#' || raw[end] == '.' {
				raw = raw[:abs] + "/" + encoded + raw[end:]
				start = abs + 1 + len(encoded)
			} else {
				start = abs + 1
			}
		}
	}

	// Replace positional params highest-index-first so "{1}" never matches inside "{10}".
	for i := len(positionalParams) - 1; i >= 0; i-- {
		placeholder := "{" + strconv.Itoa(i) + "}"
		if strings.Contains(raw, placeholder) {
			raw = strings.ReplaceAll(raw, placeholder, url.PathEscape(fmt.Sprintf("%v", positionalParams[i])))
		}
	}
	return raw
}

// appendQueryParams merges queryParams into rawURL's query string, respecting
// URLEncodingEnabled and EmptyParamsBehavior from the active Config. When URL encoding
// is disabled, parameters are appended as raw "k=v" pairs without percent-encoding.
func (r *Request) appendQueryParams(rawURL string, queryParams url.Values) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	cfg := r.config
	if cfg == nil {
		cfg = DefaultConfig()
	}
	paramBehavior := cfg.ParamConfig().EmptyParamsBehavior

	if len(queryParams) > 0 {
		if r.urlEncodingEnabled {
			q := parsedURL.Query()
			for k, values := range queryParams {
				for _, v := range values {
					if v == "" {
						if paramBehavior == EmptyParamsOmit {
							continue
						}
						if paramBehavior == EmptyParamsError {
							return "", fmt.Errorf("empty parameter %q is not allowed by ParamConfig", k)
						}
					}
					q.Add(k, v)
				}
			}
			parsedURL.RawQuery = q.Encode()
		} else {
			// Build a raw query string without percent-encoding. Existing params in the
			// base URL are preserved so they are not overwritten by the new additions.
			var parts []string
			if parsedURL.RawQuery != "" {
				parts = append(parts, parsedURL.RawQuery)
			}
			for k, values := range queryParams {
				for _, v := range values {
					if v == "" {
						if paramBehavior == EmptyParamsOmit {
							continue
						}
						if paramBehavior == EmptyParamsError {
							return "", fmt.Errorf("empty parameter %q is not allowed by ParamConfig", k)
						}
					}
					parts = append(parts, k+"="+v)
				}
			}
			if len(parts) > 0 {
				parsedURL.RawQuery = strings.Join(parts, "&")
			}
		}
	}

	return parsedURL.String(), nil
}

// decompressResponse decompresses body according to the Content-Encoding header and
// the DecoderConfig heuristics. When the Content-Encoding header is explicit and
// decompression fails, an error is returned. When only magic-byte sniffing triggered
// the attempt, failures fall back silently to the original body.
func decompressResponse(data []byte, encoding string, cfg DecoderConfig) ([]byte, error) {
	isGzipEncoded := encoding == "gzip"
	isGzipMagic := cfg.AutoDecodeGzip && len(data) >= 2 && data[0] == 0x1f && data[1] == 0x8b
	isDeflateEncoded := encoding == "deflate"
	isDeflateMagic := cfg.AutoDecodeDeflate && len(data) >= 2 && data[0] == 0x78

	if isGzipEncoded || isGzipMagic {
		gr, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			if isGzipEncoded {
				return nil, fmt.Errorf("decompressing gzip response body: %w", err)
			}
			return data, nil
		}
		decompressed, err := io.ReadAll(gr)
		closeErr := gr.Close()
		if err != nil {
			if isGzipEncoded {
				return nil, fmt.Errorf("reading gzip response body: %w", err)
			}
			return data, nil
		}
		if closeErr != nil && isGzipEncoded {
			return nil, fmt.Errorf("gzip checksum validation failed: %w", closeErr)
		}
		return decompressed, nil
	}

	if isDeflateEncoded || isDeflateMagic {
		zr, zrErr := zlib.NewReader(bytes.NewReader(data))
		if zrErr == nil {
			decompressed, err := io.ReadAll(zr)
			_ = zr.Close()
			if err != nil {
				if isDeflateEncoded {
					return nil, fmt.Errorf("reading deflate(zlib) response body: %w", err)
				}
				return data, nil
			}
			return decompressed, nil
		}
		// zlib failed; try raw DEFLATE stream
		fr := flate.NewReader(bytes.NewReader(data))
		decompressed, err := io.ReadAll(fr)
		_ = fr.Close()
		if err != nil {
			if isDeflateEncoded {
				return nil, fmt.Errorf("reading deflate response body: %w", err)
			}
			return data, nil
		}
		return decompressed, nil
	}

	return data, nil
}
