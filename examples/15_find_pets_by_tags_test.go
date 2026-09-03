package examples_test

// Feature: Find Pets by Tags
// Feature file: examples/features/15_find_pets_by_tags.feature
//
// Demonstrates: passing multiple values for the same query parameter.
//
// Run: go test ./examples/... -run TestPetstore_15 -v

import (
	"net/http"
	"testing"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
	"github.com/stretchr/testify/require"
)

// Scenario: Search for pets with multiple tags
func TestPetstore_15_FindPetsByTags(t *testing.T) {
	resp := petstore().
		Accept(rest.ContentTypeJSON).
		QueryParam("tags", "friendly", "vaccinated"). // multiple values for one parameter
		When().
		Get("/pet/findByTags")

	require.NoError(t, resp.Err(), "request must reach the server")

	resp.Then().
		StatusCode(http.StatusOK).
		ContentType(rest.ContentTypeJSON).
		AssertAllNoFail(t)
}
