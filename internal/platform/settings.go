package platform

import (
	"os"
	"path/filepath"
	"runtime"
)

// SettingsDir 返回 fuckssh 用户设置目录（~/.config/fuckssh 或 %APPDATA%\fuckssh）。
func SettingsDir() (string, error) {
	home, err := userHomeDir()
	if err != nil {
		return "", err
	}
	if runtime.GOOS == "windows" {
		if appData := os.Getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, "fuckssh"), nil
		}
		return filepath.Join(home, "AppData", "Roaming", "fuckssh"), nil
	}
	return filepath.Join(home, ".config", "fuckssh"), nil
}

// SettingsPath 返回语言等偏好设置文件路径。
func SettingsPath() (string, error) {
	dir, err := SettingsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "settings.json"), nil
}

// MkdirSettingsDir 创建设置目录（权限 0700）。
func MkdirSettingsDir() error {
	dir, err := SettingsDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(dir, 0o700)
}
