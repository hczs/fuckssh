package keys

import (
	"fmt"
	"os"

	"golang.org/x/crypto/ssh"
)

// ErrPassphraseNotSupported 表示暂不支持带口令私钥。
var ErrPassphraseNotSupported = fmt.Errorf("keys: passphrase-protected private keys are not supported")

// LoadSignerFromFile 从 OpenSSH 私钥文件加载 Signer。
func LoadSignerFromFile(path string) (ssh.Signer, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(data)
	if err != nil {
		if _, ok := err.(*ssh.PassphraseMissingError); ok {
			return nil, ErrPassphraseNotSupported
		}
		return nil, fmt.Errorf("keys: parse private key: %w", err)
	}
	return signer, nil
}
