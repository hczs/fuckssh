package update

import "testing"

func TestNormalizeTag(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"v0.6.0", "v0.6.0"},
		{"0.6.0", "v0.6.0"},
		{"  0.6.0  ", "v0.6.0"},
		{"", ""},
	}
	for _, tt := range tests {
		got := normalizeTag(tt.in)
		if got != tt.want {
			t.Errorf("normalizeTag(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestCompareVersion(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"v0.6.0", "v0.7.0", -1},
		{"v0.7.0", "v0.6.0", 1},
		{"v0.6.0", "v0.6.0", 0},
		{"0.6.0", "v0.6.0", 0},
		{"devel", "v0.6.0", -1},
	}
	for _, tt := range tests {
		got := compareVersion(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("compareVersion(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestIsReleaseVersion(t *testing.T) {
	if !isReleaseVersion("v0.6.0") {
		t.Fatal("v0.6.0 should be release version")
	}
	if isReleaseVersion("devel") {
		t.Fatal("devel should not be release version")
	}
}
