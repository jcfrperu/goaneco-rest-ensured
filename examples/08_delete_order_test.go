package examples_test

// Feature: Delete a Store Order
// Feature file: examples/features/08_delete_order.feature
//
// Demonstrates: a two-step test — POST to create, DELETE to remove —
// showing how test data is self-contained and cleaned up within the same test.
//
// Run: go test ./examples/... -run TestPetstore_08 -v

import (
	"net/http"
	"testing"

	"github.com/jcfrperu/goaneco-rest-ensured/examples/models"
	"github.com/stretchr/testify/require"
)

// Scenario: Delete a placed order
func TestPetstore_08_DeleteOrder(t *testing.T) {
	// Place an order to have something to delete.
	order := models.Order{PetID: 1, Quantity: 1, Status: "placed"}
	placed := petstore().BodyObject(order).Post("/store/order")
	require.NoError(t, placed.Err(), "place order must succeed")

	orderID := placed.JsonPath().GetInt64("id")
	require.NotZero(t, orderID, "server must return a generated order ID")

	// Delete the order.
	petstore().
		PathParam("orderId", orderID).
		When().
		Delete("/store/order/{orderId}").
		Then().
		StatusCode(http.StatusOK).
		AssertAllNoFail(t)
}
