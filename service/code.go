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

// GenerateCode generates a cryptographically secure random
// 6-character short code.
func GenerateCode() string {
	result := make([]byte, codeLength)

	for i := range result {
		n, err := rand.Int(
			rand.Reader,
			big.NewInt(int64(len(charset))),
		)

		if err != nil {
			panic(err)
		}

		result[i] = charset[n.Int64()]
	}

	return string(result)
}

// CreateCode creates a unique short code and stores the
// original URL in PostgreSQL.
//
// PostgreSQL is the source of truth.
// Redis is used as a cache.
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

	// Try multiple times in case of an extremely unlikely
	// short-code collision.
	for attempts := 0; attempts < 10; attempts++ {
		code := GenerateCode()

		var insertedCode string

		err := db.QueryRowContext(
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

		if err == sql.ErrNoRows {
			// Code collision. Generate another code.
			continue
		}

		if err != nil {
			return "", err
		}

		// Store the URL in Redis as a cache.
		//
		// Redis is not the source of truth, so a Redis
		// failure should not make URL creation fail.
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
//
// Lookup order:
//
// 1. Redis
// 2. PostgreSQL
//
// If PostgreSQL is used, the result is placed into Redis
// for subsequent requests.
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

	// ---------------------------------------------------------
	// 1. Try Redis first
	// ---------------------------------------------------------

	if rdb != nil {
		targetURL, err := rdb.Get(
			ctx,
			code,
		).Result()

		if err == nil {
			return targetURL, nil
		}

		// redis.Nil means the key does not exist.
		// For all Redis errors, fall back to PostgreSQL.
	}

	// ---------------------------------------------------------
	// 2. Fall back to PostgreSQL
	// ---------------------------------------------------------

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

	// ---------------------------------------------------------
	// 3. Cache the result in Redis
	// ---------------------------------------------------------

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