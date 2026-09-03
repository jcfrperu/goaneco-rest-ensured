Feature: Delete a User Account
  As a store administrator
  I want to remove a user account
  So that inactive users are cleaned up

  Background:
    Given the Petstore API at "https://petstore.swagger.io/v2"

  Scenario: Delete a user by username
    Given a user account has been created
    When I send a DELETE request to "/user/{username}"
    Then the response status code should be 200
