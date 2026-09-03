Feature: Full Pet Lifecycle
  As a pet store manager
  I want to manage the complete lifecycle of a pet record
  So that every phase from registration to removal is correct

  Background:
    Given the Petstore API at "https://petstore.swagger.io/v2"

  Scenario: Create, retrieve, update, and delete a pet
    When I add a new pet named "Lifecycle Dog" with status "available"
    Then the response status code should be 200
    And I capture the pet's generated ID

    When I retrieve the pet by its ID
    Then the response status code should be 200
    And the pet name should be "Lifecycle Dog"

    When I update the pet's status to "pending"
    Then the response status code should be 200
    And the pet status should be "pending"

    When I delete the pet
    Then the response status code should be 200

    When I try to retrieve the deleted pet
    Then the response status code should be 404
