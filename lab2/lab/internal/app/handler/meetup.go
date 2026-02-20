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

func (h *Handler) GetMeetup(ctx *gin.Context) {
	idStr := ctx.Param("id") // получаем id заказа из урла (то есть из /Meetup/:id)
	// через двоеточие мы указываем параметры, которые потом сможем считать через функцию выше
	id_meetup, err := strconv.Atoi(idStr) // так как функция выше возвращает нам строку, нужно ее преобразовать в int
	if err != nil {
		logrus.Error(err)
	}

	meetup, err := h.Repository.GetMeetup(id_meetup)
	if err != nil {
		logrus.Error(err)
	}

	speaker_ids, err := h.Repository.GetSpeakersIDsByMeetup((id_meetup))
	if err != nil {
		logrus.Error(err)
	}

	fmt.Print(speaker_ids)
	var data []ds.Speaker

	for _, speaker_id := range speaker_ids {
		data_piece, err := h.Repository.GetSpeaker(speaker_id)
		if err != nil {
			logrus.Error(err)
		}
		data = append(data, data_piece)
	}
	fmt.Print(data)
	ctx.HTML(http.StatusOK, "meetup.html", gin.H{
		"meetup": meetup,
		"data":   data,
	})
}

func (h *Handler) AddSpeaker(ctx *gin.Context) {
	// считываем значение из формы, которую мы добавим в наш шаблон
	// Считываем значение из URL-параметра (предполагаем маршрут POST /add-speaker/:speaker_id)
	strId := ctx.Param("id")
	if strId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "speaker_id is required in URL path",
		})
		return
	}

	id, err := strconv.Atoi(strId)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid speaker_id: must be a number",
		})
		return
	}

	// Проверяем, что ID положительный (чтобы избежать 0 или отрицательных значений)
	if id <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "speaker_id must be a positive integer",
		})
		return
	}
	err = h.Repository.AddSpeaker(uint(id))
	if err != nil && !strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
		return
	}

	// после вызова сразу произойдет обновление страницы
	ctx.Redirect(http.StatusFound, "/speakers")
}

func (h *Handler) DeleteSpeaker(ctx *gin.Context) {
	// считываем значение из формы, которую мы добавим в наш шаблон
	strId := ctx.PostForm("speaker_id")
	id, err := strconv.Atoi(strId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}
	// Вызов функции добавления чата в заявку
	err = h.Repository.DeleteSpeaker(uint(id))
	if err != nil && !strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
		return
	}

	// после вызова сразу произойдет обновление страницы
	ctx.Redirect(http.StatusFound, "/meetup/1")
}

/*
func (h *Handler) GetAllMeetups(ctx *gin.Context) {
	var meetups []ds.Meetup
	var err error

	search := ctx.Query("search")
	if search == "" {
		meetups, err = h.Repository.GetAllMeetups()
	} else {
		meetups, err = h.Repository.SearchMeetupsByName(search)
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	ctx.HTML(http.StatusOK, "meetups.html", gin.H{
		"data":       meetups,
		"cart_count": h.Repository.GetCartCount(),
		"search":     search,
	})
}
*/
