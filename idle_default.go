//go:build !windows && !linux && !darwin

package main

// GetSystemIdleTime fallback implementation for unsupported operating systems.
func GetSystemIdleTime() (uint32, error) {
	return 0, nil
}

// GetActiveWindowTitle fallback implementation for unsupported operating systems.
func GetActiveWindowTitle() string {
	return "Active"
}
