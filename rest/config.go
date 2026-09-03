package rest

import (
	"crypto/tls"
	"io"
	"maps"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"
)

// HeaderConfig controls per-header overwrite-vs-merge behavior during request building.
// By default, Content-Type and Accept are overwritable; all other headers are merged.
type HeaderConfig struct {
	headers map[string]bool // lowercase name → true=overwrite, false=merge
}

// NewHeaderConfig returns a HeaderConfig with Content-Type and Accept marked as overwritable by default.
func NewHeaderConfig() *HeaderConfig {
	return &HeaderConfig{
		headers: map[string]bool{
			"content-type": true,
			"accept":       true,
		},
	}
}

// OverwriteHeadersWithName returns a new HeaderConfig that marks the given header names as overwritable.
func (hc *HeaderConfig) OverwriteHeadersWithName(names ...string) *HeaderConfig {
	cpy := hc.cloneHeaderConfig()
	for _, name := range names {
		cpy.headers[strings.ToLower(name)] = true
	}
	return cpy
}

// MergeHeadersWithName returns a new HeaderConfig that marks the given header names as mergeable (not overwritten).
func (hc *HeaderConfig) MergeHeadersWithName(names ...string) *HeaderConfig {
	cpy := hc.cloneHeaderConfig()
	for _, name := range names {
		cpy.headers[strings.ToLower(name)] = false
	}
	return cpy
}

// ShouldOverwriteHeaderWithName reports whether the given header name should be overwritten (case-insensitive).
func (hc *HeaderConfig) ShouldOverwriteHeaderWithName(name string) bool {
	if hc == nil {
		return true
	}
	return hc.headers[strings.ToLower(name)]
}

// cloneHeaderConfig returns a new HeaderConfig with a copied headers map so that
// mutations on the clone do not affect the original.
func (hc *HeaderConfig) cloneHeaderConfig() *HeaderConfig {
	cpy := &HeaderConfig{headers: make(map[string]bool, len(hc.headers))}
	for k, v := range hc.headers {
		cpy.headers[k] = v
	}
	return cpy
}

// SSLConfig holds SSL/TLS and certificate options.
type SSLConfig struct {
	InsecureSkipVerify bool
	CertFile           string
	KeyFile            string
	TrustStorePath     string
	StrictHostnames    bool
}

// HTTPClientConfig holds configuration for the underlying HTTP transport.
type HTTPClientConfig struct {
	Timeout            time.Duration
	FollowRedirects    bool
	DisableKeepAlive   bool
	MaxIdleConns       int
	ReuseClient        bool
	TLSConfig          *tls.Config
	InsecureSkipVerify bool
	CustomClient       *http.Client
	Params             map[string]any // arbitrary key-value transport parameters
}

// SetParam returns a new HTTPClientConfig with the given key set to value, preserving all other fields.
func (c HTTPClientConfig) SetParam(key string, value any) HTTPClientConfig {
	cpy := c
	merged := make(map[string]any, len(c.Params)+1)
	for k, v := range c.Params {
		merged[k] = v
	}
	merged[key] = value
	cpy.Params = merged
	return cpy
}

// AddParams returns a new HTTPClientConfig with all entries of params merged in, preserving existing fields.
func (c HTTPClientConfig) AddParams(params map[string]any) HTTPClientConfig {
	cpy := c
	merged := make(map[string]any, len(c.Params)+len(params))
	for k, v := range c.Params {
		merged[k] = v
	}
	for k, v := range params {
		merged[k] = v
	}
	cpy.Params = merged
	return cpy
}

// ReuseHTTPClient returns a new HTTPClientConfig with ReuseClient set to true.
func (c HTTPClientConfig) ReuseHTTPClient() HTTPClientConfig {
	cpy := c
	cpy.ReuseClient = true
	return cpy
}

// LogConfig holds configuration for request and response logging.
type LogConfig struct {
	Output                         io.Writer
	Detail                         LogDetail
	EnablePrettyPrinting           bool
	EnableLoggingIfValidationFails bool
	EnableLoggingIfError           bool
	BlacklistHeaders               []string
}

