This directory does NOT use the CRE's GO library to avoid circular dependencies.

This test will not only run against the chainlink-common repository.
It will also run against the cre-sdk-* repositories.
Modules are created using a Makefile in the test directory.
The default path for tests is ./standard_tests, but it can be changed using the -path flag.
This allows us to test the module itself to ensure it works correctly, and the standard tests can pass.
Each cre-sdk-* repository would contain a standard_tests directory that produces the same binaries that chainlink-common does.

Changes to tests in this repository must be additive.
* It is ok to add new tests, or to add steps in an existing test.
* It is not ok to remove or change existing portions of tests (once a stable SDK is released), as this would allow breaking changes to the host <-> communications.
  * It is ok to refactor names of functions in the internal package (or to move the internal package), but not to modify their behaviour. 

  
Changes to the tests should be tagged with the host version so SDKs can implement them accordingly.