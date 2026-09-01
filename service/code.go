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

	// Try multiple times in the extremely unlikely event of a generated-code collision.
	for attempts := 0; attempts < 10; attempts++ {
		code := GenerateCode()

		_, err := db.ExecContext(
			ctx,
			`
			INSERT INTO urls (code, target_url)
			VALUES ($1, $2)
			ON CONFLICT (code) DO NOTHING
			`,
			code,
			targetURL,
		)

		if err != nil {
			return "", err
		}

		// Check whether our row was actually inserted.
		var storedURL string

		err = db.QueryRowContext(
			ctx,
			`
			SELECT target_url
			FROM urls
			WHERE code = $1
			`,
			code,
		).Scan(&storedURL)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}

			return "", err
		}

		// If the stored URL is ours, the insert succeeded.
		if storedURL == targetURL {
			if rdb != nil {
				_ = rdb.Set(
					ctx,
					code,
					targetURL,
					0,
				).Err()
			}

			return code, nil
		}
	}

	return "", errors.New("failed to generate a unique short code")
}

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
		targetURL, err := rdb.Get(
			ctx,
			code,
		).Result()

		if err == nil {
			return targetURL, nil
		}

		if err != redis.Nil {
			// Continue to PostgreSQL.
		}
	}

	var targetURL string

	err := db.QueryRowContext(
		ctx,
		`
		SELECT target_url
		FROM urls
		WHERE code = $1
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