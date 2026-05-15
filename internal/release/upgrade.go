package release

import (
	"fmt"
	"os"
	"os/exec"
)

func UpdateCLI() error {
	_, err := exec.LookPath("go")
	if err != nil {
		return fmt.Errorf("self-update failed: 'go' is not installed or not in your system PATH. \nDownload the latest release manually from GitHub or install 'go' to your PATH")
	}

	fmt.Println("Updating hianime-cli...")

	cmd := exec.Command("go", "install", "github.com/dhilzyi/hianime-cli/cmd/hianime-cli@latest")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("self-update failed: %w", err)
	}

	fmt.Println("Update complete.")

	return nil
}
