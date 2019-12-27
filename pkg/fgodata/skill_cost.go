package fgodata

import (
	"database/sql"

	"github.com/lib/pq"
)

type SkillCost struct {
	SkillLevelID int `json:"skillLevelId,string"`
	MaterialID   int `json:"materialId,string"`
	Amount       int `json:"amount,string"`
}
type SkillCosts map[int]SkillCost

func (skillCosts SkillCosts) BatchInsert(tx *sql.Tx) error {
	return batchInsert(tx, pq.CopyIn("skill_costs", "id", "skill_level_id", "material_id", "amount"), func(stmt *sql.Stmt) error {
		for id, skillCost := range skillCosts {
			_, err := stmt.Exec(id, skillCost.SkillLevelID, skillCost.MaterialID, skillCost.Amount)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (_ SkillCosts) DropTable(tx *sql.Tx) error {
	_, err := tx.Exec("DROP TABLE IF EXISTS skill_costs;")
	return err
}

func (_ SkillCosts) CreateTable(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE skill_costs (
			id SERIAL PRIMARY KEY,
			skill_level_id INTEGER REFERENCES skill_levels(id) NOT NULL,
			material_id INTEGER REFERENCES materials(id) NOT NULL,
			amount INTEGER NOT NULL,

			UNIQUE (skill_level_id, material_id)
		);
	`)
	return err
}
