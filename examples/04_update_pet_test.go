package examples_test

// Feature: Update an Existing Pet
// Feature file: examples/features/04_update_pet.feature
//
// Demonstrates: PUT with a JSON body — the same BodyObject pattern as POST,
// but the struct must carry the existing ID so the server updates instead of creating.
//
// Run: go test ./examples/... -run TestPetstore_04 -v

import (
	"net/http"
	"testing"

	"github.com/jcfrperu/goaneco-rest-ensured/examples/models"
	"github.com/jcfrperu/goaneco-rest-ensured/rest"
	"github.com/stretchr/testify/require"
)

// Scenario: Update a pet's name and status
func TestPetstore_04_UpdatePet(t *testing.T) {
	// Create the pet to update.
	originalPet := models.Pet{
		Name:      "Fluffy",
		Status:    "available",
		PhotoUrls: []string{},
	}
	created := petstore().BodyObject(originalPet).Post("/pet")
	require.NoError(t, created.Err(), "create pet must succeed")

	petID := created.JsonPath().GetInt64("id")
	require.NotZero(t, petID, "server must return a generated pet ID")

	// Update the pet: keep the same ID, change name and status.
	updatedPet := models.Pet{
		ID:        petID,
		Name:      "Fluffy (updated)",
		Status:    "pending",
		PhotoUrls: []string{},
	}

	petstore().
		Accept(rest.ContentTypeJSON).
		BodyObject(updatedPet).
		When().
		Put("/pet").
		Then().
		StatusCode(http.StatusOK).
		ContentType(rest.ContentTypeJSON).
		Body("name", "Fluffy (updated)").
		Body("status", "pending").
		AssertAllNoFail(t)
}
