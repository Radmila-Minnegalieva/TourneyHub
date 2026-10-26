// Package csvutil защищает текстовые ячейки экспорта от интерпретации как формул.
package csvutil

import "strings"

func EscapeCell(value string) string {
	trimmed := strings.TrimLeft(value, " \t\r\n")
	if trimmed != "" && strings.ContainsRune("=+-@", rune(trimmed[0])) {
		return "'" + value
	}
	return value
}
