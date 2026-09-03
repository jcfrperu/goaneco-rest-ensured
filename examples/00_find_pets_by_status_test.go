package examples_test

// Feature: Find Pets by Status
// Feature file: examples/features/00_find_pets_by_status.feature
//
// This example demonstrates the core Given-When-Then DSL:
//   - Fluent request building with Accept and QueryParam
//   - Status code and content-type assertions
//   - JSON body path assertions on a root-level array response
//
// Run: go test ./examples/... -run TestPetstore_00 -v

import (
	"net/http"
	"testing"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
	"github.com/stretchr/testify/require"
)

// Scenario: Find all available pets
func TestPetstore_00_FindPetsByStatus_Available(t *testing.T) {
	resp := petstore().
		Accept(rest.ContentTypeJSON).
		QueryParam("status", "available").
		When().
		Get("/pet/findByStatus")

	require.NoError(t, resp.Err(), "request must reach the server")

	resp.Then().
		StatusCode(http.StatusOK).
		ContentType(rest.ContentTypeJSON).
		Body("0.status", "available"). // first pet in the array has status "available"
		AssertAllNoFail(t)
}

// Scenario: Find pets pending adoption
func TestPetstore_00_FindPetsByStatus_Pending(t *testing.T) {
	resp := petstore().
		Accept(rest.ContentTypeJSON).
		QueryParam("status", "pending").
		When().
		Get("/pet/findByStatus")

	require.NoError(t, resp.Err(), "request must reach the server")

	resp.Then().
		StatusCode(http.StatusOK).
		ContentType(rest.ContentTypeJSON).
		AssertAllNoFail(t)
}

// Scenario: Find sold pets
func TestPetstore_00_FindPetsByStatus_Sold(t *testing.T) {
	resp := petstore().
		Accept(rest.ContentTypeJSON).
		QueryParam("status", "sold").
		When().
		Get("/pet/findByStatus")

	require.NoError(t, resp.Err(), "request must reach the server")

	resp.Then().
		StatusCode(http.StatusOK).
		ContentType(rest.ContentTypeJSON).
		AssertAllNoFail(t)
}
