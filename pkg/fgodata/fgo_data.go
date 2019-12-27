package fgodata

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"sync"
)

type FgoData struct {
	Servants        Servants        `json:"servants"`
	Classes         Classes         `json:"classes"`
	Materials       Materials       `json:"materials"`
	AscensionLevels AscensionLevels `json:"ascensionLevels"`
	AscensionCosts  AscensionCosts  `json:"ascensionCosts"`
	SkillLevels     SkillLevels     `json:"skillLevels"`
	SkillCosts      SkillCosts      `json:"skillCosts"`
}

type TableManager interface {
	DropTable(tx *sql.Tx) error
	CreateTable(tx *sql.Tx) error
	BatchInsert(tx *sql.Tx) error
}

func (fgoData *FgoData) Tables() []TableManager {
	return []TableManager{
		fgoData.Materials,
		fgoData.Classes,
		fgoData.Servants,
		fgoData.AscensionLevels,
		fgoData.SkillLevels,
		fgoData.AscensionCosts,
		fgoData.SkillCosts,
	}
}

func (fgoData *FgoData) TablesToDrop() []TableManager {
	return []TableManager{
		fgoData.SkillCosts,
		fgoData.AscensionCosts,
		fgoData.SkillLevels,
		fgoData.AscensionLevels,
		fgoData.Servants,
		fgoData.Materials,
		fgoData.Classes,
	}
}

func (fgoData *FgoData) DropSchema(tx *sql.Tx) error {
	for _, e := range fgoData.TablesToDrop() {
		if err := e.DropTable(tx); err != nil {
			return err
		}
	}

	return nil
}

func (fgoData *FgoData) CreateSchema(tx *sql.Tx) error {
	for _, e := range fgoData.Tables() {
		if err := e.CreateTable(tx); err != nil {
			return err
		}
	}

	return nil
}

func (fgoData *FgoData) ResetSchema(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if err := fgoData.DropSchema(tx); err != nil {
		tx.Rollback()
		return err
	}

	if err := fgoData.CreateSchema(tx); err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()

	return nil
}

func (fgoData *FgoData) InsertData(db *sql.DB) error {
	var wg sync.WaitGroup
	for _, e := range fgoData.Tables() {
		wg.Add(1)
		go func(e TableManager, wg *sync.WaitGroup) error {
			tx, err := db.Begin()
			if err != nil {
				wg.Done()
				return err
			}

			if err := e.BatchInsert(tx); err != nil {
				tx.Rollback()
				wg.Done()
				return err
			}

			tx.Commit()
			wg.Done()

			return nil
		}(e, &wg)
	}

	wg.Wait()

	return nil
}

func (fgoData *FgoData) clean() {
	for k, v := range fgoData.SkillLevels {
		if _, ok := fgoData.Servants[v.ServantID]; !ok {
			delete(fgoData.SkillLevels, k)
		}
	}

	for k, v := range fgoData.AscensionLevels {
		if _, ok := fgoData.Servants[v.ServantID]; !ok {
			delete(fgoData.AscensionLevels, k)
		}
	}
}

func ImportData(file string) (*FgoData, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var fgo FgoData

	err = json.Unmarshal(data, &fgo)
	if err != nil {
		return nil, err
	}

	fgo.clean()

	return &fgo, nil
}
