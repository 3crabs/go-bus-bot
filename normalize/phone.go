package normalize

import (
	"fmt"
	"strings"
)

func Phone(phone string) (string, error) {
	result, err := String(phone)
	if err != nil {
		return "", err
	}
	result = strings.ReplaceAll(result, "+", "")
	result = strings.ReplaceAll(result, "-", "")
	if len(result) == 10 {
		result = "7" + result
	}
	return fmt.Sprintf("+7-%s-%s-%s-%s", result[1:4], result[4:7], result[7:9], result[9:11]), nil
}
