package user_handlers

import (
	"go-starter/internal/core/transport/http/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

// getUserById godoc
// @Summary      Get current user
// @Tags         user
// @Produce      json
// @Security     CookieAuth
// @Success      200  {object}  response.Response{data=domain.User}
// @Failure      401  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /user/data [get]
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
