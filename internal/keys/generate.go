package keys

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"fmt"

	"golang.org/x/crypto/ssh"
)

// KeyPair 表示一对 OpenSSH 格式的 Ed25519 密钥材料。
type KeyPair struct {
	// PrivatePEM 为 OPENSSH PRIVATE KEY 格式的私钥字节。
	PrivatePEM []byte
	// PublicLine 为 authorized_keys 单行公钥（末尾含换行）。
	PublicLine string
}

// GenerateEd25519 生成新的 Ed25519 密钥对，私钥为 OpenSSH PEM，公钥为 authorized_keys 行。
func GenerateEd25519() (KeyPair, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return KeyPair{}, fmt.Errorf("keys: generate ed25519: %w", err)
	}

	pemBlock, err := ssh.MarshalPrivateKey(priv, "")
	if err != nil {
		return KeyPair{}, fmt.Errorf("keys: marshal private key: %w", err)
	}

	sshPub, err := ssh.NewPublicKey(pub)
	if err != nil {
		return KeyPair{}, fmt.Errorf("keys: marshal public key: %w", err)
	}

	pubLine := string(ssh.MarshalAuthorizedKey(sshPub))
	return KeyPair{
		PrivatePEM: pem.EncodeToMemory(pemBlock),
		PublicLine: pubLine,
	}, nil
}
