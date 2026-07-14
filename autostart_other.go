//go:build !windows

package main

// registerAutostartWindows is a no-op on non-Windows platforms.
func registerAutostartWindows(_ string) error { return nil }

// removeAutostartWindows is a no-op on non-Windows platforms.
func removeAutostartWindows() {}

