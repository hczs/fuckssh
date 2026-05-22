package keys

import (
	"fmt"
	"os"
)

// RemoveKeyPair 删除私钥与对应 .pub 公钥（add 失败回滚用，忽略不存在）。
func RemoveKeyPair(privPath string) error {
	if privPath == "" {
		return nil
	}
	pubPath := privPath + ".pub"
	for _, p := range []string{privPath, pubPath} {
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("keys: remove %q: %w", p, err)
		}
	}
	return nil
}
