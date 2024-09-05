package main

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	println("creating client")

	ctx := context.Background()

	dsn := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword("sumup", "sumup"),
		Host:   "localhost:5433",
		Path:   "sumup",
	}

	query := dsn.Query()
	query.Set("search_path", "public")
	query.Set("sslmode", "disable")
	query.Set("sslrootcert", "")
	query.Set("timezone", "UTC")

	dsn.RawQuery = query.Encode()

	poolConfig, err := pgxpool.ParseConfig(dsn.String())
	if err != nil {
		print(err)
		os.Exit(1)
	}

	poolConfig.ConnConfig.DialFunc = func(ctx context.Context, network, addr string) (net.Conn, error) {
		d := &net.Dialer{
			KeepAlive: 5 * time.Minute,
		}
		return d.DialContext(ctx, network, addr)
	}
	poolConfig.ConnConfig.ConnectTimeout = 5 * time.Second
	poolConfig.MaxConnLifetime = 1 * time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.MinConns = 10
	poolConfig.MaxConns = 10

	var tr MyTracer
	poolConfig.ConnConfig.Tracer = tr

	db, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		print(err)
		os.Exit(1)
	}

	defer db.Close()

	err = db.Ping(ctx)
	if err != nil {
		print(err)
		os.Exit(1)
	}

	println("created client")
}

type MyTracer struct{}

func (t MyTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	return ctx
}

func (t MyTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
}

func (t MyTracer) TraceConnectStart(ctx context.Context, data pgx.TraceConnectStartData) context.Context {

	fmt.Println("connecting")

	return ctx
}

func (t MyTracer) TraceConnectEnd(ctx context.Context, data pgx.TraceConnectEndData) {
	connectData, ok := ctx.Value("connect").(*traceConnectData)
	if !ok {
		connectData = &traceConnectData{}
	}

	traceData := &traceConnectEndData{
		startTime:  connectData.startTime,
		connConfig: connectData.connConfig,
		conn:       data.Conn,
		err:        data.Err,
	}

	if traceData.err == nil {
		fmt.Println("connected")
		return
	}

	fmt.Printf("Error while connecting: '%s'", traceData.err.Error())
}

type traceConnectData struct {
	startTime  time.Time
	connConfig *pgx.ConnConfig
}

type traceConnectEndData struct {
	startTime  time.Time
	connConfig *pgx.ConnConfig
	conn       *pgx.Conn
	err        error
}
