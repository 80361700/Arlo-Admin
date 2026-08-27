package security

import (
	"fmt"
	"regexp"
	"unicode/utf8"
)

var (
	hasLetter = regexp.MustCompile(`[A-Za-z]`)
	hasDigit  = regexp.MustCompile(`[0-9]`)
)

// ValidatePassword 按策略校验密码
func ValidatePassword(password string, minLength int, requireComplexity bool) error {
	if minLength <= 0 {
		minLength = 6
	}
	n := utf8.RuneCountInString(password)
	if n < minLength {
		return fmt.Errorf("密码长度不能少于 %d 位", minLength)
	}
	if n > 32 {
		return fmt.Errorf("密码长度不能超过 32 位")
	}
	if requireComplexity {
		if !hasLetter.MatchString(password) || !hasDigit.MatchString(password) {
			return fmt.Errorf("密码需同时包含字母和数字")
		}
	}
	return nil
}
