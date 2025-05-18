package grant

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

func randomPassword(size int) (string, error) {
	bytes := make([]byte, size)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

func Grant(ctx context.Context, db *sql.DB, name string, validFor time.Duration, tables, actions []string) (string, string, error) {
	password, err := randomPassword(10)
	if err != nil {
		return "", "", err
	}

	ts := time.Now().Unix()
	username := fmt.Sprintf("temp_user_%s_%d", name, ts)
	expiry := time.Now().Add(validFor).UTC().Format(time.DateTime)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", "", err
	}

	defer tx.Rollback()

	createRole := fmt.Sprintf(`CREATE ROLE %s LOGIN PASSWORD '%s' VALID UNTIL '%s'`, username, password, expiry)
	if _, err := tx.ExecContext(ctx, createRole); err != nil {
		return "", "", err
	}

	for _, table := range tables {
		grant := fmt.Sprintf(`GRANT %s ON %s TO %s`, strings.ToUpper(strings.Join(actions, ", ")), table, username)
		if _, err := tx.ExecContext(ctx, grant); err != nil {
			return "", "", fmt.Errorf("failed to grant on %s: %w", table, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return "", "", err
	}

	return username, password, nil
}
