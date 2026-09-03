package rest_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func TestResponseLogSpec(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Trace", "12345")
		http.SetCookie(w, &http.Cookie{Name: "session", Value: "active"})
		if r.URL.Path == "/err" {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":"not found"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	t.Cleanup(ts.Close)

	// Test Log().All(), Status(), Headers(), Cookies(), Body()
	valid := rest.Given().
		BaseURI(ts.URL).
		Get("/test").
		Then().
		AssertWith(t)

	logSpec := valid.Log()
	must.NotNil(logSpec)

	var customLogBuf bytes.Buffer
	logSpec.Writer(&customLogBuf)
	logSpec.All()
	logSpec.Status()
	logSpec.Headers()
	logSpec.Cookies()
	logSpec.Body()
	logSpec.IfStatusCodeIs(200)

	is.Contains(customLogBuf.String(), "--- Response Details ---")

	// Test IfError() on error response
	errValid := rest.Given().
		BaseURI(ts.URL).
		Get("/err").
		Then()

	var errLogBuf bytes.Buffer
	errValid.Log().Writer(&errLogBuf).IfError()
	is.Contains(errLogBuf.String(), "Status code:\t404")
	is.Equal(http.StatusNotFound, errValid.Extract().StatusCode())
}

func TestRequestLogSpec(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(ts.Close)

	req := rest.Given().
		BaseURI(ts.URL).
		Header("X-Client", "GoClient").
		Cookie("user", "alice").
		Body(`{"test":true}`)

	reqLog := req.Log()
	must.NotNil(reqLog)

	reqLog.All()
	reqLog.Headers()
	reqLog.Cookies()
	reqLog.Body()
	reqLog.IfValidationFails()

	resp := req.Post("/log-check")
	is.NoError(resp.Err())
}

func TestLogging_PrettyPrintJSON(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	// Compact JSON response body → indented in log output
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"name":"Alice","age":30}`))
	}))
	t.Cleanup(ts.Close)

	var logBuf bytes.Buffer
	resp := rest.Given().
		BaseURI(ts.URL).
		Filter(rest.NewResponseLoggingFilter(&logBuf, rest.LogDetailAll)).
		Get("/")
	must.NoError(resp.Err())

	out := logBuf.String()
	is.Contains(out, "\"name\": \"Alice\"")
	is.Contains(out, "\"age\": 30")

	// Non-JSON body → raw string in log output (no crash)
	logBuf.Reset()
	ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("plain text response"))
	}))
	t.Cleanup(ts2.Close)

	resp2 := rest.Given().
		BaseURI(ts2.URL).
		Filter(rest.NewResponseLoggingFilter(&logBuf, rest.LogDetailAll)).
		Get("/")
	must.NoError(resp2.Err())
	is.Contains(logBuf.String(), "plain text response")

	// Empty body → no crash, body section empty
	logBuf.Reset()
	ts3 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(ts3.Close)

	resp3 := rest.Given().
		BaseURI(ts3.URL).
		Filter(rest.NewResponseLoggingFilter(&logBuf, rest.LogDetailAll)).
		Get("/")
	must.NoError(resp3.Err())
	is.NotEmpty(logBuf.String())
}

// ── PrettifierTest ───────────────────────────────────────────────────────────

func TestLogging_PrettyPrintXML(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	compactXML := `<catalog><book id="bk101"><author>Gambardella</author><title>Go Guide</title></book></catalog>`
	// Go's encoding/xml encoder escapes apostrophes as &#39; (valid XML) — use Contains not exact equality.

	// xml_prettify: compact XML → indented output with 2-space indent
	pretty, ok := rest.PrettyPrintXML(compactXML)
	is.True(ok)
	is.Contains(pretty, "<catalog>")
	is.Contains(pretty, "  <book id=\"bk101\">")
	is.Contains(pretty, "    <author>Gambardella</author>")
	is.Contains(pretty, "    <title>Go Guide</title>")

	// xml_string_data: same input produces same result
	pretty2, ok2 := rest.PrettyPrintXML(compactXML)
	is.True(ok2)
	is.Equal(pretty, pretty2)

	// empty_data: empty string returns ("", false)
	empty, ok3 := rest.PrettyPrintXML("")
	is.False(ok3)
	is.Equal("", empty)

	// whitespace-only string treated as empty
	ws, ok4 := rest.PrettyPrintXML("   ")
	is.False(ok4)
	is.Equal("", ws)

	// invalid XML returns ("", false)
	_, ok5 := rest.PrettyPrintXML("<not-closed>")
	is.False(ok5)

	// XML body is pretty-printed in response log output
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(compactXML))
	}))
	t.Cleanup(ts.Close)

	var logBuf bytes.Buffer
	resp := rest.Given().
		BaseURI(ts.URL).
		Filter(rest.NewResponseLoggingFilter(&logBuf, rest.LogDetailAll)).
		Get("/")
	is.NoError(resp.Err())
	is.Contains(logBuf.String(), "<author>Gambardella</author>")
}

// TestPrettifier_CatalogData mirrors Java's PrettifierTest.
// Java loads multipart.json and multipart.xml from the classpath; Go uses inline strings
// with equivalent catalog content. Tests: jsonPrettify, xml_prettify, json_string_data,
// xml_string_data, empty_data.
func TestPrettifier_CatalogData(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	// json_prettify / json_string_data: compact JSON → indented output with all fields
	compactJSON := `{"catalog":{"book":[{"id":"bk101","author":"Gambardella, Matthew","title":"XML Developers Guide","genre":"Computer","price":44.95,"publish_date":"2000-10-01"}]}}`
	var jsonBuf bytes.Buffer
	err := json.Indent(&jsonBuf, []byte(compactJSON), "", "  ")
	must.NoError(err)
	out := jsonBuf.String()
	is.Contains(out, `"author": "Gambardella, Matthew"`)
	is.Contains(out, `"genre": "Computer"`)
	is.Contains(out, `"price": 44.95`)
	is.Contains(out, `"publish_date": "2000-10-01"`)

	// xml_prettify / xml_string_data: compact XML → indented with 2-space indent
	// Note: Go's encoding/xml escapes apostrophes as &#39; (valid XML) — avoid apostrophes.
	compactXML := `<catalog><book id="bk101"><author>Gambardella, Matthew</author><title>XML Developers Guide</title><genre>Computer</genre><price>44.95</price><publish_date>2000-10-01</publish_date></book></catalog>`
	pretty, ok := rest.PrettyPrintXML(compactXML)
	is.True(ok)
	is.Contains(pretty, "<catalog>")
	is.Contains(pretty, `  <book id="bk101">`)
	is.Contains(pretty, "    <author>Gambardella, Matthew</author>")
	is.Contains(pretty, "    <genre>Computer</genre>")
	is.Contains(pretty, "    <price>44.95</price>")
	is.Contains(pretty, "    <publish_date>2000-10-01</publish_date>")

	// empty_data: empty string → PrettyPrintXML returns ("", false)
	emptyOut, ok2 := rest.PrettyPrintXML("")
	is.False(ok2)
	is.Equal("", emptyOut)

	// JSON response body is pretty-printed through the logging filter
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(compactJSON))
	}))
	t.Cleanup(ts.Close)

	var logBuf bytes.Buffer
	resp := rest.Given().
		BaseURI(ts.URL).
		Filter(rest.NewResponseLoggingFilter(&logBuf, rest.LogDetailAll)).
		Get("/")
	must.NoError(resp.Err())
	is.Contains(logBuf.String(), `"author": "Gambardella, Matthew"`)
}

func TestMultiPartFileUpload(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	tempDir := t.TempDir()
	tempFilePath := filepath.Join(tempDir, "upload_test.txt")
	err := os.WriteFile(tempFilePath, []byte("Hello Multipart File"), 0644)
	must.NoError(err)

	var receivedFilename string
	var receivedContent []byte

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(10 << 20); !is.NoError(err) {
			http.Error(w, "multipart error", http.StatusBadRequest)
			return
		}

		files := r.MultipartForm.File["my_file"]
		if !is.Len(files, 1) {
			return
		}
		receivedFilename = files[0].Filename

		f, err := files[0].Open()
		if !is.NoError(err) {
			return
		}
		defer f.Close()

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(f)
		receivedContent = buf.Bytes()

		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(ts.Close)

	resp := rest.Given().
		BaseURI(ts.URL).
		MultiPartFile("my_file", tempFilePath).
		Post("/upload-file")

	must.NoError(resp.Err())
	is.Equal(http.StatusOK, resp.StatusCode())
	is.Equal("upload_test.txt", receivedFilename)
	is.Equal("Hello Multipart File", string(receivedContent))
}
