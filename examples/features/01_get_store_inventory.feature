Feature: Get Store Inventory
  As a store manager
  I want to check the current pet inventory
  So that I know how many pets are available, pending, or sold

  Background:
    Given the Petstore API at "https://petstore.swagger.io/v2"

  Scenario: Retrieve the pet inventory
    When I request the store inventory
    Then the response status code should be 200
    And the response should contain a JSON object
    And the object should map pet statuses to their quantities
