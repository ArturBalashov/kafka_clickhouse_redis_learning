package clickhouse

import (
	"github.com/jmoiron/sqlx"
)

func NewClickHouseConn() (*sqlx.DB, error) {
	clickhouseConn, err := sqlx.Open("clickhouse", "tcp://127.0.0.1:9000?password=еуые")
	if err != nil {
		return nil, err
	}
	if _, err = clickhouseConn.Exec(`
		CREATE TABLE IF NOT EXISTS users_log (
			info         String,
			action_time  DateTime
		) engine=Memory
	`); err != nil {
		return nil, err
	}

	return clickhouseConn, nil
}
