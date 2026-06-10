package vault

import (
	"bytes"
	"errors"
	"testing"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	plaintext := []byte("hello, this is a secret message for testing")
	password := "testpass123"

	// 加密
	encrypted, err := Encrypt(plaintext, password)
	if err != nil {
		t.Fatalf("Encrypt 失败: %v", err)
	}

	// 校验文件头
	if len(encrypted) < HeaderLen {
		t.Fatalf("加密数据太短: %d < %d", len(encrypted), HeaderLen)
	}
	if string(encrypted[:len(Magic)]) != Magic {
		t.Fatalf("魔数不匹配: got %q, want %q", string(encrypted[:len(Magic)]), Magic)
	}
	if encrypted[len(Magic)] != Version {
		t.Fatalf("版本号不匹配: got %d, want %d", encrypted[len(Magic)], Version)
	}

	// 解密
	decrypted, err := Decrypt(encrypted, password)
	if err != nil {
		t.Fatalf("Decrypt 失败: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Fatalf("解密结果不匹配:\n  got:  %q\n  want: %q", decrypted, plaintext)
	}
}

func TestDecryptWrongPassword(t *testing.T) {
	plaintext := []byte("secret data")
	password := "correct123"

	encrypted, err := Encrypt(plaintext, password)
	if err != nil {
		t.Fatalf("Encrypt 失败: %v", err)
	}

	// 用错误密码解密
	_, err = Decrypt(encrypted, "wrongpassword")
	if !errors.Is(err, ErrWrongPassword) {
		t.Fatalf("期望 ErrWrongPassword，got: %v", err)
	}
}

func TestDecryptTamperedData(t *testing.T) {
	plaintext := []byte("secret data")
	password := "testpass123"

	encrypted, err := Encrypt(plaintext, password)
	if err != nil {
		t.Fatalf("Encrypt 失败: %v", err)
	}

	// 篡改密文
	encrypted[len(encrypted)-1] ^= 0xFF

	_, err = Decrypt(encrypted, password)
	if err == nil {
		t.Fatal("篡改后的数据解密应该失败，但成功了")
	}
}

func TestDecryptInvalidMagic(t *testing.T) {
	data := make([]byte, HeaderLen+16)
	copy(data, "XXXX")

	_, err := Decrypt(data, "anypassword")
	if err == nil {
		t.Fatal("无效魔数应该失败，但成功了")
	}
}

func TestDecryptTooShort(t *testing.T) {
	data := []byte("short")

	_, err := Decrypt(data, "anypassword")
	if err == nil {
		t.Fatal("太短的数据应该失败，但成功了")
	}
}

func TestEncryptDifferentOutputs(t *testing.T) {
	plaintext := []byte("same input")
	password := "samepass"

	// 两次加密同样的内容，输出应该不同（因为 salt 和 nonce 随机）
	enc1, err := Encrypt(plaintext, password)
	if err != nil {
		t.Fatalf("第一次 Encrypt 失败: %v", err)
	}

	enc2, err := Encrypt(plaintext, password)
	if err != nil {
		t.Fatalf("第二次 Encrypt 失败: %v", err)
	}

	if bytes.Equal(enc1, enc2) {
		t.Fatal("两次加密同样内容输出相同，salt/nonce 应该是随机的")
	}

	// 但两次都能正确解密
	dec1, err := Decrypt(enc1, password)
	if err != nil {
		t.Fatalf("解密 enc1 失败: %v", err)
	}
	dec2, err := Decrypt(enc2, password)
	if err != nil {
		t.Fatalf("解密 enc2 失败: %v", err)
	}

	if !bytes.Equal(dec1, dec2) || !bytes.Equal(dec1, plaintext) {
		t.Fatal("两次解密结果不一致或与原文不匹配")
	}
}

func TestEncryptEmptyData(t *testing.T) {
	plaintext := []byte{}
	password := "testpass123"

	encrypted, err := Encrypt(plaintext, password)
	if err != nil {
		t.Fatalf("加密空数据失败: %v", err)
	}

	decrypted, err := Decrypt(encrypted, password)
	if err != nil {
		t.Fatalf("解密空数据失败: %v", err)
	}

	if len(decrypted) != 0 {
		t.Fatalf("解密结果应为空，got %d bytes", len(decrypted))
	}
}
