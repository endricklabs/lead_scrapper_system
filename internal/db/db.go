package db
import (
	"lead_scrapper_be/internal/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Init(connStr string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	db.AutoMigrate(&model.User{}, &model.Lead{})

	return db, nil
}
