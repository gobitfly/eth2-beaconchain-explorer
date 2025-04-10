package db

import (
	"context"
	"fmt"

	pgxdecimal "github.com/jackc/pgx-shopspring-decimal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func ClearAndCopyToTable[T []any](db *sqlx.DB, tableName string, columns []string, data []T) error {
	conn, err := db.Conn(context.Background())
	if err != nil {
		return fmt.Errorf("error retrieving raw sql connection: %w", err)
	}
	defer conn.Close()
	err = conn.Raw(func(driverConn interface{}) error {
		conn := driverConn.(*stdlib.Conn).Conn()

		pgxdecimal.Register(conn.TypeMap())
		tx, err := conn.Begin(context.Background())

		if err != nil {
			return err
		}
		defer func() {
			err := tx.Rollback(context.Background())
			if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
				logrus.Error(err, "error rolling back transaction", 0)
			}
		}()

		// clear
		_, err = tx.Exec(context.Background(), fmt.Sprintf("TRUNCATE TABLE %s", tableName))
		if err != nil {
			return errors.Wrap(err, "failed to truncate table")
		}

		// copy
		_, err = tx.CopyFrom(context.Background(), pgx.Identifier{tableName}, columns,
			pgx.CopyFromSlice(len(data), func(i int) ([]interface{}, error) {
				return data[i], nil
			}))

		if err != nil {
			return err
		}

		err = tx.Commit(context.Background())
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error copying data to %s: %w", tableName, err)
	}
	return nil
}
