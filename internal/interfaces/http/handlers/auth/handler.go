package auth

import (
	"errors"
	"net/http"

	"coin-radar-gin/internal/domain/user"
	"coin-radar-gin/internal/interfaces/http/dto"
	"coin-radar-gin/internal/interfaces/http/request"
	authsvc "coin-radar-gin/internal/services/auth"

	"github.com/gin-gonic/gin"
)

// Handler exposes authentication routes.
type Handler struct {
	svc *authsvc.Service
}

// New creates a new auth Handler.
func New(svc *authsvc.Service) *Handler {
	return &Handler{svc: svc}
}

// Register mounts the auth routes onto the given router group.
func (h *Handler) Register(rg *gin.RouterGroup) {
	g := rg.Group("/auth")
	g.POST("/register", h.register)
	g.POST("/login", h.login)
	g.POST("/telegram", h.telegram)
	g.POST("/refresh", h.refresh)
}

func (h *Handler) register(c *gin.Context) {
	var req dto.RegisterRequest
	if !request.Bind(c, &req) {
		return
	}
	u, tokens, err := h.svc.Register(c.Request.Context(), authsvc.RegisterInput{
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		writeAuthError(c, err)
		return
	}
	c.JSON(http.StatusCreated, dto.Success(tokenResponse(tokens, u)))
}

func (h *Handler) login(c *gin.Context) {
	var req dto.LoginRequest
	if !request.Bind(c, &req) {
		return
	}
	u, tokens, err := h.svc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		writeAuthError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.Success(tokenResponse(tokens, u)))
}

func (h *Handler) telegram(c *gin.Context) {
	var req dto.TelegramLoginRequest
	if !request.Bind(c, &req) {
		return
	}
	u, tokens, err := h.svc.TelegramLogin(c.Request.Context(), authsvc.TelegramAuthData{
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
	c.JSON(http.StatusOK, dto.Success(tokenResponse(tokens, u)))
}

func (h *Handler) refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if !request.Bind(c, &req) {
		return
	}
	tokens, err := h.svc.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		writeAuthError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.Success(tokenResponse(tokens, nil)))
}

func tokenResponse(t authsvc.TokenPair, u *user.User) dto.TokenResponse {
	return dto.TokenResponse{
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
	case errors.Is(err, authsvc.ErrEmailTaken):
		c.JSON(http.StatusConflict, dto.Error("email_taken", "email already registered"))
	case errors.Is(err, authsvc.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, dto.Error("invalid_credentials", "invalid email or password"))
	case errors.Is(err, authsvc.ErrTelegramSignature),
		errors.Is(err, authsvc.ErrTelegramExpired):
		c.JSON(http.StatusUnauthorized, dto.Error("telegram_auth_failed", err.Error()))
	case errors.Is(err, authsvc.ErrInvalidToken),
		errors.Is(err, authsvc.ErrWrongTokenType):
		c.JSON(http.StatusUnauthorized, dto.Error("invalid_token", "invalid or expired token"))
	default:
		c.JSON(http.StatusInternalServerError, dto.Error("internal_error", "something went wrong"))
	}
}
