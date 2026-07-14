//go:build linux

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

// linuxKnownProcesses maps common Linux process names to friendly display names.
var linuxKnownProcesses = map[string]string{
	// Browsers
	"google-chrome":         "Google Chrome",
	"google-chrome-stable":  "Google Chrome",
	"chrome":                "Google Chrome",
	"chromium":              "Chromium",
	"chromium-browser":      "Chromium",
	"firefox":               "Mozilla Firefox",
	"firefox-esr":           "Mozilla Firefox",
	"brave-browser":         "Brave",
	"opera":                 "Opera",
	"vivaldi":               "Vivaldi",
	"epiphany":              "GNOME Web",
	"microsoft-edge":        "Microsoft Edge",
	"msedge":                "Microsoft Edge",

	// IDEs & Code Editors
	"code":                  "VS Code",
	"code-oss":              "VS Code (OSS)",
	"sublime_text":          "Sublime Text",
	"subl":                  "Sublime Text",
	"atom":                  "Atom",
	"idea":                  "IntelliJ IDEA",
	"phpstorm":              "PhpStorm",
	"webstorm":              "WebStorm",
	"pycharm":               "PyCharm",
	"goland":                "GoLand",
	"clion":                 "CLion",
	"rider":                 "Rider",
	"rubymine":              "RubyMine",
	"datagrip":              "DataGrip",
	"android-studio":        "Android Studio",
	"gedit":                 "Text Editor",
	"nano":                  "Nano",
	"vim":                   "Vim",
	"nvim":                  "Neovim",
	"emacs":                 "Emacs",
	"kate":                  "Kate",
	"mousepad":              "Mousepad",
	"antigravity":           "Antigravity IDE",
	"cursor":                "Cursor",
	"windsurf":              "Windsurf",

	// Adobe Products (Linux versions / Wine)
	"photoshop":             "Adobe Photoshop",
	"illustrator":           "Adobe Illustrator",
	"acrobat":               "Adobe Acrobat",

	// Office & Productivity
	"libreoffice":           "LibreOffice",
	"soffice":               "LibreOffice",
	"lowriter":              "LibreOffice Writer",
	"localc":                "LibreOffice Calc",
	"loimpress":             "LibreOffice Impress",
	"lodraw":                "LibreOffice Draw",
	"thunderbird":           "Thunderbird",
	"evolution":             "Evolution",
	"slack":                 "Slack",
	"discord":               "Discord",
	"telegram-desktop":      "Telegram",
	"signal-desktop":        "Signal",
	"zoom":                  "Zoom",
	"teams":                 "Microsoft Teams",
	"skypeforlinux":         "Skype",

	// Design & Media
	"gimp":                  "GIMP",
	"gimp-2.10":             "GIMP",
	"inkscape":              "Inkscape",
	"blender":               "Blender",
	"kdenlive":              "Kdenlive",
	"obs":                   "OBS Studio",
	"vlc":                   "VLC Media Player",
	"spotify":               "Spotify",
	"audacity":              "Audacity",
	"shotcut":               "Shotcut",
	"darktable":             "Darktable",
	"rawtherapee":           "RawTherapee",
	"figma-linux":           "Figma",
	"krita":                 "Krita",

	// Development Tools
	"postman":               "Postman",
	"insomnia":              "Insomnia",
	"dbeaver":               "DBeaver",
	"filezilla":             "FileZilla",
	"remmina":               "Remmina",
	"docker":                "Docker",
	"pgadmin4":              "pgAdmin",
	"mysql-workbench":       "MySQL Workbench",
	"mongosh":               "MongoDB Shell",

	// Terminal Emulators
	"gnome-terminal":        "Terminal",
	"gnome-terminal-server": "Terminal",
	"xterm":                 "XTerm",
	"konsole":               "Konsole",
	"tilix":                 "Tilix",
	"alacritty":             "Alacritty",
	"kitty":                 "Kitty",
	"wezterm":               "WezTerm",
	"terminator":            "Terminator",
	"xfce4-terminal":        "XFCE Terminal",
	"lxterminal":            "LXTerminal",
	"mate-terminal":         "MATE Terminal",
	"foot":                  "Foot",
	"warp":                  "Warp",

	// File Managers
	"nautilus":              "Files",
	"thunar":                "Thunar",
	"dolphin":               "Dolphin",
	"nemo":                  "Nemo",
	"pcmanfm":               "PCManFM",
	"caja":                  "Caja",

	// System
	"gnome-system-monitor":  "System Monitor",
	"gnome-settings":        "Settings",
	"gnome-control-center":  "Settings",

	// Misc
	"notion-app":            "Notion",
	"obsidian":              "Obsidian",
	"todoist":               "Todoist",
	"evernote":              "Evernote",
	"anydesk":               "AnyDesk",
	"teamviewer":            "TeamViewer",
	"1password":             "1Password",
	"bitwarden":             "Bitwarden",
}

// getLinuxProcessName retrieves the process name for a given PID from /proc.
func getLinuxProcessName(pid string) string {
	pid = strings.TrimSpace(pid)
	if pid == "" || pid == "0" {
		return ""
	}

	// Try /proc/PID/comm first (gives the short process name)
	commPath := filepath.Join("/proc", pid, "comm")
	data, err := os.ReadFile(commPath)
	if err == nil {
		name := strings.TrimSpace(string(data))
		if name != "" {
			return name
		}
	}

	// Try /proc/PID/cmdline for the full command
	cmdlinePath := filepath.Join("/proc", pid, "cmdline")
	data, err = os.ReadFile(cmdlinePath)
	if err == nil {
		// cmdline is null-separated
		cmdline := strings.Split(string(data), "\x00")
		if len(cmdline) > 0 && cmdline[0] != "" {
			return filepath.Base(cmdline[0])
		}
	}

	return ""
}

