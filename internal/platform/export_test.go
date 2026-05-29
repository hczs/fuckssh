package platform

import "path/filepath"

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
