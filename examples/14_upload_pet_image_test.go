package examples_test

// Feature: Upload a Pet Photo
// Feature file: examples/features/14_upload_pet_image.feature
//
// Demonstrates: multipart file upload using MultiPart().
// A small synthetic JPEG byte sequence is used so the test needs no real file.
//
// Run: go test ./examples/... -run TestPetstore_14 -v

import (
	"net/http"
	"testing"

	"github.com/jcfrperu/goaneco-rest-ensured/examples/models"
	"github.com/jcfrperu/goaneco-rest-ensured/rest"
	"github.com/stretchr/testify/require"
)

// Scenario: Upload a JPEG image for a pet
func TestPetstore_14_UploadPetImage(t *testing.T) {
	// Create a pet to attach the image to.
	newPet := models.Pet{Name: "PhotoPet", Status: "available", PhotoUrls: []string{}}
	created := petstore().BodyObject(newPet).Post("/pet")
	require.NoError(t, created.Err(), "create pet must succeed")

	petID := created.JsonPath().GetInt64("id")
	require.NotZero(t, petID, "server must return a generated pet ID")

	// A minimal valid JPEG header (2 bytes) — enough for the upload endpoint.
	fakeImage := []byte{0xFF, 0xD8, 0xFF, 0xE0}

	petstore().
		PathParam("petId", petID).
		MultiPart("additionalMetadata", []byte("test photo")).
		MultiPartNamed("file", "photo.jpg", fakeImage, "image/jpeg").
		When().
		Post("/pet/{petId}/uploadImage").
		Then().
		StatusCode(http.StatusOK).
		ContentType(rest.ContentTypeJSON).
		AssertAllNoFail(t)
}
