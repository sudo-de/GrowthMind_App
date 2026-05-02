package auth

import (
	"fmt"
	"net/url"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Logout(c *fiber.Ctx) error {
	var req refreshRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.RefreshToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "refresh_token is required"})
	}
	if err := h.svc.Logout(c.Context(), req.RefreshToken); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "logged out"})
}

type registerRequest struct {
	FullName string `json:"full_name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) Register(c *fiber.Ctx) error {
	var req registerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.Email == "" || req.Password == "" || req.Username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "email, password and username are required"})
	}

	u, access, refresh, err := h.svc.Register(c.Context(), req.Email, req.Password, req.FullName, req.Username)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"user":          u,
		"access_token":  access,
		"refresh_token": refresh,
	})
}

type loginRequest struct {
	Identifier string `json:"identifier"` // email or username
	Password   string `json:"password"`
}

func (h *Handler) Login(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.Identifier == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "identifier and password are required"})
	}

	u, access, refresh, err := h.svc.Login(c.Context(), req.Identifier, req.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{
		"user":          u,
		"access_token":  access,
		"refresh_token": refresh,
	})
}

type tokenRequest struct {
	AccessToken string `json:"access_token"`
}

func (h *Handler) GoogleSignIn(c *fiber.Ctx) error {
	var req tokenRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.AccessToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "access_token is required"})
	}

	u, access, refresh, err := h.svc.GoogleSignIn(c.Context(), req.AccessToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{
		"user":          u,
		"access_token":  access,
		"refresh_token": refresh,
	})
}

type appleRequest struct {
	IdentityToken string `json:"identity_token"`
	FullName      string `json:"full_name"`
}

func (h *Handler) AppleSignIn(c *fiber.Ctx) error {
	var req appleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.IdentityToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "identity_token is required"})
	}

	u, access, refresh, err := h.svc.AppleSignIn(c.Context(), req.IdentityToken, req.FullName)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{
		"user":          u,
		"access_token":  access,
		"refresh_token": refresh,
	})
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *Handler) RefreshToken(c *fiber.Ctx) error {
	var req refreshRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.RefreshToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "refresh_token is required"})
	}

	access, refresh, err := h.svc.RefreshToken(c.Context(), req.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{
		"access_token":  access,
		"refresh_token": refresh,
	})
}

func (h *Handler) GoogleOAuthInitiate(c *fiber.Ctx) error {
	authURL, err := h.svc.GoogleOAuthURL(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Redirect(authURL, fiber.StatusTemporaryRedirect)
}

func (h *Handler) GoogleOAuthCallback(c *fiber.Ctx) error {
	if errParam := c.Query("error"); errParam != "" {
		return c.Redirect("growthmind://auth?error="+url.QueryEscape(errParam), fiber.StatusTemporaryRedirect)
	}
	code := c.Query("code")
	state := c.Query("state")
	if code == "" || state == "" {
		return c.Redirect("growthmind://auth?error=missing_params", fiber.StatusTemporaryRedirect)
	}
	_, access, refresh, err := h.svc.GoogleOAuthCallback(c.Context(), code, state)
	if err != nil {
		return c.Redirect("growthmind://auth?error="+url.QueryEscape(err.Error()), fiber.StatusTemporaryRedirect)
	}
	deepLink := fmt.Sprintf("growthmind://auth?access_token=%s&refresh_token=%s",
		url.QueryEscape(access), url.QueryEscape(refresh))
	return c.Redirect(deepLink, fiber.StatusTemporaryRedirect)
}

func (h *Handler) Me(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	u, err := h.svc.GetMe(c.Context(), userID)
	if err != nil || u == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}
	return c.JSON(fiber.Map{"user": u})
}
