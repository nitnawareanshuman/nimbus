package service

import (
	"strings"
	"testing"
)

func TestGenerateCodeLength(t *testing.T) {
	code := GenerateCode()

	if len(code) != codeLength {
		t.Fatalf("expected code length %d, got %d", codeLength, len(code))
	}
}

func TestGenerateCodeCharset(t *testing.T) {
	code := GenerateCode()

	for _, char := range code {
		if !strings.ContainsRune(charset, char) {
			t.Fatalf("generated code contains invalid character: %q", char)
		}
	}
}

func TestGenerateCodeProducesDifferentValues(t *testing.T) {
	first := GenerateCode()
	second := GenerateCode()

	if first == second {
		t.Fatalf("expected different codes, got %q twice", first)
	}
}