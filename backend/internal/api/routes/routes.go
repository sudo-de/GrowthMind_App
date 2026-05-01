package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/grownmind/backend/internal/auth"
	"github.com/grownmind/backend/internal/middleware"
)

// Register mounts all API routes onto the given group.
// The version prefix (/v1, /v2, ...) is set by the caller via API_VERSION in .env.
func Register(rg fiber.Router, authHandler *auth.Handler, jwt *middleware.JWTMiddleware) {
	a := rg.Group("/auth")
	a.Post("/register", authHandler.Register)
	a.Post("/login", authHandler.Login)
	a.Post("/google", authHandler.GoogleSignIn)
	a.Post("/apple", authHandler.AppleSignIn)
	a.Post("/refresh", authHandler.RefreshToken)
	a.Post("/logout", authHandler.Logout)
	a.Get("/me", jwt.Protect(), authHandler.Me)
}
