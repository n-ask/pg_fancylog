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

$\textcolor{#721072}{\textsf{&lt;logger-name&gt;}}$ $\textcolor{#B817B7}{\textsf{[DEBUG]}}$ $\textcolor{#0913A3}{\textsf{2026-01-08T13:58:19Z}}$ $\textcolor{#721072}{\textsf{args}}$ $\textcolor{#BBB921}{\textsf{=}}$ $\textcolor{#20AAA9}{\textsf{[admin]}}$ $\textcolor{#721072}{\textsf{duration}}$ $\textcolor{#BBB921}{\textsf{=}}$ $\textcolor{#20AAA9}{\textsf{500.615µs}}$ $\textcolor{#721072}{\textsf{rowsReturned}}$ $\textcolor{#BBB921}{\textsf{=}}$ $\textcolor{#20AAA9}{\textsf{1}}$ $\textcolor{#721072}{\textsf{sql}}$ $\textcolor{#BBB921}{\textsf{=}}$ $\textcolor{#20AAA9}{\textsf{&#115;&#8203;elect count&lpar;*&rpar; &gt; as count from users where role = &#65284;1;}}$

Errors will be logged at the `ERROR` level:

$\textcolor{#721072}{\textsf{&lt;logger-name&gt;}}$ $\textcolor{red}{\textsf{[ERROR]}}$ $\textcolor{#0913A3}{\textsf{2026-01-08T13:58:19Z}}$ $\textcolor{#721072}{\textsf{error}}$ $\textcolor{#BBB921}{\textsf{=}}$ $\textcolor{#20AAA9}{\textsf{relation testusers does not exist}}$ $\textcolor{#721072}{\textsf{args}}$ $\textcolor{#BBB921}{\textsf{=}}$ $\textcolor{#20AAA9}{\textsf{[admin]}}$ $\textcolor{#721072}{\textsf{duration}}$ $\textcolor{#BBB921}{\textsf{=}}$ $\textcolor{#20AAA9}{\textsf{500.615µs}}$ $\textcolor{#721072}{\textsf{rowsReturned}}$ $\textcolor{#BBB921}{\textsf{=}}$ $\textcolor{#20AAA9}{\textsf{1}}$ $\textcolor{#721072}{\textsf{sql}}$ $\textcolor{#BBB921}{\textsf{=}}$ $\textcolor{#20AAA9}{\textsf{&#115;&#8203;elect count&lpar;*&rpar; &gt; as count from testusers where role = &#65284;1;}}$
