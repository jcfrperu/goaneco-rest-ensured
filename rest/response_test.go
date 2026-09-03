package rest_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

type TestUser struct {
	Name  string `json:"name"`
	Score int    `json:"score"`
}

func TestResponseAccessors(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("X-Custom-Header", "Value123")
		http.SetCookie(w, &http.Cookie{Name: "auth_token", Value: "token-abc-999"})
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"name":"Diana","score":100}`))
	}))
	t.Cleanup(ts.Close)

	resp := rest.Given().BaseURI(ts.URL).Get("/user")
	must.NoError(resp.Err())

	// Status assertions
	is.Equal(http.StatusCreated, resp.StatusCode())
	is.Contains(resp.StatusLine(), "201")

	// Headers & Cookies
	is.Equal("Value123", resp.Header("X-Custom-Header"))
	is.Equal("token-abc-999", resp.Cookie("auth_token"))
	is.Contains(resp.ContentType(), "application/json")
	is.NotEmpty(resp.Headers())
	is.NotEmpty(resp.Cookies())
	is.Greater(resp.Time(), time.Duration(0))

	// Body extractions
	is.Equal(`{"name":"Diana","score":100}`, resp.AsString())
	is.Equal([]byte(`{"name":"Diana","score":100}`), resp.AsBytes())

	// Config and Raw methods
	must.NotNil(resp.Config())
	must.NotNil(resp.RawResponse())
	must.NotNil(resp.RawRequest())

	// JSON unmarshaling via As()
	var u TestUser
	err := resp.As(&u)
	must.NoError(err)
	is.Equal("Diana", u.Name)
	is.Equal(100, u.Score)

	// JSON unmarshaling via AsObject generic helper
	uGeneric, err := rest.AsObject[TestUser](resp)
	must.NoError(err)
	is.Equal("Diana", uGeneric.Name)
	is.Equal(100, uGeneric.Score)

	// Path query via gjson
	is.Equal("Diana", resp.Path("name").String())
	is.Equal(int64(100), resp.Path("score").Int())

	// JsonPath integration
	jp := resp.JsonPath()
	must.NotNil(jp)
	is.Equal("Diana", jp.GetString("name"))
	is.Equal(100, jp.GetInt("score"))

	// JsonPath from ExtractableResponse
	extJP := resp.Extract().JsonPath()
	must.NotNil(extJP)
	is.Equal("Diana", extJP.GetString("name"))
}

func TestResponse_XmlPath(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<greeting><message>Hello XML</message><count>5</count></greeting>`))
	}))
	t.Cleanup(ts.Close)

	resp := rest.Given().BaseURI(ts.URL).Get("/xml")
	must.NoError(resp.Err())

	xp, err := resp.XmlPath()
	must.NoError(err)
	must.NotNil(xp)
	is.Equal("Hello XML", xp.GetString("//greeting/message"))
	is.Equal(5, xp.GetInt("//greeting/count"))

	// XmlPath from ExtractableResponse
	extXP, err := resp.Extract().XmlPath()
	must.NoError(err)
	must.NotNil(extXP)
	is.Equal("Hello XML", extXP.GetString("//greeting/message"))
}

