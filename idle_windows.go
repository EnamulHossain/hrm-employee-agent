//go:build windows

package main

import (
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

type LASTINPUTINFO struct {
	CbSize uint32
	DwTime uint32
}

var (
	user32           = syscall.NewLazyDLL("user32.dll")
	getLastInputInfo = user32.NewProc("GetLastInputInfo")
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	getTickCount     = kernel32.NewProc("GetTickCount")
)

// GetSystemIdleTime retrieves the user idle time on Windows in seconds.
func GetSystemIdleTime() (uint32, error) {
	var lii LASTINPUTINFO
	lii.CbSize = uint32(unsafe.Sizeof(lii))

	r, _, err := getLastInputInfo.Call(uintptr(unsafe.Pointer(&lii)))
	if r == 0 {
		return 0, err
	}

	tickCount, _, _ := getTickCount.Call()
	idleMs := uint32(tickCount) - lii.DwTime
	return idleMs / 1000, nil
}

var (
	getForegroundWindow       = user32.NewProc("GetForegroundWindow")
	getWindowText             = user32.NewProc("GetWindowTextW")
	getWindowThreadProcessId  = user32.NewProc("GetWindowThreadProcessId")
	psapi                     = syscall.NewLazyDLL("psapi.dll")
)

const (
	processQueryLimitedInformation = 0x1000
	maxPath                        = 260
)

var (
	openProcess                  = kernel32.NewProc("OpenProcess")
	closeHandle                  = kernel32.NewProc("CloseHandle")
	queryFullProcessImageNameW   = kernel32.NewProc("QueryFullProcessImageNameW")
)

// getProcessName returns the executable name for a given process ID.
func getProcessName(pid uint32) string {
	// Open process with limited query rights (does NOT trigger antivirus)
	handle, _, _ := openProcess.Call(
		uintptr(processQueryLimitedInformation),
		0,
		uintptr(pid),
	)
	if handle == 0 {
		return ""
	}
	defer closeHandle.Call(handle)

	// Query the full image name
	buf := make([]uint16, maxPath)
	size := uint32(maxPath)
	r, _, _ := queryFullProcessImageNameW.Call(
		handle,
		0,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
	)
	if r == 0 {
		return ""
	}

	fullPath := syscall.UTF16ToString(buf[:size])
	return filepath.Base(fullPath)
}

// knownAppNames maps common executable names to friendly display names.
var knownAppNames = map[string]string{
	// Browsers
	"chrome.exe":            "Google Chrome",
	"firefox.exe":           "Mozilla Firefox",
	"msedge.exe":            "Microsoft Edge",
	"opera.exe":             "Opera",
	"brave.exe":             "Brave",
	"safari.exe":            "Safari",
	"vivaldi.exe":           "Vivaldi",
	"iexplore.exe":          "Internet Explorer",
	"chromium.exe":          "Chromium",
	"arc.exe":               "Arc",

	// Adobe Products
	"photoshop.exe":         "Adobe Photoshop",
	"illustrator.exe":       "Adobe Illustrator",
	"indesign.exe":          "Adobe InDesign",
	"afterfx.exe":           "Adobe After Effects",
	"premiere pro.exe":      "Adobe Premiere Pro",
	"adobe premiere pro.exe":"Adobe Premiere Pro",
	"acrobat.exe":           "Adobe Acrobat",
	"acrord32.exe":          "Adobe Reader",
	"acrobat reader.exe":    "Adobe Acrobat Reader",
	"bridge.exe":            "Adobe Bridge",
	"animate.exe":           "Adobe Animate",
	"dreamweaver.exe":       "Adobe Dreamweaver",
	"xd.exe":                "Adobe XD",
	"lightroom.exe":         "Adobe Lightroom",
	"audition.exe":          "Adobe Audition",
	"creative cloud.exe":    "Adobe Creative Cloud",
	"cep htmlengine.exe":    "Adobe CEP",
	"node.exe":              "Node.js",

	// IDEs & Code Editors
	"code.exe":              "VS Code",
	"code - insiders.exe":   "VS Code Insiders",
	"devenv.exe":            "Visual Studio",
	"idea64.exe":            "IntelliJ IDEA",
	"phpstorm64.exe":        "PhpStorm",
	"webstorm64.exe":        "WebStorm",
	"pycharm64.exe":         "PyCharm",
	"rider64.exe":           "Rider",
	"goland64.exe":          "GoLand",
	"clion64.exe":           "CLion",
	"rubymine64.exe":        "RubyMine",
	"datagrip64.exe":        "DataGrip",
	"studio64.exe":          "Android Studio",
	"sublime_text.exe":      "Sublime Text",
	"atom.exe":              "Atom",
	"notepad++.exe":         "Notepad++",
	"notepad.exe":           "Notepad",
	"antigravity.exe":       "Antigravity IDE",
	"cursor.exe":            "Cursor",
	"windsurf.exe":          "Windsurf",

	// Office & Productivity
	"winword.exe":           "Microsoft Word",
	"excel.exe":             "Microsoft Excel",
	"powerpnt.exe":          "Microsoft PowerPoint",
	"outlook.exe":           "Microsoft Outlook",
	"onenote.exe":           "Microsoft OneNote",
	"msteams.exe":           "Microsoft Teams",
	"teams.exe":             "Microsoft Teams",
	"slack.exe":             "Slack",
	"discord.exe":           "Discord",
	"zoom.exe":              "Zoom",
	"telegram.exe":          "Telegram",
	"whatsapp.exe":          "WhatsApp",
	"skype.exe":             "Skype",
	"signal.exe":            "Signal",

	// Design & Media
	"figma.exe":             "Figma",
	"sketch.exe":            "Sketch",
	"canva.exe":             "Canva",
	"blender.exe":           "Blender",
	"vlc.exe":               "VLC Media Player",
	"spotify.exe":           "Spotify",
	"obs64.exe":             "OBS Studio",
	"obs32.exe":             "OBS Studio",
	"gimp-2.10.exe":         "GIMP",
	"gimp.exe":              "GIMP",
	"inkscape.exe":          "Inkscape",
	"kdenlive.exe":          "Kdenlive",
	"audacity.exe":          "Audacity",

	// Development Tools
	"postman.exe":           "Postman",
	"insomnia.exe":          "Insomnia",
	"filezilla.exe":         "FileZilla",
	"putty.exe":             "PuTTY",
	"winscp.exe":            "WinSCP",
	"git-bash.exe":          "Git Bash",
	"mintty.exe":            "Git Bash",
	"powershell.exe":        "PowerShell",
	"windowsterminal.exe":   "Windows Terminal",
	"cmd.exe":               "Command Prompt",
	"conhost.exe":           "Console Host",
	"docker desktop.exe":    "Docker Desktop",
	"dbeaver.exe":           "DBeaver",
	"heidisql.exe":          "HeidiSQL",
	"mysqworkbench.exe":     "MySQL Workbench",
	"pgadmin4.exe":          "pgAdmin",

	// File & System
	"explorer.exe":          "File Explorer",
	"taskmgr.exe":           "Task Manager",
	"mspaint.exe":           "Paint",
	"snippingtool.exe":      "Snipping Tool",
	"calc.exe":              "Calculator",
	"mstsc.exe":             "Remote Desktop",
	"anydesk.exe":           "AnyDesk",
	"teamviewer.exe":        "TeamViewer",

	// Misc
	"notion.exe":            "Notion",
	"obsidian.exe":          "Obsidian",
	"trello.exe":            "Trello",
	"todoist.exe":           "Todoist",
	"evernote.exe":          "Evernote",
	"1password.exe":         "1Password",
	"bitwarden.exe":         "Bitwarden",
}

// GetActiveWindowTitle retrieves the active window's process name and title on Windows.
// Returns format: "AppName (Window Title)" or just "AppName" if no meaningful title.
func GetActiveWindowTitle() string {
	hwnd, _, _ := getForegroundWindow.Call()
	if hwnd == 0 {
		return "Idle"
	}

	// 1. Get the process ID for this window
	var pid uint32
	getWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&pid)))

	// 2. Get the process executable name
	processExe := ""
	if pid != 0 {
		processExe = getProcessName(pid)
	}

	// 3. Get the window title text
	buf := make([]uint16, 512)
	r, _, _ := getWindowText.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), 512)
	windowTitle := ""
	if r > 0 {
		windowTitle = syscall.UTF16ToString(buf[:r])
	}

	// 4. Resolve the friendly app name
	appName := resolveAppName(processExe, windowTitle)

	// 5. Build the final display string
	if appName == "" {
		if windowTitle != "" {
			return windowTitle
		}
		return "System"
	}

	// For known apps, also include a cleaned window title for context
	cleanTitle := cleanWindowTitle(windowTitle, appName, processExe)
	if cleanTitle != "" {
		if len(cleanTitle) > 60 {
			cleanTitle = cleanTitle[:60] + "..."
		}
		return appName + " (" + cleanTitle + ")"
	}

	return appName
}

