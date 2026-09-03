Feature: Reusable Response Specification
  As a developer
  I want to define expected response rules once
  So that I can reuse them across multiple tests without duplication

  Background:
    Given the Petstore API at "https://petstore.swagger.io/v2"

  Scenario: Apply the same response rules to two different pet requests
    Given a response spec expecting status 200 and JSON content type
    When I add a pet and apply the response spec to that request
    Then all response spec expectations should pass
    When I add a second different pet and apply the same response spec
    Then all response spec expectations should also pass
