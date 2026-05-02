package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/grownmind/backend/internal/api/routes"
	"github.com/grownmind/backend/internal/auth"
	"github.com/grownmind/backend/internal/config"
	"github.com/grownmind/backend/internal/database"
	"github.com/grownmind/backend/internal/middleware"
	"github.com/grownmind/backend/internal/user"
)

func main() {
	cfg := config.Load()

	// PostgreSQL
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer db.Close()

	if err := database.Migrate(db); err != nil {
		log.Fatalf("migration: %v", err)
	}
	log.Println("PostgreSQL connected and migrated")

	// Redis
	rdb, err := database.ConnectRedis(cfg.RedisURL)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	defer rdb.Close()
	log.Println("Redis connected")

	userRepo := user.NewRepository(db)
	authSvc := auth.NewService(userRepo, rdb, cfg.JWTSecret, cfg.JWTRefreshSecret, cfg.GoogleClientID, cfg.GoogleClientSecret, cfg.GoogleRedirectURI)
	authHandler := auth.NewHandler(authSvc)
	jwt := middleware.NewJWTMiddleware(cfg.JWTSecret)

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})

	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Next: func(c *fiber.Ctx) bool { return c.Path() == "/favicon.ico" },
	}))
	app.Get("/favicon.ico", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusNoContent) })
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.AllowOrigins,
		AllowHeaders: "Origin, Content-Type, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	api := app.Group("/api")
	routes.Register(api.Group("/"+cfg.APIVersion), authHandler, jwt)

	log.Printf("Server starting on :%s  (API /api/%s)", cfg.Port, cfg.APIVersion)
	log.Fatal(app.Listen(":" + cfg.Port))
}
