package main

import (
	"context"
	"go-starter/internal/config"
	"go-starter/internal/database"
	"go-starter/internal/handlers"
	"go-starter/internal/middleware"
	"log"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	ctx := context.Background()

	cfg := config.NewConfig()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	conn, err := database.Connect(ctx, cfg.DbUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)

	router := gin.Default()
	router.SetTrustedProxies(nil)

	router.POST("/auth/register", handlers.RegisterHandler(conn, cfg))
	router.POST("/auth/login", handlers.LoginHandler(conn, cfg))

	protected := router.Group("/protected")
	protected.Use(middleware.AuthMiddleware(cfg))

	logger.Info("Starting server")
	router.Run(":" + cfg.AppPort)
}
