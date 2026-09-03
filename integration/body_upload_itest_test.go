package integration_test

// Ported from NonMultiPartUploadITest.java, FileUploadingITest.java,
// BodyWithCustomContentTypeITest.java, GetWithContentITest.java

import (
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func TestJavaITest_NonMultiPartUpload(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("UploadByteArrayWithPost", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		payload := []byte("binary content for upload test")

		resp := rest.Given().
			BaseURI(ts.URL).
			ContentType(rest.ContentTypeBinary).
			BodyBytes(payload).
			Post("/reflect")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal(string(payload), resp.AsString())
	})

	t.Run("UploadByteArrayWithPut", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		payload := []byte("binary content for put upload")

		resp := rest.Given().
			BaseURI(ts.URL).
			ContentType(rest.ContentTypeBinary).
			BodyBytes(payload).
			Put("/reflect")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal(string(payload), resp.AsString())
	})

	t.Run("UploadByteArrayWithoutExplicitContentType", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		payload := []byte("content without explicit type")

		resp := rest.Given().
			BaseURI(ts.URL).
			BodyBytes(payload).
			Post("/reflect")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal(string(payload), resp.AsString())
	})

	t.Run("UploadTextBodyWithPost", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			ContentType(rest.ContentTypeText).
			Body("Hello World").
			Post("/reflect")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal("Hello World", resp.AsString())
	})
}

func TestJavaITest_FileUpload(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("CanUploadJSONFromFile", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		f, err := os.CreateTemp("", "*.json")
		must.NoError(err)
		t.Cleanup(func() { os.Remove(f.Name()) })
		_, err = f.WriteString(`{"message":"hello world"}`)
		must.NoError(err)
		must.NoError(f.Close())

		fileBytes, err := os.ReadFile(f.Name())
		must.NoError(err)

		resp := rest.Given().
			BaseURI(ts.URL).
			ContentType(rest.ContentTypeJSON).
			BodyBytes(fileBytes).
			Post("/reflect")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal("hello world", resp.JsonPath().GetString("message"))
	})

	t.Run("CanUploadXMLFromFile", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		xmlContent := `<tag attr="value">content</tag>`
		f, err := os.CreateTemp("", "*.xml")
		must.NoError(err)
		t.Cleanup(func() { os.Remove(f.Name()) })
		_, err = f.WriteString(xmlContent)
		must.NoError(err)
		must.NoError(f.Close())

		fileBytes, err := os.ReadFile(f.Name())
		must.NoError(err)

		resp := rest.Given().
			BaseURI(ts.URL).
			ContentType(rest.ContentTypeXML).
			BodyBytes(fileBytes).
			Post("/reflect")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Contains(resp.AsString(), "content")
	})

	t.Run("CanUploadTextFromFile", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		f, err := os.CreateTemp("", "*.txt")
		must.NoError(err)
		t.Cleanup(func() { os.Remove(f.Name()) })
		_, err = f.WriteString("Hello World")
		must.NoError(err)
		must.NoError(f.Close())

		fileBytes, err := os.ReadFile(f.Name())
		must.NoError(err)

		resp := rest.Given().
			BaseURI(ts.URL).
			ContentType(rest.ContentTypeText).
			BodyBytes(fileBytes).
			Post("/reflect")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal("Hello World", resp.AsString())
	})

	t.Run("CanUploadBinaryFromFile", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		f, err := os.CreateTemp("", "*.bin")
		must.NoError(err)
		t.Cleanup(func() { os.Remove(f.Name()) })
		_, err = f.Write([]byte{0x01, 0x02, 0x03, 0x04, 0x05})
		must.NoError(err)
		must.NoError(f.Close())

		fileBytes, err := os.ReadFile(f.Name())
		must.NoError(err)

		resp := rest.Given().
			BaseURI(ts.URL).
			ContentType(rest.ContentTypeBinary).
			BodyBytes(fileBytes).
			Post("/reflect")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal(fileBytes, resp.AsBytes())
	})

	t.Run("CanUploadFileWithCustomContentType", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		f, err := os.CreateTemp("", "*.txt")
		must.NoError(err)
		t.Cleanup(func() { os.Remove(f.Name()) })
		_, err = f.WriteString("Custom type content")
		must.NoError(err)
		must.NoError(f.Close())

		fileBytes, err := os.ReadFile(f.Name())
		must.NoError(err)

		resp := rest.Given().
			BaseURI(ts.URL).
			Header("Content-Type", "application/something").
			BodyBytes(fileBytes).
			Post("/reflect")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal("Custom type content", resp.AsString())
	})
}

func TestJavaITest_GetWithContent(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("GetMethodWithBodyContent", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Body("hullo").
			Get("/getWithContent")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
	})
}

func TestJavaITest_BodyWithCustomContentType(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("URLEncodedBodyStringIsPostedCorrectly", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// Post a raw URL-encoded body string — server parses and returns XML
		resp := rest.Given().
			BaseURI(ts.URL).
			ContentType(rest.ContentTypeURLEncoded).
			Body("firstName=John&lastName=Doe").
			Post("/greetXML")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		xp, err := resp.XmlPath()
		must.NoError(err)
		is.Equal("John", xp.GetString("//greeting/firstName"))
		is.Equal("Doe", xp.GetString("//greeting/lastName"))
	})

	t.Run("URLEncodedBodyRoundTripViaReflect", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// Verify the raw body is sent as-is when content type is URLENC
		resp := rest.Given().
			BaseURI(ts.URL).
			ContentType(rest.ContentTypeURLEncoded).
			Body("key=value&other=data").
			Post("/reflect")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Contains(resp.AsString(), "key=value")
	})
}
