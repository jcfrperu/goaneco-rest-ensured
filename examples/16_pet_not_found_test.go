package examples_test

// Feature: Pet Not Found (Negative Test)
// Feature file: examples/features/16_pet_not_found.feature
//
// Demonstrates: asserting an error response (404 Not Found).
// Negative tests are first-class citizens — the library handles non-2xx
// status codes without throwing exceptions; just assert the expected code.
//
// Run: go test ./examples/... -run TestPetstore_16 -v

import (
	"net/http"
	"testing"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
	"github.com/stretchr/testify/require"
)

// Scenario: Request a pet with a non-existent ID
func TestPetstore_16_PetNotFound(t *testing.T) {
	resp := petstore().
		Accept(rest.ContentTypeJSON).
		PathParam("petId", 999_999_999).
		When().
		Get("/pet/{petId}")

	require.NoError(t, resp.Err(), "request must reach the server (HTTP error ≠ connection error)")

	resp.Then().
		StatusCode(http.StatusNotFound).
		AssertAllNoFail(t)
}
