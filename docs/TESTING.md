# Go testing strategy

Use Go's standard test runner and inject an `httptest.Server` through `frontal.WithBaseURL` and `frontal.WithHTTPClient`. Put tests beside source files in `*_test.go` files. Cover request encoding, response decoding, API errors, retry safety, timeout behavior, path/query encoding, and event-stream parsing. Keep unit tests independent of live keys; live API checks must be explicitly opt-in.

Use Go's standard `testing` package and `net/http/httptest` to exercise request construction, responses, retries, and error conversion without a live Frontal account. Keep integration checks opt-in and place them in a clearly named integration package if they need a separate test lifecycle.
