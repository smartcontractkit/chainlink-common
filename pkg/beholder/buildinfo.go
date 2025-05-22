package beholder

import (
	"runtime/debug"
	"strings"
	"sync"
)

var (
	unknown          = "unknown"
	buidldInfoCached buidldInfo
	once             sync.Once
)

type buidldInfo struct {
	sdkVersion  string
	mainVersion string
	mainPath    string
	mainCommit  string
}

func getBuildInfoOnce() buidldInfo {
	once.Do(func() {
		buidldInfoCached = getBuildInfo()
	})
	return buidldInfoCached
}

func getBuildInfo() buidldInfo {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return buidldInfo{}
	}
	var (
		sdkVersion string
		mainCommit string
	)
	for _, mod := range info.Deps {
		if mod.Path == "github.com/smartcontractkit/chainlink-common" {
			// Extract the semantic version without metadata.
			semVer := strings.SplitN(mod.Version, "-", 2)[0]
			sdkVersion = "beholder-sdk-" + semVer
			break
		}
	}
	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" {
			mainCommit = setting.Value
			break
		}
	}
	return buidldInfo{
		sdkVersion:  sdkVersion,
		mainVersion: info.Main.Version,
		mainPath:    info.Main.Path,
		mainCommit:  mainCommit,
	}
}
