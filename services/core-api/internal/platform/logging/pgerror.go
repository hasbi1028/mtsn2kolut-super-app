package logging

import (
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5/pgconn"
)

func PgErrorAttrs(err error) []slog.Attr {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return nil
	}
	attrs := []slog.Attr{
		slog.String("error_kind", "postgres_constraint"),
		slog.String("sqlstate", pgErr.Code),
	}
	if pgErr.ConstraintName != "" {
		attrs = append(attrs, slog.String("constraint", pgErr.ConstraintName))
	}
	if pgErr.TableName != "" {
		attrs = append(attrs, slog.String("table", pgErr.TableName))
	}
	if pgErr.ColumnName != "" {
		attrs = append(attrs, slog.String("column", pgErr.ColumnName))
	}
	if pgErr.Severity != "" {
		attrs = append(attrs, slog.String("severity", pgErr.Severity))
	}
	if pgErr.Message != "" {
		attrs = append(attrs, slog.String("pg_message", pgErr.Message))
	}
	return attrs
}
