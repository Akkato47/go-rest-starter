package handlers

import (
	"go-starter/internal/common/response"
	"go-starter/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler interface {
	GetUserHandler() gin.HandlerFunc
}

type userHandler struct {
	service services.UserService
}

func NewUserHandler(service services.UserService) UserHandler {
	return &userHandler{
		service: service,
	}
}

func (h *userHandler) GetUserHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		userId := c.GetInt("user_id")

		userData, err := h.service.GetUserById(ctx, userId)
		if err != nil {
			response.SendFailResponse(c, http.StatusUnauthorized, err.Error())
			return
		}

		response.SendSuccessResponse(c, http.StatusOK, userData)
	}
}
