package integration_test

// Ported from LogIfValidationFailsITest.java

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func TestJavaITest_LogIfValidationFails_RequestSpec(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("LogsToOutputWhenValidationFails", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		var buf bytes.Buffer
		cfg := rest.DefaultConfig().WithLog(rest.LogConfig{
			EnableLoggingIfValidationFails: true,
			Output:                         &buf,
		})

		valid := rest.Given().
			Config(cfg).
			BaseURI(ts.URL).
			QueryParam("firstName", "John").
			QueryParam("lastName", "Doe").
			Get("/greet").
			Then().
			StatusCode(http.StatusBadRequest) // fails: actual is 200

		is.True(valid.HasFailures())
		is.NotEmpty(buf.String(), "log output should be written when validation fails")
	})

	t.Run("DoesNotLogWhenValidationSucceeds", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		var buf bytes.Buffer
		cfg := rest.DefaultConfig().WithLog(rest.LogConfig{
			EnableLoggingIfValidationFails: true,
			Output:                         &buf,
		})

		valid := rest.Given().
			Config(cfg).
			BaseURI(ts.URL).
			QueryParam("firstName", "John").
			QueryParam("lastName", "Doe").
			Get("/greet").
			Then().
			StatusCode(http.StatusOK) // passes

		is.False(valid.HasFailures())
		is.Empty(buf.String(), "no log output should be written when validation succeeds")
	})

	t.Run("LogsBodyWhenBodyValidationFails", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var buf bytes.Buffer
		cfg := rest.DefaultConfig().WithLog(rest.LogConfig{
			EnableLoggingIfValidationFails: true,
			Output:                         &buf,
		})

		resp := rest.Given().
			Config(cfg).
			BaseURI(ts.URL).
			PathParam("firstName", "John").
			PathParam("lastName", "Doe").
			Get("/greeting/{firstName}/{lastName}")

		must.NoError(resp.Err())

		valid := resp.Then()
		valid.Body("fullName", "John Doe2") // fails: actual is "John Doe"

		is.True(valid.HasFailures())
		is.NotEmpty(buf.String(), "log output should include response body")
		is.Contains(buf.String(), "John Doe")
	})

	t.Run("DoesNotLogBodyWhenBodyValidationSucceeds", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var buf bytes.Buffer
		cfg := rest.DefaultConfig().WithLog(rest.LogConfig{
			EnableLoggingIfValidationFails: true,
			Output:                         &buf,
		})

		resp := rest.Given().
			Config(cfg).
			BaseURI(ts.URL).
			PathParam("firstName", "John").
			PathParam("lastName", "Doe").
			Get("/greeting/{firstName}/{lastName}")

		must.NoError(resp.Err())

		valid := resp.Then()
		valid.Body("fullName", "John Doe") // passes

		is.False(valid.HasFailures())
		is.Empty(buf.String(), "no log output when validation passes")
	})
}

func TestJavaITest_LogIfValidationFails_WithResponseSpec(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("LogsWhenResponseSpecFails", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		var buf bytes.Buffer
		cfg := rest.DefaultConfig().WithLog(rest.LogConfig{
			EnableLoggingIfValidationFails: true,
			Output:                         &buf,
		})

		spec := rest.NewResponseSpecBuilder().
			ExpectStatusCode(http.StatusBadRequest). // will fail
			Build()

		valid := rest.Given().
			Config(cfg).
			BaseURI(ts.URL).
			QueryParam("firstName", "John").
			QueryParam("lastName", "Doe").
			Get("/greet").
			Then().
			Spec(spec)

		is.True(valid.HasFailures())
		is.NotEmpty(buf.String())
	})
}
