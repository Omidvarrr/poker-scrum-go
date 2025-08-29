package main

import (
	"awesomeProject1/internal/config"
	"awesomeProject1/internal/database"
	"awesomeProject1/internal/handlers"
	"awesomeProject1/internal/repositories/sqlite"
	"awesomeProject1/internal/services"
	"awesomeProject1/internal/utils"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/websocket/v2"
)

func main() {
	cfg := config.LoadConfig()

	db, err := database.Connect(cfg.DatabasePath)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	err = database.RunMigrations(cfg.DatabasePath)
	if err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	userRepo := sqlite.NewUserRepository(db)
	roomRepo := sqlite.NewRoomRepository(db)
	joinRequestRepo := sqlite.NewJoinRequestRepository(db)
	voteRepo := sqlite.NewVoteRepository(db)

	authService := services.NewAuthService(userRepo)
	greetingService := services.NewGreetingService()
	profileService := services.NewProfileService(userRepo)
	roomService := services.NewRoomService(roomRepo, joinRequestRepo, userRepo)
	voteService := services.NewVoteService(voteRepo, roomRepo)

	authHandler := handlers.NewAuthHandler(authService, greetingService)
	profileHandler := handlers.NewProfileHandler(profileService)
	roomHandler := handlers.NewRoomHandler(roomService)
	voteHandler := handlers.NewVoteHandler(voteService)
	wsHandler := handlers.NewWebSocketHandler(roomRepo)

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	app.Use(logger.New())
	app.Use(cors.New())

	err = os.MkdirAll("data/uploads", 0755)
	if err != nil {
		log.Fatal("Failed to create upload directory:", err)
	}
	app.Static("/uploads", "./data/uploads")

	api := app.Group("/api")

	api.Post("/auth/login", authHandler.Login)

	protected := api.Group("", utils.JWTMiddleware())

	protected.Get("/greeting", authHandler.GetGreeting)

	protected.Post("/profile/complete", profileHandler.CompleteProfile)
	protected.Put("/profile", profileHandler.UpdateProfile)
	protected.Get("/profile", profileHandler.GetProfile)

	protected.Get("/rooms", roomHandler.GetRoomsList)
	protected.Get("/rooms/joined", roomHandler.GetJoinedRooms)
	protected.Get("/rooms/all", roomHandler.GetAllRoomsWithStatus)
	protected.Post("/rooms", roomHandler.CreateRoom)
	protected.Post("/rooms/join", roomHandler.RequestJoinRoom)
	protected.Get("/rooms/requests", roomHandler.GetUserJoinRequests)

	protected.Get("/rooms/:roomId/requests", roomHandler.GetJoinRequests)
	protected.Post("/rooms/requests/handle", roomHandler.HandleJoinRequest)
	protected.Put("/rooms/:roomId", roomHandler.UpdateRoom)
	protected.Delete("/rooms/:roomId", roomHandler.DeleteRoom)
	protected.Get("/rooms/:roomId/members", roomHandler.GetRoomMembers)
	protected.Post("/rooms/:roomId/members/manage", roomHandler.ManageRoomMember)

	protected.Post("/rooms/:roomId/vote", voteHandler.CastVote)
	protected.Get("/rooms/:roomId/votes", voteHandler.GetVotes)
	protected.Post("/rooms/:roomId/votes/reveal", voteHandler.RevealVotes)
	protected.Post("/rooms/:roomId/votes/reset", voteHandler.ResetVotes)

	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws", websocket.New(wsHandler.HandleWebSocket))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(app.Listen(":" + port))
}
