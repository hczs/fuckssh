package wizard

import (
	"fmt"
	"strconv"
	"strings"
)

// validatePort 校验 SSH 端口为 1–65535；空字符串由调用方补默认值。
func validatePort(port string) error {
	port = strings.TrimSpace(port)
	if port == "" {
		return nil
	}
	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		return fmt.Errorf("%w: port must be between 1 and 65535", ErrInvalidInput)
	}
	return nil
}
