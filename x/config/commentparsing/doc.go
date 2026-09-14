// Package commentparsing recovers a config tree's doc comments and feeds them to code generators.
//
// A field's doc comment is the description of that field, and it is the one thing about a field
// that compilation discards: a struct tag survives into the type, so a consumer holding the type
// reads `validate` and the rest by reflection, but the comment is gone. Reading source recovers
// it, and works for as long as the source is reachable - which for a dependency it is not.
//
// [Run] closes both halves from one root value. It walks the type tree, following pointers,
// slices, maps and embedded fields; resolves every struct it reaches, parsing a type this module
// declares and reading the compiled DocComments method of a type from a dependency; writes those
// methods for the local packages so the next module downstream can do the same; and hands the
// resolved packages to each [Generator] the caller supplied:
//
//	//go:generate go run ./gen
//
//	func main() {
//		err := commentparsing.Run(commentparsing.RunArgs{
//			Roots: []any{&Config{}},
//			Tool:  "example.com/app/gen",
//		}, myGenerator)
//	}
//
// A caller names one root and never enumerates types, packages or directories. A generator
// receives [Package] values carrying the comments, whatever their origin, and returns files to
// write; [Run] writes them all through chainlink-common/pkg/utils/codegen, so every file gets the
// same formatting, import grouping and generated-by header.
//
// The generated methods are a build artifact, so they can fall behind the comments they were read
// from. A package that ships them should regenerate in CI and fail on a diff, the way any other
// generated file is guarded. A dependency that never generated is named by the walk rather than
// documented as blank.
package commentparsing
