package fgodata

import (
	"database/sql"
)

type insertFn func(stmt *sql.Stmt) error

func batchInsert(tx *sql.Tx, sql string, fn insertFn) error {
	stmt, err := tx.Prepare(sql)
	if err != nil {
		return err
	}

	err = fn(stmt)
	if err != nil {
		return err
	}

	_, err = stmt.Exec()
	if err != nil {
		return err
	}

	err = stmt.Close()
	if err != nil {
		return err
	}

	return nil
}
