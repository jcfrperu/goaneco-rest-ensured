package integration_test

// Ported from URLITest.java.
// Java uses static global RestAssured.baseURI/basePath/port; Go uses rest.BaseURI/BasePath/Port.
// Each subtest that mutates globals calls t.Cleanup(rest.Reset) to restore defaults.

import (
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func TestJavaITest_URL_FullyQualifiedOverridesGlobals(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("SpecifyingFullyQualifiedPathOverridesValues", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// Fully-qualified URL in Get() must override per-request settings.
		// Simulate wrong per-request settings; the fully-qualified URL should win.
		resp := rest.Given().
			BasePath("/something").
			BaseURI("http://www.example.com").
			Port(80).
			Get(ts.URL + "/jsonStore")
		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal("red", resp.JsonPath().GetString("store.bicycle.color"))
	})

	t.Run("FullyQualifiedURLInGetIgnoresGlobalPort", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// Passing a full URL to Get() must use that URL's port, not a per-request port override.
		resp := rest.Given().
			Port(9999).
			Get(ts.URL + "/hello")
		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
	})
}

func TestJavaITest_URL_BaseURISlashHandling(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("WhenBaseURIEndsWithSlashAndPathBeginsWithSlashThenOneSlashIsRemoved", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// BaseURI ends with "/" and path starts with "/" — result should not have "//".
		baseWithSlash := strings.TrimRight(ts.URL, "/") + "/"
		resp := rest.Given().
			BaseURI(baseWithSlash).
			Get("/lotto")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal(5, resp.JsonPath().GetInt("lotto.lottoId"))
	})

	t.Run("WhenBaseURIAndPathDoesntEndWithSlashThenOneSlashIsInserted", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("lotto") // no leading slash

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal(5, resp.JsonPath().GetInt("lotto.lottoId"))
	})

	t.Run("BaseURIPicksUpSchemeAndPort", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// Full URL including path in BaseURI; empty path in Get().
		resp := rest.Given().
			BaseURI(ts.URL + "/lotto").
			Get("")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal(5, resp.JsonPath().GetInt("lotto.lottoId"))
	})

	t.Run("BaseURIPicksUpSchemeAndPortAndBasePath", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// Combine BasePath with BaseURI — result must be baseURI + basePath + path.
		resp := rest.Given().
			BaseURI(ts.URL).
			BasePath("/lotto").
			Get("")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal(5, resp.JsonPath().GetInt("lotto.lottoId"))
	})

	t.Run("TrailingSlashInPathIsRetained", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// /hello/ is served the same as /hello by the test server.
		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/hello/")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
	})

	t.Run("AddsSingleTrailingSlashToPathWhenSlashIsUsedAsPath", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL + "/hello").
			Get("/")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
	})
}

func TestJavaITest_URL_PortHandling(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("FullyQualifiedURLIncludingPortWorks", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().Get(ts.URL + "/greet?firstName=John&lastName=Doe")
		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Contains(resp.AsString(), "John")
	})

	t.Run("TakesSpecificationPortIntoAccount", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// Extract port from the dynamic test server URL.
		parts := strings.Split(ts.URL, ":")
		portStr := parts[len(parts)-1]
		port, _ := strconv.Atoi(portStr)

		// scheme+host without port, port set separately via Port().
		schemeHost := strings.Join(parts[:len(parts)-1], ":")

		resp := rest.Given().
			BaseURI(schemeHost).
			Port(port).
			Get("/lotto")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal(5, resp.JsonPath().GetInt("lotto.lottoId"))
	})

	t.Run("TakesNonStaticSpecificationPortViaRequestSpec", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		parts := strings.Split(ts.URL, ":")
		portStr := parts[len(parts)-1]
		port, _ := strconv.Atoi(portStr)
		schemeHost := strings.Join(parts[:len(parts)-1], ":")

		spec := rest.NewRequestSpecBuilder().
			SetBaseURI(schemeHost).
			SetPort(port).
			Build()

		resp := rest.Given().
			Spec(spec).
			Get("/hello")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
	})
}

func TestJavaITest_URL_GlobalSettings(t *testing.T) {
	// NOT parallel at top level — subtests mutate globals sequentially.
	ts := integrationServer

	t.Run("GlobalBaseURIOverridesDefault", func(t *testing.T) {
		t.Cleanup(rest.Reset)
		rest.BaseURI(ts.URL)
		resp := rest.Get("/lotto")
		rest.Reset()

		is := assert.New(t)
		must := require.New(t)
		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal(5, resp.JsonPath().GetInt("lotto.lottoId"))
	})

	t.Run("GlobalBasePathIsAppliedToAllRequests", func(t *testing.T) {
		t.Cleanup(rest.Reset)
		rest.BaseURI(ts.URL)
		rest.BasePath("/lotto")
		resp := rest.Get("")
		rest.Reset()

		is := assert.New(t)
		must := require.New(t)
		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal(5, resp.JsonPath().GetInt("lotto.lottoId"))
	})

	t.Run("ResetRestoresDefaultGlobals", func(t *testing.T) {
		t.Cleanup(rest.Reset)
		rest.BaseURI("http://example.invalid")
		rest.BasePath("/bogus")
		rest.Port(1)
		rest.Reset()

		// After reset, per-request override must work correctly.
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().BaseURI(ts.URL).Get("/hello")
		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
	})

	t.Run("GlobalBasePathWithPerRequestBaseURI", func(t *testing.T) {
		t.Cleanup(rest.Reset)
		rest.BasePath("/greet")
		resp := rest.Given().
			BaseURI(ts.URL).
			QueryParam("firstName", "Jane").
			QueryParam("lastName", "Doe").
			Get("")
		rest.Reset()

		is := assert.New(t)
		must := require.New(t)
		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Contains(resp.AsString(), "Jane")
	})
}

func TestJavaITest_URL_BasePathCombinations(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("BasePathSetViaSpecBuilderWorks", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		spec := rest.NewRequestSpecBuilder().
			SetBaseURI(ts.URL).
			SetBasePath("/greet").
			Build()

		resp := rest.Given().
			Spec(spec).
			QueryParam("firstName", "Alice").
			QueryParam("lastName", "Smith").
			Get("")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Contains(resp.AsString(), "Alice")
	})

	t.Run("BasePathAndExplicitPathAreCombined", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// BasePath "/auth" + path "/basic" → "/auth/basic"
		resp := rest.Given().
			BaseURI(ts.URL).
			BasePath("/auth").
			Auth().Basic("admin", "secret").
			Get("/basic")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
	})

	t.Run("EmptyBasePathAndEmptyPathYieldsRoot", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// BaseURI already includes the path; BasePath and Get path are empty.
		resp := rest.Given().
			BaseURI(ts.URL + "/hello").
			BasePath("").
			Get("")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
	})

	t.Run("FullyQualifiedURLInGetIgnoresBasePath", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// Even with BasePath set, a fully-qualified URL passed to Get() should be used as-is.
		resp := rest.Given().
			BaseURI(ts.URL).
			BasePath("/should/be/ignored").
			Get(ts.URL + "/lotto")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal(5, resp.JsonPath().GetInt("lotto.lottoId"))
	})

	t.Run("BasePathWithQueryParamWorks", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			BasePath("/greet").
			QueryParam("firstName", "Bob").
			QueryParam("lastName", "Builder").
			Get("")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Contains(resp.AsString(), "Bob")
	})
}
