Feature: Add a New Pet
  As a pet store employee
  I want to register a new pet
  So that customers can find and adopt it

  Background:
    Given the Petstore API at "https://petstore.swagger.io/v2"

  Scenario: Successfully add a new available pet
    Given a pet named "Buddy" with status "available"
    When I send a POST request to "/pet" with the pet data as JSON
    Then the response status code should be 200
    And the response body should contain "name" equal to "Buddy"
    And the response body should contain "status" equal to "available"
    And the response body should contain a non-zero "id"
