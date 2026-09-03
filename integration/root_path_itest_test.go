package integration_test

// Ported from RootPathITest.java.
// Java uses rootPath() / appendRoot() / detachRoot() on ValidatableResponse.
// Go maps to ValidatableResponse.RootPath() / AppendRootPath() / NoRootPath().

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/matcher"
	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func TestJavaITest_RootPath_BasicUsage(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("SpecifyingRootPathAddsItForEachSubsequentBodyExpectation", func(t *testing.T) {
		t.Parallel()

		// lotto.lottoId and lotto.winning-numbers under root "lotto"
		rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			StatusCode(http.StatusOK).
			RootPath("lotto").
			Body("lottoId", 5).
			Body("winning-numbers", matcher.HasItemValue(2)).
			AssertAllNoFail(t)
	})

	t.Run("RootPathScopesAllBodyAssertions", func(t *testing.T) {
		t.Parallel()

		rest.Given().
			BaseURI(ts.URL).
			QueryParam("firstName", "John").
			QueryParam("lastName", "Doe").
			Get("/greetJSON").
			Then().
			StatusCode(http.StatusOK).
			RootPath("greeting").
			Body("firstName", "John").
			Body("lastName", "Doe").
			AssertAllNoFail(t)
	})

	t.Run("SpecifyingRootPathWithNestedObject", func(t *testing.T) {
		t.Parallel()

		rest.Given().
			BaseURI(ts.URL).
			Get("/jsonStore").
			Then().
			StatusCode(http.StatusOK).
			RootPath("store.bicycle").
			Body("color", "red").
			Body("price", matcher.GreaterThan(19.0)).
			AssertAllNoFail(t)
	})

	t.Run("WhenNotSpecifyingRootPathThenDefaultRootPathIsEmpty", func(t *testing.T) {
		t.Parallel()

		// Without rootPath, full JSON paths must be used.
		rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			Body("lotto.lottoId", 5).
			Body("lotto.winning-numbers", matcher.HasItemValue(2)).
			AssertAllNoFail(t)
	})

	t.Run("SpecifyingEmptyRootPathResetsToDefault", func(t *testing.T) {
		t.Parallel()

		// Set a rootPath, then reset it to "" — full path must be used.
		rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			RootPath("lotto").
			Body("lottoId", 5).
			NoRootPath().
			Body("lotto.lottoId", 5).
			AssertAllNoFail(t)
	})

	t.Run("NoRootPathResetsRootPathToEmptyString", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		valid := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			RootPath("lotto")

		valid.NoRootPath()

		// After NoRootPath, full path should be required.
		valid.Body("lotto.lottoId", 5)
		is.False(valid.HasFailures())
	})
}

func TestJavaITest_RootPath_AppendRootPath(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("AppendingRootPathWithoutArgumentsWorks", func(t *testing.T) {
		t.Parallel()

		// Start at "lotto", append "winners" → assertions under "lotto.winners"
		rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			RootPath("lotto").
			Body("lottoId", 5).
			AppendRootPath("winners").
			Body("0.winnerId", matcher.Anything()).
			AssertAllNoFail(t)
	})

	t.Run("CanAppendRootPathToEmptyRootPath", func(t *testing.T) {
		t.Parallel()

		// Starting from "" and appending "lotto" makes root "lotto".
		rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			AppendRootPath("lotto").
			Body("lottoId", 5).
			AssertAllNoFail(t)
	})

	t.Run("AppendingMultipleRootPathSegmentsWorks", func(t *testing.T) {
		t.Parallel()

		rest.Given().
			BaseURI(ts.URL).
			Get("/jsonStore").
			Then().
			RootPath("store").
			AppendRootPath("bicycle").
			Body("color", "red").
			AssertAllNoFail(t)
	})

	t.Run("AppendRootPathAndThenResetWorks", func(t *testing.T) {
		t.Parallel()

		rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			RootPath("lotto").
			AppendRootPath("winners").
			Body("0.winnerId", matcher.Anything()).
			NoRootPath().
			Body("lotto.lottoId", 5).
			AssertAllNoFail(t)
	})
}

