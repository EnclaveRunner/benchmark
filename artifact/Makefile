.PHONY: package generate compile

package:
	wkg wit build

generate:
	go tool wit-bindgen-go generate --world examples --out internal ./enclave:benchmark.wasm

compile:
	tinygo build -target=wasip2 -o benchmark.wasm --wit-package enclave:benchmark.wasm --wit-world benchmark main.go 2>&1
