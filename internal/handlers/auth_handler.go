package handlers

import (
	"go-starter/internal/common/response"
	"go-starter/internal/config"
	"go-starter/internal/services"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type RegisterRequest struct {
	Mail     string `json:"mail" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginRequest struct {
	Mail     string `json:"mail" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthHandler interface {
	RegisterHandler() gin.HandlerFunc
	LoginHandler() gin.HandlerFunc
}

type authHandler struct {
	service services.AuthService
	cfg     *config.Config
}

func NewAuthHandler(service services.AuthService, cfg *config.Config) AuthHandler {
	return &authHandler{
		service: service,
		cfg:     cfg,
	}
}

func (h *authHandler) RegisterHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		var registerRequest RegisterRequest

		if err := c.BindJSON(&registerRequest); err != nil {
			response.SendFailResponse(c, http.StatusBadRequest, "jsonReq"+err.Error())
			return
		}

		createdUser, accessToken, err := h.service.Register(ctx, registerRequest.Mail, registerRequest.Password)
		if err != nil {
			response.SendFailResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		c.SetCookieData(&http.Cookie{
			Name:     "access_token",
			Value:    accessToken,
			HttpOnly: true,
			Expires:  time.Now().Add(1 * time.Hour), // change with cfg
			Secure:   h.cfg.ProductionStatus,
			SameSite: http.SameSiteLaxMode,
		})
		response.SendSuccessResponse(c, http.StatusCreated, createdUser)
	}
}

func (h *authHandler) LoginHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var loginRequest LoginRequest
		ctx := c.Request.Context()

		if err := c.BindJSON(&loginRequest); err != nil {
			response.SendFailResponse(c, http.StatusBadRequest, "jsonReq"+err.Error())
			return
		}

		user, accessToken, err := h.service.Login(ctx, loginRequest.Mail, loginRequest.Password)
		if err != nil {
			response.SendFailResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		c.SetCookieData(&http.Cookie{
			Name:     "access_token",
			Value:    accessToken,
			HttpOnly: true,
			Expires:  time.Now().Add(1 * time.Hour), // change with cfg,
			Secure:   h.cfg.ProductionStatus,
			SameSite: http.SameSiteLaxMode,
		})
		response.SendSuccessResponse(c, http.StatusOK, user)
	}
}
