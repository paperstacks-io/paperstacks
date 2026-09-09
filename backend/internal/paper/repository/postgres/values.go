package postgres

import (
	"database/sql"
	"time"
)

func timestampValue(value string) (any, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, err
	}
	return parsed, nil
}

func formatTimestamp(value sql.NullTime) string {
	if !value.Valid {
		return ""
	}
	return value.Time.UTC().Format(time.RFC3339Nano)
}

func nullString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func nullInt(value int) any {
	if value == 0 {
		return nil
	}
	return value
}
