package csvutil

import "testing"

func TestCSVFormulaEscaping(t *testing.T) {
	for _, value := range []string{"=SUM(A1:A2)", "+1", "-42", "@IMPORT", " \t=HYPERLINK(\"url\")", "\r\n+1"} {
		if escaped := EscapeCell(value); escaped != "'"+value {
			t.Errorf("formula %q escaped as %q", value, escaped)
		}
	}
	for _, value := range []string{"Team", "Сборная университета", "", "  ", "quoted, team", "A = B"} {
		if EscapeCell(value) != value {
			t.Errorf("ordinary value %q changed", value)
		}
	}
}
