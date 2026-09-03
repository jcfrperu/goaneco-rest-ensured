package integration_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func TestIntegration_QueryParamsAndPathParams(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	ts := integrationServer

	// 1. Path parameters with {code} placeholder
	respPath := rest.Given().
		BaseURI(ts.URL).
		PathParam("code", 201).
		Get("/status/{code}")

	must.NoError(respPath.Err())
	is.Equal(http.StatusCreated, respPath.StatusCode())

	// 2. Positional path parameters with {0}
	respPos := rest.Given().
		BaseURI(ts.URL).
		Get("/status/{0}", 202)

	must.NoError(respPos.Err())
	is.Equal(http.StatusAccepted, respPos.StatusCode())

	// 3. Form parameters
	respForm := rest.Given().
		BaseURI(ts.URL).
		FormParam("username", "admin").
		FormParam("action", "login").
		Post("/form")

	must.NoError(respForm.Err())
	is.Equal(http.StatusOK, respForm.StatusCode())
	is.Equal("admin", respForm.JsonPath().GetString("username"))
	is.Equal("login", respForm.JsonPath().GetString("action"))
}

func TestIntegration_ParamConfigOmitBehavior(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	ts := integrationServer

	cfg := rest.DefaultConfig().WithParam(rest.ParamConfig{
		EmptyParamsBehavior: "omit",
	})

	resp := rest.Given().
		Config(cfg).
		BaseURI(ts.URL).
		QueryParam("active", "true").
		QueryParam("blank", "").
		Get("/echo")

	must.NoError(resp.Err())
	is.Equal(http.StatusOK, resp.StatusCode())
	rawQuery := resp.RawRequest().URL.RawQuery
	is.Contains(rawQuery, "active=true")
	is.NotContains(rawQuery, "blank")
}
