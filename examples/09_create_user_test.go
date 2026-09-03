package examples_test

// Feature: Create a User Account
// Feature file: examples/features/09_create_user.feature
//
// Demonstrates: POST with a User model and a uniquely-generated username
// so that repeated test runs do not collide with existing records.
//
// Run: go test ./examples/... -run TestPetstore_09 -v

import (
	"net/http"
	"testing"

	"github.com/jcfrperu/goaneco-rest-ensured/examples/models"
	"github.com/stretchr/testify/require"
)

// Scenario: Create a new user with all details
func TestPetstore_09_CreateUser(t *testing.T) {
	newUser := models.User{
		Username:  uniqueName("goaneco_user"),
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		Password:  "s3cret!",
		Phone:     "555-0100",
	}

	resp := petstore().
		BodyObject(newUser).
		When().
		Post("/user")

	require.NoError(t, resp.Err(), "request must reach the server")

	resp.Then().
		StatusCode(http.StatusOK).
		AssertAllNoFail(t)
}
