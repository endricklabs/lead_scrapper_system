package api

import (
	"errors"
	"net/http"

	api_dto "lead_scrapper_be/internal/dto/api"

	"github.com/labstack/echo/v4"
)

func SuccessfulResponse(ctx echo.Context, data string) error {
	return ctx.JSON(http.StatusOK, api_dto.ApiResponse{
		Message: data,
	})
}

func SuccessfulResponseWithData(ctx echo.Context, message string, data interface{}) error {
	return ctx.JSON(http.StatusOK, api_dto.ApiResponse{
		Message: message,
		Data:    data,
	})
}

func BadRequest(ctx echo.Context, data string) error {
	return ctx.JSON(http.StatusBadRequest, api_dto.ApiResponse{
		Message: data,
	})
}

func Unauthorized(ctx echo.Context, data string) error {
	return ctx.JSON(http.StatusUnauthorized, api_dto.ApiResponse{
		Message: data,
	})
}

func InternalServerError(ctx echo.Context, data string) error {
	return ctx.JSON(http.StatusInternalServerError, api_dto.ApiResponse{
		Message: data,
	})
}

func NewError(message string) error {
	return errors.New(message)
}
