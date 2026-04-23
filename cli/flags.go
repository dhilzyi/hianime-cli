package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/dhilzyi/hianime-cli/internal/upgrade"
)

type FlagsStruct struct {
	Debug      bool
	Version    bool
	MpvVerbose bool
	Update     bool
}

func ParseFlags() FlagsStruct {
	var flags FlagsStruct

	flag.BoolVar(&flags.Debug, "debug", false, "Enable debug")
	flag.BoolVar(&flags.Version, "version", false, "Show version")
	flag.BoolVar(&flags.MpvVerbose, "verbose", false, "Enable mpv verbose")
	flag.BoolVar(&flags.Update, "update", false, "Update to the latest version")

	flag.Parse()

	return flags
}

func HandleFlags(flags FlagsStruct, version string) {
	if flags.Version {
		fmt.Println(version)
		os.Exit(0)
	}
	if flags.Update {
		if err := upgrade.UpdateCLI(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		} else {
			os.Exit(0)
		}
	}
}
