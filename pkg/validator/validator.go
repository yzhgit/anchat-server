package validator

import (
	"fmt"
	"regexp"
	"strings"

	phonemunbers "github.com/nyaruka/phonenumbers/v2"
)

// ValidatePhone validates phone number. It accepts either a canonical E164
// string (e.g. "+8613800138000") or a bare national number (e.g.
// "13800138000"); in the latter case it normalises via NormalizePhone and
// validates the resulting E164 form.
func ValidatePhone(phone string) bool {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return false
	}
	if phone[0] == '+' {
		pn, err := phonemunbers.Parse(phone, "CN")
		return err == nil && phonemunbers.IsValidNumber(pn)
	}
	e164, err := NormalizePhone(phone)
	return err == nil && e164 != ""
}

// NormalizePhone normalises a phone number to its canonical E164 form.
// It accepts bare national numbers ("13800138000"), numbers with a
// country-code prefix ("8613800138000"), and numbers with an international
// prefix ("+8613800138000" or "+1 650-555-0199"). Numbers without an
// explicit international prefix ("+") are interpreted as Chinese
// (defaultRegion="CN"), because the API surface carries only a
// phone_number field and no region.
//
// Returns the E164 representation on success, or an error describing why
// the input could not be parsed.
func NormalizePhone(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("phone number is empty")
	}

	// If the input already carries an explicit international prefix, parse
	// it directly — the default region is irrelevant in that case.
	if len(raw) > 0 && raw[0] == '+' {
		pn, err := phonemunbers.Parse(raw, "CN")
		if err != nil {
			return "", fmt.Errorf("invalid phone number %q: %w", raw, err)
		}
		if !phonemunbers.IsValidNumber(pn) {
			return "", fmt.Errorf("invalid phone number %q", raw)
		}
		return phonemunbers.Format(pn, phonemunbers.E164), nil
	}

	// Bare digits / optional country-code prefix: normalise non-digits and
	// let libphonenumber handle the rest. Default region is CN.
	digits := phonemunbers.NormalizeDigitsOnly(raw)
	if digits == "" {
		return "", fmt.Errorf("phone number %q contains no digits", raw)
	}
	pn, err := phonemunbers.Parse(digits, "CN")
	if err != nil {
		return "", fmt.Errorf("invalid phone number %q: %w", raw, err)
	}
	if !phonemunbers.IsValidNumber(pn) {
		return "", fmt.Errorf("invalid phone number %q", raw)
	}
	return phonemunbers.Format(pn, phonemunbers.E164), nil
}

// ValidateEmail validates email
func ValidateEmail(email string) bool {
	if email == "" {
		return false
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// ValidateNickname validates nickname
func ValidateNickname(nickname string) bool {
	if nickname == "" {
		return false
	}

	// 1-20 characters
	if len([]rune(nickname)) < 1 || len([]rune(nickname)) > 20 {
		return false
	}

	return true
}

// ContainsSensitiveWords checks if text contains sensitive words
func ContainsSensitiveWords(text string) bool {
	// Simple sensitive word list, should actually be read from database or config file
	sensitiveWords := []string{"admin", "system", "customer service", "official"}

	lowerText := strings.ToLower(text)
	for _, word := range sensitiveWords {
		if strings.Contains(lowerText, strings.ToLower(word)) {
			return true
		}
	}

	return false
}

// ValidateDeviceType validates device type
func ValidateDeviceType(deviceType string) bool {
	validTypes := []string{"iOS", "Android", "Web", "PC"}
	for _, t := range validTypes {
		if deviceType == t {
			return true
		}
	}
	return false
}

// ValidateGender validates gender
func ValidateGender(gender int) bool {
	return gender >= 0 && gender <= 2
}

// SanitizeString sanitizes string
func SanitizeString(str string) string {
	return strings.TrimSpace(str)
}
