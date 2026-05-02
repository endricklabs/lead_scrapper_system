package api_dto

type ApiResponse struct {
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}
