package rest

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

// AuthScheme represents a strategy for authenticating HTTP requests.
type AuthScheme interface {
	Authenticate(req *http.Request) error
}

// NoAuthScheme represents the default state of no authentication.
type NoAuthScheme struct{}

// Authenticate performs no modification to the request.
func (n *NoAuthScheme) Authenticate(_ *http.Request) error {
	return nil
}

// BasicAuthScheme adds Basic Authentication headers to the request.
type BasicAuthScheme struct {
	Username string
	Password string
}

// Authenticate applies standard HTTP Basic authentication.
func (b *BasicAuthScheme) Authenticate(req *http.Request) error {
	if b == nil {
		return nil
	}
	req.SetBasicAuth(b.Username, b.Password)
	return nil
}

// PreemptiveBasicAuthScheme adds Basic Authentication headers unconditionally,
// without waiting for a 401 challenge. In Go's net/http client there is no
// built-in challenge-response mechanism, so this is functionally identical to
// BasicAuthScheme. Both types are kept for API compatibility and documentation
// clarity; PreemptiveBasicAuthScheme signals intent at the call site.
type PreemptiveBasicAuthScheme struct {
	Username string
	Password string
}

// Authenticate sets the Authorization header using standard HTTP Basic encoding.
func (p *PreemptiveBasicAuthScheme) Authenticate(req *http.Request) error {
	if p == nil {
		return nil
	}
	req.SetBasicAuth(p.Username, p.Password)
	return nil
}

// OAuth2Scheme adds Bearer token authorization header.
type OAuth2Scheme struct {
	AccessToken string
}

// Authenticate sets the Bearer token in the Authorization header.
func (o *OAuth2Scheme) Authenticate(req *http.Request) error {
	if o == nil || o.AccessToken == "" {
		return nil
	}
	req.Header.Set("Authorization", "Bearer "+o.AccessToken)
	return nil
}

// FormAuthConfig configures parameters for form-based authentication.
type FormAuthConfig struct {
	FormAction    string
	UsernameField string
	PasswordField string
	CsrfField     string
	LogRequest    bool
	LoginTimeout  time.Duration
}

// DefaultFormAuthConfig returns a FormAuthConfig with standard default field names.
func DefaultFormAuthConfig() *FormAuthConfig {
	return &FormAuthConfig{
		FormAction:    "/login",
		UsernameField: "username",
		PasswordField: "password",
		CsrfField:     "",
		LogRequest:    false,
		LoginTimeout:  10 * time.Second,
	}
}

// SpringSecurityFormAuth returns a FormAuthConfig configured for Spring Security defaults.
func SpringSecurityFormAuth() *FormAuthConfig {
	return &FormAuthConfig{
		FormAction:    "/login",
		UsernameField: "username",
		PasswordField: "password",
		CsrfField:     "_csrf",
		LogRequest:    false,
		LoginTimeout:  10 * time.Second,
	}
}

// FormAuthScheme specifies form-based authentication parameters.
type FormAuthScheme struct {
	Username string
	Password string
	Config   *FormAuthConfig
	mu       sync.Mutex
	filter   *FormAuthFilter // lazily created and reused so session state persists across Execute() calls
}

// Authenticate performs no direct modification on standard http.Request as form auth is orchestrated via FormAuthFilter.
func (f *FormAuthScheme) Authenticate(_ *http.Request) error {
	return nil
}

// cachedFilter returns the shared FormAuthFilter, creating it on first call.
func (f *FormAuthScheme) cachedFilter() *FormAuthFilter {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.filter == nil {
		f.filter = NewFormAuthFilter(f.Username, f.Password, f.Config)
	}
	return f.filter
}

// Form is a package-level helper that creates a FormAuthScheme with validated credentials.
// Panics if username is blank (empty or whitespace-only) or password is empty.
func Form(username, password string, cfg *FormAuthConfig) *FormAuthScheme {
	if strings.TrimSpace(username) == "" {
		panic("form auth: username must not be blank")
	}
	if password == "" {
		panic("form auth: password must not be empty")
	}
	if cfg == nil {
		cfg = DefaultFormAuthConfig()
	}
	return &FormAuthScheme{Username: username, Password: password, Config: cfg}
}

// AuthSpec provides a fluent interface for configuring authentication on a Request.
type AuthSpec struct {
	req *Request
}

// Basic sets standard Basic Authentication.
func (a *AuthSpec) Basic(username, password string) *Request {
	a.req.auth = &BasicAuthScheme{Username: username, Password: password}
	return a.req
}

// PreemptiveBasic sets preemptive Basic Authentication.
func (a *AuthSpec) PreemptiveBasic(username, password string) *Request {
	a.req.auth = &PreemptiveBasicAuthScheme{Username: username, Password: password}
	return a.req
}

// OAuth2 sets OAuth 2.0 Bearer token authentication.
func (a *AuthSpec) OAuth2(token string) *Request {
	a.req.auth = &OAuth2Scheme{AccessToken: token}
	return a.req
}

// Form sets Form-based authentication credentials and optional configuration.
func (a *AuthSpec) Form(username, password string, cfg ...*FormAuthConfig) *Request {
	var formCfg *FormAuthConfig
	if len(cfg) > 0 && cfg[0] != nil {
		formCfg = cfg[0]
	} else {
		formCfg = DefaultFormAuthConfig()
	}
	a.req.auth = &FormAuthScheme{
		Username: username,
		Password: password,
		Config:   formCfg,
	}
	return a.req
}

// None clears any configured authentication.
func (a *AuthSpec) None() *Request {
	a.req.auth = &NoAuthScheme{}
	return a.req
}
