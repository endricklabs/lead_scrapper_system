package lead_dto

type Source string

const (
	GoogleMap Source = "google_maps"
	LinkedIn  Source = "linked_in"
	Facebook  Source = "facebook"
	Instagram Source = "instagram"
)

type SourceData struct {
	Source          Source `json:"source" validate:"required"`
	NumberOfRequest int64  `json:"number_of_request" validate:"required"`
}

type LeadScrapRequest struct {
	IndustryType string `json:"industry_type" validate:"required"`
	Location     string `json:"location" validate:"required"`
	// NumberOfRequest int64        `json:"number_of_request" validate:"required"`
	Source []SourceData `json:"source" validate:"required"`
}
