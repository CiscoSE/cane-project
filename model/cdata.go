package model

import "strings"

// IsCDATA Function
func IsCDATA(s string) bool {
	startsWith := strings.HasPrefix(s, "<![CDATA[")
	endsWith := strings.HasSuffix(s, "]]>")

	if startsWith && endsWith {
		return true
	}

	return false
}
