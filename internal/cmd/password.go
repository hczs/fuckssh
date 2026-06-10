package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

// readPasswordFromTerminal 从终端读取密码（不回显）。
// 仅在终端环境下生效，否则返回错误。
func readPasswordFromTerminal() (string, error) {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return "", fmt.Errorf("not a terminal")
	}

	pw, err := term.ReadPassword(fd)
	if err != nil {
		return "", fmt.Errorf("读取密码失败: %w", err)
	}

	return string(pw), nil
}

// readPasswordMasked 从终端读取密码，每输入一个字符显示一个星号。
// 非终端环境下回退到 readLineFromStdin。
func readPasswordMasked(stderr io.Writer, prompt string) (string, error) {
	_, _ = fmt.Fprint(stderr, prompt)

	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return readLineFromStdin()
	}

	// 保存原始终端状态，结束后恢复
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		// raw 模式设置失败，回退到静默读取
		return readPasswordFromTerminal()
	}
	defer func() { _ = term.Restore(fd, oldState) }()

	var buf []byte
	tmp := make([]byte, 1)
	for {
		n, err := os.Stdin.Read(tmp)
		if n == 1 {
			switch tmp[0] {
			case '\r', '\n': // 回车结束
				_, _ = fmt.Fprintln(stderr)
				return string(buf), nil
			case 3: // Ctrl+C
				_, _ = fmt.Fprintln(stderr)
				return "", fmt.Errorf("用户取消")
			case 127, 8: // Backspace / Delete
				if len(buf) > 0 {
					buf = buf[:len(buf)-1]
					_, _ = fmt.Fprint(stderr, "\b \b")
				}
			default:
				buf = append(buf, tmp[0])
				_, _ = fmt.Fprint(stderr, "*")
			}
		}
		if err != nil {
			break
		}
	}

	_, _ = fmt.Fprintln(stderr)
	return string(buf), nil
}

// readLineFromStdin 从 stdin 读取一行（回退方案，密码会回显）。
func readLineFromStdin() (string, error) {
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return scanner.Text(), nil
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("读取输入失败: %w", err)
	}
	return "", fmt.Errorf("未读取到输入")
}
