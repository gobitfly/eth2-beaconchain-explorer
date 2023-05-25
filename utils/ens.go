package utils

import (
	"regexp"
)

var ENS_ETH_REGEXP = regexp.MustCompile(`^.{3,}\.eth$`)

func IsValidEnsDomain(text string) bool {
	return ENS_ETH_REGEXP.MatchString(text)
}
