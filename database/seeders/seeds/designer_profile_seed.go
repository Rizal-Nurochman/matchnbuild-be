package seeds

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/Rizal-Nurochman/matchnbuild/database/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type designerProfileSeed struct {
	Email             string `json:"email"`
	Bio               string `json:"bio"`
	ExperienceYears   int    `json:"experience_years"`
	IsVerified        bool   `json:"is_verified"`
	IsAvailable       bool   `json:"is_available"`
	Location          string `json:"location"`
	BankAccountNumber string `json:"bank_account_number"`
}

func ListDesignerProfileSeeder(db *gorm.DB) error {
	jsonFile, err := os.Open("./database/seeders/json/designer_profiles.json")
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	jsonData, err := io.ReadAll(jsonFile)
	if err != nil {
		return err
	}

	var seedProfiles []designerProfileSeed
	if err := json.Unmarshal(jsonData, &seedProfiles); err != nil {
		return err
	}

	seededCount := 0
	for _, seed := range seedProfiles {
		// Find user by email
		var user entities.User
		if err := db.Where("email = ?", seed.Email).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				fmt.Printf("User with email %s not found, skipping designer profile\n", seed.Email)
				continue
			}
			return err
		}

		// Check if designer profile already exists
		var existing entities.DesignerProfile
		if db.Where("user_id = ?", user.ID).First(&existing).RowsAffected > 0 {
			continue
		}

		profile := entities.DesignerProfile{
			ID:                uuid.New(),
			UserID:            user.ID,
			Bio:               seed.Bio,
			ExperienceYears:   seed.ExperienceYears,
			IsVerified:        seed.IsVerified,
			IsAvailable:       seed.IsAvailable,
			Location:          seed.Location,
			BankAccountNumber: seed.BankAccountNumber,
		}

		if err := db.Create(&profile).Error; err != nil {
			return fmt.Errorf("failed to seed designer profile for %s: %v", seed.Email, err)
		}

		seededCount++
	}

	if seededCount > 0 {
		fmt.Printf("Seeded %d designer profile(s)\n", seededCount)
	} else {
		fmt.Println("Designer profiles already seeded, skipping")
	}

	return nil
}
