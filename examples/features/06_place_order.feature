Feature: Place a Store Order
  As a customer
  I want to place an order for a pet
  So that I can reserve it for purchase

  Background:
    Given the Petstore API at "https://petstore.swagger.io/v2"

  Scenario: Place a new order for a pet
    Given an order for pet ID 1 with quantity 1
    When I send a POST request to "/store/order"
    Then the response status code should be 200
    And the response body should contain "status" equal to "placed"
    And the response body should contain a non-zero order "id"