// EncoderConfig holds configuration for payload and charset encoding.
type EncoderConfig struct {
	DefaultCharset      string
	ContentTypeCharsets map[string]string
}

// DecoderConfig holds configuration for response body decoding and decompression.
type DecoderConfig struct {
	DefaultCharset    string
	ContentDecoders   []string
	AutoDecodeGzip    bool
	AutoDecodeDeflate bool
}

// RedirectConfig configures HTTP redirection policies.
type RedirectConfig struct {
	Follow   bool
	MaxCount int
}

// FailureConfig configures callbacks invoked when response assertions fail.
type FailureConfig struct {
	Listeners []func(req *http.Request, resp *Response, failures []string)
}

// ParamConfig configures query and form parameter behaviors.
type ParamConfig struct {
	EmptyParamsBehavior string // "include" (default) | "omit" | "error"
}

// transportCache holds a lazily initialized *http.Transport shared across request instances
// when HTTPClientConfig.ReuseClient is true. sync.Once guarantees the transport is built
// exactly once even under concurrent requests.
type transportCache struct {
	once      sync.Once
	transport *http.Transport
	err       error
}

// CsrfConfig configures anti-CSRF token parameters.
type CsrfConfig struct {
	URI        string
	TokenField string
	HeaderName string
}

// SessionConfig configures automatic session identifier management.
type SessionConfig struct {
	SessionName string
	AutoTrack   bool
}

// Config is the top-level immutable configuration for REST Assured.
type Config struct {
	sslConfig        SSLConfig
	httpClientConfig HTTPClientConfig
	logConfig        LogConfig
	encoderConfig    EncoderConfig
	decoderConfig    DecoderConfig
	redirectConfig   RedirectConfig
	failureConfig    FailureConfig
	paramConfig      ParamConfig
	csrfConfig       CsrfConfig
	sessionConfig    SessionConfig
	headerConfig     *HeaderConfig
	tCache           *transportCache // shared across clones when ReuseClient=true
}

// DefaultConfig returns a new Config with standard default settings.
func DefaultConfig() *Config {
	return &Config{
		headerConfig: NewHeaderConfig(),
		tCache:       &transportCache{},
		sslConfig: SSLConfig{
			InsecureSkipVerify: false,
			StrictHostnames:    false,
		},
		httpClientConfig: HTTPClientConfig{
			Timeout:         30 * time.Second,
			FollowRedirects: true,
			ReuseClient:     true,
		},
		logConfig: LogConfig{
			EnablePrettyPrinting: true,
			BlacklistHeaders:     []string{"Authorization", "Cookie"},
		},
		encoderConfig: EncoderConfig{
			DefaultCharset:      "UTF-8",
			ContentTypeCharsets: make(map[string]string),
		},
		decoderConfig: DecoderConfig{
			DefaultCharset:    "UTF-8",
			ContentDecoders:   []string{"gzip", "deflate"},
			AutoDecodeGzip:    true,
			AutoDecodeDeflate: true,
		},
		redirectConfig: RedirectConfig{
			Follow:   true,
			MaxCount: 10,
		},
		failureConfig: FailureConfig{
			Listeners: make([]func(req *http.Request, resp *Response, failures []string), 0),
		},
		paramConfig: ParamConfig{
			EmptyParamsBehavior: "include",
		},
		csrfConfig: CsrfConfig{
			TokenField: "_csrf",
		},
		sessionConfig: SessionConfig{
			SessionName: "JSESSIONID",
			AutoTrack:   false,
		},
	}
}

