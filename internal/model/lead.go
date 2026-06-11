package model

import "github.com/google/uuid"

type Lead struct {
	BaseModel
	// uniqueIndex on website: prevents two rows with the same URL (non-empty values only).
	Name         string `json:"name" gorm:"column:name;index:idx_lead_name_industry_location"`
	IndustryType string `json:"industry_type" gorm:"column:industry_type;index:idx_lead_name_industry_location"`
	Location     string `json:"location" gorm:"column:location;index:idx_lead_name_industry_location"`
	Source       string `json:"source" gorm:"column:source"`
	Address      string `json:"address" gorm:"column:address"`
	PhoneNumber  string `json:"phone_number" gorm:"column:phone_number"`
	Website      string `json:"website" gorm:"column:website;uniqueIndex:idx_lead_website,where:website <> ''"`
	Email        string `json:"email" gorm:"column:email"`

	JobID           uuid.UUID       `json:"job_id" gorm:"column:job_id"`
	LeadScrapingJob LeadScrapingJob `json:"lead_scraping_job" gorm:"foreignKey:JobID"`
}
