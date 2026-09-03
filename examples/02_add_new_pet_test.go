package examples_test

// Feature: Add a New Pet
// Feature file: examples/features/02_add_new_pet.feature
//
// Demonstrates: POST with a JSON body built from a struct (BodyObject).
// The library serializes the struct, sets Content-Type automatically,
// and the response is validated field by field via JSON path assertions.
//
// Run: go test ./examples/... -run TestPetstore_02 -v

import (
	"net/http"
	"testing"

	"github.com/jcfrperu/goaneco-rest-ensured/examples/models"
	"github.com/jcfrperu/goaneco-rest-ensured/rest"
	"github.com/stretchr/testify/require"
)

// Scenario: Successfully add a new available pet
func TestPetstore_02_AddNewPet(t *testing.T) {
	newPet := models.Pet{
		Name:      "Buddy",
		Status:    "available",
		PhotoUrls: []string{"https://example.com/buddy.jpg"},
	}

	resp := petstore().
		Accept(rest.ContentTypeJSON).
		BodyObject(newPet). // serializes the struct to JSON; sets Content-Type automatically
		When().
		Post("/pet")

	require.NoError(t, resp.Err(), "request must reach the server")

	resp.Then().
		StatusCode(http.StatusOK).
		ContentType(rest.ContentTypeJSON).
		Body("name", "Buddy").
		Body("status", "available").
		AssertAllNoFail(t)
}
