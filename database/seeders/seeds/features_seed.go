package seeds

import (
	"encoding/json"
	"errors"
	"io"
	"os"

	"github.com/Rizal-Nurochman/matchnbuild/database/entities"
	"gorm.io/gorm"
)

func ListFeaturesSeeder(db *gorm.DB) error {
	jsonFile, err := os.Open("./database/seeders/json/features.json")
	if err != nil {
		return err
	}

	jsonData, err := io.ReadAll(jsonFile)
	if err != nil {
		return err
	}

	var listFeatures []entities.Feature
	err = json.Unmarshal(jsonData, &listFeatures)
	if err != nil {
		return err
	}

	hasTable := db.Migrator().HasTable(&entities.Feature{})
	if !hasTable {
		err := db.Migrator().CreateTable(&entities.Feature{})
		if err != nil {
			return err
		}
	}

	for _, data := range listFeatures {
		var feature entities.Feature
		err := db.Where(&entities.Feature{Name: data.Name}).First(&feature).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		isData := db.Find(&feature, "name = ?", data.Name).RowsAffected
		if isData == 0 {
			if err := db.Create(&data).Error; err != nil {
				return err
			}
		}
	}

	return nil
}