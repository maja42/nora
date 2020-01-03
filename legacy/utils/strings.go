package utils

import "strings"

func Indent(s string) string {
	if len(s) == 0 {
		return s
	}
	return "    " + strings.Replace(s, "\n", "\n    ", -1)
}
