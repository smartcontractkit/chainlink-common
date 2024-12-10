package build

import (
	"cmp"
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
		Program = cmp.Or(buildInfo.Main.Path, Program)
		if Version == Unset && buildInfo.Main.Version != "" {
			Version = buildInfo.Main.Version
		}
		if Checksum == Unset && buildInfo.Main.Sum != "" {
			Checksum = buildInfo.Main.Sum
		}
	}
	ChecksumPrefix = Checksum[:min(7, len(Checksum))]
}
