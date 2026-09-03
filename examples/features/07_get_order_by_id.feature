Feature: Get an Order by ID
  As a customer
  I want to look up my purchase order
  So that I can check its current status

  Background:
    Given the Petstore API at "https://petstore.swagger.io/v2"

  Scenario: Retrieve an order using its ID
    Given an order has been placed
    When I send a GET request to "/store/order/{orderId}"
    Then the response status code should be 200
    And the response body should contain the correct order status
