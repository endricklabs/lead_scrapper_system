package model

import (
	"github.com/google/uuid"
)

type JobStatus string

const (
	JobStatusPending   JobStatus = "PENDING"
	JobStatusCompleted JobStatus = "COMPLETED"
)

type LeadScrapingJob struct {
	BaseModel
	Source          string    `json:"source" gorm:"column:source"`
	IndustryType    string    `json:"industry_type" gorm:"column:industry_type"`
	Location        string    `json:"location" gorm:"column:location"`
	TargetRequested int       `json:"target_requested" gorm:"column:target_requested"`
	LeadsCollected  int       `json:"leads_collected" gorm:"column:leads_collected;default:0"`
	Status          string    `json:"status" gorm:"column:status;default:'PENDING'"`
	UserID          uuid.UUID `json:"user_id" gorm:"column:user_id"`
	User            User      `gorm:"foreignKey:UserID"`
}
