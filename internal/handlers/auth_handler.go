package handlers

import (
	"awesomeProject1/internal/dto"
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

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var request dto.LoginRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"error": "Invalid request body",
			},
		)
	}

	if request.Email == "" || request.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"error": "Email and password are required",
			},
		)
	}

	response, err := h.authService.Login(request)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(
			fiber.Map{
				"error": err.Error(),
			},
		)
	}

	return c.JSON(response)
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

	userName := c.Query("name", "Developer")

	greeting := h.greetingService.GetGreeting(userID, userName)
	return c.JSON(greeting)
}
