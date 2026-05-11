// Package handler contains HTTP request handlers.
package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"spark-s3-microservice/internal/domain"
	"spark-s3-microservice/internal/middleware"
	"spark-s3-microservice/internal/service"
)

// StorageHandler handles storage-related HTTP requests.
type StorageHandler struct {
	storageService *service.StorageService
}

// NewStorageHandler creates a new storage handler.
func NewStorageHandler(storageService *service.StorageService) *StorageHandler {
	return &StorageHandler{
		storageService: storageService,
	}
}

// GetUploadURL handles GET /api/v1/storage/upload-url
// @Summary Generate presigned URL for file upload
// @Description Generates a unique file_id and presigned PUT URL for uploading to S3
// @Tags storage
// @Accept json
// @Produce json
// @Param type query string true "File type: avatar, message, or sticker"
// @Param ext query string true "File extension: jpg, png, gif, webp, pdf"
// @Security BearerAuth
// @Success 200 {object} service.UploadURLResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/storage/upload-url [get]
func (h *StorageHandler) GetUploadURL(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	fileTypeStr := c.Query("type")
	extension := c.Query("ext")

	if fileTypeStr == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Parameter 'type' is required. Must be one of: avatar, message, sticker",
		})
		return
	}

	if extension == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Parameter 'ext' is required",
		})
		return
	}

	fileType, err := domain.ParseFileType(fileTypeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Parameter 'type' must be one of: avatar, message, sticker",
		})
		return
	}

	response, err := h.storageService.GenerateUploadURL(c.Request.Context(), userID, fileType, extension)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidExtension) {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_extension",
				Message: "Extension is not allowed for this file type",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to generate upload URL",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetDownloadURL handles GET /api/v1/storage/download-url
// @Summary Generate presigned URL for file download
// @Description Generates a presigned GET URL for downloading a file from S3
// @Tags storage
// @Accept json
// @Produce json
// @Param file_id query string true "File identifier"
// @Security BearerAuth
// @Success 200 {object} service.DownloadURLResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/storage/download-url [get]
func (h *StorageHandler) GetDownloadURL(c *gin.Context) {
	fileID := c.Query("file_id")

	if fileID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Parameter 'file_id' is required",
		})
		return
	}

	response, err := h.storageService.GenerateDownloadURL(c.Request.Context(), fileID)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidFileID) {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_request",
				Message: "Invalid file_id format",
			})
			return
		}
		if errors.Is(err, domain.ErrFileNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "not_found",
				Message: "File not found or access denied",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to generate download URL",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// DeleteFile handles DELETE /api/v1/storage/file
// @Summary Delete a file
// @Description Deletes a file from S3 after verifying ownership
// @Tags storage
// @Accept json
// @Produce json
// @Param file_id query string true "File identifier to delete"
// @Security BearerAuth
// @Success 200 {object} service.DeleteResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/storage/file [delete]
func (h *StorageHandler) DeleteFile(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	fileID := c.Query("file_id")

	if fileID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Parameter 'file_id' is required",
		})
		return
	}

	response, err := h.storageService.DeleteFile(c.Request.Context(), userID, fileID)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidFileID) {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_request",
				Message: "Invalid file_id format",
			})
			return
		}
		if errors.Is(err, domain.ErrForbidden) {
			c.JSON(http.StatusForbidden, ErrorResponse{
				Error:   "forbidden",
				Message: "You don't have permission to delete this file",
			})
			return
		}
		if errors.Is(err, domain.ErrFileNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "not_found",
				Message: "File not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to delete file",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ErrorResponse represents an API error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}