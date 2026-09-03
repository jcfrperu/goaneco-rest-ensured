package integration_test

import (
	"fmt"
	"net/http/httptest"
	"os"
	"testing"

	"go.uber.org/goleak"

	"github.com/jcfrperu/goaneco-rest-ensured/testserver"
)

// integrationServer is shared across all integration tests to avoid spawning
// a new TCP server per top-level test function.
var integrationServer *httptest.Server

func TestMain(m *testing.M) {
	integrationServer = testserver.NewTestServer()

	exitCode := m.Run()

	// Close server before leak detection so its goroutines are not false-positives.
	integrationServer.Close()

	if err := goleak.Find(); err != nil {
		fmt.Fprintf(os.Stderr, "goroutine leak detected: %v\n", err)
		if exitCode == 0 {
			exitCode = 1
		}
	}

	os.Exit(exitCode)
}
