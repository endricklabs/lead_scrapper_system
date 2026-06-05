package model

type SubscriptionPackage struct {
	BaseModel
	Name             string `gorm:"not null"`
	Slug             string `gorm:"not null; uniqueIndex"`
	Price            float64
	Description      string
	MaxLeadsPerMonth int
}
