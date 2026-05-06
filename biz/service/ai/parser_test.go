package ai

import (
	"strings"
	"testing"
)

func TestExtractJSONHandlesBracesInsideString(t *testing.T) {
	raw := `prefix {"message":"keep { this } text","items":[{"id":1}]} suffix`

	got := ExtractJSON(raw)
	want := `{"message":"keep { this } text","items":[{"id":1}]}`
	if got != want {
		t.Fatalf("ExtractJSON() = %q, want %q", got, want)
	}
}

func TestDecodeJSON(t *testing.T) {
	var out struct {
		Value string `json:"value"`
	}

	if !DecodeJSON("```json\n{\"value\":\"ok\"}\n```", &out) {
		t.Fatal("DecodeJSON returned false")
	}
	if out.Value != "ok" {
		t.Fatalf("Value = %q, want ok", out.Value)
	}
}

func TestStripFencedCode(t *testing.T) {
	got := StripFencedCode("```spec\nName: demo\n```", "spec", "rpm")
	if got != "Name: demo" {
		t.Fatalf("StripFencedCode() = %q", got)
	}
}

func TestRedactSecrets(t *testing.T) {
	input := "api_key=secret-value password:abc token = xyz Authorization: Bearer bearer-token sk-abcdefghijklmnopqrstuvwxyz"
	got := RedactSecrets(input)

	for _, leaked := range []string{"secret-value", "abc", "xyz", "bearer-token", "sk-abcdefghijklmnopqrstuvwxyz"} {
		if strings.Contains(got, leaked) {
			t.Fatalf("RedactSecrets leaked %q in %q", leaked, got)
		}
	}
}

func TestClampText(t *testing.T) {
	got := ClampText("abcdefghijklmnopqrstuvwxyz", 10)
	if !strings.Contains(got, "[truncated by AI input budget]") {
		t.Fatalf("ClampText() = %q, want truncation marker", got)
	}
}

func TestBuildLineDiff(t *testing.T) {
	diff := BuildLineDiff("a\nb\nc", "a\nB\nc\nd")

	if diff.AddedLines != 2 || diff.RemovedLines != 1 {
		t.Fatalf("diff counts = +%d -%d", diff.AddedLines, diff.RemovedLines)
	}
	if diff.RiskLevel != "low" {
		t.Fatalf("risk = %q, want low", diff.RiskLevel)
	}
}
