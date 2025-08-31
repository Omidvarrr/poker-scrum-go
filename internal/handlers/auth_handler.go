package handlers

import (
	"awesomeProject1/internal/services"
	"awesomeProject1/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authService     *services.AuthService
	greetingService *services.GreetingService
}

func NewAuthHandler(authService *services.AuthService, greetingService *services.GreetingService) *AuthHandler {
	return &AuthHandler{
		authService:     authService,
		greetingService: greetingService,
	}
}

func (h *AuthHandler) GetGreeting(c *fiber.Ctx) error {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(
			fiber.Map{
				"error": "Invalid token",
			},
		)
	}

	greeting, err := h.greetingService.GetGreeting(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"error": "Problem in generate greeting message",
			},
		)
	}
	return c.JSON(greeting)
}
