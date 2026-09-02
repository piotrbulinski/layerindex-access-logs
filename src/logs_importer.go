package src

import (
	"context"
	"fmt"
	"iter"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
)

const (
	// databaseName is the PostgreSQL database that stores the parsed log lines.
	databaseName = "layerindex"

	// tableName is the table that stores the parsed log lines.
	tableName = "logs"

	// maintenanceDatabase is an always-present database used to create the
	// target database when it does not yet exist.
	maintenanceDatabase = "postgres"
)

// createTableSQL creates the logs table when it does not already exist.
const createTableSQL = `CREATE TABLE IF NOT EXISTS logs (
	id              BIGSERIAL PRIMARY KEY,
	timestamp       TIMESTAMPTZ NOT NULL,
	method          TEXT NOT NULL,
	path            TEXT NOT NULL,
	route           TEXT NOT NULL,
	protocol        TEXT NOT NULL,
	response_code   INTEGER NOT NULL,
	body_bytes_sent INTEGER NOT NULL,
	http_referrer   TEXT NOT NULL,
	http_user_agent TEXT NOT NULL
)`

// logColumns lists the logs table columns in the order used when copying data.
var logColumns = []string{
	"timestamp",
	"method",
	"path",
	"route",
	"protocol",
	"response_code",
	"body_bytes_sent",
	"http_referrer",
	"http_user_agent",
}

// connString builds a PostgreSQL connection string for the given database using
// connection details from the environment, falling back to the defaults used by
// the docker compose stack.
func connString(dbname string) string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("POSTGRES_HOST", "localhost"),
		getEnv("POSTGRES_PORT", "5432"),
		getEnv("POSTGRES_USER", "metabase"),
		getEnv("POSTGRES_PASSWORD", "metabase"),
		dbname,
	)
}

// getEnv returns the value of the environment variable named by key, or fallback
// when the variable is unset or empty.
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// EnsureDatabase creates the target database if it does not already exist.
func EnsureDatabase(ctx context.Context) error {
	conn, err := pgx.Connect(ctx, connString(maintenanceDatabase))
	if err != nil {
		return fmt.Errorf("connecting to maintenance database: %w", err)
	}
	defer conn.Close(ctx)

	var exists bool
	if err := conn.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)", databaseName).Scan(&exists); err != nil {
		return fmt.Errorf("checking for database %q: %w", databaseName, err)
	}

	if exists {
		return nil
	}

	// Database names cannot be parameterized, so the identifier is quoted safely.
	if _, err := conn.Exec(ctx, "CREATE DATABASE "+pgx.Identifier{databaseName}.Sanitize()); err != nil {
		return fmt.Errorf("creating database %q: %w", databaseName, err)
	}

	log.Printf("created database %q", databaseName)
	return nil
}

// ImportLogLines loads every LogLine produced by logLines into the logs table.
// The table is created when missing and truncated before the data is copied in.
// It returns the number of rows written.
func ImportLogLines(ctx context.Context, logLines iter.Seq[LogLine]) (int64, error) {
	conn, err := pgx.Connect(ctx, connString(databaseName))
	if err != nil {
		return 0, fmt.Errorf("connecting to database %q: %w", databaseName, err)
	}
	defer conn.Close(ctx)

	if _, err := conn.Exec(ctx, createTableSQL); err != nil {
		return 0, fmt.Errorf("creating table %q: %w", tableName, err)
	}

	if _, err := conn.Exec(ctx, "TRUNCATE TABLE "+pgx.Identifier{tableName}.Sanitize()); err != nil {
		return 0, fmt.Errorf("truncating table %q: %w", tableName, err)
	}

	next, stop := iter.Pull(logLines)
	defer stop()

	count, err := conn.CopyFrom(ctx, pgx.Identifier{tableName}, logColumns, pgx.CopyFromFunc(func() ([]any, error) {
		logLine, ok := next()
		if !ok {
			return nil, nil
		}

		return []any{
			logLine.Timestamp,
			logLine.Method,
			logLine.Path,
			logLine.Route,
			logLine.Protocol,
			logLine.ResponseCode,
			logLine.BodyBytesSent,
			logLine.HTTPReferrer,
			logLine.HTTPUserAgent,
		}, nil
	}))
	if err != nil {
		return 0, fmt.Errorf("copying log lines into %q: %w", tableName, err)
	}

	return count, nil
}
