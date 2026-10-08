# Publishing to Go module proxy

Tag the repository with `vMAJOR.MINOR.PATCH`. Keep the module import path stable, verify `go list -m` resolves the tagged version, and publish release notes. The Go proxy indexes public tags; no registry upload step is required.

The repository currently has no registry publishing credentials or release action. Complete the implementation and release metadata first. Keep credentials in protected repository secrets and use the registry's recommended signing or trusted-publishing mechanism where available.
