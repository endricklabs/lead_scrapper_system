package model

import "gorm.io/gorm"

type Lead struct {
	gorm.Model
	Name         string `json:"name" gorm:"column:name"`
	IndustryType string `json:"industry_type" gorm:"column:industry_type"`
	Location     string `json:"location" gorm:"column:location"`
	Source       string `json:"source" gorm:"column:source"`
	Address      string `json:"address" gorm:"column:address"`
	PhoneNumber  string `json:"phone_number" gorm:"column:phone_number"`
	Website      string `json:"website" gorm:"column:website"`
	Email        string `json:"email" gorm:"column:email"`
}
