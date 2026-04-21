//go:generate go tool wit-bindgen-go generate --world examples --out internal ./enclave:examples.wasm

package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	testsuite "github.com/EnclaveRunner/examples-go/internal/enclave/benchmark/test-suite"
	_ "github.com/ydnar/wasi-http-go/wasihttp" // enable wasi-http
)

func init() {
	testsuite.Exports.Startuptime = func() (result [2]string) {
		measurementID := os.Getenv("MEASUREMENT_ID")
		measurementServer := os.Getenv("MEASUREMENT_SERVER")
		if measurementID == "" || measurementServer == "" {
			return [2]string{"", "MEASUREMENT_ID or MEASUREMENT_SERVER is not set"}
		}

		receiverServer := fmt.Sprintf("%s/started/?request=%s", strings.TrimRight(measurementServer, "/"), measurementID)

		resp, err := http.Get(receiverServer)
		if err != nil {
			return [2]string{"", fmt.Sprintf("Error fetching %s: %v", receiverServer, err)}
		}
		defer resp.Body.Close()

		fmt.Printf("Fetched %s with status code: %d\n", receiverServer, resp.StatusCode)
		return [2]string{"", ""}
	}
}

// main is required for the `wasi` target, even if it isn't used.
func main() {}
