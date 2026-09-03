package rest_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func TestRequestURLResolution(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Path", r.URL.Path)
		w.Header().Set("X-Query", r.URL.RawQuery)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(ts.Close)

	tests := []struct {
		name           string
		setup          func(req *rest.Request) *rest.Request
		path           string
		positionalArgs []any
		expectedPath   string
		expectedQuery  string
	}{
		{
			name: "simple path with baseURI",
			setup: func(req *rest.Request) *rest.Request {
				return req.BaseURI(ts.URL)
			},
			path:         "/users",
			expectedPath: "/users",
		},
		{
			name: "path with named path parameter curly",
			setup: func(req *rest.Request) *rest.Request {
				return req.BaseURI(ts.URL).PathParam("id", 42)
			},
			path:         "/users/{id}",
			expectedPath: "/users/42",
		},
		{
			name: "path with colon path parameter",
			setup: func(req *rest.Request) *rest.Request {
				return req.BaseURI(ts.URL).PathParam("category", "books")
			},
			path:         "/store/:category",
			expectedPath: "/store/books",
		},
		{
			name: "path with positional parameter",
			setup: func(req *rest.Request) *rest.Request {
				return req.BaseURI(ts.URL)
			},
			path:           "/items/{0}",
			positionalArgs: []any{99},
			expectedPath:   "/items/99",
		},
		{
			name: "query parameters encoded",
			setup: func(req *rest.Request) *rest.Request {
				return req.BaseURI(ts.URL).
					QueryParam("filter", "active").
					QueryParam("page", 1)
			},
			path:          "/search",
			expectedPath:  "/search",
			expectedQuery: "filter=active&page=1",
		},
		{
			name: "baseURI with subpath and port",
			setup: func(req *rest.Request) *rest.Request {
				return req.BaseURI(ts.URL + "/api/v1")
			},
			path:         "/orders",
			expectedPath: "/api/v1/orders",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)
			must := require.New(t)

			req := rest.Given()
			req = tt.setup(req)

			resp := req.Get(tt.path, tt.positionalArgs...)
			must.NoError(resp.Err())
			is.Equal(tt.expectedPath, resp.Header("X-Path"))
			if tt.expectedQuery != "" {
				is.Equal(tt.expectedQuery, resp.Header("X-Query"))
			}
		})
	}
}

func TestRequestContextCancellation(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(ts.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	resp := rest.Given().
		Context(ctx).
		BaseURI(ts.URL).
		Get("/slow")

	must.NotNil(resp)
	is.NotNil(resp.Err())
}

func TestRequestBodyObjectMarshaling(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	type UserPayload struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	var receivedBody string
	var receivedCT string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedCT = r.Header.Get("Content-Type")
		buf := make([]byte, 1024)
		n, _ := r.Body.Read(buf)
		receivedBody = string(buf[:n])
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(ts.Close)

	payload := UserPayload{
		Name:  "Charlie",
		Email: "charlie@example.com",
	}

	resp := rest.Given().
		BaseURI(ts.URL).
		BodyObject(payload).
		Post("/users")

	must.NoError(resp.Err())
	is.Equal(string(rest.ContentTypeJSON), receivedCT)
	is.JSONEq(`{"name":"Charlie","email":"charlie@example.com"}`, receivedBody)
}

func TestRequestParamsAndHeaders(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		is.Equal("application/xml", r.Header.Get("Accept"))
		is.Equal("Val1", r.Header.Get("X-Header-1"))
		cookie, err := r.Cookie("user_session")
		if !is.NoError(err) {
			http.Error(w, "cookie missing", http.StatusBadRequest)
			return
		}
		is.Equal("sess_val", cookie.Value)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`ok`))
	}))
	t.Cleanup(ts.Close)

	resp := rest.Given().
		BaseURI(ts.URL).
		BasePath("/docs").
		URLEncodingEnabled(true).
		Accept(rest.ContentTypeXML).
		ContentType(rest.ContentTypeJSON).
		Header("X-Header-1", "Val1").
		Cookie("user_session", "sess_val").
		Param("general_param", "p_val").
		FormParams(map[string]any{"form_k": "form_v"}).
		QueryParams(map[string]any{"q": "golang"}).
		PathParams(map[string]any{"section": "overview"}).
		Body("raw text body").
		BodyBytes([]byte("raw text body")).
		Config(rest.DefaultConfig()).
		RelaxedHTTPSValidation().
		Given().
		When().
		Get("/{section}")

	must.NoError(resp.Err())
	is.Equal(http.StatusOK, resp.StatusCode())
}

