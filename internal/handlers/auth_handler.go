package handlers

import (
	"go-starter/internal/config"
	"go-starter/internal/model"
	"go-starter/internal/repository"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Mail     string `json:"mail" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginRequest struct {
	Mail     string `json:"mail" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func RegisterHandler(conn *pgx.Conn, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var registerRequest RegisterRequest

		if err := c.BindJSON(&registerRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "jsonReq" + err.Error()})
			return
		}

		if len(registerRequest.Password) < 6 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 6 characters long"})
			return
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(registerRequest.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to hash password " + err.Error()})
			return
		}

		user := &model.User{
			Mail:     registerRequest.Mail,
			Password: string(hashedPassword),
		}

		createdUser, err := repository.CreateUser(conn, user)
		if err != nil {
			if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Mail already registred"})
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to create user " + err.Error()})
			return
		}

		tokenExp := time.Now().Add(1 * time.Hour)

		claims := jwt.MapClaims{
			"user_id": createdUser.ID,
			"exp":     tokenExp.Unix(),
		}

		accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		accessTokenString, err := accessToken.SignedString([]byte(cfg.JwtSecret))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token: " + err.Error()})
			return
		}

		c.SetCookieData(&http.Cookie{
			Name:     "access_token",
			Value:    accessTokenString,
			HttpOnly: true, Expires: tokenExp,
		})
		c.JSON(http.StatusCreated, createdUser)
	}
}

func LoginHandler(conn *pgx.Conn, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var loginRequest LoginRequest

		if err := c.BindJSON(&loginRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "jsonReq" + err.Error()})
			return
		}

		user, err := repository.GetuserByMail(conn, loginRequest.Mail)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginRequest.Password))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		tokenExp := time.Now().Add(1 * time.Hour)

		claims := jwt.MapClaims{
			"user_id": user.ID,
			"exp":     tokenExp.Unix(),
		}

		accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		accessTokenString, err := accessToken.SignedString([]byte(cfg.JwtSecret))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token: " + err.Error()})
			return
		}

		c.SetCookieData(&http.Cookie{
			Name:     "access_token",
			Value:    accessTokenString,
			HttpOnly: true, Expires: tokenExp,
		})
		c.JSON(http.StatusOK, user)
	}
}
