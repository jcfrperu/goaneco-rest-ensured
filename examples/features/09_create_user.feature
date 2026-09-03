Feature: Create a User Account
  As a new customer
  I want to create an account on the Petstore
  So that I can log in and manage my orders

  Background:
    Given the Petstore API at "https://petstore.swagger.io/v2"

  Scenario: Create a new user with all details
    Given a user with username, first name, last name, email and password
    When I send a POST request to "/user"
    Then the response status code should be 200
