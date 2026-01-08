package pg_fancylog

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/n-ask/fancylog"
)

var (
	re_leadclose_whtsp = regexp.MustCompile(`^[\s\p{Zs}]+|[\s\p{Zs}]+$`)
	re_inside_whtsp    = regexp.MustCompile(`[\s\p{Zs}]{2,}`)
)

// LoggingQueryTracer implements the pgx.QueryTracer interface using fancylog.
type LoggingQueryTracer struct {
	logger fancylog.FancyLogger
}

// NewLoggingQueryTracer creates a new LoggingQueryTracer with the provided fancylog.FancyLogger.
func NewLoggingQueryTracer(logger fancylog.FancyLogger) *LoggingQueryTracer {
	return &LoggingQueryTracer{logger: logger}
}

type sqlTracer struct {
	data    *pgx.TraceQueryStartData
	startTs time.Time
}

// GetSQL returns the SQL query string with normalized whitespace.
func (t *sqlTracer) GetSQL() string {
	if t.data != nil {
		final := re_leadclose_whtsp.ReplaceAllString(t.data.SQL, "")
		final = re_inside_whtsp.ReplaceAllString(final, " ")
		return final
	}
	return ""
}

// TraceQueryStart is called at the beginning of a query.
func (l *LoggingQueryTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	return context.WithValue(ctx, "fancylog", &sqlTracer{data: &data, startTs: time.Now()})
}

// TraceQueryEnd is called at the end of a query and logs the query details.
func (l *LoggingQueryTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	start, ok := ctx.Value("fancylog").(*sqlTracer)
	finish := time.Now()
	args := make(map[string]any)
	if ok {
		args["duration"] = finish.Sub(start.startTs).String()
		args["args"] = start.data.Args
		args["sql"] = start.GetSQL()
		if data.Err != nil {
			args["error"] = data.Err.Error()
			l.logger.ErrorMap(args)
		} else {
			if data.CommandTag.Delete() || data.CommandTag.Insert() || data.CommandTag.Update() {
				args["rowAffected"] = data.CommandTag.RowsAffected()
			} else {
				args["rowsReturned"] = data.CommandTag.RowsAffected()
			}
			l.logger.DebugMap(args)
		}
	}
}

// NewPoolWithTrace creates a new pgxpool.Pool with the LoggingQueryTracer configured.
func NewPoolWithTrace(ctx context.Context, log fancylog.FancyLogger, databaseURL string) (*pgxpool.Pool, error) {
	connConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}
	connConfig.ConnConfig.Tracer = NewLoggingQueryTracer(log)

	pool, err := pgxpool.NewWithConfig(ctx, connConfig)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	return pool, nil
}

// NewTracePoolWithConfig creates a new pgxpool.Pool with the provided config and LoggingQueryTracer configured.
// It also logs the pool configuration details.
func NewTracePoolWithConfig(ctx context.Context, log fancylog.FancyLogger, options *pgxpool.Config) (*pgxpool.Pool, error) {
	if options.ConnConfig.Tracer != nil {
		return nil, fmt.Errorf("tracer already set, cannot set fancylog tracer")
	}
	options.ConnConfig.Tracer = NewLoggingQueryTracer(log)
	log.InfoMap(map[string]any{
		"message":               "fancy tracer pool options",
		"MaxConns":              options.MaxConns,
		"MinConns":              options.MinConns,
		"MaxConnLifetime":       options.MaxConnLifetime,
		"MaxConnIdleTime":       options.MaxConnIdleTime,
		"HealthCheckPeriod":     options.HealthCheckPeriod,
		"MinIdleConns":          options.MinIdleConns,
		"MaxConnLifetimeJitter": options.MaxConnLifetimeJitter,
	})

	pool, err := pgxpool.NewWithConfig(ctx, options)

	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	return pool, nil
}
