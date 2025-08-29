package handlers

import (
	"awesomeProject1/internal/dto"
	"awesomeProject1/internal/services"
	"awesomeProject1/internal/utils"
	"github.com/gofiber/fiber/v2"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ProfileHandler struct {
	profileService *services.ProfileService
}

func NewProfileHandler(profileService *services.ProfileService) *ProfileHandler {
	return &ProfileHandler{
		profileService: profileService,
	}
}

func (h *ProfileHandler) CompleteProfile(c *fiber.Ctx) error {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token",
		})
	}

	var request dto.CompleteProfileRequest
	request.Name = c.FormValue("name")

	if request.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name is required",
		})
	}

	file, err := c.FormFile("avatar")
	if err == nil {
		avatarPath, err := h.saveAvatar(c, file, userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to save avatar",
			})
		}
		request.Avatar = avatarPath
	}

	err = h.profileService.CompleteProfile(userID, request)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Profile completed successfully",
	})
}

func (h *ProfileHandler) UpdateProfile(c *fiber.Ctx) error {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token",
		})
	}

	var request dto.UpdateProfileRequest
	request.Name = c.FormValue("name")

	if request.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name is required",
		})
	}

	file, err := c.FormFile("avatar")
	if err == nil {
		avatarPath, err := h.saveAvatar(c, file, userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to save avatar",
			})
		}
		request.Avatar = avatarPath
	}

	err = h.profileService.UpdateProfile(userID, request)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Profile updated successfully",
	})
}

func (h *ProfileHandler) GetProfile(c *fiber.Ctx) error {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token",
		})
	}

	baseURL := c.Protocol() + "://" + c.Get("Host")
	profile, err := h.profileService.GetProfile(userID, baseURL)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(profile)
}

func (h *ProfileHandler) saveAvatar(c *fiber.Ctx, file *multipart.FileHeader, userID int) (string, error) {
	uploadDir := "data/uploads"
	err := os.MkdirAll(uploadDir, 0755)
	if err != nil {
		return "", err
	}

	ext := filepath.Ext(file.Filename)
	if ext == "" {
		ext = ".jpg"
	}

	filename := filepath.Join(uploadDir,
		"avatar_"+strings.ReplaceAll(time.Now().Format(time.RFC3339), ":", "-")+"_"+
			filepath.Base(file.Filename))

	err = c.SaveFile(file, filename)
	if err != nil {
		return "", err
	}

	return filename, nil
}
