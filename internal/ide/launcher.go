package ide

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// OpenInRider opens a directory in JetBrains Rider.
func OpenInRider(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		// Search PATH first, then known install locations
		candidates := []string{"rider64.exe", "rider.cmd", "rider"}
		knownPaths := []string{
			os.Getenv("LOCALAPPDATA") + "\\Programs\\Rider\\bin\\rider64.exe",
			os.Getenv("PROGRAMFILES") + "\\JetBrains\\JetBrains Rider\\bin\\rider64.exe",
		}
		// Also check Toolbox installs
		toolboxBase := os.Getenv("LOCALAPPDATA") + "\\JetBrains\\Toolbox\\apps\\Rider"
		if entries, err := os.ReadDir(toolboxBase); err == nil {
			for _, ch := range entries {
				if ch.IsDir() {
					subs, _ := os.ReadDir(toolboxBase + "\\" + ch.Name())
					for _, s := range subs {
						if s.IsDir() {
							knownPaths = append(knownPaths, toolboxBase+"\\"+ch.Name()+"\\"+s.Name()+"\\bin\\rider64.exe")
						}
					}
				}
			}
		}

		// Try PATH first
		for _, name := range candidates {
			if p, err := exec.LookPath(name); err == nil {
				cmd = exec.Command(p, path)
				break
			}
		}
		// Try known locations
		if cmd == nil {
			for _, p := range knownPaths {
				if _, err := os.Stat(p); err == nil {
					cmd = exec.Command(p, path)
					break
				}
			}
		}
		if cmd == nil {
			return fmt.Errorf("Rider not found. Install it or add to PATH")
		}
	default:
		cmd = exec.Command("rider", path)
	}
	return cmd.Start()
}

// OpenInVSCode opens a directory in Visual Studio Code.
func OpenInVSCode(path string) error {
	name := "code"
	if runtime.GOOS == "windows" {
		name = "code.cmd"
		if _, err := exec.LookPath(name); err != nil {
			name = "code"
		}
	}
	p, err := exec.LookPath(name)
	if err != nil {
		return fmt.Errorf("VS Code not found in PATH: %w", err)
	}
	return exec.Command(p, path).Start()
}

// OpenInExplorer opens a directory in the system file manager.
func OpenInExplorer(path string) error {
	switch runtime.GOOS {
	case "windows":
		// explorer requires backslashes, git returns forward slashes
		winPath := strings.ReplaceAll(path, "/", "\\")
		return exec.Command("explorer", winPath).Start()
	case "darwin":
		return exec.Command("open", path).Start()
	default:
		return exec.Command("xdg-open", path).Start()
	}
}
