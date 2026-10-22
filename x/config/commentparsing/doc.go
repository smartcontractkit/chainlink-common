// Package commentparsing recovers a config tree's doc comments and feeds them to code generators.
//
// [Run] closes that gap from one root value: it parses the comments of types this module declares,
// reads the compiled DocComments method of types from a dependency, and writes those methods for
// the local packages so the next module downstream can do the same. See the examples below, the
// examples directory for whole working commands, and the README for the wider arrangement.
//
// The generated methods are a build artifact, so they can fall behind the comments they were read
// from. A package that ships them should regenerate in CI and fail on a diff.
package commentparsing
