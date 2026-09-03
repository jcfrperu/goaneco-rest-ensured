Feature: Reusable Request Specification
  As a developer writing multiple tests
  I want to define shared request configuration once
  So that I avoid repeating headers, base URLs, and auth in every test

  Background:
    Given the Petstore API at "https://petstore.swagger.io/v2"

  Scenario: Two requests share the same base specification
    Given a request spec with the Petstore base URL and JSON Accept header
    When I use the spec to search for "available" pets
    Then the response status code should be 200
    When I reuse the same spec to search for "sold" pets
    Then the response status code should also be 200
