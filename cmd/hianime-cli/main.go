package main

import (
	"bufio"
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/dhilzyi/hianime-cli/internal/app"
	"github.com/dhilzyi/hianime-cli/internal/config"
	"github.com/dhilzyi/hianime-cli/internal/release"
	"github.com/dhilzyi/hianime-cli/internal/state"
)

//go:embed version.txt
var embedVersion string

func main() {
	cleanEmbedVersion := strings.TrimSpace(embedVersion)
	flags := ParseFlags()
	HandleFlags(flags, cleanEmbedVersion)

	appDir, err := config.InitPath()
	if err != nil {
		fmt.Println("Fail to initialize path: " + err.Error())
	}
	configSes, err := config.LoadConfig(appDir.ConfigDir, cleanEmbedVersion)
	if err != nil {
		fmt.Println("Fail to load config file: " + err.Error())
	}

	if updated, err := release.Run(cleanEmbedVersion, appDir.DataDir, configSes); err != nil {
		fmt.Println("Fail to run version module:", err)
	} else {
		if updated {
			fmt.Printf("Info: Set to new config complete.\n")
		}
	}

	history, err := state.LoadHistory(appDir.DataDir)
	if err != nil {
		fmt.Println("Fail to load history file: " + err.Error())
	}

	scanner := bufio.NewScanner(os.Stdin)
	myApp := app.New(configSes, history, &appDir, &flags, *scanner)

	myApp.Start()
}
