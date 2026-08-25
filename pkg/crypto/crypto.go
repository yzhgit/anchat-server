package crypto

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword encrypts password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword verifies password
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ValidatePasswordStrength validates password strength
// Requirements: 8-32 characters, contains numbers and letters
func ValidatePasswordStrength(password string) bool {
	if len(password) < 8 || len(password) > 32 {
		return false
	}

	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)

	return hasNumber && hasLetter
}

// MD5 calculates MD5 hash
func MD5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

// GenerateRandomString generates random string
func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}

// GenerateVerifyCode generates verification code
func GenerateVerifyCode(length int) (string, error) {
	if length <= 0 {
		length = 6
	}

	const digits = "0123456789"
	code := make([]byte, length)

	for i := 0; i < length; i++ {
		b := make([]byte, 1)
		if _, err := rand.Read(b); err != nil {
			return "", err
		}
		code[i] = digits[int(b[0])%len(digits)]
	}

	return string(code), nil
}

// GenerateQRCodeToken generates QR code token
func GenerateQRCodeToken(userID string) (string, error) {
	randomStr, err := GenerateRandomString(16)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s_%s", userID, randomStr), nil
}
