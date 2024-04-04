package build

import (
	"os"
	"runtime/debug"
)

// Unset is a sentinel value.
const Unset = "unset"

// Version and Checksum are set at compile time via build arguments.
var (
	// Program is updated to the full main program path if [debug.BuildInfo] is available.
	Program = os.Args[0]
	// Version is the semantic version of the build or Unset.
	Version = Unset
	// Checksum is the commit hash of the build or Unset.
	Checksum       = Unset
	ChecksumPrefix = Unset
)

func init() {
	buildInfo, ok := debug.ReadBuildInfo()
	if ok {
		Program = buildInfo.Main.Path
		if Version == Unset {
			Version = buildInfo.Main.Version
		}
		if Checksum == Unset {
			Checksum = buildInfo.Main.Sum
		}
	}
	if Version == "" {
		Version = Unset
	}
	if Checksum == "" {
		Checksum = Unset
	}
	ChecksumPrefix = Checksum
	if len(ChecksumPrefix) > 7 {
		ChecksumPrefix = ChecksumPrefix[:7]
	}
}
