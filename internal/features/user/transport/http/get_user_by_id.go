package user_handlers

import (
	"go-starter/internal/core/transport/http/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *userHandler) getUserById(c *gin.Context) {
	ctx := c.Request.Context()
	userId := c.GetInt("user_id")

	userData, err := h.service.GetUserById(ctx, userId)
	if err != nil {
		response.SendFailResponse(c, http.StatusInternalServerError, "failed to get user data")
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, userData)
}
