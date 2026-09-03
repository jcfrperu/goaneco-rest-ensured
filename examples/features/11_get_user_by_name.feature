Feature: Get User by Username
  As a store administrator
  I want to look up a user by their username
  So that I can view or manage their account

  Background:
    Given the Petstore API at "https://petstore.swagger.io/v2"

  Scenario: Retrieve a user using their username
    Given a user account has been created
    When I send a GET request to "/user/{username}"
    Then the response status code should be 200
    And the response body should contain the correct username
    And the response body should contain the correct email
