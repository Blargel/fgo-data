package fgodata

import (
	"database/sql"

	"github.com/lib/pq"
)

type AscensionCost struct {
	AscensionLevelID int `json:"ascensionLevelId,string"`
	MaterialID       int `json:"materialId,string"`
	Amount           int `json:"amount,string"`
}
type AscensionCosts map[int]AscensionCost

func (ascensionCosts AscensionCosts) BatchInsert(tx *sql.Tx) error {
	return batchInsert(tx, pq.CopyIn("ascension_costs", "id", "ascension_level_id", "material_id", "amount"), func(stmt *sql.Stmt) error {
		for id, ascensionCost := range ascensionCosts {
			_, err := stmt.Exec(id, ascensionCost.AscensionLevelID, ascensionCost.MaterialID, ascensionCost.Amount)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (_ AscensionCosts) DropTable(tx *sql.Tx) error {
	_, err := tx.Exec("DROP TABLE IF EXISTS ascension_costs;")
	return err
}

func (_ AscensionCosts) CreateTable(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE ascension_costs (
			id SERIAL PRIMARY KEY,
			ascension_level_id INTEGER REFERENCES ascension_levels(id) NOT NULL,
			material_id INTEGER REFERENCES materials(id) NOT NULL,
			amount INTEGER NOT NULL,

			UNIQUE (ascension_level_id, material_id)
		);
	`)
	return err
}
