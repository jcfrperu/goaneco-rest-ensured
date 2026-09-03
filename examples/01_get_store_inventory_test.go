package examples_test

// Feature: Get Store Inventory
// Feature file: examples/features/01_get_store_inventory.feature
//
// Demonstrates: simplest possible GET request — no parameters, no body.
// The inventory endpoint returns a JSON object mapping pet status to count.
//
// Run: go test ./examples/... -run TestPetstore_01 -v

import (
	"net/http"
	"testing"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

// Scenario: Retrieve the pet inventory
func TestPetstore_01_GetStoreInventory(t *testing.T) {
	petstore().
		Accept(rest.ContentTypeJSON).
		When().
		Get("/store/inventory").
		Then().
		StatusCode(http.StatusOK).
		ContentType(rest.ContentTypeJSON).
		AssertAllNoFail(t)
}
