package integration_test

// Ported from RequestSpecificationITest.java and RequestSpecMergingITest.java

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/matcher"
	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func TestJavaITest_RequestSpecification_Headers(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("AllowsSpecifyingHeader", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Header("X-Custom-Header", "CustomValue").
			Get("/headers")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal("CustomValue", resp.JsonPath().GetString("X-Custom-Header.0"))
	})

	t.Run("AllowsSpecifyingMultipleHeaders", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Header("X-Header-1", "Val1").
			Header("X-Header-2", "Val2").
			Get("/headers")

		must.NoError(resp.Err())
		is.Equal("Val1", resp.JsonPath().GetString("X-Header-1.0"))
		is.Equal("Val2", resp.JsonPath().GetString("X-Header-2.0"))
	})

	t.Run("AllowsSpecifyingHeadersAsMap", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Headers(map[string]string{
				"X-Map-Header-A": "mapValA",
				"X-Map-Header-B": "mapValB",
			}).
			Get("/headers")

		must.NoError(resp.Err())
		is.Equal("mapValA", resp.JsonPath().GetString("X-Map-Header-A.0"))
		is.Equal("mapValB", resp.JsonPath().GetString("X-Map-Header-B.0"))
	})

	t.Run("AllowsSpecifyingHeadersViaRequestSpec", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		spec := rest.NewRequestSpecBuilder().
			SetBaseURI(ts.URL).
			AddHeader("X-Spec-Header", "SpecValue").
			Build()

		resp := rest.Given().
			Spec(spec).
			Get("/headers")

		must.NoError(resp.Err())
		is.Equal("SpecValue", resp.JsonPath().GetString("X-Spec-Header.0"))
	})

	t.Run("SpecHeaderMergesWithPerRequestHeader", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		spec := rest.NewRequestSpecBuilder().
			SetBaseURI(ts.URL).
			AddHeader("X-Spec-Common", "SpecCommon").
			Build()

		resp := rest.Given().
			Spec(spec).
			Header("X-Per-Request", "PerRequestVal").
			Get("/headers")

		must.NoError(resp.Err())
		is.Equal("SpecCommon", resp.JsonPath().GetString("X-Spec-Common.0"))
		is.Equal("PerRequestVal", resp.JsonPath().GetString("X-Per-Request.0"))
	})
}

func TestJavaITest_RequestSpecification_Cookies(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("AllowsSpecifyingCookieInSpec", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		spec := rest.NewRequestSpecBuilder().
			SetBaseURI(ts.URL).
			AddCookie("spec_cookie", "spec_cookie_value").
			Build()

		resp := rest.Given().
			Spec(spec).
			Get("/cookies")

		must.NoError(resp.Err())
		is.Equal("spec_cookie_value", resp.JsonPath().GetString("spec_cookie"))
	})

	t.Run("AllowsSpecifyingCookieDirect", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Cookie("direct_cookie", "direct_val").
			Get("/cookies")

		must.NoError(resp.Err())
		is.Equal("direct_val", resp.JsonPath().GetString("direct_cookie"))
	})

	t.Run("MultipleCookiesAreConcatenated", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Cookie("cookie_a", "val_a").
			Cookie("cookie_b", "val_b").
			Get("/cookies")

		must.NoError(resp.Err())
		jp := resp.JsonPath()
		is.Equal("val_a", jp.GetString("cookie_a"))
		is.Equal("val_b", jp.GetString("cookie_b"))
	})
}

func TestJavaITest_RequestSpecification_QueryParams(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("AllowsSpecifyingQueryParamsViaSpec", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			QueryParam("firstName", "John").
			QueryParam("lastName", "Doe").
			Get("/greet").
			Then().
			StatusCode(http.StatusOK).
			Body("greeting", matcher.EqualTo("Greetings John Doe")).
			AssertAllNoFail(t)
	})

	t.Run("AllowsSpecifyingQueryParamsViaRequestSpec", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		spec := rest.NewRequestSpecBuilder().
			SetBaseURI(ts.URL).
			AddQueryParam("firstName", "Jane").
			AddQueryParam("lastName", "Smith").
			Build()

		resp := rest.Given().
			Spec(spec).
			Get("/greet")

		must.NoError(resp.Err())
		is.Equal("Greetings Jane Smith", resp.JsonPath().GetString("greeting"))
	})

	t.Run("MultipleParamsAreConcatenated", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			QueryParam("firstName", "A").
			QueryParam("lastName", "B").
			Get("/greet").
			Then().
			Body("greeting", matcher.ContainsString("A")).
			Body("greeting", matcher.ContainsString("B")).
			AssertAllNoFail(t)
	})
}

