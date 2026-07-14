package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// ConfigureAutostart sets up the desktop agent to automatically start on OS boot/login.
// It uses the most reliable native mechanism for each platform with fallback strategies.
func ConfigureAutostart() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not determine executable path: %w", err)
	}

	// Resolve symlinks to get the real path
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return fmt.Errorf("could not resolve executable symlinks: %w", err)
	}

	switch runtime.GOOS {
	case "windows":
		return registerAutostartWindows(exePath)
	case "linux":
		return registerAutostartLinux(exePath)
	case "darwin":
		return registerAutostartDarwin(exePath)
	default:
		log.Printf("[WARNING] Autostart not supported on %s", runtime.GOOS)
		return nil
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Linux: systemd user service (primary) + XDG desktop autostart (fallback)
// ──────────────────────────────────────────────────────────────────────────────

func registerAutostartLinux(exePath string) error {
	var firstErr error

	// 1. systemd user service (most reliable for modern Linux)
	if err := createSystemdUserService(exePath); err != nil {
		log.Printf("[WARNING] Failed to create systemd user service: %v", err)
		firstErr = err
	} else {
		log.Println("[INFO] Linux autostart: systemd user service created and enabled.")
	}

	// 2. XDG Desktop autostart (fallback for GNOME/KDE desktop environments)
	if err := createDesktopAutostart(exePath); err != nil {
		log.Printf("[WARNING] Failed to create XDG desktop autostart entry: %v", err)
		if firstErr == nil {
			firstErr = err
		}
	} else {
		log.Println("[INFO] Linux autostart: XDG desktop autostart entry created.")
	}

	return firstErr
}

func createSystemdUserService(exePath string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	serviceDir := filepath.Join(home, ".config", "systemd", "user")
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		return err
	}

	serviceContent := fmt.Sprintf(`[Unit]
Description=HRM Employee Desktop Agent
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=%s
Restart=on-failure
RestartSec=10
StartLimitIntervalSec=300
StartLimitBurst=5
Environment=DISPLAY=:0

[Install]
WantedBy=default.target
`, exePath)

	servicePath := filepath.Join(serviceDir, "employee-agent.service")
	if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
		return err
	}

	// Reload systemd and enable the service
	_ = exec.Command("systemctl", "--user", "daemon-reload").Run()
	if err := exec.Command("systemctl", "--user", "enable", "employee-agent.service").Run(); err != nil {
		return fmt.Errorf("failed to enable systemd service: %w", err)
	}

	return nil
}

func createDesktopAutostart(exePath string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	autostartDir := filepath.Join(home, ".config", "autostart")
	if err := os.MkdirAll(autostartDir, 0755); err != nil {
		return err
	}

	desktopFileContent := fmt.Sprintf(`[Desktop Entry]
Type=Application
Version=1.0
Name=HRM Employee Desktop Agent
Comment=Background monitoring and heartbeat agent for HRM system
Exec=%s
Icon=utilities-system-monitor
Terminal=false
StartupNotify=false
X-GNOME-Autostart-enabled=true
X-GNOME-Autostart-Delay=10
`, exePath)

	desktopFilePath := filepath.Join(autostartDir, "employee-agent.desktop")
	return os.WriteFile(desktopFilePath, []byte(desktopFileContent), 0644)
}

// ──────────────────────────────────────────────────────────────────────────────
// macOS: LaunchAgent plist with KeepAlive
// ──────────────────────────────────────────────────────────────────────────────

func registerAutostartDarwin(exePath string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	launchAgentsDir := filepath.Join(home, "Library", "LaunchAgents")
	if err := os.MkdirAll(launchAgentsDir, 0755); err != nil {
		return err
	}

	// Determine log file path
	logDir := filepath.Join(home, ".employee-agent")
	_ = os.MkdirAll(logDir, 0755)

	plistContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.hrm.employee-agent</string>

    <key>ProgramArguments</key>
    <array>
        <string>%s</string>
    </array>

    <key>RunAtLoad</key>
    <true/>

    <key>KeepAlive</key>
    <dict>
        <key>SuccessfulExit</key>
        <false/>
    </dict>

    <key>ThrottleInterval</key>
    <integer>15</integer>

    <key>StandardOutPath</key>
    <string>%s/agent-stdout.log</string>

    <key>StandardErrorPath</key>
    <string>%s/agent-stderr.log</string>

    <key>ProcessType</key>
    <string>Background</string>
</dict>
</plist>
`, exePath, logDir, logDir)

	plistPath := filepath.Join(launchAgentsDir, "com.hrm.employee-agent.plist")

	// Unload existing if present (ignore errors)
	_ = exec.Command("launchctl", "unload", plistPath).Run()

	if err := os.WriteFile(plistPath, []byte(plistContent), 0644); err != nil {
		return err
	}

	// Load the new plist
	if err := exec.Command("launchctl", "load", plistPath).Run(); err != nil {
		log.Printf("[WARNING] Failed to load LaunchAgent (may need manual load): %v", err)
	}

	log.Println("[INFO] macOS autostart: LaunchAgent plist created and loaded.")
	return nil
}

// ──────────────────────────────────────────────────────────────────────────────
// RemoveAutostart cleans up all autostart registrations (used during uninstall)
// ──────────────────────────────────────────────────────────────────────────────

func RemoveAutostart() {
	switch runtime.GOOS {
	case "windows":
		removeAutostartWindows()

	case "linux":
		home, _ := os.UserHomeDir()
		if home != "" {
			_ = os.Remove(filepath.Join(home, ".config", "autostart", "employee-agent.desktop"))
			_ = exec.Command("systemctl", "--user", "disable", "employee-agent.service").Run()
			_ = os.Remove(filepath.Join(home, ".config", "systemd", "user", "employee-agent.service"))
			_ = exec.Command("systemctl", "--user", "daemon-reload").Run()
		}

	case "darwin":
		home, _ := os.UserHomeDir()
		if home != "" {
			plistPath := filepath.Join(home, "Library", "LaunchAgents", "com.hrm.employee-agent.plist")
			_ = exec.Command("launchctl", "unload", plistPath).Run()
			_ = os.Remove(plistPath)
		}
	}
}
