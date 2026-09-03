Feature: Delete a Pet
  As a pet store manager
  I want to remove a pet from the store
  So that sold or unavailable pets are no longer listed

  Background:
    Given the Petstore API at "https://petstore.swagger.io/v2"

  Scenario: Delete a pet using its ID
    Given a pet has been added to the store
    When I send a DELETE request to "/pet/{petId}" with the api_key header
    Then the response status code should be 200
