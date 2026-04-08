//go:build linux
// +build linux

package data

import "os"

// CheckChrome checks common Linux Chromium/Chrome install paths.
func CheckChrome() (string, bool) {
	locations := []string{
		"/usr/bin/google-chrome",
		"/usr/bin/google-chrome-stable",
		"/usr/bin/chromium",
		"/usr/bin/chromium-browser",
		"/snap/bin/chromium",
	}
	for _, location := range locations {
		if _, err := os.Stat(location); err == nil {
			return location, true
		}
	}
	return "", false
}

// CheckBrowser checks whether a supported browser is available on Linux.
func CheckBrowser() (string, bool) {
	return CheckChrome()
}
