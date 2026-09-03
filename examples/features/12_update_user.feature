Feature: Update a User Account
  As a registered customer
  I want to update my profile information
  So that my contact details stay current

  Background:
    Given the Petstore API at "https://petstore.swagger.io/v2"

  Scenario: Update a user's email and phone
    Given a user account has been created
    When I send a PUT request to "/user/{username}" with updated email and phone
    Then the response status code should be 200
