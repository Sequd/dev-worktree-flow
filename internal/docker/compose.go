package docker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ekoru/dev-flow-tui/internal/process"
)

// HasComposeFile checks if a docker-compose file exists in the directory.
func HasComposeFile(dir string) bool {
	candidates := []string{
		"docker-compose.yml",
		"docker-compose.yaml",
		"compose.yml",
		"compose.yaml",
	}
	for _, name := range candidates {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			return true
		}
	}
	return false
}

// ComposeUp runs docker compose up -d in the given directory.
func ComposeUp(pm *process.Manager, wtPath string) error {
	if !HasComposeFile(wtPath) {
		return fmt.Errorf("no docker-compose file found in %s", wtPath)
	}

	cmd := exec.Command("docker", "compose", "up", "-d")
	cmd.Dir = wtPath

	return pm.Start(process.KindDocker, "docker-compose", wtPath, cmd)
}

// ComposeDown runs docker compose down in the given directory.
func ComposeDown(wtPath string) error {
	if !HasComposeFile(wtPath) {
		return fmt.Errorf("no docker-compose file found in %s", wtPath)
	}

	cmd := exec.Command("docker", "compose", "down")
	cmd.Dir = wtPath
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker compose down: %w\n%s", err, out)
	}
	return nil
}
