package examples_test

// Feature: User Login
// Feature file: examples/features/10_user_login.feature
//
// Demonstrates: GET with multiple query parameters and extracting
// a plain-text response body (the session token) for further verification.
//
// Run: go test ./examples/... -run TestPetstore_10 -v

import (
	"net/http"
	"testing"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
	"github.com/stretchr/testify/require"
)

// Scenario: Log in with valid credentials
func TestPetstore_10_UserLogin(t *testing.T) {
	resp := petstore().
		Accept(rest.ContentTypeJSON).
		QueryParam("username", "user1").
		QueryParam("password", "XXXXXXXXXXX").
		When().
		Get("/user/login")

	require.NoError(t, resp.Err(), "request must reach the server")

	resp.Then().
		StatusCode(http.StatusOK).
		AssertAllNoFail(t)

	// The response body is a string containing the session token.
	body := resp.AsString()
	require.NotEmpty(t, body, "response must contain a session token")
}
