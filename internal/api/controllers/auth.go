package controllers

import (
	time "time"
	userservice "vm-controller/internal/services/user_service"

	"net/http"

	gin "github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
	os "os"
)

type AuthController struct {
}

func (authController *AuthController) Login(c *gin.Context) {
	userService := &userservice.UserService{}
	user, err := userService.AuthenticateUser(c.PostForm("student_id"), c.PostForm("password"))

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "인증 실패"})
		return
	}

	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}).SignedString([]byte(os.Getenv("JWT_SECRET")))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "토큰 생성 실패"})
		return
	}

	c.SetCookie("authorization", "Bearer "+tokenString, 86400, "/", "", true, true)
	c.JSON(http.StatusOK, gin.H{"message": "로그인 성공"})
}
