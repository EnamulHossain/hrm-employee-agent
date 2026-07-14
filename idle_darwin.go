//go:build darwin

package main

import (
	"bytes"
	"os/exec"
	"strconv"
	"strings"
)

// GetSystemIdleTime retrieves the user idle time on macOS in seconds.
// Uses IOKit's HIDIdleTime via ioreg (no CGo required).
func GetSystemIdleTime() (uint32, error) {
	// ioreg -c IOHIDSystem -d 4 outputs HIDIdleTime in nanoseconds
	cmd := exec.Command("ioreg", "-c", "IOHIDSystem", "-d", "4")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return 0, nil // Fallback: assume active
	}

	// Parse the output looking for HIDIdleTime
	for _, line := range strings.Split(out.String(), "\n") {
		if strings.Contains(line, "HIDIdleTime") {
			// The line looks like: "HIDIdleTime" = 1234567890
			parts := strings.Split(line, "=")
			if len(parts) >= 2 {
				valStr := strings.TrimSpace(parts[len(parts)-1])
				idleNs, err := strconv.ParseUint(valStr, 10, 64)
				if err == nil {
					// Convert nanoseconds to seconds
					return uint32(idleNs / 1_000_000_000), nil
				}
			}
		}
	}

	return 0, nil
}

// GetActiveWindowTitle retrieves the active (frontmost) application name
// and window title on macOS using AppleScript via osascript.
func GetActiveWindowTitle() string {
	// First, get the frontmost application name
	appScript := `tell application "System Events" to get name of first application process whose frontmost is true`
	appCmd := exec.Command("osascript", "-e", appScript)
	var appOut bytes.Buffer
	appCmd.Stdout = &appOut
	appErr := appCmd.Run()

	appName := ""
	if appErr == nil {
		appName = strings.TrimSpace(appOut.String())
	}

	if appName == "" {
		return "macOS Desktop"
	}

	// Try to get the window title of the frontmost application
	winScript := `tell application "System Events"
	set frontApp to first application process whose frontmost is true
	try
		set winTitle to name of front window of frontApp
		return winTitle
	on error
		return ""
	end try
end tell`
	winCmd := exec.Command("osascript", "-e", winScript)
	var winOut bytes.Buffer
	winCmd.Stdout = &winOut
	winErr := winCmd.Run()

	winTitle := ""
	if winErr == nil {
		winTitle = strings.TrimSpace(winOut.String())
	}

	// If we got a window title, include it
	if winTitle != "" && winTitle != appName {
		if len(winTitle) > 50 {
			winTitle = winTitle[:50] + "..."
		}
		return appName + " (" + winTitle + ")"
	}

	return appName
}
