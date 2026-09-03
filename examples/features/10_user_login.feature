Feature: User Login
  As a registered customer
  I want to log in to the Petstore
  So that I can access protected resources

  Background:
    Given the Petstore API at "https://petstore.swagger.io/v2"

  Scenario: Log in with valid credentials
    Given a registered user with username and password
    When I send a GET request to "/user/login" with credentials as query parameters
    Then the response status code should be 200
    And the response should contain a session token
