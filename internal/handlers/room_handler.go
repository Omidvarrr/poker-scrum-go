package handlers

import (
	"awesomeProject1/internal/dto"
	"awesomeProject1/internal/services"
	"awesomeProject1/internal/utils"
	pages "awesomeProject1/web/templates/pages"
	"bytes"
	"context"

	"github.com/gofiber/fiber/v2"
)

type RoomHandler struct {
	roomService *services.RoomService
}

func NewRoomHandler(roomService *services.RoomService) *RoomHandler {
	return &RoomHandler{
		roomService: roomService,
	}
}

func (h *RoomHandler) CreateRoom(c *fiber.Ctx) error {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(
			fiber.Map{
				"error": "Invalid token",
			},
		)
	}

	var request dto.CreateRoomRequest
	request.Name = c.FormValue("name")

	// Handle optional file upload
	file, err := c.FormFile("room_image")
	if err == nil && file != nil {
		// Save the uploaded file
		savedPath, saveErr := utils.SaveUploadedFile(file, "rooms")
		if saveErr != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				fiber.Map{
					"error": "Failed to save image",
				},
			)
		}
		request.Avatar = savedPath
	}

	if request.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"error": "Room name is required",
			},
		)
	}

	room, err := h.roomService.CreateRoom(userID, request)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{
				"error": err.Error(),
			},
		)
	}

	// Redirect immediately to the created room
	return c.Redirect("/rooms/"+room.ID, fiber.StatusFound)
}

func (h *RoomHandler) GetRoomsList(c *fiber.Ctx) error {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(
			fiber.Map{
				"error": "Invalid token",
			},
		)
	}

	rooms, err := h.roomService.GetRoomsList(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{
				"error": err.Error(),
			},
		)
	}

	return c.JSON(rooms)
}

func (h *RoomHandler) UpdateRoom(c *fiber.Ctx) error {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(
			fiber.Map{
				"error": "Invalid token",
			},
		)
	}

	roomID := c.Params("roomId")
	if roomID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"error": "Room ID is required",
			},
		)
	}

	var request dto.UpdateRoomRequest
	request.Name = c.FormValue("name")

	// Handle optional file upload
	file, err := c.FormFile("room_image")
	if err == nil && file != nil {
		// Save the uploaded file
		savedPath, saveErr := utils.SaveUploadedFile(file, "rooms")
		if saveErr != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				fiber.Map{
					"error": "Failed to save image",
				},
			)
		}
		request.Avatar = savedPath
	}

	if request.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"error": "Room name is required",
			},
		)
	}

	err = h.roomService.UpdateRoom(userID, roomID, request)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(
			fiber.Map{
				"error": err.Error(),
			},
		)
	}

	return c.JSON(
		fiber.Map{
			"message": "Room updated successfully",
		},
	)
}

func (h *RoomHandler) DeleteRoom(c *fiber.Ctx) error {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(
			fiber.Map{
				"error": "Invalid token",
			},
		)
	}

	roomID := c.Params("roomId")
	if roomID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"error": "Room ID is required",
			},
		)
	}

	err = h.roomService.DeleteRoom(userID, roomID)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(
			fiber.Map{
				"error": err.Error(),
			},
		)
	}

	return c.JSON(
		fiber.Map{
			"message": "Room deleted successfully",
		},
	)
}

func (h *RoomHandler) GetRoomMembers(c *fiber.Ctx) error {
	roomID := c.Params("roomId")
	if roomID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"error": "Room ID is required",
			},
		)
	}

	baseURL := c.Protocol() + "://" + c.Get("Host")
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(
			fiber.Map{
				"error": "Invalid token",
			},
		)
	}

	members, err := h.roomService.GetRoomMembers(roomID, userID, baseURL)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(
			fiber.Map{
				"error": err.Error(),
			},
		)
	}

	return c.JSON(members)
}

func (h *RoomHandler) GetAllRoomsWithStatus(c *fiber.Ctx) error {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(
			fiber.Map{
				"error": "Invalid token",
			},
		)
	}

	baseURL := c.Protocol() + "://" + c.Get("Host")
	rooms, err := h.roomService.GetAllRooms(userID, baseURL)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{
				"error": err.Error(),
			},
		)
	}

	return c.JSON(rooms)
}

func (h *RoomHandler) GetLiveRooms(c *fiber.Ctx) error {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		userID = 0 // Allow anonymous users to see live rooms
	}

	baseURL := c.Protocol() + "://" + c.Get("Host")
	rooms, err := h.roomService.GetAllRoomsWithStatus(userID, baseURL)
	if err != nil {
		// Return an empty list or an error indicator partial
		return c.Status(fiber.StatusInternalServerError).SendString("<div>Error loading rooms</div>")
	}

	// Render the RoomList component directly
	var buf bytes.Buffer
	component := pages.RoomList(rooms, true) // `true` indicates it's for the "live" view
	if err := component.Render(context.Background(), &buf); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("<div>Render error</div>")
	}
	return c.Type("html").Send(buf.Bytes())
}
