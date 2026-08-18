This package allows the CRE module and host to be tested without relying on the Go SDK, which would create a circular dependency.
It allows for the host to be developed and tested independently of the Go SDK, which would allow us to parallelize and prioritize development of language SDKs.
It is also used in the standard_tests directory (see the readme there for more information).
