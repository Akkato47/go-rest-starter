package handlers

import (
	"go-starter/internal/common/response"
	"go-starter/internal/config"
	"go-starter/internal/model"
	"go-starter/internal/repository"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Mail     string `json:"mail" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginRequest struct {
	Mail     string `json:"mail" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func RegisterHandler(pool *pgxpool.Pool, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		var registerRequest RegisterRequest

		if err := c.BindJSON(&registerRequest); err != nil {
			response.SendFailResponse(c, http.StatusBadRequest, "jsonReq"+err.Error())
			return
		}

		if len(registerRequest.Password) < 6 {
			response.SendFailResponse(c, http.StatusBadRequest, "Password must be at least 6 characters long")
			return
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(registerRequest.Password), bcrypt.DefaultCost)
		if err != nil {
			response.SendFailResponse(c, http.StatusBadRequest, "Failed to hash password "+err.Error())
			return
		}

		user := &model.User{
			Mail:     registerRequest.Mail,
			Password: string(hashedPassword),
		}

		createdUser, err := repository.CreateUser(ctx, pool, user)
		if err != nil {
			if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
				response.SendFailResponse(c, http.StatusBadRequest, "Mail already registered")
				return
			}
			response.SendFailResponse(c, http.StatusBadRequest, "failed to create user "+err.Error())
			return
		}

		tokenExp := time.Now().Add(1 * time.Hour)

		accessTokenString, err := generateJWT(createdUser.ID, tokenExp, cfg)
		if err != nil {
			response.SendFailResponse(c, http.StatusInternalServerError, "Failed to generate token: "+err.Error())
			return
		}

		c.SetCookieData(&http.Cookie{
			Name:     "access_token",
			Value:    accessTokenString,
			HttpOnly: true, Expires: tokenExp,
			Secure:   cfg.ProductionStatus,
			SameSite: http.SameSiteLaxMode,
		})
		response.SendSuccessResponse(c, http.StatusCreated, createdUser)
	}
}

func LoginHandler(pool *pgxpool.Pool, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var loginRequest LoginRequest
		ctx := c.Request.Context()

		if err := c.BindJSON(&loginRequest); err != nil {
			response.SendFailResponse(c, http.StatusBadRequest, "jsonReq"+err.Error())
			return
		}

		user, err := repository.GetUserByMail(ctx, pool, loginRequest.Mail)
		if err != nil {
			response.SendFailResponse(c, http.StatusUnauthorized, "Something went wrong: "+err.Error())
			return
		}
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginRequest.Password))
		if err != nil {
			response.SendFailResponse(c, http.StatusUnauthorized, "Invalid credentials")
			return
		}

		tokenExp := time.Now().Add(1 * time.Hour)

		accessTokenString, err := generateJWT(user.ID, tokenExp, cfg)
		if err != nil {
			response.SendFailResponse(c, http.StatusInternalServerError, "Failed to generate token: "+err.Error())
			return
		}

		c.SetCookieData(&http.Cookie{
			Name:     "access_token",
			Value:    accessTokenString,
			HttpOnly: true, Expires: tokenExp,
			Secure:   cfg.ProductionStatus,
			SameSite: http.SameSiteLaxMode,
		})
		response.SendSuccessResponse(c, http.StatusOK, user)
	}
}

func generateJWT(userId string, expTime time.Time, cfg *config.Config) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userId,
		"exp":     expTime.Unix(),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return accessToken.SignedString([]byte(cfg.JwtSecret))
}
