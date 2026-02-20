package handler

import (
	"lab1/internal/app/ds"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

const jwtPrefix = "Bearer "

func (h *Handler) WithAuthCheck(gCtx *gin.Context) {
	// 1. Получаем заголовок
	header := gCtx.GetHeader("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		gCtx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing Bearer token"})
		return
	}

	// 2. Чистый токен (строго без префикса)
	jwtStr := strings.TrimPrefix(header, "Bearer ")

	// 3. ПРОВЕРКА В REDIS
	// Если в терминале ключ совпадет с тем, что в docker — всё заработает
	err := h.redis.CheckJWTInBlacklist(gCtx.Request.Context(), jwtStr)
	if err == nil {
		// Ошибки нет = Ключ найден = Токен забанен
		gCtx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token is blacklisted (logged out)"})
		return
	}

	// 4. ПАРСИНГ JWT
	claims := &ds.JWTClaims{}
	key := h.config.JWT.Token
	if key == "" {
		key = "mysecretkey123"
	}

	token, err := jwt.ParseWithClaims(jwtStr, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(key), nil
	})

	if err != nil || !token.Valid {
		gCtx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// 5. УСПЕХ
	gCtx.Set("user_id", claims.UserUUID)
	gCtx.Next()
}
