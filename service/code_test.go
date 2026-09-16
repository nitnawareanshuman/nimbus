package service

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestGenerateCode(t *testing.T) {
	for i := 0; i < 100; i++ {
		code := GenerateCode()

		if len(code) != codeLength {
			t.Fatalf("expected code length %d, got %d", codeLength, len(code))
		}

		for _, char := range code {
			if !strings.ContainsRune(charset, char) {
				t.Fatalf("generated code contains invalid character %q", char)
			}
		}
	}
}

func TestCreateCodeRejectsBlankURL(t *testing.T) {
	code, err := CreateCode(context.Background(), nil, nil, "   ")

	if err == nil {
		t.Fatal("expected an error for a blank target URL")
	}

	if code != "" {
		t.Fatalf("expected no code, got %q", code)
	}
}

func TestGetURLRejectsBlankCode(t *testing.T) {
	targetURL, err := GetURL(context.Background(), nil, nil, "   ")

	if !errors.Is(err, ErrCodeNotFound) {
		t.Fatalf("expected ErrCodeNotFound, got %v", err)
	}

	if targetURL != "" {
		t.Fatalf("expected no target URL, got %q", targetURL)
	}
}
