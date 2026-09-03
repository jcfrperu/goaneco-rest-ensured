package integration_test

// Ported from MultiPartUploadITest.java

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func TestJavaITest_MultiPart_ByteArrayAndString(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("MultiPartUploadingWorksForByteArrays", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			MultiPartNamed("file", "test.txt", []byte("Hello, World!"), "text/plain").
			Post("/upload")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Contains(resp.AsString(), "test.txt")
	})

	t.Run("MultiPartUploadingWorksForStrings", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			MultiPart("content", []byte("text content"), "text/plain").
			Post("/upload")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.NotEmpty(resp.AsString())
	})

	t.Run("MultiPartUploadingWorksForJsonObjects", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		jsonData := []byte(`{"key":"value","count":42}`)
		resp := rest.Given().
			BaseURI(ts.URL).
			MultiPartNamed("json_payload", "data.json", jsonData, "application/json").
			Post("/upload")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Contains(resp.AsString(), "data.json")
	})

	t.Run("MultiPartUploadingWorksForXmlObjects", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		xmlData := []byte(`<?xml version="1.0"?><root><item>value</item></root>`)
		resp := rest.Given().
			BaseURI(ts.URL).
			MultiPartNamed("xml_data", "data.xml", xmlData, "application/xml").
			Post("/upload")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Contains(resp.AsString(), "data.xml")
	})

	t.Run("MultiPartUploadingWorksForMultipleStrings", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			MultiPartNamed("file1", "first.txt", []byte("first content"), "text/plain").
			MultiPartNamed("file2", "second.txt", []byte("second content"), "text/plain").
			Post("/upload")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		body := resp.AsString()
		is.Contains(body, "first.txt")
		is.Contains(body, "second.txt")
	})

	t.Run("MultiPartUploadingWorksForByteArrayAndStrings", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			MultiPartNamed("text_file", "readme.txt", []byte("text content"), "text/plain").
			MultiPartNamed("binary_file", "data.bin", []byte{0x00, 0x01, 0x02, 0x03}, "application/octet-stream").
			Post("/upload")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		body := resp.AsString()
		is.Contains(body, "readme.txt")
		is.Contains(body, "data.bin")
	})
}

func TestJavaITest_MultiPart_WithFormParams(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("MultiPartUploadingWorksForByteArrayAndFormParams", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			FormParam("description", "test upload").
			MultiPartNamed("file", "document.pdf", []byte{0x25, 0x50, 0x44, 0x46}, "application/pdf").
			Post("/upload")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		body := resp.AsString()
		is.Contains(body, "document.pdf")
		is.Contains(body, "test upload")
	})

	t.Run("MultiPartUploadingWorksForFormParamsAndByteArray", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			MultiPart("text_field", []byte("field value"), "text/plain").
			FormParam("extra", "extra-value").
			Post("/upload")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Contains(resp.AsString(), "extra-value")
	})

	t.Run("MultiPartUploadingWorksForByteArrayAndNumberFormParams", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			FormParam("count", "5").
			FormParam("offset", "0").
			MultiPartNamed("data", "numbers.bin", []byte{1, 2, 3, 4, 5}, "application/octet-stream").
			Post("/upload")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		body := resp.AsString()
		is.Contains(body, "5")
		is.Contains(body, "numbers.bin")
	})

	t.Run("MultiPartFormFieldsAreReturnedInResponse", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			FormParam("title", "My Upload").
			FormParam("author", "Test User").
			MultiPartNamed("file", "content.txt", []byte("file content"), "text/plain").
			Post("/upload")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		body := resp.AsString()
		is.Contains(body, "My Upload")
		is.Contains(body, "Test User")
		is.Contains(body, "content.txt")
	})

	t.Run("BytesAndFormParamUploadingWorkUsingRequestBuilder", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		spec := rest.NewRequestSpecBuilder().
			SetBaseURI(ts.URL).
			AddFormParam("category", "test").
			Build()

		resp := rest.Given().
			Spec(spec).
			MultiPartNamed("file", "spec-test.txt", []byte("spec file content"), "text/plain").
			Post("/upload")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		body := resp.AsString()
		is.Contains(body, "spec-test.txt")
		is.Contains(body, "test")
	})
}

func TestJavaITest_MultiPart_ContentTypeAndCharset(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("MultiPartSupportsSpecifyingCharset", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		utf8Data := []byte("Hello, 世界!")
		resp := rest.Given().
			BaseURI(ts.URL).
			MultiPartNamed("file", "utf8.txt", utf8Data, "text/plain; charset=utf-8").
			Post("/upload")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Contains(resp.AsString(), "utf8.txt")
	})

	t.Run("MultiPartBinaryUploadWithOctetStream", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		binaryData := make([]byte, 100)
		for i := range binaryData {
			binaryData[i] = byte(i)
		}

		resp := rest.Given().
			BaseURI(ts.URL).
			MultiPartNamed("binary", "data.bin", binaryData, "application/octet-stream").
			Post("/upload")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Contains(resp.AsString(), "data.bin")
	})

	t.Run("MultiPartUploadWithExplicitImageContentType", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		imageData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A}
		resp := rest.Given().
			BaseURI(ts.URL).
			MultiPartNamed("image", "photo.png", imageData, "image/png").
			Post("/upload")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Contains(resp.AsString(), "photo.png")
	})

	t.Run("MultiPartUploadRequestBodyIsNotEmpty", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			MultiPart("field1", []byte("value1"), "text/plain").
			MultiPart("field2", []byte("value2"), "text/plain").
			Post("/upload")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.NotEmpty(resp.AsString())
	})
}

func TestJavaITest_MultiPart_ResponseParsing(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("MultiPartResponseContainsUploadedFileNames", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			MultiPartNamed("document", "my-document.pdf", []byte("PDF content"), "application/pdf").
			Post("/upload")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())

		jp := resp.JsonPath()
		files := jp.GetStringList("files")
		is.Contains(files, "my-document.pdf")
	})

	t.Run("MultiPartResponseContainsFieldValues", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			FormParam("name", "upload-test").
			MultiPartNamed("file", "test.txt", []byte("content"), "text/plain").
			Post("/upload")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())

		jp := resp.JsonPath()
		is.Equal("upload-test", jp.GetString("fields.name"))
	})

	t.Run("MultiPartSingleFileUploadAndVerify", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			MultiPartNamed("attachment", "report.csv", []byte("col1,col2\nval1,val2"), "text/csv").
			Post("/upload")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())

		var result struct {
			Fields map[string]string `json:"fields"`
			Files  []string          `json:"files"`
		}
		must.NoError(resp.As(&result))
		is.Contains(result.Files, "report.csv")
	})

	t.Run("MultiPartMultipleFilesAreParsedInResponse", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			MultiPartNamed("file1", "a.txt", []byte("aaa"), "text/plain").
			MultiPartNamed("file2", "b.txt", []byte("bbb"), "text/plain").
			MultiPartNamed("file3", "c.txt", []byte("ccc"), "text/plain").
			Post("/upload")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())

		jp := resp.JsonPath()
		files := jp.GetStringList("files")
		is.Len(files, 3)
		is.Contains(files, "a.txt")
		is.Contains(files, "b.txt")
		is.Contains(files, "c.txt")
	})
}
