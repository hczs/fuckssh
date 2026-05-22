package platform

import (
	"path/filepath"
	"testing"
)

func TestSettingsPathFromHome(t *testing.T) {
	got := SettingsPathFromHome("/home/user", "linux")
	want := filepath.Join("/home/user", ".config", "fuckssh", "settings.json")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	gotWin := SettingsPathFromHome(`C:\Users\u`, "windows")
	wantWin := filepath.Join(`C:\Users\u`, "AppData", "Roaming", "fuckssh", "settings.json")
	if gotWin != wantWin {
		t.Errorf("got %q, want %q", gotWin, wantWin)
	}
}
