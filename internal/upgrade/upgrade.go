package upgrade

import (
	"fmt"
	"os"
	"os/exec"
)

func UpdateCLI() error {
	fmt.Println("Upgrading hianime-cli...")

	cmd := exec.Command("go", "install", "github.com/dhilzyi/hianime-cli@latest")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to update: %w", err)
	}

	fmt.Println("Update complete. Please restart the application.")

	return nil
}
