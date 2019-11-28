package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"io/ioutil"

	"github.com/lib/pq"
)

type Servant struct {
	Name    string `json:"name"`
	Rarity  int    `json:"rarity,string"`
	Icon    string `json:"icon"`
	ClassID int    `json:"classId,string"`
}

type Class struct {
	Name string `json:"name"`
	Icon string `json:"icon"`
}

type Material struct {
	Name  string `json:"name"`
	Icon  string `json:"icon"`
	Order int    `json:"order,string"`
}

type AscensionLevel struct {
	AscendTo  int `json:"ascendTo,string"`
	ServantID int `json:"servantId,string"`
}

type AscensionCost struct {
	AscensionLevelID int `json:"ascensionLevelId,string"`
	MaterialID       int `json:"materialId,string"`
	Amount           int `json:"amount,string"`
}

type SkillLevel struct {
	LevelTo   int `json:"levelTo,string"`
	ServantID int `json:"servantId,string"`
}

type SkillCost struct {
	SkillLevelID int `json:"skillLevelId,string"`
	MaterialID   int `json:"materialId,string"`
	Amount       int `json:"amount,string"`
}

type Servants map[int]Servant
type Classes map[int]Class
type Materials map[int]Material
type AscensionLevels map[int]AscensionLevel
type AscensionCosts map[int]AscensionCost
type SkillLevels map[int]SkillLevel
type SkillCosts map[int]SkillCost

type Fgo struct {
	Servants        Servants        `json:"servants"`
	Classes         Classes         `json:"classes"`
	Materials       Materials       `json:"materials"`
	AscensionLevels AscensionLevels `json:"ascensionLevels"`
	AscensionCosts  AscensionCosts  `json:"ascensionCosts"`
	SkillLevels     SkillLevels     `json:"skillLevels"`
	SkillCosts      SkillCosts      `json:"skillCosts"`
}

var input string
var dburl string

func init() {
	flag.StringVar(&input, "input", "", "input json file")
	flag.StringVar(&dburl, "dburl", "postgres://localhost/fgo_data?sslmode=disable", "database name to create")
	flag.Parse()
}

func createSchema(db *sql.Tx) error {
	_, err := db.Exec(`
		DROP TABLE IF EXISTS ascension_costs;
		DROP TABLE IF EXISTS skill_costs;
		DROP TABLE IF EXISTS ascension_levels;
		DROP TABLE IF EXISTS skill_levels;
		DROP TABLE IF EXISTS servants;
		DROP TABLE IF EXISTS classes;
		DROP TABLE IF EXISTS materials;

		CREATE TABLE materials (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			icon TEXT NOT NULL UNIQUE,
			position INTEGER UNIQUE
		);

		CREATE TABLE classes (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			icon TEXT NOT NULL UNIQUE
		);

    CREATE TABLE servants (
      id INTEGER PRIMARY KEY NOT NULL,
      name TEXT NOT NULL,
      icon TEXT NOT NULL UNIQUE,
			rarity INTEGER NOT NULL,
      class_id INTEGER REFERENCES classes(id) NOT NULL,

			UNIQUE (class_id, rarity, name)
    );

		CREATE TABLE skill_levels (
			id SERIAL PRIMARY KEY,
			servant_id INTEGER REFERENCES servants(id) NOT NULL,
			level_to INTEGER NOT NULL,

			UNIQUE (servant_id, level_to)
		);

		CREATE TABLE ascension_levels (
			id SERIAL PRIMARY KEY,
			servant_id INTEGER REFERENCES servants(id) NOT NULL,
			ascend_to INTEGER NOT NULL,

			UNIQUE (servant_id, ascend_to)
		);

		CREATE TABLE skill_costs (
			id SERIAL PRIMARY KEY,
			skill_level_id INTEGER REFERENCES skill_levels(id) NOT NULL,
			material_id INTEGER REFERENCES materials(id) NOT NULL,
			amount INTEGER NOT NULL,

			UNIQUE (skill_level_id, material_id)
		);

		CREATE TABLE ascension_costs (
			id SERIAL PRIMARY KEY,
			ascension_level_id INTEGER REFERENCES ascension_levels(id) NOT NULL,
			material_id INTEGER REFERENCES materials(id) NOT NULL,
			amount INTEGER NOT NULL,

			UNIQUE (ascension_level_id, material_id)
		);
  `)

	if err != nil {
		return err
	}

	return nil
}

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

