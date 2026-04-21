//go:generate go tool wit-bindgen-go generate --world examples --out internal ./enclave:examples.wasm

package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	testsuite "github.com/EnclaveRunner/examples-go/internal/enclave/benchmark/test-suite"
	_ "github.com/ydnar/wasi-http-go/wasihttp" // enable wasi-http
)

func perform_measurement() error {
	measurementID := os.Getenv("MEASUREMENT_ID")
	measurementServer := os.Getenv("MEASUREMENT_SERVER")
	if measurementID == "" || measurementServer == "" {
		return errors.New("MEASUREMENT_ID or MEASUREMENT_SERVER is not set")
	}

	receiverServer := fmt.Sprintf("%s/started/?request=%s", strings.TrimRight(measurementServer, "/"), measurementID)

	resp, err := http.Get(receiverServer)
	if err != nil {
		return fmt.Errorf("Error fetching %s: %v", receiverServer, err)
	}

	defer resp.Body.Close()

	fmt.Printf("Fetched %s with status code: %d\n", receiverServer, resp.StatusCode)

	return nil
}

func init() {
	testsuite.Exports.Startuptime = func() (result [2]string) {
		err := perform_measurement()
		if err != nil {
			return [2]string{"", err.Error()}
		}

		return [2]string{"", ""}
	}
}

// main is required for the `wasi` target, even if it isn't used.
func main() {
	err := perform_measurement()
	if err != nil {

		fmt.Print(err)
	}
}
