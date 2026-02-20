package handler

import (
	"lab1/internal/app/config"
	"lab1/internal/app/redis"
	"lab1/internal/app/repository"

	//"lab1/internal/app/role"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Repository *repository.Repository
	config     config.Config
	redis      *redis.Client
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

type pingReq struct{}
type pingResp struct {
	Status string `json:"status"`
}

// RegisterHandler Функция, в которой мы отдельно регистрируем маршруты, чтобы не писать все в одном месте
func (h *Handler) RegisterHandler(router *gin.Engine) {
	router.GET("/speakers", h.GetSpeakers)
	router.GET("/speaker/:id", h.GetSpeaker)
	router.GET("/meetup/:id", h.GetMeetup)
	router.POST("/delete-speaker", h.DeleteSpeaker)
	router.POST("/add-speaker/:id", h.AddSpeaker)
	router.POST("/sign_up", h.Register)
	router.POST("/login", h.Login) // там где мы ранее уже заводили эндпоинты
	router.POST("/logout", h.Logout)
	// или ниженаписанное значит что доступ имеют менеджер и админ
	//router.Use(h.WithAuthCheck(role.Manager, role.Buyer)).GET("/ping", h.Ping)
	router.Use(h.WithAuthCheck).GET("/ping", h.Ping)

}

// RegisterStatic То же самое, что и с маршрутами, регистрируем статику
func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*")
	router.Static("/styles", "./resources/styles")
	router.Static("/img", "./resources/img")
}
