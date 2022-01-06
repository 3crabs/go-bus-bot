package normalize

import (
	"errors"
	"strings"
)

func String(s string) (string, error) {
	result := strings.TrimSpace(s)
	if result == "" {
		return "", errors.New("empty string")
	}
	return result, nil
}
