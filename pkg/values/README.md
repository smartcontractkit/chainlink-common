# About this package

This package is separate from the main module, so it can be imported by the CRE SDK without unnecessary dependencies.

To update the chainlink-protos commit, update the variable chainlinkProtosVersion in installer/pkg/consts.go.

# installer
Use the installer package to ensure that the chainlink-protos commit always matches the version used in this package.

## ProtocGen
The `ProtocGen` struct allows you to directly compile a file or to specify a CRE capabilities using a helper `CapabilityConfig`.

Use of the `CapabilityConfig` requires the `ProtocHelper` field to be set.

The CRE uses the chainlink-protos repository, especially, and capabilities are used in both the cre-sdk-go and capabilities repositories.
Adding a different package with its own go.mod to isolate that logic would require an extra commit in this repository to update the protos repository.
The first would be to update this package, the second to update the consumer.

## InstallProtocGenToDir

This function allows you to install additional plugins for protoc. See comments on the function for more information.