func TestRequestMultipartUpload(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(10 << 20); !is.NoError(err) {
			http.Error(w, "multipart error", http.StatusBadRequest)
			return
		}
		is.Equal("doc_title", r.MultipartForm.Value["title"][0])

		files := r.MultipartForm.File["file_upload"]
		if !is.Len(files, 1) {
			return
		}
		is.Equal("sample.txt", files[0].Filename)

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"uploaded"}`))
	}))
	t.Cleanup(ts.Close)

	resp := rest.Given().
		BaseURI(ts.URL).
		FormParam("title", "doc_title").
		MultiPartNamed("file_upload", "sample.txt", []byte("Hello World File Content"), "text/plain").
		Post("/upload")

	must.NoError(resp.Err())
	is.Equal(http.StatusOK, resp.StatusCode())
}

func TestRequest_URIValidation(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	// Mirrors UriValidatorTest.returns_false_when_uri_is_empty
	is.Error(rest.Given().BaseURI("").Get("/path").Err())

	// Mirrors returns_false_when_uri_is_blank ("   " has no scheme or host)
	is.Error(rest.Given().BaseURI("   ").Get("/path").Err())

	// Mirrors returns_false_when_uri_doesnt_contain_scheme
	is.Error(rest.Given().BaseURI("localhost").Get("/path").Err())

	// Mirrors returns_false_when_uri_doesnt_contain_host
	is.Error(rest.Given().BaseURI("http://").Get("/path").Err())

	// Mirrors returns_false_when_uri_is_malformed
	is.Error(rest.Given().BaseURI("://not-valid").Get("/path").Err())

	// Unreachable port → connection refused (not a URI format error)
	is.Error(rest.Given().BaseURI("http://127.0.0.1:1").Get("/path").Err())

	// Mirrors returns_true_when_uri_contains_scheme_and_host: valid server succeeds
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(ts.Close)
	must.NoError(rest.Given().BaseURI(ts.URL).Get("/").Err())
}

// TestIOUtils_ReadAll mirrors Java's IOUtilsTest.toByteArray_small and toByteArray_big.
// Java uses Apache Commons IOUtils.toByteArray(InputStream); Go uses io.ReadAll.
func TestIOUtils_ReadAll(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	// toByteArray_small: 3-byte input
	small := []byte{1, 2, 3}
	out, err := io.ReadAll(bytes.NewReader(small))
	must.NoError(err)
	is.Equal(small, out)

	// toByteArray_big: 100 000 bytes (larger than any internal buffer)
	const size = 100_000
	big := make([]byte, size)
	for i := range big {
		big[i] = 1
	}
	out2, err := io.ReadAll(bytes.NewReader(big))
	must.NoError(err)
	is.Equal(big, out2)
}

func TestRequest_BodySerializationCandidates(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	var capturedBody string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 4096)
		n, _ := r.Body.Read(buf)
		capturedBody = string(buf[:n])
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(ts.Close)

	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	// Struct → marshaled to JSON
	must.NoError(rest.Given().BaseURI(ts.URL).BodyObject(User{Name: "Alice", Age: 30}).Post("/").Err())
	is.JSONEq(`{"name":"Alice","age":30}`, capturedBody)

	// Map → marshaled to JSON
	must.NoError(rest.Given().BaseURI(ts.URL).BodyObject(map[string]any{"k": "v", "n": 1}).Post("/").Err())
	is.JSONEq(`{"k":"v","n":1}`, capturedBody)

	// nil → marshals to literal "null"
	must.NoError(rest.Given().BaseURI(ts.URL).BodyObject(nil).Post("/").Err())
	is.Equal("null", capturedBody)

	// Non-marshalable type (channel) → buildErr set, request returns error
	ch := make(chan int)
	is.Error(rest.Given().BaseURI(ts.URL).BodyObject(ch).Post("/").Err())

	// String via Body() → sent as-is, no marshaling
	must.NoError(rest.Given().BaseURI(ts.URL).Body("plain text").Post("/").Err())
	is.Equal("plain text", capturedBody)

	// No body → empty request body
	must.NoError(rest.Given().BaseURI(ts.URL).Get("/").Err())
	is.Empty(capturedBody)
}

func TestConfigImmutability(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	cfg1 := rest.DefaultConfig()
	cfg2 := cfg1.HTTPClient(rest.HTTPClientConfig{
		Timeout: 5 * time.Second,
	})

	is.NotEqual(cfg1.HTTPClientConfig().Timeout, cfg2.HTTPClientConfig().Timeout)
	is.Equal(30*time.Second, cfg1.HTTPClientConfig().Timeout)
	is.Equal(5*time.Second, cfg2.HTTPClientConfig().Timeout)

	cfg3 := cfg1.Log(rest.LogConfig{EnableLoggingIfError: true})
	is.False(cfg1.LogConfig().EnableLoggingIfError)
	is.True(cfg3.LogConfig().EnableLoggingIfError)

	cfg4 := cfg1.Encoder(rest.EncoderConfig{DefaultCharset: "ISO-8859-1"})
	is.Equal("UTF-8", cfg1.EncoderConfig().DefaultCharset)
	is.Equal("ISO-8859-1", cfg4.EncoderConfig().DefaultCharset)

	cfg5 := cfg1.Decoder(rest.DecoderConfig{DefaultCharset: "ISO-8859-1"})
	is.Equal("UTF-8", cfg1.DecoderConfig().DefaultCharset)
	is.Equal("ISO-8859-1", cfg5.DecoderConfig().DefaultCharset)
}
