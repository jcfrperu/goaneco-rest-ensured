package examples_test

// Feature: Place a Store Order
// Feature file: examples/features/06_place_order.feature
//
// Demonstrates: POST with a struct body and chained response assertions.
// Shows how to extract a value (orderId) from the response for later use.
//
// Run: go test ./examples/... -run TestPetstore_06 -v

import (
	"net/http"
	"testing"

	"github.com/jcfrperu/goaneco-rest-ensured/examples/models"
	"github.com/jcfrperu/goaneco-rest-ensured/rest"
	"github.com/stretchr/testify/require"
)

// Scenario: Place a new order for a pet
func TestPetstore_06_PlaceOrder(t *testing.T) {
	order := models.Order{
		PetID:    1,
		Quantity: 1,
		Status:   "placed",
		Complete: false,
	}

	resp := petstore().
		Accept(rest.ContentTypeJSON).
		BodyObject(order).
		When().
		Post("/store/order")

	require.NoError(t, resp.Err(), "request must reach the server")

	resp.Then().
		StatusCode(http.StatusOK).
		ContentType(rest.ContentTypeJSON).
		Body("status", "placed").
		AssertAllNoFail(t)

	orderID := resp.JsonPath().GetInt64("id")
	require.NotZero(t, orderID, "server must return a generated order ID")
}
