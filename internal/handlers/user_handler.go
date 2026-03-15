package handlers

import (
	"go-starter/internal/common"
	"go-starter/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func GetUserHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		userId := c.GetInt("user_id")

		userData, err := repository.GetUserById(ctx, pool, userId)
		if err != nil {
			common.SendFailResponse(c, http.StatusUnauthorized, "Error while get user data: "+err.Error())
			return
		}

		common.SendSuccessResponse(c, http.StatusOK, userData)
	}
}
