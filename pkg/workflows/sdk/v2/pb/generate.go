//go:generate go run ./gen
//go:generate mv sdk/v1alpha/sdk.pb.go sdk.pb.go
//go:generate rm -rf sdk
//go:generate mv tools/generator/v1alpha/cre_metadata.pb.go cre_metadata.pb.go
//go:generate rm -rf tools
package pb
