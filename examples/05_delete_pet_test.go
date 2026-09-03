package examples_test

// Feature: Delete a Pet
// Feature file: examples/features/05_delete_pet.feature
//
// Demonstrates: DELETE request with a path parameter and a custom request header.
//
// Run: go test ./examples/... -run TestPetstore_05 -v

import (
	"net/http"
	"testing"

	"github.com/jcfrperu/goaneco-rest-ensured/examples/models"
	"github.com/stretchr/testify/require"
)

// Scenario: Delete a pet using its ID
func TestPetstore_05_DeletePet(t *testing.T) {
	// Create a pet so we have something to delete.
	newPet := models.Pet{
		Name:      "ToDelete",
		Status:    "available",
		PhotoUrls: []string{},
	}
	created := petstore().BodyObject(newPet).Post("/pet")
	require.NoError(t, created.Err(), "create pet must succeed")

	petID := created.JsonPath().GetInt64("id")
	require.NotZero(t, petID, "server must return a generated pet ID")

	// Delete the pet. The api_key header is required by the Petstore spec.
	petstore().
		Header("api_key", "special-key").
		PathParam("petId", petID).
		When().
		Delete("/pet/{petId}").
		Then().
		StatusCode(http.StatusOK).
		AssertAllNoFail(t)
}
