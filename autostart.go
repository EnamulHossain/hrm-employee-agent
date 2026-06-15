package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// ConfigureAutostart sets up the desktop agent to automatically start on OS boot.
func ConfigureAutostart() error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	switch runtime.GOOS {
	case "windows":
		return registerAutostartWindows(exePath)
	case "linux":
		return registerAutostartLinux(exePath)
	}
	return nil
}

func registerAutostartWindows(exePath string) error {
	// Use Windows built-in 'reg' tool to add to HKCU Run folder
	cmd := exec.Command("reg", "add", "HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Run", "/v", "EmployeeAgent", "/t", "REG_SZ", "/d", exePath, "/f")
	return cmd.Run()
}

func registerAutostartLinux(exePath string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	autostartDir := filepath.Join(home, ".config", "autostart")
	err = os.MkdirAll(autostartDir, 0755)
	if err != nil {
		return err
	}

	desktopFileContent := fmt.Sprintf(`[Desktop Entry]
Type=Application
Version=1.0
Name=Employee Desktop Agent
Comment=Background tracking and heartbeat monitor for HRM
Exec="%s"
Icon=utilities-system-monitor
Terminal=false
StartupNotify=false
X-GNOME-Autostart-enabled=true
`, exePath)

	desktopFilePath := filepath.Join(autostartDir, "employee-agent.desktop")
	return os.WriteFile(desktopFilePath, []byte(desktopFileContent), 0644)
}
