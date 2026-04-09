package runtimepaths

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

var (
	once       sync.Once
	appBaseDir string
)

// AppBaseDir returns the writable root for runtime-generated files.
// In development it prefers the project working directory; in packaged builds
// it prefers the executable directory.
func AppBaseDir() string {
	once.Do(func() {
		appBaseDir = resolveAppBaseDir()
	})
	return appBaseDir
}

func DataDir() string {
	return filepath.Join(AppBaseDir(), "data")
}

func LogsDir() string {
	return filepath.Join(AppBaseDir(), "logs")
}

func DBPath() string {
	return filepath.Join(DataDir(), "stock.db") + "?cache_size=-524288&journal_mode=WAL"
}

func QuantTemplatesDir() string {
	return filepath.Join(DataDir(), "quant_templates")
}

func resolveAppBaseDir() string {
	exeDir := executableDir()
	wd, err := os.Getwd()
	if err == nil && looksLikeProjectRoot(wd) && isLikelyDevExecutableDir(exeDir) {
		return wd
	}
	if runtime.GOOS == "darwin" {
		if dir := userRuntimeBaseDir(); dir != "" {
			return dir
		}
	}
	if exeDir != "" {
		return exeDir
	}
	if err == nil && wd != "" {
		return wd
	}
	return "."
}

func executableDir() string {
	exePath, err := os.Executable()
	if err != nil {
		return ""
	}
	return filepath.Dir(exePath)
}

func looksLikeProjectRoot(dir string) bool {
	if dir == "" {
		return false
	}
	_, err := os.Stat(filepath.Join(dir, "wails.json"))
	return err == nil
}

func isLikelyDevExecutableDir(dir string) bool {
	if dir == "" {
		return false
	}
	lowerDir := strings.ToLower(filepath.Clean(dir))
	tempDir := strings.ToLower(filepath.Clean(os.TempDir()))
	return strings.Contains(lowerDir, tempDir) || strings.Contains(lowerDir, "go-build")
}

func userRuntimeBaseDir() string {
	configDir, err := os.UserConfigDir()
	if err != nil || configDir == "" {
		return ""
	}
	return filepath.Join(configDir, "Rubin Investment")
}
