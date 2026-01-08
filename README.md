# pg_fancylog

`pg_fancylog` is a query tracer for [`pgx/v5`](https://github.com/jackc/pgx) that uses [`fancylog`](https://github.com/n-ask/fancylog) for beautiful, structured logging of database queries.

## Features

- **Query Tracing**: Automatically logs SQL queries, arguments, and execution duration.
- **Visual Clarity**: Uses `fancylog` to provide stylized and readable log output.
- **Result Tracking**: Logs rows affected for DML statements (INSERT, UPDATE, DELETE) and rows returned for others.
- **Error Logging**: Captures and logs database errors with full query context.
- **Easy Integration**: Simple wrapper functions to create `pgxpool` instances with tracing enabled.

## Installation

```bash
go get gitlab.wg.nask.world/nask/pg_fancylog.git
```

## Usage

### Simple Pool with Tracing

The easiest way to get started is using `NewPoolWithTrace`:

```go
package main

import (
	"context"
	"os"

	"github.com/n-ask/fancylog"
	"gitlab.wg.nask.world/nask/pg_fancylog.git"
)

func main() {
	ctx := context.Background()
	log := fancylog.New(os.Stdout)

	dbURL := "postgres://user:password@localhost:5432/dbname"
	pool, err := pg_fancylog.NewPoolWithTrace(ctx, log, dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	// All queries executed via this pool will now be logged by fancylog
}
```

### Advanced Configuration

If you need to customize the pool configuration, use `NewTracePoolWithConfig`:

```go
package main

import (
	"context"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/n-ask/fancylog"
	"gitlab.wg.nask.world/nask/pg_fancylog.git"
)

func main() {
	ctx := context.Background()
	log := fancylog.New(os.Stdout)

	config, _ := pgxpool.ParseConfig("postgres://user:password@localhost:5432/dbname")
	config.MaxConns = 10
	config.MaxConnIdleTime = 5 * time.Minute

	pool, err := pg_fancylog.NewTracePoolWithConfig(ctx, log, config)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
}
```

## Output Example

When a query is executed, `pg_fancylog` will output a structured log map similar to:

<span><span style="color: #6136D6">&lt;logger-name&gt;</span> <span style="color: #B817B7">[DEBUG]</span> <span style="color: #0913A3">2026-01-08T13:58:19Z</span> <span style="color: #721072">args</span><span style="color: #BBB921">=</span><span style="color: #20AAA9">[admin]</span> <span style="color: #721072">duration</span><span style="color: #BBB921">=</span><span style="color: #20AAA9">500.615µs</span> <span style="color: #721072">rowsReturned</span><span style="color: #BBB921">=</span><span style="color: #20AAA9">1</span> <span style="color: #721072">sql</span><span style="color: #BBB921">=</span><span style="color: #20AAA9">select count(*) > 0 as num_users from users where role = $1; </span></span>

Errors will be logged at the `ERROR` level:

<span><span style="color: #6136D6">&lt;logger-name&gt;</span> <span style="color: red">[ERROR]</span> <span style="color: #0913A3">2026-01-08T13:58:19Z</span> <span style="color: #721072">error</span><span style="color: #BBB921">=</span><span style="color: #20AAA9">relation test_users does not exist</span> <span style="color: #721072">args</span><span style="color: #BBB921">=</span><span style="color: #20AAA9">[admin]</span> <span style="color: #721072">duration</span><span style="color: #BBB921">=</span><span style="color: #20AAA9">500.615µs</span> <span style="color: #721072">rowsReturned</span><span style="color: #BBB921">=</span><span style="color: #20AAA9">1</span> <span style="color: #721072">sql</span><span style="color: #BBB921">=</span><span style="color: #20AAA9">select count(*) > 0 as num_users from test_users where role = $1; </span></span>
