package database

import (
	"github.com/Rizal-Nurochman/matchnbuild/database/seeders/seeds"
	"gorm.io/gorm"
)

func Seeder(db *gorm.DB) error {
	if err := seeds.ListUserSeeder(db); err != nil {
		return err
	}

	if err := seeds.ListFeaturesSeeder(db); err != nil {
		return err
	}

	return nil
}
