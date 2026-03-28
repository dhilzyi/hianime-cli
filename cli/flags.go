package cli

import (
	"flag"
	"fmt"
	"os"
)

type FlagsStruct struct {
	Debug      bool
	Version    bool
	MpvVerbose bool
}

func ParseFlags() FlagsStruct {
	var flags FlagsStruct

	flag.BoolVar(&flags.Debug, "debug", false, "Enable debug")
	flag.BoolVar(&flags.Version, "version", false, "Show version")
	flag.BoolVar(&flags.MpvVerbose, "verbose", false, "Enable mpv verbose")

	flag.Parse()

	return flags
}

func HandleFlags(flags FlagsStruct, version string) {
	if flags.Version == true {
		fmt.Println(version)
		os.Exit(0)
	}
}
