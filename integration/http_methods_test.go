package integration_test

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func TestIntegration_HTTPMethods(t *testing.T) {
	t.Parallel()

	ts := integrationServer

	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		expectedSt int
	}{
		{
			name:       "GET /hello",
			method:     "GET",
			path:       "/hello",
			expectedSt: http.StatusOK,
		},
		{
			name:       "POST /echo",
			method:     "POST",
			path:       "/echo",
			body:       `{"action":"create"}`,
			expectedSt: http.StatusOK,
		},
		{
			name:       "PUT /echo",
			method:     "PUT",
			path:       "/echo",
			body:       `{"action":"update"}`,
			expectedSt: http.StatusOK,
		},
		{
			name:       "PATCH /echo",
			method:     "PATCH",
			path:       "/echo",
			body:       `{"action":"patch"}`,
			expectedSt: http.StatusOK,
		},
		{
			name:       "DELETE /status/204",
			method:     "DELETE",
			path:       "/status/204",
			expectedSt: http.StatusNoContent,
		},
		{
			name:       "HEAD /status/200",
			method:     "HEAD",
			path:       "/status/200",
			expectedSt: http.StatusOK,
		},
		{
			name:       "OPTIONS /status/200",
			method:     "OPTIONS",
			path:       "/status/200",
			expectedSt: http.StatusOK,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)
			must := require.New(t)

			req := rest.Given().
				BaseURI(ts.URL).
				ContentType(rest.ContentTypeJSON)

			if tt.body != "" {
				req.Body(tt.body)
			}

			var resp *rest.Response
			switch tt.method {
			case "GET":
				resp = req.Get(tt.path)
			case "POST":
				resp = req.Post(tt.path)
			case "PUT":
				resp = req.Put(tt.path)
			case "PATCH":
				resp = req.Patch(tt.path)
			case "DELETE":
				resp = req.Delete(tt.path)
			case "HEAD":
				resp = req.Head(tt.path)
			case "OPTIONS":
				resp = req.Options(tt.path)
			}

			must.NotNil(resp)
			must.NoError(resp.Err())
			is.Equal(tt.expectedSt, resp.StatusCode())
			if tt.method == "HEAD" {
				is.Empty(resp.AsString())
			}
		})
	}
}

func TestIntegration_MultipartUpload(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	ts := integrationServer

	tempDir := t.TempDir()
	sampleFile := filepath.Join(tempDir, "sample.txt")
	err := os.WriteFile(sampleFile, []byte("hello multipart file"), 0600)
	must.NoError(err)

	resp := rest.Given().
		BaseURI(ts.URL).
		FormParam("field1", "value1").
		MultiPartFile("attachment", sampleFile).
		MultiPartNamed("inmemory", "doc.json", []byte(`{"nested":true}`), "application/json").
		Post("/upload")

	must.NoError(resp.Err())
	is.Equal(http.StatusOK, resp.StatusCode())
	is.Equal("value1", resp.JsonPath().GetString("fields.field1"))

	files := resp.JsonPath().GetList("files")
	is.Len(files, 2)
}
