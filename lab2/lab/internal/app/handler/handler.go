package handler

import (
	"lab1/internal/app/repository"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

// RegisterHandler Функция, в которой мы отдельно регистрируем маршруты, чтобы не писать все в одном месте
func (h *Handler) RegisterHandler(router *gin.Engine) {
	router.GET("/speakers", h.GetSpeakers)
	router.GET("/speaker/:id", h.GetSpeaker)
	router.GET("/meetup/:id", h.GetMeetup)
	router.POST("/delete-speaker", h.DeleteSpeaker)
	router.POST("/add-speaker/:id", h.AddSpeaker)
}

// RegisterStatic То же самое, что и с маршрутами, регистрируем статику
func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*")
	router.Static("/styles", "./resources/styles")
	router.Static("/img", "./resources/img")
}
