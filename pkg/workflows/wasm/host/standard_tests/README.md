This directory does NOT use the CRE's GO library to avoid circular dependencies.

Instead, it makes raw callbacks to allow the host to be tested on its own for functionality.
The SDK repository will run the same set of standard tests.