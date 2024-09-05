package main

import (
	"context"
	"net"
	"net/url"
	"os"
	"time"

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
