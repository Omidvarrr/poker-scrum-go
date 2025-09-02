package handlers

import (
	"awesomeProject1/internal/dto"
	"awesomeProject1/internal/services"
	templates "awesomeProject1/web/templates"
	pages "awesomeProject1/web/templates/pages"
	"bytes"
	"context"

	"awesomeProject1/internal/utils"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
)

type PageHandler struct {
	roomService     *services.RoomService
	greetingService *services.GreetingService
	voteService     *services.VoteService
	profileService  *services.ProfileService
}

func NewPageHandler(
	roomService *services.RoomService, greetingService *services.GreetingService, voteService *services.VoteService,
	profileService *services.ProfileService,
) *PageHandler {
	return &PageHandler{
		roomService:     roomService,
		greetingService: greetingService,
		voteService:     voteService,
		profileService:  profileService,
	}
}

func (h *PageHandler) render(c *fiber.Ctx, title string, content templ.Component) error {
	var buf bytes.Buffer
	page := templates.Base(title, templates.AppShell(c.Path(), content))
	if err := page.Render(context.Background(), &buf); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("render error")
	}
	return c.Type("html").SendString(buf.String())
}

func (h *PageHandler) Home(c *fiber.Ctx) error {
	userID, _ := getUserIDIfAny(c)
	if userID == 0 {
		return c.Redirect("/login", fiber.StatusFound)
	}
	var greetingResp dto.GreetingResponse
	if userID > 0 {
		gr, err := h.greetingService.GetGreeting(userID)
		if err == nil {
			greetingResp = gr
		}
	}
	baseURL := c.Protocol() + "://" + c.Get("Host")
	rooms, _ := h.roomService.GetAllRoomsWithStatus(userID, baseURL)
	return h.render(c, "Home", pages.HomePage(greetingResp, rooms))
}

func (h *PageHandler) Login(c *fiber.Ctx) error {
	// Render login without AppShell (no header/nav)
	var buf bytes.Buffer
	page := templates.Base("Login", pages.LoginPage())
	if err := page.Render(context.Background(), &buf); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("render error")
	}
	return c.Type("html").SendString(buf.String())
}

func (h *PageHandler) RoomVote(c *fiber.Ctx) error {
	roomID := c.Params("roomId")
	userID, err := getUserIDIfAny(c)

	if err != nil || userID == 0 {
		return c.Redirect("/login", fiber.StatusFound)
	}

	baseURL := c.Protocol() + "://" + c.Get("Host")
	membersResp, err := h.roomService.GetRoomMembers(roomID, userID, baseURL)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Could not load room members: " + err.Error())
	}

	voteResp, err := h.voteService.GetVotes(userID, roomID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Could not load votes: " + err.Error())
	}

	// Get all rooms to find room name and check ownership
	allRooms, err := h.roomService.GetAllRooms(userID, baseURL)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Could not load room info: " + err.Error())
	}

	var currentRoom *dto.RoomResponse
	canManageVotes := false
	for _, r := range allRooms {
		if r.ID == roomID {
			currentRoom = &r
			canManageVotes = r.IsOwner
			break
		}
	}

	if currentRoom == nil {
		return c.Status(fiber.StatusNotFound).SendString("Room not found")
	}

	return h.render(
		c, "Vote", pages.VotingPage(currentRoom, membersResp.Members, voteResp, canManageVotes, userID),
	)
}

func (h *PageHandler) Rooms(c *fiber.Ctx) error {
	userID, _ := getUserIDIfAny(c)
	if userID == 0 {
		return c.Redirect("/login", fiber.StatusFound)
	}
	baseURL := c.Protocol() + "://" + c.Get("Host")
	rooms, _ := h.roomService.GetAllRooms(userID, baseURL)
	return h.render(c, "Rooms", pages.RoomsPage(rooms))
}

func (h *PageHandler) Create(c *fiber.Ctx) error {
	userID, _ := getUserIDIfAny(c)
	if userID == 0 {
		return c.Redirect("/login", fiber.StatusFound)
	}
	return h.render(c, "Create", pages.CreateRoomPage())
}

func (h *PageHandler) Profile(c *fiber.Ctx) error {
	userID, _ := getUserIDIfAny(c)
	if userID == 0 {
		return c.Redirect("/login", fiber.StatusFound)
	}
	baseURL := c.Protocol() + "://" + c.Get("Host")
	profile, _ := h.profileService.GetProfile(userID, baseURL)
	return h.render(c, "Profile", pages.ProfilePage(profile))
}

func (h *PageHandler) RoomSettings(c *fiber.Ctx) error {
	roomID := c.Params("roomId")
	userID, err := getUserIDIfAny(c)

	if err != nil || userID == 0 {
		return c.Redirect("/login", fiber.StatusFound)
	}

	baseURL := c.Protocol() + "://" + c.Get("Host")
	room, err := h.roomService.GetRoomByID(roomID, userID, baseURL)
	if err != nil {
		return c.Status(fiber.StatusForbidden).SendString("Access denied")
	}

	if !room.IsOwner {
		return c.Status(fiber.StatusForbidden).SendString("Only owners can access settings")
	}

	return h.render(c, "Room Settings", pages.RoomSettingsPage(room))
}

func getUserIDIfAny(c *fiber.Ctx) (int, error) {
	if id, err := utils.GetUserIDFromToken(c); err == nil {
		return id, nil
	}
	if token := c.Cookies("token"); token != "" {
		// Temporarily inject header to reuse the same validator
		c.Request().Header.Set("Authorization", "Bearer "+token)
		if id, err := utils.GetUserIDFromToken(c); err == nil {
			c.Request().Header.Del("Authorization")
			return id, nil
		}
		c.Request().Header.Del("Authorization")
	}
	return 0, nil
}
