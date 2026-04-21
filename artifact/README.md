# Artifact `enclave:examples-go`

This is an example artifact containing the package `enclave:examples-go`

`wit/examples-go.wit` defines the package.

1. Run `make package` to create `enclave:examples-go.wasm` from the wit.
2. Run `make generate` to generate the go boilerplate from the `enclave-examples-go.wasm`.
3. `main.go` implements the exported function.
4. Run `make compile` to compile `examples-go.wasm` using `tinygo`.