package response

import (
	"net/http"

	"touchdown-tally/internal/models"

	"github.com/gin-gonic/gin"
)

// JSON sends a JSON response with the given status code and data
func JSON(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, data)
}

// Success sends a successful JSON response
func Success(c *gin.Context, data interface{}, message ...string) {
	response := models.SuccessResponse{
		Success: true,
		Data:    data,
	}

	if len(message) > 0 {
		response.Message = message[0]
	}

	c.JSON(http.StatusOK, response)
}

// Created sends a 201 Created response
func Created(c *gin.Context, data interface{}, message ...string) {
	response := models.SuccessResponse{
		Success: true,
		Data:    data,
	}

	if len(message) > 0 {
		response.Message = message[0]
	}

	c.JSON(http.StatusCreated, response)
}

// Error sends an error response with the given status code
func Error(c *gin.Context, statusCode int, err string, message ...string) {
	response := models.ErrorResponse{
		Error: err,
		Code:  statusCode,
	}

	if len(message) > 0 {
		response.Message = message[0]
	}

	c.JSON(statusCode, response)
}

// BadRequest sends a 400 Bad Request response
func BadRequest(c *gin.Context, err string, message ...string) {
	Error(c, http.StatusBadRequest, err, message...)
}

// Unauthorized sends a 401 Unauthorized response
func Unauthorized(c *gin.Context, err string, message ...string) {
	Error(c, http.StatusUnauthorized, err, message...)
}

// Forbidden sends a 403 Forbidden response
func Forbidden(c *gin.Context, err string, message ...string) {
	Error(c, http.StatusForbidden, err, message...)
}

// NotFound sends a 404 Not Found response
func NotFound(c *gin.Context, err string, message ...string) {
	Error(c, http.StatusNotFound, err, message...)
}

// Conflict sends a 409 Conflict response
func Conflict(c *gin.Context, err string, message ...string) {
	Error(c, http.StatusConflict, err, message...)
}

// InternalServerError sends a 500 Internal Server Error response
func InternalServerError(c *gin.Context, err string, message ...string) {
	Error(c, http.StatusInternalServerError, err, message...)
}

// ValidationError sends a 422 Unprocessable Entity response for validation errors
func ValidationError(c *gin.Context, err string, message ...string) {
	Error(c, http.StatusUnprocessableEntity, err, message...)
}
