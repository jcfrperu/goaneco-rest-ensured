package examples_test

// Feature: Delete a User Account
// Feature file: examples/features/13_delete_user.feature
//
// Demonstrates: DELETE with a path parameter (username).
//
// Run: go test ./examples/... -run TestPetstore_13 -v

import (
	"net/http"
	"testing"

	"github.com/jcfrperu/goaneco-rest-ensured/examples/models"
	"github.com/stretchr/testify/require"
)

// Scenario: Delete a user by username
func TestPetstore_13_DeleteUser(t *testing.T) {
	username := uniqueName("goaneco_del")

	// Create a user to delete.
	user := models.User{
		Username: username,
		Email:    username + "@example.com",
		Password: "pass",
	}
	created := petstore().BodyObject(user).Post("/user")
	require.NoError(t, created.Err(), "create user must succeed")

	// Delete the user.
	petstore().
		PathParam("username", username).
		When().
		Delete("/user/{username}").
		Then().
		StatusCode(http.StatusOK).
		AssertAllNoFail(t)
}
