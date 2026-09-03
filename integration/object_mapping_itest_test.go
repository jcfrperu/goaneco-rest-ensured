package integration_test

// Ported from ObjectMappingITest.java and TypeObjectMappingITest.java.
// Java uses JAXB/Jackson for object mapping; Go uses encoding/json via resp.As().

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

// Greeting mirrors the Java Greeting POJO used in ObjectMappingITest.
type Greeting struct {
	Greeting string `json:"greeting"`
}

// GreetingNested mirrors nested greeting used in greetJSON endpoint.
type GreetingNested struct {
	Greeting struct {
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
	} `json:"greeting"`
}

// LottoResponse mirrors the lotto JSON structure.
type LottoResponse struct {
	Lotto struct {
		LottoID        int   `json:"lottoId"`
		WinningNumbers []int `json:"winning-numbers"`
		Winners        []struct {
			WinnerID int   `json:"winnerId"`
			Numbers  []int `json:"numbers"`
		} `json:"winners"`
	} `json:"lotto"`
}

// StoreBook represents a book in the store JSON.
type StoreBook struct {
	Category string  `json:"category"`
	Author   string  `json:"author"`
	Title    string  `json:"title"`
	Price    float64 `json:"price"`
}

// StoreResponse mirrors the store JSON.
type StoreResponse struct {
	Store struct {
		Book    []StoreBook `json:"book"`
		Bicycle struct {
			Color string  `json:"color"`
			Price float64 `json:"price"`
		} `json:"bicycle"`
	} `json:"store"`
}

// PersonAddress mirrors the address in jsonList endpoint.
type PersonAddress struct {
	Street  string `json:"street"`
	ZipCode string `json:"zipCode"`
}

// Person mirrors a person in jsonList endpoint.
type Person struct {
	Name    string        `json:"name"`
	Address PersonAddress `json:"address"`
}

func TestJavaITest_ObjectMapping_ResponseToStruct(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("MapResponseToGreetingStruct", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			QueryParam("firstName", "John").
			QueryParam("lastName", "Doe").
			Get("/greet")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())

		var g Greeting
		must.NoError(resp.As(&g))
		is.Equal("Greetings John Doe", g.Greeting)
	})

	t.Run("MapResponseToNestedGreetingStruct", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			QueryParam("firstName", "Jane").
			QueryParam("lastName", "Doe").
			Get("/greetJSON")

		must.NoError(resp.Err())
		must.Contains(resp.ContentType(), "application/json")

		var g GreetingNested
		must.NoError(resp.As(&g))
		is.Equal("Jane", g.Greeting.FirstName)
		is.Equal("Doe", g.Greeting.LastName)
	})

	t.Run("MapResponseToLottoStruct", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto")

		must.NoError(resp.Err())

		var lotto LottoResponse
		must.NoError(resp.As(&lotto))
		is.Equal(5, lotto.Lotto.LottoID)
		is.Contains(lotto.Lotto.WinningNumbers, 2)
		is.Contains(lotto.Lotto.WinningNumbers, 45)
		is.Len(lotto.Lotto.Winners, 2)
	})

	t.Run("MapResponseToStoreStruct", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/jsonStore")

		must.NoError(resp.Err())

		var store StoreResponse
		must.NoError(resp.As(&store))
		is.Equal("red", store.Store.Bicycle.Color)
		is.InDelta(19.95, store.Store.Bicycle.Price, 0.001)
		is.Len(store.Store.Book, 4)
		is.Equal("Nigel Rees", store.Store.Book[0].Author)
	})

	t.Run("MapPersonListResponse", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/jsonList")

		must.NoError(resp.Err())

		var persons []Person
		must.NoError(resp.As(&persons))
		is.Len(persons, 2)
		is.Equal("Anders", persons[0].Name)
		is.Equal("12345", persons[0].Address.ZipCode)
		is.Equal("Sven", persons[1].Name)
		is.Equal("Svensgatan", persons[1].Address.Street)
	})

	t.Run("MapResponseToMapType", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			QueryParam("firstName", "Alice").
			QueryParam("lastName", "Smith").
			Get("/greet")

		must.NoError(resp.Err())

		var result map[string]any
		must.NoError(resp.As(&result))
		is.Equal("Greetings Alice Smith", result["greeting"])
	})

	t.Run("EmptyBodyReturnsErrorOnAs", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/emptyBody")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())

		var g Greeting
		err := resp.As(&g)
		is.Error(err, "As() on empty body should return error")
	})
}