// Clone creates a deep copy of Config.
func (c *Config) Clone() *Config {
	if c == nil {
		return DefaultConfig()
	}
	cpy := *c

	// Deep copy maps and slices
	if c.encoderConfig.ContentTypeCharsets != nil {
		cpy.encoderConfig.ContentTypeCharsets = maps.Clone(c.encoderConfig.ContentTypeCharsets)
	} else {
		cpy.encoderConfig.ContentTypeCharsets = make(map[string]string)
	}

	if c.logConfig.BlacklistHeaders != nil {
		cpy.logConfig.BlacklistHeaders = slices.Clone(c.logConfig.BlacklistHeaders)
	} else {
		cpy.logConfig.BlacklistHeaders = make([]string, 0)
	}

	if c.decoderConfig.ContentDecoders != nil {
		cpy.decoderConfig.ContentDecoders = slices.Clone(c.decoderConfig.ContentDecoders)
	} else {
		cpy.decoderConfig.ContentDecoders = make([]string, 0)
	}

	if c.failureConfig.Listeners != nil {
		cpy.failureConfig.Listeners = slices.Clone(c.failureConfig.Listeners)
	} else {
		cpy.failureConfig.Listeners = make([]func(req *http.Request, resp *Response, failures []string), 0)
	}

	if c.headerConfig != nil {
		cpy.headerConfig = c.headerConfig.cloneHeaderConfig()
	} else {
		cpy.headerConfig = NewHeaderConfig()
	}

	// Share transport cache across clones when ReuseClient=true so TCP connections are pooled.
	// When ReuseClient=false, clear the cache so each config builds its own transport.
	if !c.httpClientConfig.ReuseClient {
		cpy.tCache = nil
	}

	return &cpy
}

// WithSSL returns a copy of Config with updated SSLConfig.
// The transport cache is always reset so that the new TLS parameters are applied
// on the next request — sharing the old cached transport would ignore the new certs.
func (c *Config) WithSSL(cfg SSLConfig) *Config {
	cpy := c.Clone()
	cpy.sslConfig = cfg
	cpy.tCache = &transportCache{}
	return cpy
}

// SSLConfig returns the current SSLConfig.
func (c *Config) SSLConfig() SSLConfig {
	if c == nil {
		return DefaultConfig().sslConfig
	}
	return c.sslConfig
}

// WithHTTPClient returns a copy of Config with updated HTTPClientConfig.
func (c *Config) WithHTTPClient(cfg HTTPClientConfig) *Config {
	cpy := c.Clone()
	cpy.httpClientConfig = cfg
	if cfg.ReuseClient {
		if cpy.tCache == nil {
			cpy.tCache = &transportCache{}
		}
	} else {
		cpy.tCache = nil
	}
	return cpy
}

// HTTPClient is an alias for WithHTTPClient for backward compatibility.
func (c *Config) HTTPClient(cfg HTTPClientConfig) *Config {
	return c.WithHTTPClient(cfg)
}

// HTTPClientConfig returns the current HTTPClientConfig.
func (c *Config) HTTPClientConfig() HTTPClientConfig {
	if c == nil {
		return DefaultConfig().httpClientConfig
	}
	return c.httpClientConfig
}

// WithLog returns a copy of Config with updated LogConfig.
func (c *Config) WithLog(cfg LogConfig) *Config {
	cpy := c.Clone()
	cpy.logConfig = cfg
	if cfg.BlacklistHeaders != nil {
		cpy.logConfig.BlacklistHeaders = slices.Clone(cfg.BlacklistHeaders)
	} else {
		cpy.logConfig.BlacklistHeaders = make([]string, 0)
	}
	return cpy
}

// Log is an alias for WithLog for backward compatibility.
func (c *Config) Log(cfg LogConfig) *Config {
	return c.WithLog(cfg)
}

// LogConfig returns the current LogConfig.
func (c *Config) LogConfig() LogConfig {
	if c == nil {
		return DefaultConfig().logConfig
	}
	return c.logConfig
}

// WithEncoder returns a copy of Config with updated EncoderConfig.
func (c *Config) WithEncoder(cfg EncoderConfig) *Config {
	cpy := c.Clone()
	cpy.encoderConfig = cfg
	if cfg.ContentTypeCharsets != nil {
		cpy.encoderConfig.ContentTypeCharsets = maps.Clone(cfg.ContentTypeCharsets)
	} else {
		cpy.encoderConfig.ContentTypeCharsets = make(map[string]string)
	}
	return cpy
}

// Encoder is an alias for WithEncoder for backward compatibility.
func (c *Config) Encoder(cfg EncoderConfig) *Config {
	return c.WithEncoder(cfg)
}

