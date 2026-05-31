package model

import "github.com/google/uuid"

type Lead struct {
	BaseModel
	Name         string `json:"name" gorm:"column:name"`
	IndustryType string `json:"industry_type" gorm:"column:industry_type"`
	Location     string `json:"location" gorm:"column:location"`
	Source       string `json:"source" gorm:"column:source"`
	Address      string `json:"address" gorm:"column:address"`
	PhoneNumber  string `json:"phone_number" gorm:"column:phone_number"`
	Website      string `json:"website" gorm:"column:website"`
	Email        string `json:"email" gorm:"column:email"`

	JobID           uuid.UUID       `json:"job_id" gorm:"column:job_id"`
	LeadScrapingJob LeadScrapingJob `json:"lead_scraping_job" gorm:"foreignKey:JobID"`
}