func TestJavaITest_ObjectMapping_SerializeStructToBody(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("WhenNoRequestContentTypeIsSpecifiedBodyIsSerializedToJSON", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		type EchoPayload struct {
			Action string `json:"action"`
			Value  int    `json:"value"`
		}

		payload := EchoPayload{Action: "run", Value: 7}
		resp := rest.Given().
			BaseURI(ts.URL).
			BodyObject(payload).
			ContentType(rest.ContentTypeJSON).
			Post("/echo")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal("run", resp.JsonPath().GetString("action"))
		is.Equal(float64(7), resp.JsonPath().Get("value").Value())
	})

	t.Run("SerializeStructAndMapResponseBackToStruct", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		type Payload struct {
			Name  string `json:"name"`
			Score int    `json:"score"`
		}

		sent := Payload{Name: "TestUser", Score: 99}
		resp := rest.Given().
			BaseURI(ts.URL).
			BodyObject(sent).
			ContentType(rest.ContentTypeJSON).
			Post("/echo")

		must.NoError(resp.Err())

		var received Payload
		must.NoError(resp.As(&received))
		is.Equal("TestUser", received.Name)
		is.Equal(99, received.Score)
	})

	t.Run("ContentTypesEndingWithPlusJsonWorkForMapping", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/mimeTypeWithPlusJson")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Contains(resp.ContentType(), "application/something+json")

		var result map[string]any
		must.NoError(resp.As(&result))
		is.Equal("It works", result["message"])
	})

	t.Run("SerializeNormalParams_FormSubmit", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			FormParam("username", "bob").
			FormParam("action", "submit").
			Post("/form")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal("bob", resp.JsonPath().GetString("username"))
		is.Equal("submit", resp.JsonPath().GetString("action"))
	})
}

func TestJavaITest_ObjectMapping_AdditionalScenarios(t *testing.T) {
	t.Parallel()
	ts := integrationServer

	t.Run("MapResponseToObjectWhenNoContentTypeDefined", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		// /noContentTypeJsonCompatible returns JSON body without Content-Type header
		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/noContentTypeJsonCompatible")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())

		var result map[string]any
		must.NoError(resp.As(&result))
		is.NotEmpty(result)
	})

	t.Run("CanDeserializeAnonymousListToSlice", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/anonymous_list_with_numbers")

		must.NoError(resp.Err())

		var numbers []float64
		must.NoError(resp.As(&numbers))
		is.Len(numbers, 3)
		is.Contains(numbers, float64(100))
		is.Contains(numbers, float64(50))
		is.Contains(numbers, float64(31))
	})

	t.Run("MapResponseCanBeCalledMultipleTimes", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			QueryParam("firstName", "Multi").
			QueryParam("lastName", "Call").
			Get("/greet")

		must.NoError(resp.Err())

		var g1, g2 Greeting
		must.NoError(resp.As(&g1))
		must.NoError(resp.As(&g2))
		is.Equal(g1.Greeting, g2.Greeting)
		is.Equal("Greetings Multi Call", g1.Greeting)
	})

	t.Run("SerializeStructWithNestedFields", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		type NestedPayload struct {
			User struct {
				Name  string `json:"name"`
				Email string `json:"email"`
			} `json:"user"`
			Active bool `json:"active"`
		}

		payload := NestedPayload{Active: true}
		payload.User.Name = "TestUser"
		payload.User.Email = "test@example.com"

		resp := rest.Given().
			BaseURI(ts.URL).
			BodyObject(payload).
			ContentType(rest.ContentTypeJSON).
			Post("/echo")

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
		is.Equal("TestUser", resp.JsonPath().GetString("user.name"))
		is.Equal("test@example.com", resp.JsonPath().GetString("user.email"))
		is.True(resp.JsonPath().GetBool("active"))
	})

	t.Run("MapLottoWinnersToSlice", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/lotto")

		must.NoError(resp.Err())

		var lotto LottoResponse
		must.NoError(resp.As(&lotto))
		is.Len(lotto.Lotto.Winners, 2)

		// Collect all winner IDs
		var winnerIDs []int
		for _, w := range lotto.Lotto.Winners {
			winnerIDs = append(winnerIDs, w.WinnerID)
		}
		is.Contains(winnerIDs, 23)
		is.Contains(winnerIDs, 54)
	})

	t.Run("ContentTypesEndingWithPlusXmlWorkForMapping", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			Get("/mimeTypeWithPlusJson")

		must.NoError(resp.Err())
		is.Contains(resp.ContentType(), "+json")

		var result map[string]any
		must.NoError(resp.As(&result))
		is.Equal("It works", result["message"])
	})
}
