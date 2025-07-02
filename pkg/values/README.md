# About this package

This package is separate from the main module, so it can be imported by the CRE SDK without unnecessary dependencies.

To update the chainlink-protos commit, update the variable chainlinkProtosVersion in installer/pkg/consts.go.

# installer
Use the installer package to ensure that the chainlink-protos commit always matches the version used in this package.

## ProtocGen
The `ProtocGen` struct allows you to directly compile a file or to specify a CRE capabilities using a helper `CapabilityConfig`.
Use of the `CapabilityConfig` requires the `ProtocHelper` field to be set.


## InstallProtocGenToDir

This function allows you to install additional plugins for protoc. See comments on the function for more information. 