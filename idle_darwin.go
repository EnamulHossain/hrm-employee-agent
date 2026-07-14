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

// macKnownApps maps common macOS application names to friendly display names.
// AppleScript returns the application process name which is usually the app name.
var macKnownApps = map[string]string{
	// Browsers
	"Google Chrome":          "Google Chrome",
	"Safari":                 "Safari",
	"Firefox":                "Mozilla Firefox",
	"Microsoft Edge":         "Microsoft Edge",
	"Opera":                  "Opera",
	"Brave Browser":          "Brave",
	"Vivaldi":                "Vivaldi",
	"Arc":                    "Arc",
	"Chromium":               "Chromium",
	"Orion":                  "Orion",

	// Adobe Products
	"Adobe Photoshop":        "Adobe Photoshop",
	"Adobe Photoshop 2024":   "Adobe Photoshop",
	"Adobe Photoshop 2025":   "Adobe Photoshop",
	"Adobe Illustrator":      "Adobe Illustrator",
	"Adobe InDesign":         "Adobe InDesign",
	"Adobe Premiere Pro":     "Adobe Premiere Pro",
	"Adobe After Effects":    "Adobe After Effects",
	"Adobe Acrobat Reader":   "Adobe Acrobat Reader",
	"Adobe Acrobat":          "Adobe Acrobat",
	"Adobe XD":               "Adobe XD",
	"Adobe Lightroom":        "Adobe Lightroom",
	"Adobe Lightroom Classic":"Adobe Lightroom Classic",
	"Adobe Bridge":           "Adobe Bridge",
	"Adobe Animate":          "Adobe Animate",
	"Adobe Dreamweaver":      "Adobe Dreamweaver",
	"Adobe Audition":         "Adobe Audition",
	"Adobe Media Encoder":    "Adobe Media Encoder",
	"Creative Cloud":         "Adobe Creative Cloud",

	// IDEs & Code Editors
	"Code":                   "VS Code",
	"Visual Studio Code":     "VS Code",
	"Sublime Text":           "Sublime Text",
	"Xcode":                  "Xcode",
	"IntelliJ IDEA":          "IntelliJ IDEA",
	"PhpStorm":               "PhpStorm",
	"WebStorm":               "WebStorm",
	"PyCharm":                "PyCharm",
	"GoLand":                 "GoLand",
	"CLion":                  "CLion",
	"Rider":                  "Rider",
	"RubyMine":               "RubyMine",
	"DataGrip":               "DataGrip",
	"Android Studio":         "Android Studio",
	"Atom":                   "Atom",
	"Nova":                   "Nova",
	"BBEdit":                 "BBEdit",
	"TextMate":               "TextMate",
	"TextEdit":               "TextEdit",
	"Antigravity":            "Antigravity IDE",
	"Cursor":                 "Cursor",
	"Windsurf":               "Windsurf",

	// Office & Productivity
	"Microsoft Word":         "Microsoft Word",
	"Microsoft Excel":        "Microsoft Excel",
	"Microsoft PowerPoint":   "Microsoft PowerPoint",
	"Microsoft Outlook":      "Microsoft Outlook",
	"Microsoft OneNote":      "Microsoft OneNote",
	"Microsoft Teams":        "Microsoft Teams",
	"Pages":                  "Pages",
	"Numbers":                "Numbers",
	"Keynote":                "Keynote",
	"Notes":                  "Notes",
	"Reminders":              "Reminders",
	"Calendar":               "Calendar",
	"Mail":                   "Apple Mail",
	"Slack":                  "Slack",
	"Discord":                "Discord",
	"Zoom":                   "Zoom",
	"zoom.us":                "Zoom",
	"Telegram":               "Telegram",
	"WhatsApp":               "WhatsApp",
	"Skype":                  "Skype",
	"Signal":                 "Signal",
	"Messages":               "Messages",
	"FaceTime":               "FaceTime",

	// Design & Media
	"Figma":                  "Figma",
	"Sketch":                 "Sketch",
	"Canva":                  "Canva",
	"Blender":                "Blender",
	"GIMP":                   "GIMP",
	"Inkscape":               "Inkscape",
	"Pixelmator Pro":         "Pixelmator Pro",
	"Affinity Photo":         "Affinity Photo",
	"Affinity Designer":      "Affinity Designer",
	"Affinity Publisher":     "Affinity Publisher",
	"DaVinci Resolve":        "DaVinci Resolve",
	"Final Cut Pro":          "Final Cut Pro",
	"iMovie":                 "iMovie",
	"Logic Pro":              "Logic Pro",
	"GarageBand":             "GarageBand",
	"VLC":                    "VLC Media Player",
	"Spotify":                "Spotify",
	"Music":                  "Apple Music",
	"QuickTime Player":       "QuickTime Player",
	"Preview":                "Preview",
	"Photos":                 "Photos",
	"OBS":                    "OBS Studio",
	"Audacity":               "Audacity",

	// Development Tools
	"Terminal":               "Terminal",
	"iTerm2":                 "iTerm2",
	"iTerm":                  "iTerm2",
	"Alacritty":              "Alacritty",
	"kitty":                  "Kitty",
	"WezTerm":                "WezTerm",
	"Warp":                   "Warp",
	"Hyper":                  "Hyper",
	"Postman":                "Postman",
	"Insomnia":               "Insomnia",
	"Docker Desktop":         "Docker Desktop",
	"Docker":                 "Docker Desktop",
	"DBeaver":                "DBeaver",
	"TablePlus":              "TablePlus",
	"Sequel Pro":             "Sequel Pro",
	"pgAdmin 4":              "pgAdmin",
	"Cyberduck":              "Cyberduck",
	"Transmit":               "Transmit",
	"Tower":                  "Tower (Git)",
	"Fork":                   "Fork (Git)",
	"SourceTree":             "SourceTree",
	"GitHub Desktop":         "GitHub Desktop",
	"Charles":                "Charles Proxy",

	// File & System
	"Finder":                 "Finder",
	"System Preferences":     "System Settings",
	"System Settings":        "System Settings",
	"Activity Monitor":       "Activity Monitor",
	"Disk Utility":           "Disk Utility",
	"Console":                "Console",
	"Keychain Access":        "Keychain Access",
	"App Store":              "App Store",

	// Remote Access
	"AnyDesk":                "AnyDesk",
	"TeamViewer":             "TeamViewer",
	"Microsoft Remote Desktop": "Remote Desktop",
	"Screens":                "Screens",
	"Jump Desktop":           "Jump Desktop",

	// Misc
	"Notion":                 "Notion",
	"Obsidian":               "Obsidian",
	"Todoist":                "Todoist",
	"Evernote":               "Evernote",
	"Bear":                   "Bear",
	"Craft":                  "Craft",
	"Things":                 "Things",
	"1Password":              "1Password",
	"Bitwarden":              "Bitwarden",
	"CleanMyMac":             "CleanMyMac",
	"Alfred":                 "Alfred",
	"Raycast":                "Raycast",
	"Bartender":              "Bartender",
	"Magnet":                 "Magnet",
	"Rectangle":              "Rectangle",
}

