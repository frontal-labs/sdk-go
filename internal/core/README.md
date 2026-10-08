# Internal HTTP transport

`internal/core` owns the low-level HTTP client used by the root `frontal` package. It applies the default request timeout and exposes idle connection cleanup while keeping transport details out of the public API.
