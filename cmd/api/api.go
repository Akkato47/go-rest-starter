package main

import (
	"go-starter/internal/config"
	"go-starter/internal/handlers"
	"go-starter/internal/middleware"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Referrer-Policy", "strict-origin")
		c.Header("Permissions-Policy", "geolocation=(), camera=(), microphone=()")
		c.Next()
	}
}

func Cors(cfg *config.Config) gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

func SetupHandlers(cfg *config.Config, conn *pgx.Conn) *gin.Engine {
	router := gin.Default()
	gin.SetMode(gin.ReleaseMode)

	router.Use(Cors(cfg))
	router.Use(SecurityHeaders())
	router.SetTrustedProxies(nil)

	router.Use(func(c *gin.Context) {
		if isShuttingDown.Load() {
			c.AbortWithStatusJSON(503, gin.H{"error": "server shutting down"})
			return
		}
		c.Next()
	})

	router.GET("/healthz", func(c *gin.Context) {
		if isShuttingDown.Load() {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "shutting down",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	router.GET("/test-grace", func(c *gin.Context) {
		select {
		case <-time.After(5 * time.Second):
			c.String(http.StatusOK, "Hello, world!")
		case <-c.Request.Context().Done():
			c.String(http.StatusRequestTimeout, "Request cancelled")
		}
	})

	router.POST("/auth/register", handlers.RegisterHandler(conn, cfg))
	router.POST("/auth/login", handlers.LoginHandler(conn, cfg))

	protectedUserRouter := router.Group("/user")
	protectedUserRouter.Use(middleware.AuthMiddleware(cfg))

	protectedUserRouter.GET("/data", handlers.GetUserHandler(conn))

	return router
}
