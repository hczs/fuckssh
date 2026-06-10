package vault

import "testing"

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name    string
		pw      string
		wantErr bool
	}{
		{"合法密码", "abc123", false},
		{"纯字母", "abcdef", false},
		{"带特殊字符", "p@ss!1", false},
		{"太短", "ab1", true},
		{"纯数字", "123456", true},
		{"空密码", "", true},
		{"首尾空格", " abc123 ", true},
		{"正好6位", "abc123", false},
		{"5位", "ab12d", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.pw)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePassword(%q) error = %v, wantErr %v", tt.pw, err, tt.wantErr)
			}
		})
	}
}
