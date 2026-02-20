package handler

import (
	"lab1/internal/app/ds"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

const jwtPrefix = "Bearer "

/*
	func (h *Handler) WithAuthCheck(assignedRoles ...role.Role) func(ctx *gin.Context) {
		return func(gCtx *gin.Context) {
			jwtStr := gCtx.GetHeader("Authorization")
			if !strings.HasPrefix(jwtStr, jwtPrefix) { // если нет префикса то нас дурят!
				gCtx.AbortWithStatus(http.StatusForbidden) // отдаем что нет доступа

				return // завершаем обработку
			}
			// отрезаем префикс
			jwtStr = jwtStr[len(jwtPrefix):]
			log.Println(jwtStr)
			token, err := jwt.ParseWithClaims(jwtStr, &ds.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
				return []byte(h.config.JWT.Token), nil
			})
			if err != nil {
				gCtx.AbortWithStatus(http.StatusForbidden)
				log.Println(err)

				return
			}

			myClaims := token.Claims.(*ds.JWTClaims)

			for _, oneOfAssignedRole := range assignedRoles {
				if myClaims.Role == oneOfAssignedRole {
					gCtx.AbortWithStatus(http.StatusForbidden)

					return
				}
			}

		}

}
*/
func (h *Handler) WithAuthCheck(gCtx *gin.Context) {
	jwtStr := gCtx.GetHeader("Authorization")
	if !strings.HasPrefix(jwtStr, jwtPrefix) { // если нет префикса то нас дурят!
		gCtx.AbortWithStatus(http.StatusForbidden) // отдаем что нет доступа
		return                                     // завершаем обработку
	}

	// отрезаем префикс
	jwtStr = jwtStr[len(jwtPrefix):]

	_, err := jwt.ParseWithClaims(jwtStr, &ds.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.config.JWT.Token), nil
	})
	if err != nil {
		gCtx.AbortWithStatus(http.StatusForbidden)
		log.Println(err)
		return
	}
}
