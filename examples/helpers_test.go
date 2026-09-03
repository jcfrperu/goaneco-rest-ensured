package examples_test

import (
	"fmt"
	"time"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

const petstoreURL = "https://petstore.swagger.io/v2"

// petstore returns a request pre-configured for the Petstore API.
// Port(0) avoids appending the library's default port 8080.
func petstore() *rest.Request {
	return rest.Given().
		BaseURI(petstoreURL).
		Port(0)
}

// uniqueName returns a short unique name safe for use as usernames or pet names.
func uniqueName(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano()%1_000_000)
}
