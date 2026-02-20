package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"lab1/internal/app/ds"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handler) GetSpeakers(ctx *gin.Context) {
	var speakers []ds.Speaker
	var err error
	var count int64

	searchQuery := ctx.Query("query") // получаем значение из нашего поля
	if searchQuery == "" {            // если поле поиска пусто, то просто получаем из репозитория все записи
		speakers, err = h.Repository.GetSpeakers()
		if err != nil {
			logrus.Error(err)
		}
	} else {
		speakers, err = h.Repository.GetSpeakersByTitle(searchQuery) // в ином случае ищем заказ по заголовку
		if err != nil {
			logrus.Error(err)
		}
	}

	count = h.Repository.GetCartCount()

	fmt.Print(count)

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"speakers":   speakers,
		"cart_count": count,
	})
}

func (h *Handler) GetSpeaker(ctx *gin.Context) {
	idStr := ctx.Param("id") // получаем id заказа из урла (то есть из /speaker/:id)
	// через двоеточие мы указываем параметры, которые потом сможем считать через функцию выше
	id, err := strconv.Atoi(idStr) // так как функция выше возвращает нам строку, нужно ее преобразовать в int
	if err != nil {
		logrus.Error(err)
	}

	speaker, err := h.Repository.GetSpeaker(id)
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "speaker.html", gin.H{
		"speaker": speaker,
	})
}
