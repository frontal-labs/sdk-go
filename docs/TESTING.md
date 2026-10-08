# Go testing strategy

Use Go's standard test runner and mock HTTP at the shared transport boundary. Put package tests beside source files in `*_test.go` files. Cover request encoding, response decoding, API errors, retry safety, timeout behavior, pagination, and streaming. Keep unit tests independent of live keys; live API checks must be explicitly opt-in.

Use Go's standard `testing` package and `net/http/httptest` to exercise request construction, responses, retries, and error conversion without a live Frontal account. Keep integration checks opt-in and place them in a clearly named integration package if they need a separate test lifecycle. No runtime tests have been added yet.
