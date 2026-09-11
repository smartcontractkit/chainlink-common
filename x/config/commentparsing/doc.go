// Package commentparsing collects a config tree's own documentation, from one root value.
//
// A field's doc comment and its `validate` tag are the most accurate description of that field
// that exists, but only one of them survives compilation: reflection reaches the tag and never the
// comment. Reading source recovers the comment, and works for as long as the source is reachable -
// which for a dependency it is not, so a config field whose type comes from another module would
// simply lose its documentation.
//
// [Discover] closes both halves. Given a root value it walks the type tree, following pointers,
// slices, maps and embedded fields, and resolves every struct it reaches: a type this module
// declares is parsed from its source, and a type from a dependency is read from the DocComments
// method compiled into it. The caller names one root and never enumerates types, packages or
// directories.
//
// [Docs.GenerateDocCommentFiles] writes those methods into each discovered directory this module
// owns, which is what lets the next module downstream read them:
//
//	//go:generate go run ./gen
//
// A dependency's own package is never written to - its source is in the read-only module cache,
// and its documentation is its repository's to generate. A dependency that never generated is
// named by Discover rather than documented as blank.
//
// The generated methods are a build artifact, so they can fall behind the comments they were read
// from. A package that ships them should regenerate in CI and fail on a diff, the way any other
// generated file is guarded.
package commentparsing
