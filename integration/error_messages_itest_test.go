package integration_test

// Ported from ErrorMessageITest.java, GivenWhenThenErrorITest.java, GivenWhenThenResponseSpecITest.java

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func TestJavaITest_ErrorMessages_MultipleFailures(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("ShowsAllFailingJSONPathExpectations", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		valid := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			Body("lotto.lottoId", 2).           // fails: actual is 5
			Body("lotto.winning-numbers.0", 99) // fails: actual is 2

		is.True(valid.HasFailures())
		is.Len(valid.Failures(), 2, "both failing expectations should be reported")
	})

	t.Run("ShowsOnlyFailingExpectationsNotSuccessful", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		valid := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			Body("lotto.lottoId", 5).           // passes: actual is 5
			Body("lotto.lottoId", 2).           // fails
			Body("lotto.winning-numbers.0", 99) // fails

		is.True(valid.HasFailures())
		is.Len(valid.Failures(), 2, "only the 2 failing expectations should appear")
	})

	t.Run("MixesBodyAndStatusCodeErrors", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		valid := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			StatusCode(http.StatusCreated). // fails: actual 200
			Body("lotto.lottoId", 2)        // fails: actual 5

		is.True(valid.HasFailures())
		is.Len(valid.Failures(), 2)
	})

	t.Run("AllExpectationsPassMeansNoFailures", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		valid := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			StatusCode(http.StatusOK).
			Body("lotto.lottoId", 5)

		is.False(valid.HasFailures())
		is.Empty(valid.Failures())
	})
}

func TestJavaITest_ErrorMessages_FailureMessageContent(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("FailureMessageContainsExpectedAndActual", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		valid := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			Body("lotto.lottoId", 2) // fails: actual is 5

		is.True(valid.HasFailures())
		is.Len(valid.Failures(), 1)
		failure := valid.Failures()[0]
		is.Contains(failure, "lotto.lottoId")
	})

	t.Run("StatusCodeFailureMessageContainsExpectedAndActual", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		valid := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			StatusCode(http.StatusCreated) // fails: actual is 200

		is.True(valid.HasFailures())
		failure := valid.Failures()[0]
		is.Contains(failure, "201")
		is.Contains(failure, "200")
	})

	t.Run("MissingJSONPathProducesFailure", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		valid := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto").
			Then().
			StatusCode(http.StatusOK).
			Body("nonExistentPath.value", 42)

		is.True(valid.HasFailures())
	})

	t.Run("HeaderFailureMessageContainsName", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		valid := rest.Given().
			BaseURI(ts.URL).
			QueryParam("firstName", "John").
			QueryParam("lastName", "Doe").
			Get("/greet").
			Then().
			Header("X-Nonexistent", "somevalue")

		is.True(valid.HasFailures())
		failure := valid.Failures()[0]
		is.Contains(failure, "X-Nonexistent")
	})

	t.Run("CookieFailureMessageContainsName", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		valid := rest.Given().
			BaseURI(ts.URL).
			QueryParam("firstName", "John").
			QueryParam("lastName", "Doe").
			Get("/greet").
			Then().
			Cookie("mycookie", "jux")

		is.True(valid.HasFailures())
		failure := valid.Failures()[0]
		is.Contains(failure, "mycookie")
	})
}

func TestJavaITest_GivenWhenThenErrors(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("BodyAssertionIncorrectProducesFailure", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		valid := rest.Given().
			BaseURI(ts.URL).
			QueryParam("firstName", "John").
			QueryParam("lastName", "Doe").
			Get("/greet").
			Then().
			StatusCode(http.StatusOK).
			Body("greeting", "Greetings John Doe!")

		is.True(valid.HasFailures())
		failure := valid.Failures()[0]
		is.Contains(failure, "greeting")
	})

	t.Run("StatusAssertionIncorrectProducesFailure", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		valid := rest.Given().
			BaseURI(ts.URL).
			QueryParam("firstName", "John").
			QueryParam("lastName", "Doe").
			Get("/greet").
			Then().
			StatusCode(http.StatusAccepted)

		is.True(valid.HasFailures())
		failure := valid.Failures()[0]
		is.Contains(failure, "202")
		is.Contains(failure, "200")
	})

	t.Run("ContentTypeAssertionIncorrectProducesFailure", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		valid := rest.Given().
			BaseURI(ts.URL).
			QueryParam("firstName", "John").
			QueryParam("lastName", "Doe").
			Get("/greet").
			Then().
			StatusCode(http.StatusOK).
			ContentType(rest.ContentTypeXML)

		is.True(valid.HasFailures())
		failure := valid.Failures()[0]
		is.Contains(failure, "application/xml")
	})

	t.Run("HeaderAssertionFailsWhenHeaderMissing", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		valid := rest.Given().
			BaseURI(ts.URL).
			QueryParam("firstName", "John").
			QueryParam("lastName", "Doe").
			Get("/greet").
			Then().
			Header("Ikk", "jux")

		is.True(valid.HasFailures())
		failure := valid.Failures()[0]
		is.Contains(failure, "Ikk")
	})

	t.Run("CookieAssertionFailsWhenNoCookiesInResponse", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		valid := rest.Given().
			BaseURI(ts.URL).
			QueryParam("firstName", "John").
			QueryParam("lastName", "Doe").
			Get("/greet").
			Then().
			Cookie("mycookie", "jux")

		is.True(valid.HasFailures())
	})
}

func TestJavaITest_GivenWhenThenResponseSpec(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("ResponseSpecWithMultipleWrongExpectationsReportsAllFailures", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		spec := rest.NewResponseSpecBuilder().
			ExpectStatusCode(http.StatusCreated).
			ExpectBody("greeting", "Greetings John Doo"). // wrong value
			Build()

		valid := rest.Given().
			BaseURI(ts.URL).
			QueryParam("firstName", "John").
			QueryParam("lastName", "Doe").
			Get("/greet").
			Then().
			Spec(spec)

		is.True(valid.HasFailures())
		failures := valid.Failures()
		is.GreaterOrEqual(len(failures), 2, "both status code and body failures should be reported")
	})

	t.Run("ResponseSpecPassesWhenExpectationsMatch", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		spec := rest.NewResponseSpecBuilder().
			ExpectStatusCode(http.StatusOK).
			ExpectBody("greeting", "Greetings John Doe").
			Build()

		valid := rest.Given().
			BaseURI(ts.URL).
			QueryParam("firstName", "John").
			QueryParam("lastName", "Doe").
			Get("/greet").
			Then().
			Spec(spec)

		is.False(valid.HasFailures())
	})
}
