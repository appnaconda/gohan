package gohan

import (
	"crypto/rand"
	"encoding/base64"
	"os"
	"reflect"
	"runtime"
	"strings"

	"github.com/google/uuid"
)

// Getenv uses the os.Getenv function to retrieve the value of the environment variable named by the key.
// If not value is found, the defaultValue parameter is return.
func Getenv(key string, defaultValue string) string {
	value := os.Getenv(key)

	if value != "" {
		return value
	}

	return defaultValue
}

// Creates a new UUID using google uuid library.
func NewUUID() (string, error) {
	uuid, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}

	return uuid.String(), nil
}

// Gets the string representation of s atruct// Gets function name using reflection
func GetFuncName(i interface{}) string {
	fullFuncName := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()

	nameArray := strings.Split(fullFuncName, ".")
	return nameArray[len(nameArray)-1]
}

// Check if a string is empty
func IsStringEmpty(s string) bool {
	if s == "" {
		return true
	}

	if len(s) == 0 {
		return true
	}

	if len(strings.TrimSpace(s)) == 0 {
		return true
	}

	return false
}

// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

// GenerateRandomString returns a URL-safe, base64 encoded
// securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomString(s int) (string, error) {
	b, err := GenerateRandomBytes(s)
	return base64.URLEncoding.EncodeToString(b), err
}
