Feature: Find Pets by Status
  As a pet store customer
  I want to search for pets by their availability status
  So that I can see which pets I can adopt, reserve, or have already been sold

  Background:
    Given the Petstore API at "https://petstore.swagger.io/v2"

  Scenario: Find all available pets
    When I search for pets with status "available"
    Then the response status code should be 200
    And the content type should be "application/json"
    And the response should contain at least one pet
    And every pet in the response should have a non-empty name
    And every pet in the response should have status "available"

  Scenario: Find pets pending adoption
    When I search for pets with status "pending"
    Then the response status code should be 200
    And the content type should be "application/json"

  Scenario: Find sold pets
    When I search for pets with status "sold"
    Then the response status code should be 200
    And the content type should be "application/json"
