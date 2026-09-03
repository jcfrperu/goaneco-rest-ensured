package examples_test

// Feature: Update a User Account
// Feature file: examples/features/12_update_user.feature
//
// Demonstrates: PUT with a path parameter (username) and a JSON body.
//
// Run: go test ./examples/... -run TestPetstore_12 -v

import (
	"net/http"
	"testing"

	"github.com/jcfrperu/goaneco-rest-ensured/examples/models"
	"github.com/jcfrperu/goaneco-rest-ensured/rest"
	"github.com/stretchr/testify/require"
)

// Scenario: Update a user's email and phone
func TestPetstore_12_UpdateUser(t *testing.T) {
	username := uniqueName("goaneco_upd")

	// Create the user first.
	user := models.User{
		Username: username,
		Email:    "original@example.com",
		Password: "pass",
	}
	created := petstore().BodyObject(user).Post("/user")
	require.NoError(t, created.Err(), "create user must succeed")

	// Update the user's email and phone.
	updatedUser := models.User{
		Username: username,
		Email:    "updated@example.com",
		Phone:    "555-9999",
		Password: "pass",
	}

	petstore().
		Accept(rest.ContentTypeJSON).
		BodyObject(updatedUser).
		PathParam("username", username).
		When().
		Put("/user/{username}").
		Then().
		StatusCode(http.StatusOK).
		AssertAllNoFail(t)
}
