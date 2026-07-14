//go:build windows

package main

import (
	"log"
	"os/exec"
	"syscall"
)

// registerAutostartWindows sets up autostart on Windows using both the Registry Run key
// and Windows Task Scheduler for maximum reliability.
func registerAutostartWindows(exePath string) error {
	var firstErr error

	// 1. Registry Run key (user-level, no admin required)
	regCmd := exec.Command("reg", "add",
		`HKCU\Software\Microsoft\Windows\CurrentVersion\Run`,
		"/v", "HRM Employee Agent",
		"/t", "REG_SZ",
		"/d", `"`+exePath+`"`,
		"/f",
	)
	regCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := regCmd.Run(); err != nil {
		log.Printf("[WARNING] Failed to set Registry Run key: %v", err)
		firstErr = err
	} else {
		log.Println("[INFO] Windows autostart: Registry Run key set successfully.")
	}

	// 2. Scheduled Task (survives antivirus cleanup of Run keys, triggers on logon)
	taskName := "HRM_Employee_Desktop_Agent"

	// Delete existing task first (ignore errors if it doesn't exist)
	delCmd := exec.Command("schtasks", "/Delete", "/TN", taskName, "/F")
	delCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_ = delCmd.Run()

	// Create task that triggers on any user logon with a 15-second delay
	createCmd := exec.Command("schtasks", "/Create",
		"/TN", taskName,
		"/TR", `"`+exePath+`"`,
		"/SC", "ONLOGON",
		"/DELAY", "0000:15",
		"/RL", "LIMITED",
		"/F",
	)
	createCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := createCmd.Run(); err != nil {
		log.Printf("[WARNING] Failed to create Scheduled Task: %v", err)
		if firstErr == nil {
			firstErr = err
		}
	} else {
		log.Println("[INFO] Windows autostart: Scheduled Task created successfully.")
	}

	return firstErr
}

// removeAutostartWindows removes all Windows autostart registrations.
func removeAutostartWindows() {
	regCmd := exec.Command("reg", "delete",
		`HKCU\Software\Microsoft\Windows\CurrentVersion\Run`,
		"/v", "HRM Employee Agent", "/f")
	regCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_ = regCmd.Run()

	delCmd := exec.Command("schtasks", "/Delete", "/TN", "HRM_Employee_Desktop_Agent", "/F")
	delCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_ = delCmd.Run()
}