// resolveAppName maps a process exe name to a friendly display name.
func resolveAppName(processExe, windowTitle string) string {
	if processExe == "" {
		return ""
	}

	lowerExe := strings.ToLower(processExe)

	// Check exact match in known names
	if friendly, ok := knownAppNames[lowerExe]; ok {
		return friendly
	}

	// Check partial matches for versioned executables (e.g., "gimp-2.10.exe")
	for exe, friendly := range knownAppNames {
		baseName := strings.TrimSuffix(exe, ".exe")
		if strings.Contains(lowerExe, baseName) {
			return friendly
		}
	}

	// Electron apps often run as their app name .exe
	// Clean up the exe name as a display name
	name := strings.TrimSuffix(processExe, ".exe")
	name = strings.TrimSuffix(name, ".EXE")

	// Capitalize first letter of each word
	words := strings.Fields(strings.ReplaceAll(strings.ReplaceAll(name, "-", " "), "_", " "))
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}

	if len(words) > 0 {
		return strings.Join(words, " ")
	}

	return processExe
}

// cleanWindowTitle strips the app name from the window title to avoid redundancy.
func cleanWindowTitle(windowTitle, appName, processExe string) string {
	if windowTitle == "" {
		return ""
	}

	// Don't show title if it's just the app name
	if strings.EqualFold(windowTitle, appName) {
		return ""
	}

	// Strip common suffixes that repeat the app name
	title := windowTitle
	suffixesToRemove := []string{
		" - " + appName,
		" — " + appName,
		" | " + appName,
		" – " + appName,
	}

	lowerTitle := strings.ToLower(title)
	for _, suffix := range suffixesToRemove {
		if strings.HasSuffix(lowerTitle, strings.ToLower(suffix)) {
			title = title[:len(title)-len(suffix)]
			break
		}
	}

	// Also try stripping known browser/app suffixes
	extraSuffixes := []string{
		" - Google Chrome", " - Mozilla Firefox", " - Microsoft Edge",
		" - Brave", " - Opera", " - Safari", " - Vivaldi",
		" - Visual Studio Code", " - Sublime Text",
		" - Adobe Photoshop", " - Adobe Illustrator", " - Adobe Premiere Pro",
		" - Adobe After Effects", " - Adobe InDesign", " - Adobe Acrobat",
	}
	for _, suffix := range extraSuffixes {
		if strings.HasSuffix(lowerTitle, strings.ToLower(suffix)) {
			title = title[:len(title)-len(suffix)]
			break
		}
	}

	title = strings.TrimSpace(title)

	// If after cleaning the title is empty or same as app, don't show it
	if title == "" || strings.EqualFold(title, appName) {
		return ""
	}

	return title
}
