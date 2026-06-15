//go:build !windows && !linux

package main

// GetSystemIdleTime fallback implementation for other operating systems.
func GetSystemIdleTime() (uint32, error) {
	return 0, nil
}

// GetActiveWindowTitle fallback implementation for other operating systems.
func GetActiveWindowTitle() string {
	return "Active"
}
