package vault

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
)

// ErrWrongPassword 密码错误。
var ErrWrongPassword = errors.New("vault: 密码错误")

// 文件格式常量
const (
	Magic   = "FSSH" // 文件头魔数，用于校验
	Version = 1      // 格式版本号

	SaltLen   = 16                                  // Argon2id 盐长度
	NonceLen  = 12                                  // AES-GCM nonce 长度
	HeaderLen = len(Magic) + 1 + SaltLen + NonceLen // 文件头总长度
)

// Argon2id 参数（平衡安全性和速度）
const (
	Argon2Time    = 3         // 迭代次数
	Argon2Memory  = 64 * 1024 // 64MB 内存
	Argon2Threads = 4         // 并行度
	Argon2KeyLen  = 32        // 派生密钥长度（AES-256 需要 32 字节）
)

// deriveKey 用 Argon2id 从主密码派生 AES-256 密钥。
func deriveKey(password string, salt []byte) []byte {
	return argon2.IDKey([]byte(password), salt, Argon2Time, Argon2Memory, Argon2Threads, Argon2KeyLen)
}

// Encrypt 使用 AES-256-GCM 加密明文。
// 返回文件头 + 密文（含 GCM tag），可直接写入文件。
func Encrypt(plaintext []byte, password string) ([]byte, error) {
	// 生成随机 salt 和 nonce
	salt := make([]byte, SaltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("vault: 生成 salt 失败: %w", err)
	}

	nonce := make([]byte, NonceLen)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("vault: 生成 nonce 失败: %w", err)
	}

	// 派生密钥
	key := deriveKey(password, salt)

	// 创建 AES-GCM 加密器
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("vault: 创建 AES 加密器失败: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("vault: 创建 GCM 失败: %w", err)
	}

	// 加密：GCM 自动附加 tag
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	// 组装文件：magic + version + salt + nonce + ciphertext
	out := make([]byte, 0, HeaderLen+len(ciphertext))
	out = append(out, Magic...)
	out = append(out, Version)
	out = append(out, salt...)
	out = append(out, nonce...)
	out = append(out, ciphertext...)

	return out, nil
}

// Decrypt 解密 vault 文件内容。
// 输入为完整的文件字节（含文件头），返回解密后的明文。
func Decrypt(data []byte, password string) ([]byte, error) {
	if len(data) < HeaderLen {
		return nil, fmt.Errorf("vault: 文件太小，不是有效的 vault 文件")
	}

	// 校验魔数
	if string(data[:len(Magic)]) != Magic {
		return nil, fmt.Errorf("vault: 文件头校验失败，不是 fuckssh vault 文件")
	}

	// 校验版本
	if data[len(Magic)] != Version {
		return nil, fmt.Errorf("vault: 不支持的版本号 %d", data[len(Magic)])
	}

	// 提取 salt 和 nonce
	offset := len(Magic) + 1
	salt := data[offset : offset+SaltLen]
	offset += SaltLen
	nonce := data[offset : offset+NonceLen]
	offset += NonceLen
	ciphertext := data[offset:]

	// 派生密钥（同样的 salt + password = 同样的密钥）
	key := deriveKey(password, salt)

	// 创建 AES-GCM 解密器
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("vault: 创建 AES 解密器失败: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("vault: 创建 GCM 失败: %w", err)
	}

	// 解密：前面已校验文件格式，走到这里说明文件结构完整，GCM 失败大概率是密码错误
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, ErrWrongPassword
	}

	return plaintext, nil
}
