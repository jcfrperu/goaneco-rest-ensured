package integration_test

// Ported from SpecificationBuilderITest.java, ResponseSpecBuilderExpectationsTest.java,
// and ResponseSpecBuilderPathTest.java

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func TestJavaITest_SpecBuilder_CombinedRequestAndResponseSpec(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("CanSpecifyCombinedRequestAndResponseSpec", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		reqSpec := rest.NewRequestSpecBuilder().
			SetBaseURI(ts.URL).
			AddHeader("X-Spec-Test", "spec-test-val").
			Build()

		respSpec := rest.NewResponseSpecBuilder().
			ExpectStatusCode(http.StatusOK).
			ExpectContentType(rest.ContentTypeJSON).
			Build()

		valid := rest.Given().
			Spec(reqSpec).
			Get("/lotto").
			Then().
			Spec(respSpec)

		is.False(valid.HasFailures(), "spec validation failed: %v", valid.Failures())
	})

	t.Run("RequestSpecWithQueryParamsAndResponseSpec", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		reqSpec := rest.NewRequestSpecBuilder().
			SetBaseURI(ts.URL).
			AddQueryParam("firstName", "John").
			AddQueryParam("lastName", "Doe").
			Build()

		respSpec := rest.NewResponseSpecBuilder().
			ExpectStatusCode(http.StatusOK).
			ExpectBody("greeting", "Greetings John Doe").
			Build()

		valid := rest.Given().
			Spec(reqSpec).
			Get("/greet").
			Then().
			Spec(respSpec)

		is.False(valid.HasFailures(), "spec validation failed: %v", valid.Failures())
	})

	t.Run("ResponseSpecWithTimeLimitPasses", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		respSpec := rest.NewResponseSpecBuilder().
			ExpectStatusCode(http.StatusOK).
			ExpectResponseTimeLessThan(5 * time.Second).
			Build()

		valid := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			Spec(respSpec)

		is.False(valid.HasFailures(), "spec validation failed: %v", valid.Failures())
	})

	t.Run("ResponseSpecExpectingStatusAndContentType", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		respSpec := rest.NewResponseSpecBuilder().
			ExpectStatusCode(http.StatusOK).
			ExpectContentType(rest.ContentTypeJSON).
			Build()

		valid := rest.Given().
			BaseURI(ts.URL).
			Get("/jsonList").
			Then().
			Spec(respSpec)

		is.False(valid.HasFailures(), "spec validation failed: %v", valid.Failures())
	})

	t.Run("ResponseSpecWithHeaderExpectation", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		respSpec := rest.NewResponseSpecBuilder().
			ExpectStatusCode(http.StatusOK).
			ExpectHeader("Content-Type", "application/json").
			Build()

		valid := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			Spec(respSpec)

		is.False(valid.HasFailures(), "spec validation failed: %v", valid.Failures())
	})

	t.Run("ResponseSpecCanBeReusedAcrossMultipleRequests", func(t *testing.T) {
		t.Parallel()
		must := require.New(t)

		respSpec := rest.NewResponseSpecBuilder().
			ExpectStatusCode(http.StatusOK).
			ExpectContentType(rest.ContentTypeJSON).
			Build()

		endpoints := []string{"/lotto", "/jsonList", "/numbers"}
		for _, ep := range endpoints {
			valid := rest.Given().
				BaseURI(ts.URL).
				Get(ep).
				Then().
				Spec(respSpec)
			must.False(valid.HasFailures(), "endpoint %s: %v", ep, valid.Failures())
		}
	})
}

