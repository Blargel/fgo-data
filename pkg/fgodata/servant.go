package fgodata

import (
	"database/sql"

	"github.com/lib/pq"
)

type Servant struct {
	Name    string `json:"name"`
	Rarity  int    `json:"rarity,string"`
	Icon    string `json:"icon"`
	ClassID int    `json:"classId,string"`
}

type Servants map[int]Servant

func (servants Servants) BatchInsert(tx *sql.Tx) error {
	return batchInsert(tx, pq.CopyIn("servants", "id", "name", "icon", "rarity", "class_id"), func(stmt *sql.Stmt) error {
		for id, servant := range servants {
			_, err := stmt.Exec(id, servant.Name, servant.Icon, servant.Rarity, servant.ClassID)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (_ Servants) DropTable(tx *sql.Tx) error {
	_, err := tx.Exec("DROP TABLE IF EXISTS servants;")
	return err
}

func (_ Servants) CreateTable(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE servants (
			id INTEGER PRIMARY KEY NOT NULL,
			name TEXT NOT NULL,
			icon TEXT NOT NULL UNIQUE,
			rarity INTEGER NOT NULL,
			class_id INTEGER REFERENCES classes(id) NOT NULL,

			UNIQUE (class_id, rarity, name)
		);
	`)
	return err
}
