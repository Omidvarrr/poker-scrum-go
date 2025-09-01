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
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/websocket/v2"
)

func main() {
	cfg := config.LoadConfig()

	// inject secrets from config
	utils.SetJWTSecret(cfg.JWTSecret)
	utils.SetOAuthStateSecret([]byte(cfg.OAuthStateSecret))

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
	voteRepo := sqlite.NewVoteRepository(db)

	authService := services.NewAuthService(userRepo)
	greetingService := services.NewGreetingService(userRepo)
	profileService := services.NewProfileService(userRepo)
	roomService := services.NewRoomService(roomRepo, userRepo)
	voteService := services.NewVoteService(voteRepo, roomRepo)

	authHandler := handlers.NewAuthHandler(authService, greetingService)
	profileHandler := handlers.NewProfileHandler(profileService)
	roomHandler := handlers.NewRoomHandler(roomService)
	wsHandler := handlers.NewWebSocketHandler(roomRepo).WithVoteService(voteService)

	// page handler for templ rendering
	pageHandler := handlers.NewPageHandler(roomService, greetingService, voteService, profileService)
	// oauth handler with tuned session store
	sessionStore := session.New(
		session.Config{
			CookieName:     "sid",
			CookieHTTPOnly: true,
			CookieSecure:   false,
			Expiration:     15 * time.Minute,
		},
	)
	oauthHandler := handlers.NewOAuthHandler(cfg, userRepo).WithSessionStore(sessionStore)

	app := fiber.New(
		fiber.Config{
			ErrorHandler: func(c *fiber.Ctx, err error) error {
				code := fiber.StatusInternalServerError
				if e, ok := err.(*fiber.Error); ok {
					code = e.Code
				}
				return c.Status(code).JSON(
					fiber.Map{
						"error": err.Error(),
					},
				)
			},
		},
	)

	app.Use(logger.New())
	app.Use(cors.New())

	err = os.MkdirAll("data/uploads", 0755)
	if err != nil {
		log.Fatal("Failed to create upload directory:", err)
	}
	app.Static("/uploads", "./data/uploads")
	app.Static("/static", "./web/static")

	api := app.Group("/api")

	api.Get("/auth/google/login", oauthHandler.GoogleLogin)
	api.Get("/auth/google/callback", oauthHandler.GoogleCallback)
	api.Post("/auth/logout", oauthHandler.Logout)

	protected := api.Group("", utils.JWTMiddleware())

	protected.Get("/greeting", authHandler.GetGreeting)

	protected.Post("/profile/complete", profileHandler.CompleteProfile)
	protected.Put("/profile", profileHandler.UpdateProfile)
	protected.Get("/profile", profileHandler.GetProfile)

	protected.Get("/rooms", roomHandler.GetRoomsList)
	protected.Get("/rooms/live", roomHandler.GetLiveRooms)
	protected.Get("/rooms/all", roomHandler.GetAllRoomsWithStatus)
	protected.Post("/rooms", roomHandler.CreateRoom)
	protected.Put("/rooms/:roomId", roomHandler.UpdateRoom)
	protected.Delete("/rooms/:roomId", roomHandler.DeleteRoom)
	protected.Get("/rooms/:roomId/members", roomHandler.GetRoomMembers)

	// Pages (templ)
	app.Get("/", pageHandler.Home)
	app.Get("/home", pageHandler.Home)
	app.Get("/login", pageHandler.Login)
	app.Get("/rooms", pageHandler.Rooms)
	app.Get("/create", pageHandler.Create)
	app.Get("/profile", pageHandler.Profile)
	app.Get("/rooms/:roomId", pageHandler.RoomVote)

	app.Use(
		"/ws", func(c *fiber.Ctx) error {
			if websocket.IsWebSocketUpgrade(c) {
				c.Locals("allowed", true)
				return c.Next()
			}
			return fiber.ErrUpgradeRequired
		},
	)

	app.Get("/ws", websocket.New(wsHandler.HandleWebSocket))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(app.Listen(":" + port))
}
