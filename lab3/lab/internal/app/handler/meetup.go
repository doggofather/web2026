package handler

import (
	"fmt"
	"lab1/internal/app/ds"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handler) GetMeetup(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	idStr := c.Param("id")
	id_meetup, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	meetup, err := h.Repository.GetMeetup(id_meetup)
	if err != nil {
		logrus.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	speaker_ids, err := h.Repository.GetSpeakersIDsByMeetup(id_meetup)
	if err != nil {
		logrus.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fmt.Print(speaker_ids)
	var data []ds.Speaker

	for _, speaker_id := range speaker_ids {
		data_piece, err := h.Repository.GetSpeaker(speaker_id)
		if err != nil {
			logrus.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		data = append(data, data_piece)
	}
	fmt.Print(data)

	c.JSON(http.StatusOK, gin.H{
		"meetup": meetup,
		"data":   data,
	})
}

func (h *Handler) AddSpeaker(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	strId := c.Param("id")
	if strId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "speaker_id is required in URL path"})
		return
	}

	id, err := strconv.Atoi(strId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid speaker_id: must be a number"})
		return
	}

	if id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "speaker_id must be a positive integer"})
		return
	}

	err = h.Repository.AddSpeaker(uint(id))
	if err != nil && !strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Speaker added successfully"})
}

func (h *Handler) DeleteSpeaker(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	strId := c.PostForm("speaker_id")
	id, err := strconv.Atoi(strId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid speaker_id: must be a number"})
		return
	}

	err = h.Repository.DeleteSpeaker(uint(id))
	if err != nil && !strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Speaker deleted successfully"})
}

func (h *Handler) GetCartIcon(c *gin.Context) {
	// 1. Получаем ID пользователя из контекста (JWT)
	val, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := val.(uint)

	// 2. Получаем данные из БД
	meetupID, count, err := h.Repository.GetCartIconData(userID)

	if err != nil {
		// Если записи нет (RecordNotFound)
		// По заданию: возвращаем 0 и -1 (или как у тебя в логике)
		c.JSON(http.StatusOK, gin.H{
			"cart_count": -1,
			"meetup_id":  0,
		})
		return
	}

	// 3. Возвращаем успешный ответ
	c.JSON(http.StatusOK, gin.H{
		"cart_count": count,
		"meetup_id":  meetupID,
	})
}

func (h *Handler) GetMeetups(c *gin.Context) {
	status := c.Query("status")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	meetups, err := h.Repository.GetMeetups(status, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, meetups)
}

func (h *Handler) GetMeetupAPI(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	meetup, err := h.Repository.GetMeetupWithSpeakers(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"meetup": meetup})
}

func (h *Handler) UpdateMeetup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid meetup ID"})
		return
	}

	var meetup ds.Meetup
	if err := c.ShouldBindJSON(&meetup); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON or missing fields: " + err.Error()})
		return
	}
	meetup.ID = uint(id)

	if err := h.Repository.UpdateMeetup(&meetup); err != nil {
		// Log error here if needed
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update meetup"})
		return
	}

	// Optionally fetch and return the updated meetup
	c.JSON(http.StatusOK, gin.H{"message": "Meetup updated!"})
}

func (h *Handler) FormMeetup(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	val, _ := c.Get("user_id")
	userID := val.(uint)

	// Вызываем обновленный метод репозитория
	_, err := h.Repository.FormMeetup(uint(id), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем весь объект митапа
	c.JSON(http.StatusOK, gin.H{"message": "Form created"})
}

func (h *Handler) CompleteOrRejectMeetup(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	action := c.Query("action")
	userID, _ := c.Get("user_id")

	if action != "complete" && action != "reject" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action"})
		return
	}

	// Получаем обновленный объект
	_, err := h.Repository.CompleteOrRejectMeetup(uint(id), action, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем митап в JSON
	c.JSON(http.StatusOK, gin.H{"message": "Completed!"})
}

func (h *Handler) DeleteMeetup(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	// Step 1: First, delete all related records in speaker_meetups for this meetup
	// Assuming your repository has a method like this; if not, add it (see below)
	if err := h.Repository.DeleteSpeakerMeetupsByMeetupID(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete associated speakers: " + err.Error()})
		return
	}

	// Step 2: Now delete the meetup itself
	if err := h.Repository.DeleteMeetup(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Meetup deleted successfully"})
}

// Новые функции для SpeakerMeetup
func (h *Handler) DeleteFromMeetup(c *gin.Context) {
	meetupIDStr := c.Param("meetup_id")
	speakerIDStr := c.Param("speaker_id")

	meetupID, err := strconv.Atoi(meetupIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid meetup_id"})
		return
	}
	speakerID, err := strconv.Atoi(speakerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid speaker_id"})
		return
	}

	if err := h.Repository.DeleteFromMeetup(uint(speakerID), uint(meetupID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Deleted from meetup"})
}

func (h *Handler) UpdateSpeakerMeetup(c *gin.Context) {
	// Bind JSON as before
	var sm ds.SpeakerMeetup
	if err := c.ShouldBindJSON(&sm); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse IDs from URL
	meetupIDStr := c.Param("meetup_id")
	speakerIDStr := c.Param("speaker_id")

	meetupID, err := strconv.Atoi(meetupIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid meetup_id"})
		return
	}
	speakerID, err := strconv.Atoi(speakerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid speaker_id"})
		return
	}

	// Prepare updates map if updating specific fields; for example:
	updates := map[string]interface{}{}
	// assuming sm has fields like Quantity, Order, etc.
	// populate as needed, e.g.,
	// updates["quantity"] = sm.Quantity

	// For demonstration, suppose you want to update all non-zero fields
	// or you can directly update entire struct, but your repo expects map.

	if err := h.Repository.UpdateSpeakerMeetup(uint(speakerID), uint(meetupID), updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Updated successfully"})
}
func (h *Handler) AddMeetupToDraft(c *gin.Context) {
	// Достаем ID создателя из контекста
	val, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authorized"})
		return
	}
	userID := val.(uint)

	// Создаем черновик через репозиторий
	meetup, err := h.Repository.CreateDraftMeetup(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create draft: " + err.Error()})
		return
	}

	// Возвращаем созданный митап (там будет ID, который нужен для PUT)
	c.JSON(http.StatusCreated, meetup)
}