func TestResponse_ErrorsAndEdgeCases(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	// 1. Empty body — use an in-process server that returns 204 No Content
	emptyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(emptyServer.Close)

	emptyResp := rest.Given().
		BaseURI(emptyServer.URL).
		Get("/")

	var u TestUser
	is.Error(emptyResp.As(&u))
	_, err := rest.AsObject[TestUser](emptyResp)
	is.Error(err)

	// 2. Malformed JSON body
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{not valid json`))
	}))
	t.Cleanup(ts.Close)

	badResp := rest.Given().BaseURI(ts.URL).Get("/bad")
	is.Error(badResp.As(&u))

	// 3. Malformed XML body
	badXMLResp := rest.Given().BaseURI(ts.URL).Get("/bad")
	_, err = badXMLResp.XmlPath()
	is.Error(err)
}

func TestNilResponseSafety(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	var nilResp *rest.Response

	is.Equal("", nilResp.AsString())
	is.Empty(nilResp.AsBytes())
	is.Equal(0, nilResp.StatusCode())
	is.Equal("", nilResp.StatusLine())
	is.Equal("", nilResp.Header("X-Any"))
	is.Empty(nilResp.Headers())
	is.Equal("", nilResp.Cookie("session"))
	is.Empty(nilResp.Cookies())
	is.Equal("", nilResp.ContentType())
	is.Equal(time.Duration(0), nilResp.Time())
	is.ErrorIs(nilResp.Err(), rest.ErrNilResponse)
	is.False(nilResp.Path("foo").Exists())
	is.Nil(nilResp.RawResponse())
	is.Nil(nilResp.RawRequest())
	is.NotNil(nilResp.Config())

	var target TestUser
	is.ErrorIs(nilResp.As(&target), rest.ErrNilResponse)
	_, err := rest.AsObject[TestUser](nilResp)
	is.ErrorIs(err, rest.ErrNilResponse)

	_, err = nilResp.XmlPath()
	is.ErrorIs(err, rest.ErrNilResponse)

	ext := nilResp.Extract()
	is.Equal("", ext.AsString())
	is.Equal(0, ext.StatusCode())
	is.Equal("", ext.Header("X-Any"))
}

func TestResponse_ContentTypeExtraction(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	// application/json without charset → returned as-is
	ts1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(ts1.Close)

	resp1 := rest.Given().BaseURI(ts1.URL).Get("/")
	must.NoError(resp1.Err())
	is.Equal("application/json", resp1.ContentType())

	// application/json; charset=utf-8 → full value including charset
	ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(ts2.Close)

	resp2 := rest.Given().BaseURI(ts2.URL).Get("/")
	must.NoError(resp2.Err())
	is.Contains(resp2.ContentType(), "application/json")
	is.Contains(resp2.ContentType(), "charset=utf-8")

	// ValidatableResponse.ContentTypeContains() matches despite charset suffix
	valid := resp2.Then().ContentTypeContains("application/json")
	is.False(valid.HasFailures())
}

// ── ContentTypeTest ──────────────────────────────────────────────────────────

func TestContentType_WithCharsetAndMatches(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	// WithCharset appends "; charset=<name>"
	is.Equal("application/json; charset=UTF-8", rest.ContentTypeJSON.WithCharset("UTF-8"))
	is.Equal("application/json; charset=ISO-8859-1", rest.ContentTypeJSON.WithCharset("ISO-8859-1"))

	// Matches is case-insensitive
	is.True(rest.ContentTypeJSON.Matches("appliCatIon/JSON"))
	is.True(rest.ContentTypeJSON.Matches("application/json"))

	// Matches returns false for non-equal content type
	is.False(rest.ContentTypeJSON.Matches("application/json2"))

	// Matches returns false for empty string
	is.False(rest.ContentTypeJSON.Matches(""))

	// FromContentType finds any known ContentType by its MIME string
	ct, ok := rest.FromContentType("application/json")
	is.True(ok)
	is.Equal(rest.ContentTypeJSON, ct)

	ct2, ok2 := rest.FromContentType("APPLICATION/XML")
	is.True(ok2)
	is.Equal(rest.ContentTypeXML, ct2)

	// FromContentType returns ("", false) for unknown MIME type
	unknown, ok3 := rest.FromContentType("application/unknown-type")
	is.False(ok3)
	is.Equal(rest.ContentType(""), unknown)
}

// TestContentType_FromContentType_AllTypes mirrors the parameterized
// should_find_content_type_from_string Java test: every known ContentType constant
// must be discoverable by its MIME string, case-insensitively.
func TestContentType_FromContentType_AllTypes(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	cases := []struct {
		mime     string
		expected rest.ContentType
	}{
		{"application/json", rest.ContentTypeJSON},
		{"APPLICATION/JSON", rest.ContentTypeJSON},
		{"application/xml", rest.ContentTypeXML},
		{"APPLICATION/XML", rest.ContentTypeXML},
		{"text/html", rest.ContentTypeHTML},
		{"TEXT/HTML", rest.ContentTypeHTML},
		{"text/plain", rest.ContentTypeText},
		{"application/x-www-form-urlencoded", rest.ContentTypeURLEncoded},
		{"multipart/form-data", rest.ContentTypeMultipart},
		{"application/octet-stream", rest.ContentTypeBinary},
		{"*/*", rest.ContentTypeAny},
	}

	for _, tc := range cases {
		ct, ok := rest.FromContentType(tc.mime)
		is.True(ok, "expected to find ContentType for %q", tc.mime)
		is.Equal(tc.expected, ct, "wrong ContentType for %q", tc.mime)
	}

	_, ok := rest.FromContentType("application/unknown-type-xyz")
	is.False(ok)
}

// TestContentType_NullAndCharsetVariants mirrors the remaining Java ContentTypeTest cases.
func TestContentType_NullAndCharsetVariants(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	// content_type_doesnt_match_when_expected_content_type_is_null:
	// Java passes null; Go equivalent is an empty string. Matches() returns false.
	is.False(rest.ContentTypeJSON.Matches(""))

	// content_type_with_java_charset_returns_the_content_type_with_the_given_charset:
	// Java uses Charset.forName("ISO-8859-1"). Go's WithCharset accepts the charset name string directly.
	is.Equal("application/json; charset=ISO-8859-1", rest.ContentTypeJSON.WithCharset("ISO-8859-1"))
	is.Equal("application/json; charset=UTF-8", rest.ContentTypeJSON.WithCharset("UTF-8"))
}

func TestCookie_NegativeMaxAge(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	// Mirrors Java CookieTest.can_use_negative_values_as_max_age.
	// net/http.Cookie.MaxAge accepts any int; negative signals "delete immediately".
	c := &http.Cookie{Name: "hello", Value: "world", MaxAge: -3600}
	is.Equal(-3600, c.MaxAge)
}

func TestHeader_SameNameCaseInsensitive(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	// Mirrors Java HeaderTest: hasSameNameAs uses case-insensitive comparison.
	// In Go, http.CanonicalHeaderKey normalises names to Title-Case.
	is.Equal(http.CanonicalHeaderKey("foo"), http.CanonicalHeaderKey("Foo"))
	is.Equal(http.CanonicalHeaderKey("foo"), http.CanonicalHeaderKey("FOO"))
	is.NotEqual(http.CanonicalHeaderKey("foo"), http.CanonicalHeaderKey("bar"))

	// Accessing a header via a different case still works
	h := http.Header{}
	h.Set("foo", "bar")
	h.Set("Foo", "baz") // overwrites — canonical form is the same key
	is.Equal("baz", h.Get("FOO"))
}
