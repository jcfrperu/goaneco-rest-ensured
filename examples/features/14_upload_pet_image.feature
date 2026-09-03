Feature: Upload a Pet Photo
  As a pet store employee
  I want to attach a photo to a pet listing
  So that customers can see what the pet looks like

  Background:
    Given the Petstore API at "https://petstore.swagger.io/v2"

  Scenario: Upload a JPEG image for a pet
    Given a pet has been added to the store
    When I send a multipart POST request to "/pet/{petId}/uploadImage" with an image file
    Then the response status code should be 200
    And the response body should confirm the upload
