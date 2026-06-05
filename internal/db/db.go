package db

import (
	"encoding/json"
	"log"
	"os"

	"lead_scrapper_be/internal/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Init(connStr string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	db.AutoMigrate(&model.User{}, &model.Lead{}, &model.LeadScrapingJob{}, &model.SubscriptionPackage{}, &model.UserSubscription{})

	if err := seedData(db); err != nil {
		log.Printf("failed to seed data: %v", err)
	}

	return db, nil
}

func seedData(db *gorm.DB) error {
	path := "configs/seed/subscription_packages.json"
	file, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var packages []model.SubscriptionPackage
	if err := json.Unmarshal(file, &packages); err != nil {
		return err
	}

	for _, p := range packages {
		err := db.Where(model.SubscriptionPackage{Slug: p.Slug}).FirstOrCreate(&p).Error
		if err != nil {
			return err
		}
	}

	return nil
}
