package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/crypto/argon2"
)

// Argon2 配置参数
const (
	memory      uint32 = 64 * 1024 // 64MB 内存使用
	iterations  uint32 = 3         // 迭代次数
	parallelism uint8  = 2         // 并行度
	saltLength  uint32 = 16        // 盐值长度
	keyLength   uint32 = 32        // 密钥长度
)

// GeneratePasswordHash 生成密码哈希值
func GeneratePasswordHash(password string) (string, error) {
	// 生成随机盐值
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	// 使用Argon2id算法生成密钥
	hash := argon2.IDKey(
		[]byte(password),
		salt,
		iterations,
		memory,
		parallelism,
		keyLength,
	)

	// Base64编码
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// 返回格式化的哈希字符串
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		memory,
		iterations,
		parallelism,
		b64Salt,
		b64Hash,
	), nil
}

// VerifyPassword 验证密码
func VerifyPassword(password, encodedHash string) (bool, error) {
	// 解析哈希字符串
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, errors.New("无效的哈希格式")
	}

	// 验证算法
	if parts[1] != "argon2id" {
		return false, errors.New("不支持的哈希算法")
	}

	// 解析参数
	var version int
	var memory, iterations uint32
	var parallelism uint8
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return false, err
	}
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism)
	if err != nil {
		return false, err
	}

	// 解码盐值和哈希值
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}
	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}

	// 使用相同参数生成新哈希
	newHash := argon2.IDKey(
		[]byte(password),
		salt,
		iterations,
		memory,
		parallelism,
		keyLength,
	)

	// 安全比较两个哈希值
	if subtle.ConstantTimeCompare(hash, newHash) == 1 {
		return true, nil
	}
	return false, nil
}

func ValidatePasswordPolicy(password string) error {
	// 1. 长度验证
	if len(password) < 8 {
		return errors.New("密码长度至少8位")
	}
	if len(password) > 64 {
		return errors.New("密码长度不能超过64位")
	}

	// 2. 字符类型验证
	var (
		hasUpper   = false
		hasLower   = false
		hasDigit   = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	// 3. 组合要求验证
	if !hasUpper {
		return errors.New("密码必须包含至少一个大写字母")
	}
	if !hasLower {
		return errors.New("密码必须包含至少一个小写字母")
	}
	if !hasDigit {
		return errors.New("密码必须包含至少一个数字")
	}
	if !hasSpecial {
		return errors.New("密码必须包含至少一个特殊字符")
	}

	// 4. 常见弱密码检查
	weakPasswords := []string{
		"password", "12345678", "qwertyui", "admin123", "letmein",
		"welcome", "passw0rd", "abc12345", "changeme", "iloveyou",
	}
	for _, weak := range weakPasswords {
		if password == weak {
			return errors.New("密码过于简单，请选择更复杂的密码")
		}
	}

	// 5. 重复字符检查（可选）
	if hasRepeatingChars(password, 4) {
		return errors.New("密码包含过多重复字符")
	}

	return nil
}

// hasRepeatingChars 检查是否有过多重复字符
func hasRepeatingChars(s string, maxRepeat int) bool {
	count := 1
	prev := rune(-1)

	for _, char := range s {
		if char == prev {
			count++
			if count > maxRepeat {
				return true
			}
		} else {
			count = 1
			prev = char
		}
	}
	return false
}
