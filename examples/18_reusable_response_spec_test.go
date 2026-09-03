package examples_test

// Feature: Reusable Response Specification
// Feature file: examples/features/18_reusable_response_spec.feature
//
// Demonstrates: ResponseSpecBuilder — encode the set of expected response
// properties once and reuse it across multiple assertions.
// Great for enforcing a consistent API contract across all endpoints.
//
// Run: go test ./examples/... -run TestPetstore_18 -v

import (
	"net/http"
	"testing"

	"github.com/jcfrperu/goaneco-rest-ensured/examples/models"
	"github.com/jcfrperu/goaneco-rest-ensured/rest"
	"github.com/stretchr/testify/require"
)

// Scenario: Apply the same response rules to two different pet requests
func TestPetstore_18_ReusableResponseSpec(t *testing.T) {
	// Define the response contract once: status 200 + JSON content type.
	petResponseSpec := rest.NewResponseSpecBuilder().
		ExpectStatusCode(http.StatusOK).
		ExpectContentType(rest.ContentTypeJSON).
		Build()

	// First pet.
	pet1 := models.Pet{Name: "Spec Pet One", Status: "available", PhotoUrls: []string{}}
	resp1 := petstore().BodyObject(pet1).Post("/pet")
	require.NoError(t, resp1.Err(), "first request must reach the server")

	resp1.Then().
		Spec(petResponseSpec).
		Body("name", "Spec Pet One").
		AssertAllNoFail(t)

	// Second pet — same response spec, different data.
	pet2 := models.Pet{Name: "Spec Pet Two", Status: "pending", PhotoUrls: []string{}}
	resp2 := petstore().BodyObject(pet2).Post("/pet")
	require.NoError(t, resp2.Err(), "second request must reach the server")

	resp2.Then().
		Spec(petResponseSpec).
		Body("name", "Spec Pet Two").
		AssertAllNoFail(t)
}
