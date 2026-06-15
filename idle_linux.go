//go:build linux

package main

import (
	"bytes"
	"os/exec"
	"strconv"
	"strings"
)

var lastMouseCoords string
var simulatedIdleSeconds uint32

// GetSystemIdleTime retrieves the user idle time on Linux (Ubuntu) in seconds.
func GetSystemIdleTime() (uint32, error) {
	// 1. Try xprintidle (Standard X11 idle tracker)
	cmd := exec.Command("xprintidle")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err == nil {
		idleMsStr := strings.TrimSpace(out.String())
		idleMs, err := strconv.ParseUint(idleMsStr, 10, 64)
		if err == nil {
			return uint32(idleMs / 1000), nil
		}
	}

	// 2. Try xdotool as fallback to monitor mouse coordinate changes
	cmd = exec.Command("xdotool", "getmouselocation")
	out.Reset()
	cmd.Stdout = &out
	err = cmd.Run()
	if err == nil {
		currentCoords := strings.TrimSpace(out.String())
		if currentCoords == lastMouseCoords {
			// Coordinates didn't change: add to idle timer
			simulatedIdleSeconds += 2
		} else {
			lastMouseCoords = currentCoords
			simulatedIdleSeconds = 0
		}
		return simulatedIdleSeconds, nil
	}

	// 3. Fallback: if running on headless server or no X11 utilities are installed.
	return 0, nil
}

// GetActiveWindowTitle retrieves the active window's title on Linux.
func GetActiveWindowTitle() string {
	// 1. Try xdotool getwindowfocus getwindowname
	cmd := exec.Command("xdotool", "getwindowfocus", "getwindowname")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err == nil {
		title := strings.TrimSpace(out.String())
		if title != "" {
			return title
		}
	}

	// 2. Try xprop as fallback to retrieve the window class/app name
	cmd = exec.Command("sh", "-c", "xprop -id $(xprop -root _NET_ACTIVE_WINDOW | awk '{print $5}') WM_CLASS 2>/dev/null")
	out.Reset()
	cmd.Stdout = &out
	err = cmd.Run()
	if err == nil {
		class := strings.TrimSpace(out.String())
		if strings.Contains(class, "=") {
			parts := strings.Split(class, "=")
			if len(parts) > 1 {
				val := strings.Trim(strings.TrimSpace(parts[1]), "\"")
				subparts := strings.Split(val, ",")
				if len(subparts) > 0 {
					return strings.Trim(strings.TrimSpace(subparts[len(subparts)-1]), "\"")
				}
			}
		}
	}

	return "Linux Desktop"
}
