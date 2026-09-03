package examples_test

// Feature: Get an Order by ID
// Feature file: examples/features/07_get_order_by_id.feature
//
// Demonstrates: using Extract().PathInt() to pull a value from one response
// and feed it as a PathParam into the next request.
//
// Run: go test ./examples/... -run TestPetstore_07 -v

import (
	"net/http"
	"testing"

	"github.com/jcfrperu/goaneco-rest-ensured/examples/models"
	"github.com/jcfrperu/goaneco-rest-ensured/rest"
	"github.com/stretchr/testify/require"
)

// Scenario: Retrieve an order using its ID
func TestPetstore_07_GetOrderById(t *testing.T) {
	// Place an order to ensure the ID exists.
	order := models.Order{PetID: 1, Quantity: 2, Status: "placed"}
	placed := petstore().BodyObject(order).Post("/store/order")
	require.NoError(t, placed.Err(), "place order must succeed")

	orderID := placed.JsonPath().GetInt64("id")
	require.NotZero(t, orderID, "server must return a generated order ID")

	// Retrieve the order by its ID.
	petstore().
		Accept(rest.ContentTypeJSON).
		PathParam("orderId", orderID).
		Get("/store/order/{orderId}").
		Then().
		StatusCode(http.StatusOK).
		ContentType(rest.ContentTypeJSON).
		Body("status", "placed").
		AssertAllNoFail(t)
}
