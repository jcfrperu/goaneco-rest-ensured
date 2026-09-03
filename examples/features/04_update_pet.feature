Feature: Update an Existing Pet
  As a pet store employee
  I want to update a pet's information
  So that the listing reflects the latest details

  Background:
    Given the Petstore API at "https://petstore.swagger.io/v2"

  Scenario: Update a pet's name and status
    Given a pet has been added to the store
    When I send a PUT request to "/pet" with updated name and status "pending"
    Then the response status code should be 200
    And the response body should reflect the updated name
    And the response body should reflect status "pending"
