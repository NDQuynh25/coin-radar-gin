package auth

import (
	"errors"
	"net/http"

	"coin-radar-gin/internal/modules/user"
	"coin-radar-gin/internal/shared/response"
	"coin-radar-gin/internal/shared/validator"

	"github.com/gin-gonic/gin"
)

// Handler exposes authentication routes.
type AuthHandler struct {
	svc *Service
}

// New creates a new auth Handler.
func NewAuthHandler(svc *Service) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// Register mounts the auth routes onto the given router group.
func (h *AuthHandler) Register(rg *gin.RouterGroup) {
	g := rg.Group("/auth")
	g.POST("/register", h.register)
	g.POST("/login", h.login)
	g.POST("/telegram", h.telegram)
	g.POST("/refresh", h.refresh)
}

func (h *AuthHandler) register(c *gin.Context) {
	var req RegisterRequest
	if !validator.Bind(c, &req) {
		return
	}
	u, tokens, err := h.svc.Register(c.Request.Context(), RegisterInput{
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		writeAuthError(c, err)
		return
	}
	c.JSON(http.StatusCreated, response.Success(tokenResponse(tokens, u)))
}

func (h *AuthHandler) login(c *gin.Context) {
	var req LoginRequest
	if !validator.Bind(c, &req) {
		return
	}
	u, tokens, err := h.svc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		writeAuthError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(tokenResponse(tokens, u)))
}

func (h *AuthHandler) telegram(c *gin.Context) {
	var req TelegramLoginRequest
	if !validator.Bind(c, &req) {
		return
	}
	u, tokens, err := h.svc.TelegramLogin(c.Request.Context(), TelegramAuthData{
		ID:        req.ID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Username:  req.Username,
		PhotoURL:  req.PhotoURL,
		AuthDate:  req.AuthDate,
		Hash:      req.Hash,
	})
	if err != nil {
		writeAuthError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(tokenResponse(tokens, u)))
}

func (h *AuthHandler) refresh(c *gin.Context) {
	var req RefreshRequest
	if !validator.Bind(c, &req) {
		return
	}
	tokens, err := h.svc.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		writeAuthError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(tokenResponse(tokens, nil)))
}

func tokenResponse(t TokenPair, u *user.User) TokenResponse {
	return TokenResponse{
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    t.ExpiresIn,
		User:         u,
	}
}

// writeAuthError maps service errors to appropriate HTTP responses.
func writeAuthError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrEmailTaken):
		c.JSON(http.StatusConflict, response.Error("email_taken", "email already registered"))
	case errors.Is(err, ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, response.Error("invalid_credentials", "invalid email or password"))
	case errors.Is(err, ErrTelegramSignature),
		errors.Is(err, ErrTelegramExpired):
		c.JSON(http.StatusUnauthorized, response.Error("telegram_auth_failed", err.Error()))
	case errors.Is(err, ErrInvalidToken),
		errors.Is(err, ErrWrongTokenType):
		c.JSON(http.StatusUnauthorized, response.Error("invalid_token", "invalid or expired token"))
	default:
		c.JSON(http.StatusInternalServerError, response.Error("internal_error", "something went wrong"))
	}
}
