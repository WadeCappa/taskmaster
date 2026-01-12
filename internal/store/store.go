package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func Call(
	ctx context.Context,
	postgresUrl string,
	query func(*pgx.Conn) error,
) error {
	// this is probably insecure. Will want to change how we access this in the future
	conn, err := pgx.Connect(ctx, postgresUrl)
	if err != nil {
		return fmt.Errorf("connecting to postgres: %w", err)
	}
	defer conn.Close(ctx)

	if err := query(conn); err != nil {
		return fmt.Errorf("querying postgres: %w", err)
	}
	return nil
}

func CallAndReturn[T any](
	ctx context.Context,
	postgresUrl string,
	query func(*pgx.Conn) (*T, error),
) (*T, error) {
	// this is probably insecure. Will want to change how we access this in the future
	conn, err := pgx.Connect(ctx, postgresUrl)
	if err != nil {
		return nil, fmt.Errorf("connecting to postgres: %w", err)
	}
	defer conn.Close(ctx)
	res, err := query(conn)
	if err != nil {
		return nil, fmt.Errorf("querying postgres: %w", err)
	}
	return res, nil
}
