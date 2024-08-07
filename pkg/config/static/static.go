package static

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
	Checksum = Unset
)

func init() {
	buildInfo, ok := debug.ReadBuildInfo()
	if ok {
		Program = buildInfo.Main.Path
		if Version == "" {
			Version = buildInfo.Main.Version
		}
		if Checksum == "" {
			Checksum = buildInfo.Main.Sum
		}
	}
}

// Short returns a 7-character sha prefix and version, or Unset if blank.
func Short() (shaPre string, ver string) {
	return short(Checksum, Version)
}

func short(sha, ver string) (string, string) {
	if sha == "" {
		sha = Unset
	} else if len(sha) > 7 {
		sha = sha[:7]
	}
	if ver == "" {
		ver = Unset
	}
	return sha, ver
}
