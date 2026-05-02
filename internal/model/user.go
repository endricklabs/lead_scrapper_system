package model

type User struct {
	BaseModel
	Email        string `gorm:"uniqueIndex;not null"`
	Password     string `gorm:"not null"`
	TokenVersion int    `gorm:"default:0;not null"`
}
