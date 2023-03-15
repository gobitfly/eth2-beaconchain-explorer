package utils

import (
	"strings"
)

func IsValidEnsDomain(text string) bool {
	return strings.HasSuffix(text, ".eth") && len(text) > 4
}
