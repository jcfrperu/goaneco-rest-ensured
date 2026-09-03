package examples_test

// Feature: Get User by Username
// Feature file: examples/features/11_get_user_by_name.feature
//
// Demonstrates: creating a resource and immediately reading it back —
// a common test pattern that keeps data self-contained within the test.
//
// Run: go test ./examples/... -run TestPetstore_11 -v

import (
	"net/http"
	"testing"

	"github.com/jcfrperu/goaneco-rest-ensured/examples/models"
	"github.com/jcfrperu/goaneco-rest-ensured/rest"
	"github.com/stretchr/testify/require"
)

// Scenario: Retrieve a user using their username
func TestPetstore_11_GetUserByName(t *testing.T) {
	username := uniqueName("goaneco_get")
	email := username + "@example.com"

	// Create the user first.
	user := models.User{
		Username:  username,
		FirstName: "Jane",
		LastName:  "Smith",
		Email:     email,
		Password:  "password",
	}
	created := petstore().BodyObject(user).Post("/user")
	require.NoError(t, created.Err(), "create user must succeed")

	// Retrieve the user by username.
	petstore().
		Accept(rest.ContentTypeJSON).
		PathParam("username", username).
		Get("/user/{username}").
		Then().
		StatusCode(http.StatusOK).
		ContentType(rest.ContentTypeJSON).
		Body("username", username).
		Body("email", email).
		AssertAllNoFail(t)
}
