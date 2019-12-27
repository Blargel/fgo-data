package fgodata

import (
	"database/sql"

	"github.com/lib/pq"
)

type SkillLevel struct {
	LevelTo   int `json:"levelTo,string"`
	ServantID int `json:"servantId,string"`
}
type SkillLevels map[int]SkillLevel

func (skillLevels SkillLevels) BatchInsert(tx *sql.Tx) error {
	return batchInsert(tx, pq.CopyIn("skill_levels", "id", "servant_id", "level_to"), func(stmt *sql.Stmt) error {
		for id, skillLevel := range skillLevels {
			_, err := stmt.Exec(id, skillLevel.ServantID, skillLevel.LevelTo)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (_ SkillLevels) DropTable(tx *sql.Tx) error {
	_, err := tx.Exec("DROP TABLE IF EXISTS skill_levels;")
	return err
}

func (_ SkillLevels) CreateTable(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE skill_levels (
			id SERIAL PRIMARY KEY,
			servant_id INTEGER REFERENCES servants(id) NOT NULL,
			level_to INTEGER NOT NULL,

			UNIQUE (servant_id, level_to)
		);
	`)
	return err
}
