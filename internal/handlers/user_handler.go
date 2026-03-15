package handlers

import (
	"go-starter/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func GetUserHandler(conn *pgx.Conn) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		userId := c.GetInt("user_id")

		userData, err := repository.GetUserById(ctx, conn, userId)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Error while get user data: " + err.Error()})
		}

		c.JSON(http.StatusOK, userData)
	}
}
