package accept_fixture

import (
	"fmt"
	"strings"
)

func validateAndFormatA(input string) (string, error) {
	//art-dupl:accept
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "", fmt.Errorf("input cannot be empty")
	}
	upper := strings.ToUpper(trimmed)
	lower := strings.ToLower(trimmed)
	if len(upper) > 100 {
		return "", fmt.Errorf("input too long")
	}
	result := fmt.Sprintf("%s|%s", upper, lower)
	return result, nil
}
