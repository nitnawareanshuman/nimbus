package service

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestGenerateCode(t *testing.T) {
	for i := 0; i < 100; i++ {
		code, err := GenerateCode()
		if err != nil {
			t.Fatalf("GenerateCode returned error: %v", err)
		}

		if len(code) != codeLength {
			t.Fatalf(
				"expected code length %d, got %d",
				codeLength,
				len(code),
			)
		}

		for _, char := range code {
			if !strings.ContainsRune(charset, char) {
				t.Fatalf(
					"generated code contains invalid character %q",
					char,
				)
			}
		}
	}
}

func TestGenerateCodeProducesDifferentValues(t *testing.T) {
	first, err := GenerateCode()
	if err != nil {
		t.Fatal(err)
	}

	second, err := GenerateCode()
	if err != nil {
		t.Fatal(err)
	}

	if first == second {
		t.Fatalf(
			"expected different random codes, both were %q",
			first,
		)
	}
}

func TestCreateCodeRejectsBlankURL(t *testing.T) {
	code, err := CreateCode(
		context.Background(),
		nil,
		nil,
		"   ",
	)

	if err == nil {
		t.Fatal("expected an error for a blank target URL")
	}

	if code != "" {
		t.Fatalf("expected no code, got %q", code)
	}
}

func TestCreateCodeRequiresDatabase(t *testing.T) {
	code, err := CreateCode(
		context.Background(),
		nil,
		nil,
		"https://example.com",
	)

	if err == nil {
		t.Fatal("expected an error when database is nil")
	}

	if code != "" {
		t.Fatalf("expected no code, got %q", code)
	}
}

func TestGetURLRejectsBlankCode(t *testing.T) {
	targetURL, err := GetURL(
		context.Background(),
		nil,
		nil,
		"   ",
	)

	if !errors.Is(err, ErrCodeNotFound) {
		t.Fatalf(
			"expected ErrCodeNotFound, got %v",
			err,
		)
	}

	if targetURL != "" {
		t.Fatalf(
			"expected no target URL, got %q",
			targetURL,
		)
	}
}

func TestGetURLRequiresDatabase(t *testing.T) {
	targetURL, err := GetURL(
		context.Background(),
		nil,
		nil,
		"abc123",
	)

	if err == nil {
		t.Fatal("expected an error when database is nil")
	}

	if targetURL != "" {
		t.Fatalf(
			"expected no target URL, got %q",
			targetURL,
		)
	}
}