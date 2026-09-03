package examples_test

// Feature: Full Pet Lifecycle
// Feature file: examples/features/19_full_pet_lifecycle.feature
//
// Demonstrates: a complete CRUD workflow in a single test — create, read,
// update, delete — and verify the resource is truly gone at the end.
// This is the most representative example of how the library is used in practice.
//
// Run: go test ./examples/... -run TestPetstore_19 -v

import (
	"net/http"
	"testing"

	"github.com/jcfrperu/goaneco-rest-ensured/examples/models"
	"github.com/jcfrperu/goaneco-rest-ensured/rest"
	"github.com/stretchr/testify/require"
)

// Scenario: Create, retrieve, update, and delete a pet
func TestPetstore_19_FullPetLifecycle(t *testing.T) {
	// ── Step 1: CREATE ────────────────────────────────────────────────────────
	newPet := models.Pet{
		Name:      "Lifecycle Dog",
		Status:    "available",
		PhotoUrls: []string{"https://example.com/dog.jpg"},
	}

	created := petstore().
		Accept(rest.ContentTypeJSON).
		BodyObject(newPet).
		Post("/pet")

	require.NoError(t, created.Err())

	created.Then().
		StatusCode(http.StatusOK).
		Body("name", "Lifecycle Dog").
		Body("status", "available").
		AssertAllNoFail(t)

	petID := created.JsonPath().GetInt64("id")
	require.NotZero(t, petID, "server must assign an ID to the new pet")

	// ── Step 2: READ ──────────────────────────────────────────────────────────
	petstore().
		Accept(rest.ContentTypeJSON).
		PathParam("petId", petID).
		Get("/pet/{petId}").
		Then().
		StatusCode(http.StatusOK).
		Body("name", "Lifecycle Dog").
		Body("status", "available").
		AssertAllNoFail(t)

	// ── Step 3: UPDATE ────────────────────────────────────────────────────────
	updatedPet := models.Pet{
		ID:        petID,
		Name:      "Lifecycle Dog",
		Status:    "pending", // ← new status
		PhotoUrls: []string{},
	}

	petstore().
		Accept(rest.ContentTypeJSON).
		BodyObject(updatedPet).
		Put("/pet").
		Then().
		StatusCode(http.StatusOK).
		Body("status", "pending").
		AssertAllNoFail(t)

	// ── Step 4: DELETE ────────────────────────────────────────────────────────
	petstore().
		Header("api_key", "special-key").
		PathParam("petId", petID).
		Delete("/pet/{petId}").
		Then().
		StatusCode(http.StatusOK).
		AssertAllNoFail(t)

	// ── Step 5: VERIFY DELETION ───────────────────────────────────────────────
	petstore().
		Accept(rest.ContentTypeJSON).
		PathParam("petId", petID).
		Get("/pet/{petId}").
		Then().
		StatusCode(http.StatusNotFound).
		AssertAllNoFail(t)
}
