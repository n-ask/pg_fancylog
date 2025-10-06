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

type LoggingQueryTracer struct {
	logger fancylog.FancyLogger
}

func NewLoggingQueryTracer(logger fancylog.FancyLogger) *LoggingQueryTracer {
	return &LoggingQueryTracer{logger: logger}
}

type sqlTracer struct {
	data    *pgx.TraceQueryStartData
	startTs time.Time
}

func (t *sqlTracer) GetSQL() string {
	if t.data != nil {
		final := re_leadclose_whtsp.ReplaceAllString(t.data.SQL, "")
		final = re_inside_whtsp.ReplaceAllString(final, " ")
		return final
	}
	return ""
}

func (l *LoggingQueryTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	return context.WithValue(ctx, "fancylog", &sqlTracer{data: &data, startTs: time.Now()})
}

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
			if data.CommandTag.Delete() || data.CommandTag.Insert() || data.CommandTag.Insert() {
				args["rowAffected"] = data.CommandTag.RowsAffected()
			} else {
				args["rowsReturned"] = data.CommandTag.RowsAffected()
			}
			l.logger.DebugMap(args)
		}
	}
}

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

func NewTracePoolWithConfig(ctx context.Context, log fancylog.FancyLogger, databaseURL string, options *pgxpool.Config) (*pgxpool.Pool, error) {
	var err error
	if options == nil {
		options, err = pgxpool.ParseConfig(databaseURL)
		if err != nil {
			return nil, err
		}
		options.ConnConfig.Tracer = NewLoggingQueryTracer(log)
	} else {
		if options.ConnConfig.Tracer != nil {
			return nil, fmt.Errorf("tracer already set, cannot set fancylog tracer")
		}
		options.ConnConfig.Tracer = NewLoggingQueryTracer(log)
	}
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
