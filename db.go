package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func Insert(conn *pgx.Conn, str string) error {
	tx, err := conn.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("begin: %d,\n %s,\n\n %s", err, str[:1000], str[len(str)-500:])
	}
	_, err = conn.Exec(context.Background(), str)
	if err != nil {
		return fmt.Errorf("exec:, %d,\n %s,\n\n %s", err, str[:1000], str[len(str)-500:])
	}
	err = tx.Commit(context.Background())
	if err != nil {
		return fmt.Errorf("commit:, %d,\n %s,\n\n %s", err, str[:1000], str[len(str)-500:])
	}
	return nil
}
