package main

import (
	"fmt"
	"go-starter/internal/config"
	"go-starter/internal/handlers"
	"go-starter/internal/middleware"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

const (
	_rateLimitWindow  = time.Minute
	_rateLimitMaxReqs = 10
)

func SetupSecurityHeaders() gin.HandlerFunc {
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

func SetupCors(cfg *config.Config) gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

func SetupRateLimiter(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		key := fmt.Sprintf("rate_limit:%s", c.ClientIP())

		pipe := rdb.Pipeline()
		incr := pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, _rateLimitWindow)

		if _, err := pipe.Exec(ctx); err != nil {
			c.Next()
			return
		}

		if incr.Val() > _rateLimitMaxReqs {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			return
		}

		c.Next()
	}
}

func SetupSlogMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		logger.Info("request",
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.String("query", query),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("latency", time.Since(start)),
			slog.String("client_ip", c.ClientIP()),
			slog.Int("body_size", c.Writer.Size()),
		)

		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				logger.Error("request error", slog.String("error", err.Error()))
			}
		}
	}
}

func SetupRouter(
	cfg *config.Config, pool *pgxpool.Pool, redisClient *redis.Client, logger *slog.Logger, authH handlers.AuthHandler, userH handlers.UserHandler) *gin.Engine {

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(SetupSlogMiddleware(logger))
	router.Use(SetupCors(cfg))
	router.Use(SetupSecurityHeaders())
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

	authRouter := router.Group("/auth")
	authRouter.Use(SetupRateLimiter(redisClient))
	authRouter.POST("/register", authH.RegisterHandler())
	authRouter.POST("/login", authH.LoginHandler())

	protectedUserRouter := router.Group("/user")
	protectedUserRouter.Use(middleware.AuthMiddleware(cfg))

	protectedUserRouter.GET("/data", userH.GetUserHandler())

	return router
}
