package seeds

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/Rizal-Nurochman/matchnbuild/database/entities"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type designItemSeed struct {
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	Category        string   `json:"category"`
	Style           string   `json:"style"`
	LandAreaMin     *float64 `json:"land_area_min"`
	LandAreaMax     *float64 `json:"land_area_max"`
	BuildingArea    *float64 `json:"building_area"`
	NumFloors       *int     `json:"num_floors"`
	NumBedrooms     *int     `json:"num_bedrooms"`
	RoomType        *string  `json:"room_type"`
	RoomArea        *float64 `json:"room_area"`
	EstimatedBudget float64  `json:"estimated_budget"`
	PriceStartFrom  float64  `json:"price_start_from"`
	Location        string   `json:"location"`
	ImageURL        string   `json:"image_url"`
}

func ListDesignItemSeeder(db *gorm.DB) error {
	// Find designer users to assign design items
	var designers []entities.DesignerProfile
	if err := db.Find(&designers).Error; err != nil {
		return fmt.Errorf("failed to find designers: %v", err)
	}

	if len(designers) == 0 {
		fmt.Println("No designers found, skipping design item seeder. Please seed users and designer profiles first.")
		return nil
	}

	jsonFile, err := os.Open("./database/seeders/json/design_items.json")
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	jsonData, err := io.ReadAll(jsonFile)
	if err != nil {
		return err
	}

	var seedItems []designItemSeed
	if err := json.Unmarshal(jsonData, &seedItems); err != nil {
		return err
	}

	seededCount := 0
	for i, seed := range seedItems {
		// Check if already seeded by title
		var existing entities.DesignItem
		err := db.Where("title = ?", seed.Title).First(&existing).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if db.Find(&existing, "title = ?", seed.Title).RowsAffected > 0 {
			continue
		}

		// Round-robin assign to available designers
		designer := designers[i%len(designers)]

		item := entities.DesignItem{
			ID:              uuid.New(),
			DesignerID:      designer.ID,
			Title:           seed.Title,
			Description:     seed.Description,
			Category:        seed.Category,
			Style:           seed.Style,
			LandAreaMin:     seed.LandAreaMin,
			LandAreaMax:     seed.LandAreaMax,
			BuildingArea:    seed.BuildingArea,
			NumFloors:       seed.NumFloors,
			NumBedrooms:     seed.NumBedrooms,
			RoomType:        seed.RoomType,
			RoomArea:        seed.RoomArea,
			EstimatedBudget: decimal.NewFromFloat(seed.EstimatedBudget),
			PriceStartFrom:  decimal.NewFromFloat(seed.PriceStartFrom),
			Location:        seed.Location,
			ImageURL:        seed.ImageURL,
		}

		if err := db.Create(&item).Error; err != nil {
			return fmt.Errorf("failed to seed design item '%s': %v", seed.Title, err)
		}

		seededCount++
	}

	if seededCount > 0 {
		fmt.Printf("Seeded %d design item(s)\n", seededCount)
	} else {
		fmt.Println("Design items already seeded, skipping")
	}

	return nil
}
