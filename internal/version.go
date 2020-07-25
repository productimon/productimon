package internal

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

var printVersion bool

var (
	BinaryName = filepath.Base(os.Args[0])

	GitCommit      = "unknown"
	GitVersion     = "unknown"
	Builder        = "unknown"
	BuildMode      = "prod"
	BuildTimestamp = "0"
	BuildTime      = time.Unix(0, 0)
)

func init() {
	flag.BoolVar(&printVersion, "version", false, "Print version information")
	if buildTimeNum, err := strconv.Atoi(BuildTimestamp); err == nil {
		BuildTime = time.Unix(int64(buildTimeNum), 0)
	}
}

func ParseFlags() {
	flag.Parse()
	if printVersion {
		PrintVersion()
		os.Exit(0)
	}
}

func PrintVersion() {
	fmt.Printf("%s (%s)\n", BinaryName, BuildMode)
	fmt.Printf("Built on %s (%d)\n", BuildTime.Format(time.RFC1123), BuildTime.Unix())
	fmt.Printf("Built by %s\n", Builder)
	fmt.Printf("Built from git checkout %s (%s)\n", GitVersion, GitCommit)
}

func IsDebugBuild() bool {
	return BuildMode == "debug"
}
