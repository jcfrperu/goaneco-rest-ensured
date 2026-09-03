Feature: Pet Not Found (Negative Test)
  As a developer
  I want to verify that requesting a non-existent pet returns a proper error
  So that the API handles invalid input gracefully

  Background:
    Given the Petstore API at "https://petstore.swagger.io/v2"

  Scenario: Request a pet with a non-existent ID
    When I send a GET request to "/pet/999999999"
    Then the response status code should be 404
    And the response body should indicate the pet was not found
