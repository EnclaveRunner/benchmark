//go:generate go tool wit-bindgen-go generate --world examples --out internal ./enclave:examples.wasm

package main

import (
	"fmt"
	"net/http"

	testsuite "github.com/EnclaveRunner/examples-go/internal/enclave/benchmark/test-suite"
	_ "github.com/ydnar/wasi-http-go/wasihttp" // enable wasi-http
)

func init() {
	testsuite.Exports.Startuptime = func(receiverserver string) (result [2]string) {
		fmt.Printf("Welcome to Enclave, %s\n", receiverserver)

		fmt.Println("### Starting start-up benchmark: fetching " + receiverserver)
		resp, err := http.Get(receiverserver)
		if err != nil {
			return [2]string{"", fmt.Sprintf("Error fetching %s: %v\n", receiverserver, err)}
		}
		defer resp.Body.Close()
		fmt.Printf("Fetched %s with status code: %d\n", receiverserver, resp.StatusCode)
		return [2]string{"", ""}
	}
}

// main is required for the `wasi` target, even if it isn't used.
func main() {}