func insertClasses(tx *sql.Tx, classes Classes) error {
	err := batchInsert(tx, pq.CopyIn("classes", "id", "name", "icon"), func(stmt *sql.Stmt) error {
		for id, class := range classes {
			_, err := stmt.Exec(id, class.Name, class.Icon)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func insertMaterials(tx *sql.Tx, materials Materials) error {
	err := batchInsert(tx, pq.CopyIn("materials", "id", "name", "icon", "position"), func(stmt *sql.Stmt) error {
		for id, material := range materials {
			_, err := stmt.Exec(id, material.Name, material.Icon, material.Order)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func insertServants(tx *sql.Tx, servants Servants) error {
	err := batchInsert(tx, pq.CopyIn("servants", "id", "name", "icon", "rarity", "class_id"), func(stmt *sql.Stmt) error {
		for id, servant := range servants {
			_, err := stmt.Exec(id, servant.Name, servant.Icon, servant.Rarity, servant.ClassID)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func insertSkillLevels(tx *sql.Tx, skillLevels SkillLevels) error {
	err := batchInsert(tx, pq.CopyIn("skill_levels", "id", "servant_id", "level_to"), func(stmt *sql.Stmt) error {
		for id, skillLevel := range skillLevels {
			_, err := stmt.Exec(id, skillLevel.ServantID, skillLevel.LevelTo)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func insertAscensionLevels(tx *sql.Tx, ascensionLevels AscensionLevels) error {
	err := batchInsert(tx, pq.CopyIn("ascension_levels", "id", "servant_id", "ascend_to"), func(stmt *sql.Stmt) error {
		for id, ascensionLevel := range ascensionLevels {
			_, err := stmt.Exec(id, ascensionLevel.ServantID, ascensionLevel.AscendTo)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func insertSkillCosts(tx *sql.Tx, skillCosts SkillCosts) error {
	err := batchInsert(tx, pq.CopyIn("skill_costs", "id", "skill_level_id", "material_id", "amount"), func(stmt *sql.Stmt) error {
		for id, skillCost := range skillCosts {
			_, err := stmt.Exec(id, skillCost.SkillLevelID, skillCost.MaterialID, skillCost.Amount)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func insertAscensionCosts(tx *sql.Tx, ascensionCosts AscensionCosts) error {
	err := batchInsert(tx, pq.CopyIn("ascension_costs", "id", "ascension_level_id", "material_id", "amount"), func(stmt *sql.Stmt) error {
		for id, ascensionCost := range ascensionCosts {
			_, err := stmt.Exec(id, ascensionCost.AscensionLevelID, ascensionCost.MaterialID, ascensionCost.Amount)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func cleanFgoData(fgo *Fgo) {
	for k, v := range fgo.SkillLevels {
		if _, ok := fgo.Servants[v.ServantID]; !ok {
			delete(fgo.SkillLevels, k)
		}
	}

	for k, v := range fgo.AscensionLevels {
		if _, ok := fgo.Servants[v.ServantID]; !ok {
			delete(fgo.AscensionLevels, k)
		}
	}
}

func main() {
	if input == "" || dburl == "" {
		flag.PrintDefaults()
		return
	}

	data, err := ioutil.ReadFile(input)
	if err != nil {
		panic(err)
	}

	var fgo Fgo

	err = json.Unmarshal(data, &fgo)
	if err != nil {
		panic(err)
	}

	cleanFgoData(&fgo)

	db, err := sql.Open("postgres", dburl)
	if err != nil {
		panic(err)
	}

	txn, err := db.Begin()
	if err != nil {
		panic(err)
	}

	err = createSchema(txn)
	if err != nil {
		txn.Rollback()
		panic(err)
	}

	err = insertClasses(txn, fgo.Classes)
	if err != nil {
		txn.Rollback()
		panic(err)
	}

	err = insertMaterials(txn, fgo.Materials)
	if err != nil {
		txn.Rollback()
		panic(err)
	}

	err = insertServants(txn, fgo.Servants)
	if err != nil {
		txn.Rollback()
		panic(err)
	}

	err = insertSkillLevels(txn, fgo.SkillLevels)
	if err != nil {
		txn.Rollback()
		panic(err)
	}

	err = insertAscensionLevels(txn, fgo.AscensionLevels)
	if err != nil {
		txn.Rollback()
		panic(err)
	}

	err = insertSkillCosts(txn, fgo.SkillCosts)
	if err != nil {
		txn.Rollback()
		panic(err)
	}

	err = insertAscensionCosts(txn, fgo.AscensionCosts)
	if err != nil {
		txn.Rollback()
		panic(err)
	}

	txn.Commit()
}
