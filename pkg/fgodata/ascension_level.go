package fgodata

import (
	"database/sql"

	"github.com/lib/pq"
)

type AscensionLevel struct {
	AscendTo  int `json:"ascendTo,string"`
	ServantID int `json:"servantId,string"`
}
type AscensionLevels map[int]AscensionLevel

func (ascensionLevels AscensionLevels) BatchInsert(tx *sql.Tx) error {
	return batchInsert(tx, pq.CopyIn("ascension_levels", "id", "servant_id", "ascend_to"), func(stmt *sql.Stmt) error {
		for id, ascensionLevel := range ascensionLevels {
			_, err := stmt.Exec(id, ascensionLevel.ServantID, ascensionLevel.AscendTo)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (_ AscensionLevels) DropTable(tx *sql.Tx) error {
	_, err := tx.Exec("DROP TABLE IF EXISTS ascension_levels;")
	return err
}

func (_ AscensionLevels) CreateTable(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE ascension_levels (
			id SERIAL PRIMARY KEY,
			servant_id INTEGER REFERENCES servants(id) NOT NULL,
			ascend_to INTEGER NOT NULL,

			UNIQUE (servant_id, ascend_to)
		);
	`)
	return err
}
