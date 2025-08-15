package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"
)

// GenerateRandomString generates a random string of the specified length
func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// StringPtr returns a pointer to the string value
func StringPtr(s string) *string {
	return &s
}

// BoolPtr returns a pointer to the bool value
func BoolPtr(b bool) *bool {
	return &b
}

// IntPtr returns a pointer to the int value
func IntPtr(i int) *int {
	return &i
}

// TimePtr returns a pointer to the time value
func TimePtr(t time.Time) *time.Time {
	return &t
}

// StringValue returns the value of a string pointer or empty string if nil
func StringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// BoolValue returns the value of a bool pointer or false if nil
func BoolValue(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

// IntValue returns the value of an int pointer or 0 if nil
func IntValue(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

// IsValidEmail validates email format
func IsValidEmail(email string) bool {
	const emailRegex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailRegex)
	return re.MatchString(email)
}

// TruncateString truncates a string to the specified length
func TruncateString(str string, length int) string {
	if len(str) <= length {
		return str
	}
	return str[:length] + "..."
}

// SliceContains checks if a slice contains a specific item
func SliceContains[T comparable](slice []T, item T) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// RemoveFromSlice removes an item from a slice
func RemoveFromSlice[T comparable](slice []T, item T) []T {
	var result []T
	for _, v := range slice {
		if v != item {
			result = append(result, v)
		}
	}
	return result
}

// UniqueSlice returns a slice with unique elements
func UniqueSlice[T comparable](slice []T) []T {
	keys := make(map[T]bool)
	var result []T
	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}
	return result
}

// ToSnakeCase converts a string to snake_case
func ToSnakeCase(str string) string {
	re := regexp.MustCompile("([a-z0-9])([A-Z])")
	snake := re.ReplaceAllString(str, "${1}_${2}")
	return strings.ToLower(snake)
}

// ToCamelCase converts a string to camelCase
func ToCamelCase(str string) string {
	parts := strings.Split(strings.ToLower(str), "_")
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

// ToPascalCase converts a string to PascalCase
func ToPascalCase(str string) string {
	camel := ToCamelCase(str)
	if len(camel) > 0 {
		return strings.ToUpper(camel[:1]) + camel[1:]
	}
	return camel
}

// IsZeroValue checks if a value is zero value
func IsZeroValue(v interface{}) bool {
	return reflect.ValueOf(v).IsZero()
}

// Paginate calculates pagination values
func Paginate(page, limit int, total int64) (offset int, pages int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	
	offset = (page - 1) * limit
	pages = int((total + int64(limit) - 1) / int64(limit))
	
	return offset, pages
}

// FormatDuration formats a duration in a human-readable way
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	}
	return fmt.Sprintf("%.1fh", d.Hours())
}

// Retry executes a function with retries
func Retry(attempts int, delay time.Duration, fn func() error) error {
	var err error
	for i := 0; i < attempts; i++ {
		if err = fn(); err == nil {
			return nil
		}
		if i < attempts-1 {
			time.Sleep(delay)
		}
	}
	return err
}

// CoalesceString returns the first non-empty string
func CoalesceString(strings ...string) string {
	for _, s := range strings {
		if s != "" {
			return s
		}
	}
	return ""
}