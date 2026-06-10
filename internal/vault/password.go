package vault

import (
	"fmt"
	"strings"
	"unicode"
)

// ValidatePassword 校验主密码强度。
// 规则：最少 6 位，不能纯数字。
func ValidatePassword(password string) error {
	if len(password) < 6 {
		return fmt.Errorf("密码至少需要 6 位")
	}

	allDigits := true
	for _, r := range password {
		if !unicode.IsDigit(r) {
			allDigits = false
			break
		}
	}
	if allDigits {
		return fmt.Errorf("密码不能是纯数字")
	}

	// 检查空白字符
	if strings.TrimSpace(password) != password {
		return fmt.Errorf("密码首尾不能有空格")
	}

	return nil
}
