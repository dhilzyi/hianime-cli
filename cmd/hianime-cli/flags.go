package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/dhilzyi/hianime-cli/internal/app"
	"github.com/dhilzyi/hianime-cli/internal/release"
)

func ParseFlags() app.Flags {
	var flags app.Flags

	flag.BoolVar(&flags.Debug, "debug", false, "Enable debug")
	flag.BoolVar(&flags.Version, "version", false, "Show version")
	flag.BoolVar(&flags.MpvVerbose, "verbose", false, "Enable mpv verbose")
	flag.BoolVar(&flags.Update, "update", false, "Update to the latest version")

	flag.Parse()

	return flags
}

func HandleFlags(flags app.Flags, version string) {
	if flags.Version {
		fmt.Println(version)
		os.Exit(0)
	}
	if flags.Update {
		if err := release.UpdateCLI(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		} else {
			os.Exit(0)
		}
	}
}
