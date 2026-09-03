package examples_test

// Feature: Reusable Request Specification
// Feature file: examples/features/17_reusable_request_spec.feature
//
// Demonstrates: RequestSpecBuilder — define shared headers, base URL and
// configuration once and apply them to multiple requests. This avoids
// copy-pasting common setup across every test in a suite.
//
// Run: go test ./examples/... -run TestPetstore_17 -v

import (
	"net/http"
	"testing"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
	"github.com/stretchr/testify/require"
)

// Scenario: Two requests share the same base specification
func TestPetstore_17_ReusableRequestSpec(t *testing.T) {
	// Build the spec once — it captures base URL, port, and Accept header.
	sharedSpec := rest.NewRequestSpecBuilder().
		SetBaseURI(petstoreURL).
		AddHeader("Accept", string(rest.ContentTypeJSON)).
		Build()

	// First request: find available pets.
	resp1 := rest.Given().
		Spec(sharedSpec).
		Port(0). // override global default port
		QueryParam("status", "available").
		Get("/pet/findByStatus")

	require.NoError(t, resp1.Err(), "first request must reach the server")

	resp1.Then().
		StatusCode(http.StatusOK).
		ContentType(rest.ContentTypeJSON).
		AssertAllNoFail(t)

	// Second request: reuse the same spec to find sold pets.
	resp2 := rest.Given().
		Spec(sharedSpec).
		Port(0).
		QueryParam("status", "sold").
		Get("/pet/findByStatus")

	require.NoError(t, resp2.Err(), "second request must reach the server")

	resp2.Then().
		StatusCode(http.StatusOK).
		ContentType(rest.ContentTypeJSON).
		AssertAllNoFail(t)
}
