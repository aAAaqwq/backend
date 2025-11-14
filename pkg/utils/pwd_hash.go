package utils

import (
	"errors"
	"fmt"
	"regexp"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

const (
	// DefaultBcryptCost bcrypt默认成本因子（推荐值：10-12）
	DefaultBcryptCost = 12
	// MinBcryptCost bcrypt最小成本因子
	MinBcryptCost = 4
	// MaxBcryptCost bcrypt最大成本因子
	MaxBcryptCost = 31
)

var (
	// bcryptCost 当前使用的bcrypt成本因子
	bcryptCost = DefaultBcryptCost
)

// SetBcryptCost 设置bcrypt成本因子
// cost值越大，哈希计算越慢，但更安全（范围：4-31）
func SetBcryptCost(cost int) error {
	if cost < MinBcryptCost || cost > MaxBcryptCost {
		return fmt.Errorf("bcrypt cost必须在%d-%d之间", MinBcryptCost, MaxBcryptCost)
	}
	bcryptCost = cost
	return nil
}

// GetBcryptCost 获取当前bcrypt成本因子
func GetBcryptCost() int {
	return bcryptCost
}

// CreatePasswordHash 创建密码哈希
// 使用bcrypt算法对密码进行哈希处理
func CreatePasswordHash(password string) (string, error) {
	if password == "" {
		return "", errors.New("密码不能为空")
	}

	// 生成bcrypt哈希
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("生成密码哈希失败: %w", err)
	}

	return string(hash), nil
}

