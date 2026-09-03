package integration_test

// Ported from AcceptHeaderITest.java

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func TestJavaITest_AcceptHeader(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("AcceptStringHeaderIsPassedToServer", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Header("Accept", "application/json").
			ContentType(rest.ContentTypeJSON).
			Body(`{ "message" : "hello world"}`).
			Post("/jsonBodyAcceptHeader")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal("hello world", resp.AsString())
	})

	t.Run("AcceptContentTypeHeaderIsPassedToServer", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Accept(rest.ContentTypeJSON).
			ContentType(rest.ContentTypeJSON).
			Body(`{ "message" : "hello world"}`).
			Post("/jsonBodyAcceptHeader")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal("hello world", resp.AsString())
	})

	t.Run("AcceptFromRequestSpecIsApplied", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		spec := rest.NewRequestSpecBuilder().
			SetAccept(rest.ContentTypeJSON).
			Build()

		resp := rest.Given().
			Spec(spec).
			BaseURI(ts.URL).
			ContentType(rest.ContentTypeJSON).
			Body(`{ "message" : "hello world"}`).
			Post("/jsonBodyAcceptHeader")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal("hello world", resp.AsString())
	})
}