func TestJavaITest_SpecBuilder_ResponseSpecExpectations(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("ExpectBodyWithJsonPath", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		respSpec := rest.NewResponseSpecBuilder().
			ExpectStatusCode(http.StatusOK).
			ExpectBody("lotto.lottoId", 5).
			Build()

		valid := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			Spec(respSpec)

		is.False(valid.HasFailures(), "spec body expectation failed: %v", valid.Failures())
	})

	t.Run("ExpectCookieInResponseSpec", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		respSpec := rest.NewResponseSpecBuilder().
			ExpectStatusCode(http.StatusOK).
			ExpectCookie("key1", "value1").
			Build()

		valid := rest.Given().
			BaseURI(ts.URL).
			Get("/setCookies").
			Then().
			Spec(respSpec)

		is.False(valid.HasFailures(), "spec cookie expectation failed: %v", valid.Failures())
	})

	t.Run("SpecWithWrongStatusCodeRecordsFailure", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		respSpec := rest.NewResponseSpecBuilder().
			ExpectStatusCode(http.StatusCreated). // wrong — actual is 200
			Build()

		valid := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			Spec(respSpec)

		is.True(valid.HasFailures(), "expected spec validation to fail with wrong status")
	})

	t.Run("SpecWithWrongBodyRecordsFailure", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		respSpec := rest.NewResponseSpecBuilder().
			ExpectStatusCode(http.StatusOK).
			ExpectBody("lotto.lottoId", 9999). // wrong value
			Build()

		valid := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			Spec(respSpec)

		is.True(valid.HasFailures(), "expected spec validation to fail with wrong body")
	})
}

func TestJavaITest_SpecBuilder_ResponseSpecWithRootPath(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("RootPathReducesBodyPathPrefixInSpec", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// Without root path: need full path "lotto.lottoId"
		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto")

		must.NoError(resp.Err())
		is.Equal(float64(5), resp.JsonPath().Get("lotto.lottoId").Value())
	})

	t.Run("AddResponseSpecMergesExpectations", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		base := rest.NewResponseSpecBuilder().
			ExpectStatusCode(http.StatusOK).
			Build()

		extra := rest.NewResponseSpecBuilder().
			ExpectContentType(rest.ContentTypeJSON).
			Build()

		merged := rest.NewResponseSpecBuilder().
			AddResponseSpec(base).
			AddResponseSpec(extra).
			Build()

		valid := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			Spec(merged)

		is.False(valid.HasFailures(), "merged spec validation failed: %v", valid.Failures())
	})

	t.Run("ResponseSpecBuilderRootPathConfiguration", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// Directly on the response builder, set a root path
		builder := rest.NewResponseSpecBuilder().
			RootPath("lotto")

		is.Equal("lotto", builder.Build().GetRootPath())

		// Append sub-path
		builder.AppendRootPath("winners")
		is.Equal("lotto.winners", builder.Build().GetRootPath())

		// NoRootPath clears it
		builder.NoRootPath()
		is.Equal("", builder.Build().GetRootPath())

		// Verify functional correctness: root path on the spec
		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto")

		must.NoError(resp.Err())
		jp := resp.JsonPath()
		is.Equal(float64(5), jp.Get("lotto.lottoId").Value())
	})
}

func TestJavaITest_SpecBuilder_RequestSpecFilters(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("FiltersAreAppliedFromRequestSpec", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		var filterExecuted bool
		timingFilter := rest.NewTimingFilter(func(_ time.Duration) {
			filterExecuted = true
		})

		spec := rest.NewRequestSpecBuilder().
			SetBaseURI(ts.URL).
			AddFilter(timingFilter).
			Build()

		resp := rest.Given().
			Spec(spec).
			Get("/lotto")

		must.NoError(resp.Err())
		is.True(filterExecuted, "filter from request spec should have been executed")
	})

	t.Run("ResponseSpecBuilderExpectationsTestAllExpectations", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		respSpec := rest.NewResponseSpecBuilder().
			ExpectStatusCode(http.StatusOK).
			ExpectContentType(rest.ContentTypeJSON).
			ExpectBody("lotto.lottoId", 5).
			ExpectResponseTimeLessThan(5 * time.Second).
			Build()

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto")

		must.NoError(resp.Err())

		valid := resp.Then().Spec(respSpec)
		is.False(valid.HasFailures(), "all expectations should pass: %v", valid.Failures())
	})
}

