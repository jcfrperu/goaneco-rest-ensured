package rest_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func TestPathSupport_GetPath(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	// Fully qualified URIs without an explicit path — must return "/"
	is.Equal("/", rest.GetPath("http://localhost:8080"))
	is.Equal("/", rest.GetPath("http://localhost"))

	// Relative paths — returned as-is with "/" prefix when missing
	is.Equal("/path", rest.GetPath("/path"))
	is.Equal("/path", rest.GetPath("path"))

	// Query parameters stripped from relative paths
	is.Equal("/path", rest.GetPath("path?q=r&u=2"))

	// Query parameters stripped from fully qualified URIs with a path
	is.Equal("/path", rest.GetPath("http://localhost:808/path?u=4"))
	is.Equal("/path", rest.GetPath("http://localhost/path?u=4"))

	// Fully qualified URI with query but no path — must return "/"
	is.Equal("/", rest.GetPath("http://localhost?u=4"))

	// Query parameter whose value contains a scheme — path component only
	is.Equal("/path", rest.GetPath("/path?uri=http://localhost"))
	is.Equal("/path", rest.GetPath("http://localhost/path?uri=http://localhost"))
}
