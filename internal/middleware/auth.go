package middleware

import (
	strings "strings"

	gin "github.com/gin-gonic/gin"

	http "net/http"

	fmt "fmt"
	jwt "github.com/golang-jwt/jwt/v5"
)

func AuthGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 쿠키에서 "authorization" 값 가져오기
		tokenString, err := c.Cookie("authorization")

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "로그인 토큰이 없습니다."})
			c.Abort() // 중요: 이후의 핸들러 함수 호출을 중단함 (Guard의 핵심)
			return
		}

		// 2. "Bearer " 접두사 제거 및 토큰 검증 로직
		// (예: jwt.Parse 등을 활용한 실제 검증)
		if !strings.HasPrefix(tokenString, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "유효하지 않은 토큰 형식입니다."})
			c.Abort()
			return
		}

		tokenString = tokenString[7:] // "Bearer " 접두사 제거
		// 3. 검증 통과 시 사용자 정보를 Context에 저장 (Next 핸들러에서 사용 가능)

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte("secret"), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "유효하지 않은 토큰입니다."})
			c.Abort()
			return
		}

		// 4. 토큰 클레임에서 사용자 식별 정보(user_id) 추출
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "토큰 클레임을 읽을 수 없습니다."})
			c.Abort()
			return
		}

		userID, ok := claims["user_id"]
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "토큰에 사용자 정보가 누락되었습니다."})
			c.Abort()
			return
		}

		// 5. Context에 user_id를 저장하여 이후 핸들러에서 c.Get("user_id")로 접근 가능하게 함
		c.Set("user_id", userID)

		c.Next() // 다음 미들웨어 또는 핸들러로 진행
	}
}
