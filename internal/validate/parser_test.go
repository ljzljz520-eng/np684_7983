package validate

import (
	"strings"
	"testing"
)

func TestParseCodeLinesNormalizesInput(t *testing.T) {
	codes, issues, err := ParseCodeLines(strings.NewReader(" abc123 \n\nXYZ789\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 0 || len(codes) != 2 || codes[0] != "ABC123" {
		t.Fatalf("codes=%v issues=%v", codes, issues)
	}
	if got := DistinctCodes([]string{"a", "A", "b"}); len(got) != 2 {
		t.Fatalf("distinct=%v", got)
	}
}