// CheckPasswordHash 检查密码是否匹配哈希值
// 返回true表示密码匹配，false表示不匹配
func CheckPasswordHash(password, hash string) bool {
	if password == "" || hash == "" {
		return false
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ValidatePasswordStrength 验证密码强度
// 返回错误信息，如果密码符合要求则返回nil
func ValidatePasswordStrength(password string, minLength int) error {
	if minLength <= 0 {
		minLength = 8 // 默认最小长度
	}

	if len(password) < minLength {
		return fmt.Errorf("密码长度至少需要%d个字符", minLength)
	}

	// 检查是否包含数字
	hasDigit := false
	// 检查是否包含小写字母
	hasLower := false
	// 检查是否包含大写字母
	hasUpper := false
	// 检查是否包含特殊字符
	hasSpecial := false

	for _, char := range password {
		switch {
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	var missing []string
	if !hasDigit {
		missing = append(missing, "数字")
	}
	if !hasLower {
		missing = append(missing, "小写字母")
	}
	if !hasUpper {
		missing = append(missing, "大写字母")
	}
	if !hasSpecial {
		missing = append(missing, "特殊字符")
	}

	if len(missing) > 0 {
		return fmt.Errorf("密码必须包含: %v", missing)
	}

	return nil
}

// ValidatePasswordStrengthSimple 简单密码强度验证（仅检查长度和基本字符类型）
// 返回错误信息，如果密码符合要求则返回nil
func ValidatePasswordStrengthSimple(password string, minLength int) error {
	if minLength <= 0 {
		minLength = 6 // 默认最小长度
	}

	if len(password) < minLength {
		return fmt.Errorf("密码长度至少需要%d个字符", minLength)
	}

	if len(password) > 128 {
		return errors.New("密码长度不能超过128个字符")
	}

	// 检查是否包含至少一个字母和一个数字
	hasLetter := false
	hasDigit := false

	for _, char := range password {
		if unicode.IsLetter(char) {
			hasLetter = true
		}
		if unicode.IsDigit(char) {
			hasDigit = true
		}
		if hasLetter && hasDigit {
			break
		}
	}

	if !hasLetter {
		return errors.New("密码必须包含至少一个字母")
	}

	if !hasDigit {
		return errors.New("密码必须包含至少一个数字")
	}

	return nil
}

// IsCommonPassword 检查是否为常见弱密码
// 返回true表示是常见弱密码
func IsCommonPassword(password string) bool {
	commonPasswords := []string{
		"123456", "password", "123456789", "12345678", "12345",
		"1234567", "1234567890", "qwerty", "abc123", "111111",
		"123123", "admin", "letmein", "welcome", "monkey",
		"1234", "password1", "qwerty123", "000000", "123321",
	}

	passwordLower := toLower(password)
	for _, common := range commonPasswords {
		if passwordLower == common {
			return true
		}
	}

	return false
}

// toLower 将字符串转换为小写（处理中文等）
func toLower(s string) string {
	runes := []rune(s)
	for i, r := range runes {
		runes[i] = unicode.ToLower(r)
	}
	return string(runes)
}

// HashPasswordWithCost 使用指定的成本因子创建密码哈希
func HashPasswordWithCost(password string, cost int) (string, error) {
	if password == "" {
		return "", errors.New("密码不能为空")
	}

	if cost < MinBcryptCost || cost > MaxBcryptCost {
		return "", fmt.Errorf("bcrypt cost必须在%d-%d之间", MinBcryptCost, MaxBcryptCost)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", fmt.Errorf("生成密码哈希失败: %w", err)
	}

	return string(hash), nil
}

// NeedsRehash 检查哈希是否需要重新计算（成本因子升级）
// 如果当前哈希使用的成本因子低于指定值，返回true
func NeedsRehash(hash string, minCost int) bool {
	if hash == "" {
		return true
	}

	// 解析bcrypt哈希获取成本因子
	cost, err := bcrypt.Cost([]byte(hash))
	if err != nil {
		return true // 无法解析，需要重新哈希
	}

	return cost < minCost
}

// ValidateAndHashPassword 验证密码强度并创建哈希
// 这是一个便捷函数，结合了密码验证和哈希创建
func ValidateAndHashPassword(password string, minLength int, requireStrong bool) (string, error) {
	// 验证密码强度
	if requireStrong {
		if err := ValidatePasswordStrength(password, minLength); err != nil {
			return "", err
		}
	} else {
		if err := ValidatePasswordStrengthSimple(password, minLength); err != nil {
			return "", err
		}
	}

	// 检查是否为常见弱密码
	if IsCommonPassword(password) {
		return "", errors.New("密码过于简单，请使用更复杂的密码")
	}

	// 创建哈希
	return CreatePasswordHash(password)
}

// CheckPasswordWithRehash 检查密码并返回是否需要重新哈希
// 返回: (是否匹配, 是否需要重新哈希, 错误)
func CheckPasswordWithRehash(password, hash string) (bool, bool, error) {
	if password == "" || hash == "" {
		return false, false, errors.New("密码或哈希不能为空")
	}

	// 检查密码是否匹配
	match := CheckPasswordHash(password, hash)
	if !match {
		return false, false, nil
	}

	// 检查是否需要重新哈希（成本因子升级）
	needsRehash := NeedsRehash(hash, bcryptCost)

	return true, needsRehash, nil
}

// PasswordStrengthScore 计算密码强度分数（0-100）
// 分数越高表示密码越强
func PasswordStrengthScore(password string) int {
	if password == "" {
		return 0
	}

	score := 0

	// 长度评分（最多30分）
	length := len(password)
	if length >= 12 {
		score += 30
	} else if length >= 8 {
		score += 20
	} else if length >= 6 {
		score += 10
	}

	// 字符类型评分（最多40分）
	hasDigit := false
	hasLower := false
	hasUpper := false
	hasSpecial := false

	for _, char := range password {
		if unicode.IsDigit(char) {
			hasDigit = true
		} else if unicode.IsLower(char) {
			hasLower = true
		} else if unicode.IsUpper(char) {
			hasUpper = true
		} else if unicode.IsPunct(char) || unicode.IsSymbol(char) {
			hasSpecial = true
		}
	}

	charTypeCount := 0
	if hasDigit {
		charTypeCount++
	}
	if hasLower {
		charTypeCount++
	}
	if hasUpper {
		charTypeCount++
	}
	if hasSpecial {
		charTypeCount++
	}

	score += charTypeCount * 10 // 每种字符类型10分

	// 复杂度评分（最多30分）
	// 检查是否有重复字符模式
	hasPattern := checkPasswordPattern(password)
	if !hasPattern {
		score += 15
	}

	// 检查是否不是常见密码
	if !IsCommonPassword(password) {
		score += 15
	}

	// 确保分数在0-100范围内
	if score > 100 {
		score = 100
	}

	return score
}

// checkPasswordPattern 检查密码是否有明显的重复模式
func checkPasswordPattern(password string) bool {
	// 检查连续相同字符
	for i := 0; i < len(password)-2; i++ {
		if password[i] == password[i+1] && password[i] == password[i+2] {
			return true
		}
	}

	// 检查连续数字或字母
	digitPattern := regexp.MustCompile(`012|123|234|345|456|567|678|789|890`)
	letterPattern := regexp.MustCompile(`abc|bcd|cde|def|efg|fgh|ghi|hij|ijk|jkl|klm|lmn|mno|nop|opq|pqr|qrs|rst|stu|tuv|uvw|vwx|wxy|xyz`)

	if digitPattern.MatchString(toLower(password)) || letterPattern.MatchString(toLower(password)) {
		return true
	}

	return false
}

// GetPasswordStrengthLevel 获取密码强度等级
// 返回: "weak", "medium", "strong", "very_strong"
func GetPasswordStrengthLevel(password string) string {
	score := PasswordStrengthScore(password)

	if score >= 80 {
		return "very_strong"
	} else if score >= 60 {
		return "strong"
	} else if score >= 40 {
		return "medium"
	}
	return "weak"
}