func TestJavaITest_RequestSpecification_Auth(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("AllowsSpecifyingAuthViaSpec", func(t *testing.T) {
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
		is.Equal("authenticated", resp.JsonPath().GetString("status"))
	})

	t.Run("AllowsSpecifyingBearerTokenViaSpec", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		spec := rest.NewRequestSpecBuilder().
			SetBaseURI(ts.URL).
			SetAuth(&rest.OAuth2Scheme{AccessToken: "secret-token-123"}).
			Build()

		resp := rest.Given().
			Spec(spec).
			Get("/auth/bearer")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal("authenticated", resp.JsonPath().GetString("status"))
	})
}

func TestJavaITest_RequestSpecification_BaseURIAndPath(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("AllowsSpecifyingBaseURIAndBasePath", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		spec := rest.NewRequestSpecBuilder().
			SetBaseURI(ts.URL).
			SetBasePath("/json").
			Build()

		resp := rest.Given().
			Spec(spec).
			Get("/users")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal(float64(2), resp.JsonPath().Get("total").Value())
	})

	t.Run("AllowsOverridingBaseURIPerRequest", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		spec := rest.NewRequestSpecBuilder().
			SetBaseURI("http://should-be-overridden.example.com").
			Build()

		resp := rest.Given().
			Spec(spec).
			BaseURI(ts.URL). // per-request override
			Get("/lotto")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
	})
}

func TestJavaITest_RequestSpecification_SpecMerging(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("SpecificationsAreMergedCorrectly", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		baseSpec := rest.NewRequestSpecBuilder().
			SetBaseURI(ts.URL).
			AddHeader("X-Base", "BaseVal").
			Build()

		resp := rest.Given().
			Spec(baseSpec).
			Header("X-Extra", "ExtraVal").
			Get("/headers")

		must.NoError(resp.Err())
		jp := resp.JsonPath()
		is.Equal("BaseVal", jp.GetString("X-Base.0"))
		is.Equal("ExtraVal", jp.GetString("X-Extra.0"))
	})

	t.Run("SpecWithResponseSpecCombined", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		reqSpec := rest.NewRequestSpecBuilder().
			SetBaseURI(ts.URL).
			AddHeader("X-Req-Spec", "ReqSpecVal").
			Build()

		respSpec := rest.NewResponseSpecBuilder().
			ExpectStatusCode(http.StatusOK).
			ExpectResponseTimeLessThan(5 * time.Second).
			Build()

		valid := rest.Given().
			Spec(reqSpec).
			Get("/lotto").
			Then().
			Spec(respSpec)

		is.False(valid.HasFailures(), "spec validation should pass: %v", valid.Failures())
	})

	t.Run("SpecCanBeReusedAcrossMultipleRequests", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		spec := rest.NewRequestSpecBuilder().
			SetBaseURI(ts.URL).
			AddHeader("X-Reuse", "ReuseValue").
			Build()

		for i := range 3 {
			resp := rest.Given().
				Spec(spec).
				Get("/lotto")
			must.NoError(resp.Err(), "iteration %d", i)
			is.Equal(http.StatusOK, resp.StatusCode(), "iteration %d", i)
		}
	})
}

func TestJavaITest_RequestSpecification_Body(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("AllowsSpecifyingStringBodyForPost", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Body(`{"message":"spec body"}`).
			ContentType(rest.ContentTypeJSON).
			Post("/echo")

		must.NoError(resp.Err())
		is.Equal("spec body", resp.JsonPath().GetString("message"))
	})

	t.Run("BodySpecifiedInRequestSpecIsUsed", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		spec := rest.NewRequestSpecBuilder().
			SetBaseURI(ts.URL).
			SetBody(`{"from":"spec"}`).
			SetContentType(rest.ContentTypeJSON).
			Build()

		resp := rest.Given().
			Spec(spec).
			Post("/echo")

		must.NoError(resp.Err())
		is.Equal("spec", resp.JsonPath().GetString("from"))
	})
}
