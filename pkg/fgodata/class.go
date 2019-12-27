package fgodata

import (
	"database/sql"

	"github.com/lib/pq"
)

type Class struct {
	Name string `json:"name"`
	Icon string `json:"icon"`
}
type Classes map[int]Class

func (classes Classes) BatchInsert(tx *sql.Tx) error {
	return batchInsert(tx, pq.CopyIn("classes", "id", "name", "icon"), func(stmt *sql.Stmt) error {
		for id, class := range classes {
			_, err := stmt.Exec(id, class.Name, class.Icon)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (_ Classes) DropTable(tx *sql.Tx) error {
	_, err := tx.Exec("DROP TABLE IF EXISTS classes;")
	return err
}

func (_ Classes) CreateTable(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE classes (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			icon TEXT NOT NULL UNIQUE
		);
	`)
	return err
}