func TestJavaITest_RootPath_GlobalRootPath(t *testing.T) {
	// NOT parallel — mutates global root path.
	ts := integrationServer

	t.Run("GlobalRootPathIsAppliedToAllRequests", func(t *testing.T) {
		rest.RootPath("lotto")
		t.Cleanup(func() {
			rest.RootPath("")
		})

		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto")

		must.NoError(resp.Err())
		must.Equal(http.StatusOK, resp.StatusCode())

		// rootPath "lotto" means Body("lottoId") == 5, not Body("lotto.lottoId")
		valid := resp.Then().Body("lottoId", 5)
		must.False(valid.HasFailures())
	})

	t.Run("GlobalRootPathCanBeReset", func(t *testing.T) {
		rest.RootPath("something")
		t.Cleanup(func() {
			rest.RootPath("")
		})
		rest.RootPath("") // reset immediately

		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto")
		must.NoError(resp.Err())

		valid := resp.Then().Body("lotto.lottoId", 5)
		must.False(valid.HasFailures())
	})
}

func TestJavaITest_RootPath_WithResponseSpec(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("RootPathInResponseSpecIsApplied", func(t *testing.T) {
		t.Parallel()

		spec := rest.NewResponseSpecBuilder().
			ExpectStatusCode(http.StatusOK).
			RootPath("lotto").
			ExpectBody("lottoId", 5).
			Build()

		rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			Spec(spec).
			AssertAllNoFail(t)
	})

	t.Run("AppendRootPathInResponseSpecWorks", func(t *testing.T) {
		t.Parallel()

		spec := rest.NewResponseSpecBuilder().
			ExpectStatusCode(http.StatusOK).
			RootPath("store").
			AppendRootPath("bicycle").
			ExpectBody("color", "red").
			Build()

		rest.Given().
			BaseURI(ts.URL).
			Get("/jsonStore").
			Then().
			Spec(spec).
			AssertAllNoFail(t)
	})

	t.Run("ResponseSpecBuilderRootPathConfigurationRespectsLottoStructure", func(t *testing.T) {
		t.Parallel()

		spec := rest.NewResponseSpecBuilder().
			RootPath("lotto.winners").
			ExpectBody("0.winnerId", matcher.Anything()).
			Build()

		rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			Spec(spec).
			AssertAllNoFail(t)
	})
}

func TestJavaITest_RootPath_BodyWithoutPath(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("CanUseBodyExpectationWithoutPathWhenRootPathIsSet", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		// With a rootPath set, Body("lottoId", ...) works.
		valid := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			RootPath("lotto").
			Body("lottoId", 5)

		is.False(valid.HasFailures())
	})

	t.Run("BodyAssertionWithoutRootPathRequiresFullPath", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		// Without rootPath, only the full path matches.
		valid := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			Body("lotto.lottoId", 5)

		is.False(valid.HasFailures())
	})

	t.Run("RootPathWithArrayIndexWorks", func(t *testing.T) {
		t.Parallel()

		rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			RootPath("lotto.winners").
			Body("0.winnerId", matcher.Anything()).
			Body("1.winnerId", matcher.Anything()).
			AssertAllNoFail(t)
	})

	t.Run("RootPathScopedBodyAssertionFailsForWrongValue", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		valid := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			RootPath("lotto").
			Body("lottoId", 999)

		is.True(valid.HasFailures(), "wrong lottoId should record a failure")
	})

	t.Run("MultipleRootPathAssertionsOnSameResponse", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		valid := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			RootPath("lotto").
			Body("lottoId", 5).
			Body("winning-numbers", matcher.HasItemValue(45)).
			NoRootPath().
			Body("lotto.winners", matcher.Anything())

		is.False(valid.HasFailures())
	})
}
