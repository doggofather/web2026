package handler

import (
	"fmt"
	"lab1/internal/app/config"
	"lab1/internal/app/redis"
	"lab1/internal/app/repository"
	"lab1/internal/app/role"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go"
)

type Handler struct {
	Repository *repository.Repository
	config     config.Config
	redis      *redis.Client
}

func NewHandler(rep *repository.Repository, rdb *redis.Client) *Handler {
	return &Handler{
		Repository: rep,
		redis:      rdb, // сохраняем в структуру
	}
}

var minioClient *minio.Client

func initMinIO() {
	var err error
	// want (string, string, string, bool)
	minioClient, err = minio.New(
		"localhost:9000",
		"minioadmin",
		"minioadmin",
		false, // это useSSL
	)
	if err != nil {
		fmt.Println("Error initializing MinIO:", err)
	}
}

// RegisterHandler Функция, в которой мы отдельно регистрируем маршруты, чтобы не писать все в одном месте
func (h *Handler) RegisterHandler(router *gin.Engine) {
	// --- ПУБЛИЧНЫЕ РОУТЫ ---
	// Эти роуты доступны без токена
	router.POST("/sign_up", h.Register)
	router.POST("/login", h.Login)

	// --- ЗАЩИЩЕННЫЕ РОУТЫ ---
	// Создаем группу и применяем к ней ваш middleware проверки авторизации
	protected := router.Group("/")
	protected.Use(h.WithAuthCheck)
	{
		// Роуты для Speakers
		protected.GET("/speakers", h.GetSpeakers)
		protected.GET("/speaker/:id", h.GetSpeaker)
		protected.POST("/delete-speaker", h.DeleteSpeaker)
		protected.POST("/add-speaker/:id", h.AddSpeaker)
		protected.POST("/speakers", h.CreateSpeaker)
		protected.PUT("/speakers/:id", h.UpdateSpeaker)
		protected.DELETE("/speakers/:id", h.CheckRole(role.Manager), h.DeleteSpeakerAPI)
		protected.POST("/speakers/:id/add-to-draft", h.AddSpeakerToDraft)
		protected.POST("/speakers/:id/image/:img_id", h.UploadSpeakerImage)

		// Роуты для Meetups
		protected.GET("/meetup/:id", h.GetMeetup)
		protected.GET("/cart-icon", h.GetCartIcon)
		protected.GET("/meetups", h.CheckRole(role.Organiser, role.Manager), h.GetMeetups)
		protected.GET("/meetups/:id", h.GetMeetupAPI)
		protected.PUT("/meetups/:id", h.UpdateMeetup)
		protected.PUT("/meetups/:id/form", h.FormMeetup)
		protected.PUT("/meetups/:id/complete", h.CheckRole(role.Organiser, role.Manager), h.CompleteOrRejectMeetup)
		protected.DELETE("/meetups/:id", h.DeleteMeetup)
		protected.POST("/meetups/add-to-draft", h.AddMeetupToDraft)

		// Роуты для SpeakerMeetup (м-м)
		protected.DELETE("/speaker-meetup/:meetup_id/:speaker_id", h.DeleteFromMeetup)
		protected.PUT("/speaker-meetup/:meetup_id/:speaker_id", h.UpdateSpeakerMeetup)
		protected.POST("/speaker-meetup/:meetup_id/:speaker_id", h.AddSpeakerToMeetup)

		// Профиль пользователя
		protected.GET("/profile/:id", h.GetUserProfile)
		protected.PUT("/profile/:id", h.UpdateUserProfile)
		router.POST("/logout", h.Logout)
	}
}

// RegisterStatic То же самое, что и с маршрутами, регистрируем статику
func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*")
	router.Static("/styles", "./resources/styles")
	router.Static("/img", "./resources/img")
}

func (h *Handler) CheckRole(allowedRoles ...role.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Достаем ID пользователя, который положил туда WithAuthCheck
		val, exists := c.Get("user_id")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user id not found in context"})
			return
		}
		userID := val.(uint)

		// 2. Получаем пользователя из БД, чтобы проверить его актуальную роль
		user, err := h.Repository.GetUserProfile(userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user role"})
			return
		}

		// 3. Проверяем, подходит ли роль пользователя под разрешенные
		isAllowed := false
		for _, r := range allowedRoles {
			if user.Role == r {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "access denied: you don't have enough permissions"})
			return
		}

		// Если роль совпала, идем дальше к хендлеру
		c.Next()
	}
}
