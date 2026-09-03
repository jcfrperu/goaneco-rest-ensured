package examples_test

// Feature: Find a Pet by ID
// Feature file: examples/features/03_find_pet_by_id.feature
//
// Demonstrates: path parameters ({petId}) and extracting a value from
// the response to use as input for the next request.
//
// Run: go test ./examples/... -run TestPetstore_03 -v

import (
	"net/http"
	"testing"

	"github.com/jcfrperu/goaneco-rest-ensured/examples/models"
	"github.com/jcfrperu/goaneco-rest-ensured/rest"
	"github.com/stretchr/testify/require"
)

// Scenario: Retrieve a pet using its ID
func TestPetstore_03_FindPetById(t *testing.T) {
	// First, create a pet so we have a known ID to look up.
	newPet := models.Pet{
		Name:      "Rex",
		Status:    "available",
		PhotoUrls: []string{},
	}
	created := petstore().BodyObject(newPet).Post("/pet")
	require.NoError(t, created.Err(), "create pet must succeed")

	petID := created.JsonPath().GetInt64("id")
	require.NotZero(t, petID, "server must return a generated pet ID")

	// Now retrieve the pet by its ID using a path parameter.
	petstore().
		Accept(rest.ContentTypeJSON).
		PathParam("petId", petID).
		Get("/pet/{petId}").
		Then().
		StatusCode(http.StatusOK).
		ContentType(rest.ContentTypeJSON).
		Body("name", "Rex").
		Body("status", "available").
		AssertAllNoFail(t)
}