// resolveMacAppName maps a macOS application name to a friendly display name.
func resolveMacAppName(appName string) string {
	if appName == "" {
		return ""
	}

	// Exact match
	if friendly, ok := macKnownApps[appName]; ok {
		return friendly
	}

	// Case-insensitive match
	lowerName := strings.ToLower(appName)
	for key, friendly := range macKnownApps {
		if strings.ToLower(key) == lowerName {
			return friendly
		}
	}

	// Partial match for versioned apps (e.g., "Adobe Photoshop 2025")
	for key, friendly := range macKnownApps {
		if strings.Contains(lowerName, strings.ToLower(key)) {
			return friendly
		}
	}

	// Return the original name — macOS process names are usually human-readable
	return appName
}

// GetActiveWindowTitle retrieves the active (frontmost) application name
// and window title on macOS using AppleScript via osascript.
func GetActiveWindowTitle() string {
	// Get the frontmost application name using a single combined AppleScript
	// This is more efficient than two separate osascript calls
	script := `tell application "System Events"
	set frontApp to first application process whose frontmost is true
	set appName to name of frontApp
	try
		set winTitle to name of front window of frontApp
	on error
		set winTitle to ""
	end try
	return appName & "|||" & winTitle
end tell`

	cmd := exec.Command("osascript", "-e", script)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()

	if err != nil {
		return "macOS Desktop"
	}

	result := strings.TrimSpace(out.String())
	parts := strings.SplitN(result, "|||", 2)

	appName := ""
	winTitle := ""
	if len(parts) >= 1 {
		appName = strings.TrimSpace(parts[0])
	}
	if len(parts) >= 2 {
		winTitle = strings.TrimSpace(parts[1])
	}

	if appName == "" {
		return "macOS Desktop"
	}

	// Resolve to friendly name
	friendlyName := resolveMacAppName(appName)

	// Clean window title
	if winTitle != "" && !strings.EqualFold(winTitle, friendlyName) && !strings.EqualFold(winTitle, appName) {
		// Strip app name suffix from window title
		cleanTitle := winTitle
		lowerTitle := strings.ToLower(cleanTitle)
		suffixes := []string{
			" - " + strings.ToLower(friendlyName),
			" — " + strings.ToLower(friendlyName),
			" | " + strings.ToLower(friendlyName),
			" - " + strings.ToLower(appName),
		}
		for _, suffix := range suffixes {
			if strings.HasSuffix(lowerTitle, suffix) {
				cleanTitle = cleanTitle[:len(cleanTitle)-len(suffix)]
				break
			}
		}
		cleanTitle = strings.TrimSpace(cleanTitle)

		if cleanTitle != "" && !strings.EqualFold(cleanTitle, friendlyName) {
			if len(cleanTitle) > 60 {
				cleanTitle = cleanTitle[:60] + "..."
			}
			return friendlyName + " (" + cleanTitle + ")"
		}
	}

	return friendlyName
}