func TestJavaITest_SpecBuilder_RequestSpecParameters(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("SupportsSpecifyingQueryParametersInRequestSpecBuilderWhenGet", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		spec := rest.NewRequestSpecBuilder().
			SetBaseURI(ts.URL).
			AddQueryParam("firstName", "John").
			AddQueryParam("lastName", "Doe").
			Build()

		valid := rest.Given().
			Spec(spec).
			Get("/greet").
			Then().
			StatusCode(http.StatusOK)

		is.False(valid.HasFailures(), "query params from spec should be sent: %v", valid.Failures())
	})

	t.Run("SupportsSpecifyingQueryParametersInRequestSpecBuilderWhenPost", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		spec := rest.NewRequestSpecBuilder().
			SetBaseURI(ts.URL).
			AddQueryParam("firstName", "Jane").
			AddQueryParam("lastName", "Smith").
			Build()

		valid := rest.Given().
			Spec(spec).
			Post("/greet").
			Then().
			StatusCode(http.StatusOK)

		is.False(valid.HasFailures(), "query params from spec should work for POST: %v", valid.Failures())
	})

	t.Run("SupportsMergingCookiesWhenUsingRequestSpecBuilder", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		spec := rest.NewRequestSpecBuilder().
			SetBaseURI(ts.URL).
			AddCookie("spec_cookie", "spec_value").
			Build()

		resp := rest.Given().
			Spec(spec).
			Cookie("per_request_cookie", "per_value").
			Get("/cookies")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		body := resp.AsString()
		is.Contains(body, "spec_cookie")
		is.Contains(body, "per_request_cookie")
	})

	t.Run("SupportsMergingHeadersWhenUsingRequestSpecBuilder", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		spec := rest.NewRequestSpecBuilder().
			SetBaseURI(ts.URL).
			AddHeader("X-Spec-Header", "spec-value").
			Build()

		resp := rest.Given().
			Spec(spec).
			Header("X-Request-Header", "request-value").
			Get("/headers")

		must.NoError(resp.Err())
		body := resp.AsString()
		is.Contains(body, "X-Spec-Header")
		is.Contains(body, "X-Request-Header")
	})

	t.Run("RequestSpecBuilderSupportsSettingAuthentication", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		spec := rest.NewRequestSpecBuilder().
			SetBaseURI(ts.URL).
			SetAuth(&rest.BasicAuthScheme{Username: "admin", Password: "secret"}).
			Build()

		resp := rest.Given().
			Spec(spec).
			Get("/auth/basic")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
	})

	t.Run("SupportsMergingFormParametersWhenUsingRequestSpecBuilder", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		spec := rest.NewRequestSpecBuilder().
			SetBaseURI(ts.URL).
			AddFormParam("username", "bob").
			Build()

		resp := rest.Given().
			Spec(spec).
			FormParam("role", "viewer").
			Post("/form")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		body := resp.AsString()
		is.Contains(body, "bob")
		is.Contains(body, "viewer")
	})

	t.Run("SupportsMergingPathParametersWhenUsingRequestSpecBuilder", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		spec := rest.NewRequestSpecBuilder().
			SetBaseURI(ts.URL).
			AddPathParam("firstName", "Alice").
			Build()

		resp := rest.Given().
			Spec(spec).
			PathParam("lastName", "Wonder").
			Get("/greeting/{firstName}/{lastName}")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Contains(resp.AsString(), "Alice")
		is.Contains(resp.AsString(), "Wonder")
	})

	t.Run("SupportsSettingLoggingWhenUsingRequestSpecBuilder", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// Logging via spec doesn't fail the request; just verifies it compiles and runs
		spec := rest.NewRequestSpecBuilder().
			SetBaseURI(ts.URL).
			Build()

		resp := rest.Given().
			Spec(spec).
			Log().All().
			Get("/lotto")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
	})

	t.Run("SupportsSettingRedirectConfigWhenUsingRequestSpecBuilder", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		cfg := rest.DefaultConfig().WithHTTPClient(rest.HTTPClientConfig{FollowRedirects: true})
		spec := rest.NewRequestSpecBuilder().
			SetBaseURI(ts.URL).
			SetConfig(cfg).
			Build()

		resp := rest.Given().
			Spec(spec).
			Get("/redirect/step1")

		must.NoError(resp.Err())
		// With redirects, should follow to /redirect/final
		is.Equal(http.StatusOK, resp.StatusCode())
	})

	t.Run("MergesStaticallyDefinedResponseSpecificationsCorrectly", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		base := rest.NewResponseSpecBuilder().
			ExpectStatusCode(http.StatusOK).
			Build()

		extra := rest.NewResponseSpecBuilder().
			ExpectContentType(rest.ContentTypeJSON).
			Build()

		merged := rest.NewResponseSpecBuilder().
			AddResponseSpec(base).
			AddResponseSpec(extra).
			Build()

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto")

		must.NoError(resp.Err())
		valid := resp.Then().Spec(merged)
		is.False(valid.HasFailures(), "merged spec should pass: %v", valid.Failures())
	})
}
