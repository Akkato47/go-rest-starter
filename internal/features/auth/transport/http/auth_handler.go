package auth_handlers

import (
	"go-starter/internal/core/config"
	"go-starter/internal/core/transport/http/response"
	"go-starter/internal/core/transport/http/server"
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

type authHandler struct {
	service AuthService
	cfg     *config.Config
}

func NewAuthHandler(service AuthService, cfg *config.Config) AuthHandler {
	return &authHandler{service: service, cfg: cfg}
}

func (h *authHandler) Routes() []server.Route {
	return []server.Route{
		{Method: http.MethodPost, Path: "/register", Handler: h.register},
		{Method: http.MethodPost, Path: "/login", Handler: h.login},
	}
}

// register godoc
// @Summary      Register new user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      RegisterRequest  true  "Registration data"
// @Success      201      {object}  response.Response{data=domain.User}
// @Failure      400      {object}  response.Response
// @Failure      409      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Router       /auth/register [post]
func (h *authHandler) register(c *gin.Context) {
	ctx := c.Request.Context()
	var req RegisterRequest

	if err := c.BindJSON(&req); err != nil {
		response.SendFailResponse(c, http.StatusBadRequest, "invalid request body")
		return
	}

	createdUser, accessToken, err := h.service.Register(ctx, req.Mail, req.Password)
	if err != nil {
		response.SendFailResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	c.SetCookieData(&http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		HttpOnly: true,
		Expires:  time.Now().Add(1 * time.Hour),
		Secure:   h.cfg.ProductionStatus,
		SameSite: http.SameSiteLaxMode,
	})
	response.SendSuccessResponse(c, http.StatusCreated, createdUser)
}

// login godoc
// @Summary      Login
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      LoginRequest  true  "Login credentials"
// @Success      200      {object}  response.Response{data=domain.User}
// @Failure      400      {object}  response.Response
// @Failure      401      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Router       /auth/login [post]
func (h *authHandler) login(c *gin.Context) {
	ctx := c.Request.Context()
	var req LoginRequest

	if err := c.BindJSON(&req); err != nil {
		response.SendFailResponse(c, http.StatusBadRequest, "invalid request body")
		return
	}

	user, accessToken, err := h.service.Login(ctx, req.Mail, req.Password)
	if err != nil {
		response.SendFailResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	c.SetCookieData(&http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		HttpOnly: true,
		Expires:  time.Now().Add(1 * time.Hour),
		Secure:   h.cfg.ProductionStatus,
		SameSite: http.SameSiteLaxMode,
	})
	response.SendSuccessResponse(c, http.StatusOK, user)
}
