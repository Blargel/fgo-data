package fgodata

import (
	"database/sql"

	"github.com/lib/pq"
)

type Material struct {
	Name  string `json:"name"`
	Icon  string `json:"icon"`
	Order int    `json:"order,string"`
}
type Materials map[int]Material

func (materials Materials) BatchInsert(tx *sql.Tx) error {
	return batchInsert(tx, pq.CopyIn("materials", "id", "name", "icon", "position"), func(stmt *sql.Stmt) error {
		for id, material := range materials {
			_, err := stmt.Exec(id, material.Name, material.Icon, material.Order)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (_ Materials) DropTable(tx *sql.Tx) error {
	_, err := tx.Exec("DROP TABLE IF EXISTS materials;")
	return err
}

func (_ Materials) CreateTable(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE materials (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			icon TEXT NOT NULL UNIQUE,
			position INTEGER UNIQUE
		);
	`)
	return err
}
