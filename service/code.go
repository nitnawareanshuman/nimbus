package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"math/big"
	"strings"

	"github.com/redis/go-redis/v9"
)

const charset = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

const codeLength = 6

var ErrCodeNotFound = errors.New("short code not found")

// GenerateCode generates a cryptographically secure random 6-character short code.
func GenerateCode() (string, error) {
	result := make([]byte, codeLength)

	for i := range result {
		n, err := rand.Int(
			rand.Reader,
			big.NewInt(int64(len(charset))),
		)
		if err != nil {
			return "", err
		}

		result[i] = charset[n.Int64()]
	}

	return string(result), nil
}

// CreateCode creates a unique short code and stores the original URL in PostgreSQL.
func CreateCode(
	ctx context.Context,
	db *sql.DB,
	rdb *redis.Client,
	targetURL string,
) (string, error) {
	targetURL = strings.TrimSpace(targetURL)

	if targetURL == "" {
		return "", errors.New("target URL cannot be empty")
	}

	if db == nil {
		return "", errors.New("database is required")
	}

	// Try multiple times in case of an extremely unlikely short-code collision.
	for attempts := 0; attempts < 10; attempts++ {
		code, err := GenerateCode()
		if err != nil {
			return "", err
		}

		var insertedCode string

		err = db.QueryRowContext(
			ctx,
			`
			INSERT INTO codes (short_code, original_url)
			VALUES ($1, $2)
			ON CONFLICT (short_code) DO NOTHING
			RETURNING short_code
			`,
			code,
			targetURL,
		).Scan(&insertedCode)

		if errors.Is(err, sql.ErrNoRows) {
			continue
		}

		if err != nil {
			return "", err
		}

		// Redis is only a cache. Redis failure must not fail URL creation.
		if rdb != nil {
			_ = rdb.Set(
				ctx,
				insertedCode,
				targetURL,
				0,
			).Err()
		}

		return insertedCode, nil
	}

	return "", errors.New("failed to generate a unique short code")
}

// GetURL retrieves the original URL for a short code.
// Redis is checked first. PostgreSQL remains the source of truth.
func GetURL(
	ctx context.Context,
	db *sql.DB,
	rdb *redis.Client,
	code string,
) (string, error) {
	code = strings.TrimSpace(code)

	if code == "" {
		return "", ErrCodeNotFound
	}

	if rdb != nil {
		targetURL, err := rdb.Get(ctx, code).Result()

		if err == nil {
			return targetURL, nil
		}

		// On redis.Nil or any Redis failure, fall back to PostgreSQL.
	}

	if db == nil {
		return "", errors.New("database is required")
	}

	var targetURL string

	err := db.QueryRowContext(
		ctx,
		`
		SELECT original_url
		FROM codes
		WHERE short_code = $1
		`,
		code,
	).Scan(&targetURL)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrCodeNotFound
		}

		return "", err
	}

	if rdb != nil {
		_ = rdb.Set(
			ctx,
			code,
			targetURL,
			0,
		).Err()
	}

	return targetURL, nil
}