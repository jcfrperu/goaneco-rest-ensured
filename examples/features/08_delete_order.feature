Feature: Delete a Store Order
  As a customer
  I want to cancel my order
  So that I am no longer charged for a pet I decided not to buy

  Background:
    Given the Petstore API at "https://petstore.swagger.io/v2"

  Scenario: Delete a placed order
    Given an order has been placed
    When I send a DELETE request to "/store/order/{orderId}"
    Then the response status code should be 200