// getWindowPID tries to get the PID of the active window using multiple methods.
func getWindowPID() string {
	// Method 1: xdotool getwindowfocus getwindowpid
	cmd := exec.Command("xdotool", "getwindowfocus", "getwindowpid")
	var out bytes.Buffer
	cmd.Stdout = &out
	if cmd.Run() == nil {
		pid := strings.TrimSpace(out.String())
		if pid != "" && pid != "0" {
			return pid
		}
	}

	// Method 2: xprop _NET_WM_PID on the active window
	cmd = exec.Command("sh", "-c", `xprop -id $(xprop -root _NET_ACTIVE_WINDOW | awk '{print $5}') _NET_WM_PID 2>/dev/null | awk '{print $3}'`)
	out.Reset()
	cmd.Stdout = &out
	if cmd.Run() == nil {
		pid := strings.TrimSpace(out.String())
		if pid != "" && pid != "0" {
			return pid
		}
	}

	return ""
}

// resolveLinuxAppName maps a process name to a friendly display name.
func resolveLinuxAppName(processName string) string {
	if processName == "" {
		return ""
	}

	lower := strings.ToLower(processName)

	// Exact match
	if friendly, ok := linuxKnownProcesses[lower]; ok {
		return friendly
	}

	// Partial match (for versioned binaries like gimp-2.10)
	for proc, friendly := range linuxKnownProcesses {
		if strings.Contains(lower, proc) {
			return friendly
		}
	}

	// Electron apps — the process might be the app name directly
	// Clean up dashes/underscores and capitalize
	name := strings.ReplaceAll(processName, "-", " ")
	name = strings.ReplaceAll(name, "_", " ")
	words := strings.Fields(name)
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	if len(words) > 0 {
		return strings.Join(words, " ")
	}

	return processName
}

// GetActiveWindowTitle retrieves the active window's process name and title on Linux.
// Uses process detection via /proc for reliable identification of all applications.
func GetActiveWindowTitle() string {
	// 1. Get the window title
	windowTitle := ""
	cmd := exec.Command("xdotool", "getwindowfocus", "getwindowname")
	var out bytes.Buffer
	cmd.Stdout = &out
	if cmd.Run() == nil {
		windowTitle = strings.TrimSpace(out.String())
	}

	// 2. Get the PID of the active window and resolve the process name
	pid := getWindowPID()
	processName := ""
	if pid != "" {
		processName = getLinuxProcessName(pid)
	}

	// 3. Resolve to a friendly app name
	appName := resolveLinuxAppName(processName)

	// 4. If we couldn't get the process name, try WM_CLASS as fallback
	if appName == "" {
		cmd = exec.Command("sh", "-c", `xprop -id $(xprop -root _NET_ACTIVE_WINDOW | awk '{print $5}') WM_CLASS 2>/dev/null`)
		out.Reset()
		cmd.Stdout = &out
		if cmd.Run() == nil {
			class := strings.TrimSpace(out.String())
			if strings.Contains(class, "=") {
				parts := strings.Split(class, "=")
				if len(parts) > 1 {
					val := strings.Trim(strings.TrimSpace(parts[1]), `"`)
					subparts := strings.Split(val, ",")
					if len(subparts) > 0 {
						className := strings.Trim(strings.TrimSpace(subparts[len(subparts)-1]), `" `)
						appName = resolveLinuxAppName(className)
					}
				}
			}
		}
	}

	// 5. Build the final display string
	if appName == "" {
		if windowTitle != "" {
			return windowTitle
		}
		return "Linux Desktop"
	}

	// Clean window title to avoid redundancy with app name
	cleanTitle := cleanLinuxWindowTitle(windowTitle, appName)
	if cleanTitle != "" {
		if len(cleanTitle) > 60 {
			cleanTitle = cleanTitle[:60] + "..."
		}
		return fmt.Sprintf("%s (%s)", appName, cleanTitle)
	}

	return appName
}

// cleanLinuxWindowTitle strips the app name from the window title.
func cleanLinuxWindowTitle(windowTitle, appName string) string {
	if windowTitle == "" {
		return ""
	}
	if strings.EqualFold(windowTitle, appName) {
		return ""
	}

	title := windowTitle
	lowerTitle := strings.ToLower(title)

	// Strip common suffixes
	suffixes := []string{
		" - " + appName, " — " + appName, " | " + appName, " – " + appName,
		" - Google Chrome", " - Mozilla Firefox", " - Microsoft Edge",
		" - Brave", " - Opera", " - Vivaldi", " - Chromium",
		" - Visual Studio Code", " - Sublime Text",
	}
	for _, suffix := range suffixes {
		if strings.HasSuffix(lowerTitle, strings.ToLower(suffix)) {
			title = title[:len(title)-len(suffix)]
			break
		}
	}

	title = strings.TrimSpace(title)
	if title == "" || strings.EqualFold(title, appName) {
		return ""
	}

	return title
}
