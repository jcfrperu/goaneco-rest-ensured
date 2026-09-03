Feature: Find Pets by Tags
  As a customer
  I want to search for pets using specific tags
  So that I can find pets matching my preferences

  Background:
    Given the Petstore API at "https://petstore.swagger.io/v2"

  Scenario: Search for pets with multiple tags
    When I search for pets tagged with "friendly" and "vaccinated"
    Then the response status code should be 200
    And the content type should be "application/json"