// EncoderConfig returns the current EncoderConfig.
func (c *Config) EncoderConfig() EncoderConfig {
	if c == nil {
		return DefaultConfig().encoderConfig
	}
	return c.encoderConfig
}

// WithDecoder returns a copy of Config with updated DecoderConfig.
func (c *Config) WithDecoder(cfg DecoderConfig) *Config {
	cpy := c.Clone()
	cpy.decoderConfig = cfg
	if cfg.ContentDecoders != nil {
		cpy.decoderConfig.ContentDecoders = slices.Clone(cfg.ContentDecoders)
	} else {
		cpy.decoderConfig.ContentDecoders = make([]string, 0)
	}
	return cpy
}

// Decoder is an alias for WithDecoder for backward compatibility.
func (c *Config) Decoder(cfg DecoderConfig) *Config {
	return c.WithDecoder(cfg)
}

// DecoderConfig returns the current DecoderConfig.
func (c *Config) DecoderConfig() DecoderConfig {
	if c == nil {
		return DefaultConfig().decoderConfig
	}
	return c.decoderConfig
}

// WithRedirect returns a copy of Config with updated RedirectConfig.
func (c *Config) WithRedirect(cfg RedirectConfig) *Config {
	cpy := c.Clone()
	cpy.redirectConfig = cfg
	return cpy
}

// RedirectConfig returns the current RedirectConfig.
func (c *Config) RedirectConfig() RedirectConfig {
	if c == nil {
		return DefaultConfig().redirectConfig
	}
	return c.redirectConfig
}

// WithFailure returns a copy of Config with updated FailureConfig.
func (c *Config) WithFailure(cfg FailureConfig) *Config {
	cpy := c.Clone()
	cpy.failureConfig = cfg
	if cfg.Listeners != nil {
		cpy.failureConfig.Listeners = slices.Clone(cfg.Listeners)
	} else {
		cpy.failureConfig.Listeners = make([]func(req *http.Request, resp *Response, failures []string), 0)
	}
	return cpy
}

// FailureConfig returns the current FailureConfig.
func (c *Config) FailureConfig() FailureConfig {
	if c == nil {
		return DefaultConfig().failureConfig
	}
	return c.failureConfig
}

// WithParam returns a copy of Config with updated ParamConfig.
func (c *Config) WithParam(cfg ParamConfig) *Config {
	cpy := c.Clone()
	cpy.paramConfig = cfg
	return cpy
}

// ParamConfig returns the current ParamConfig.
func (c *Config) ParamConfig() ParamConfig {
	if c == nil {
		return DefaultConfig().paramConfig
	}
	return c.paramConfig
}

// WithCsrf returns a copy of Config with updated CsrfConfig.
func (c *Config) WithCsrf(cfg CsrfConfig) *Config {
	cpy := c.Clone()
	cpy.csrfConfig = cfg
	return cpy
}

// CsrfConfig returns the current CsrfConfig.
func (c *Config) CsrfConfig() CsrfConfig {
	if c == nil {
		return DefaultConfig().csrfConfig
	}
	return c.csrfConfig
}

// WithSession returns a copy of Config with updated SessionConfig.
func (c *Config) WithSession(cfg SessionConfig) *Config {
	cpy := c.Clone()
	cpy.sessionConfig = cfg
	return cpy
}

// SessionConfig returns the current SessionConfig.
func (c *Config) SessionConfig() SessionConfig {
	if c == nil {
		return DefaultConfig().sessionConfig
	}
	return c.sessionConfig
}

// WithHeaderConfig returns a copy of Config with the given HeaderConfig.
func (c *Config) WithHeaderConfig(hc *HeaderConfig) *Config {
	cpy := c.Clone()
	if hc != nil {
		cpy.headerConfig = hc.cloneHeaderConfig()
	} else {
		cpy.headerConfig = NewHeaderConfig()
	}
	return cpy
}

// HeaderConfig returns the current HeaderConfig.
func (c *Config) HeaderConfig() *HeaderConfig {
	if c == nil || c.headerConfig == nil {
		return NewHeaderConfig()
	}
	return c.headerConfig
}
