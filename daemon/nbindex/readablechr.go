package nbindex

import (
	"regexp"
	"strings"
)

func IsReableByte(b byte) bool {
	return ' ' /*0x20*/ <= b && b <= '~' /*0x7f*/
}

var (
	alphanumeric = regexp.MustCompile("[a-zA-Z0-9_\\.]+")
)

func toAlphaNumeric(org string) []string {
	return alphanumeric.FindAllString(org, -1)
}

func extractAlphaNumeric(org string) string {
	return strings.Join(toAlphaNumeric(org), " ")
}
