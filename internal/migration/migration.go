package migration

import (
	"fmt"
	"log"

	"github.com/dhilzyi/hianime-cli/internal/config"
	"github.com/dhilzyi/hianime-cli/internal/player"
)

func Run(dataDir, embedVer string, oldCfg *config.Config) error {
	fmt.Println("\nUpdate started...")
	scriptPath := player.TrackScriptPath(dataDir, player.ScriptName)
	if err := player.WriteLuaScript(scriptPath); err != nil {
		log.Println(err)
	} else {
		fmt.Println("	Update lua script completed.")
	}
	err := config.BumpConfig(oldCfg, embedVer)
	if err != nil {
		return err
	}
	fmt.Println("	Update config completed.")

	return nil
}
