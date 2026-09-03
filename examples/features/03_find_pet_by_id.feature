Feature: Find a Pet by ID
  As a customer
  I want to look up a specific pet by its ID
  So that I can view its details before visiting the store

  Background:
    Given the Petstore API at "https://petstore.swagger.io/v2"

  Scenario: Retrieve a pet using its ID
    Given a pet has been added to the store
    When I send a GET request to "/pet/{petId}" with the pet's ID
    Then the response status code should be 200
    And the response content type should be "application/json"
    And the response body should contain the correct pet name
    And the response body should contain the correct pet status
