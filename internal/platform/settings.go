package platform

import (
	"fmt"
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

// ExpandSettingsPath 展开设置路径中的 ~（测试用）。
func ExpandSettingsPath(path string) (string, error) {
	return ExpandPath(path)
}

// SettingsDirFromHome 根据给定 home 构造设置目录（单测用）。
func SettingsDirFromHome(home string, goos string) string {
	if goos == "windows" {
		return filepath.Join(home, "AppData", "Roaming", "fuckssh")
	}
	return filepath.Join(home, ".config", "fuckssh")
}

// SettingsPathFromHome 根据给定 home 构造 settings.json 路径（单测用）。
func SettingsPathFromHome(home string, goos string) string {
	return filepath.Join(SettingsDirFromHome(home, goos), "settings.json")
}

// ValidateSettingsDir 检查目录名是否符合预期（单测辅助）。
func ValidateSettingsDir(dir string) error {
	if dir == "" {
		return fmt.Errorf("platform: empty settings dir")
	}
	return nil
}
