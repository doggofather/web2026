package repository

import (
	"errors"
	"lab1/internal/app/ds"
	"time"

	"github.com/sirupsen/logrus"
	//"github.com/sirupsen/logrus"
)

func (r *Repository) GetMeetup(id int) (ds.Meetup, error) {
	Meetup := ds.Meetup{}
	err := r.db.Where("id = ?", id).First(&Meetup).Error
	if err != nil {
		return ds.Meetup{}, err
	}

	return Meetup, nil
}

func (r *Repository) GetSpeakersIDsByMeetup(id int) ([]int, error) {
	var SpeakerIDs []int

	err := r.db.Model(&ds.SpeakerMeetup{}).Where("meetup_id = ? AND is_delete = ?", id, false).Pluck("speaker_id", &SpeakerIDs).Error
	if err != nil {
		return SpeakerIDs, err
	}
	return SpeakerIDs, nil
}

// GetCartCount для получения количества услуг в заявке (чатов в сообщении в моем случае)
func (r *Repository) GetCartCount() int64 {
	var meetupID uint
	var count int64
	creatorID := 1
	// пока что мы захардкодили id создателя заявки, в последующем вы сделаете авторизацию и будете получать его из JWT

	err := r.db.Model(&ds.Meetup{}).Where("creator_id = ?", creatorID).Select("id").First(&meetupID).Error
	if err != nil {
		return 0
	}

	err = r.db.Model(&ds.SpeakerMeetup{}).Where("meetup_id = ? AND is_delete = ?", meetupID, false).Count(&count).Error
	if err != nil {
		logrus.Println("Error counting records in lists_chats:", err)
	}

	return count
}

func (r *Repository) AddSpeaker(speakerID uint) error {
	meetupID := uint(1)
	speakerMeetup := ds.SpeakerMeetup{
		SpeakerID: speakerID,
		MeetupID:  meetupID,
	}
	// Use Create to insert a new row
	err := r.db.Create(&speakerMeetup).Error
	if err != nil {
		logrus.Println("Error adding speaker to meetup:", err)
		return err // Return the error so the caller can handle it
	}

	return nil
}

func (r *Repository) UpdateSpeaker(speakerID uint) error {
	meetupID := 1
	err := r.db.Model(&ds.SpeakerMeetup{}).Where("speaker_id = ? AND meetup_id = ?", speakerID, meetupID).UpdateColumn("is_delete", false).Error
	if err != nil {
		logrus.Println("Error counting records in lists_chats:", err)
	}

	return nil
}

func (r *Repository) DeleteSpeaker(speakerID uint) error {
	meetupID := 1
	err := r.db.Where("speaker_id = ? AND meetup_id = ?", speakerID, meetupID).Delete(&ds.SpeakerMeetup{}).Error
	//err := r.db.Model(&ds.SpeakerMeetup{}).Where("speaker_id = ? AND meetup_id = ?", speakerID, meetupID).UpdateColumn("is_delete", true).Error
	if err != nil {
		logrus.Println("Error counting records in lists_chats:", err)
	}

	return nil
}

// Новые методы для Meetups
func (r *Repository) GetCartIcon(userID uint) (uint, int) {
	var meetup ds.Meetup
	r.db.Where("creator_id = ? AND status = ?", userID, "draft").First(&meetup)
	if meetup.ID == 0 {
		return 0, 0
	}
	var count int64
	r.db.Model(&ds.SpeakerMeetup{}).Where("meetup_id = ?", meetup.ID).Count(&count)
	return meetup.ID, int(count)
}

func (r *Repository) GetMeetups(status, startDate, endDate string) ([]ds.Meetup, error) {
	var meetups []ds.Meetup
	query := r.db.Preload("Creator").Preload("Moderator")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if startDate != "" && endDate != "" {
		query = query.Where("created_at BETWEEN ? AND ?", startDate, endDate)
	}
	return meetups, query.Find(&meetups).Error
}

func (r *Repository) GetMeetupWithSpeakers(id int) (*ds.Meetup, error) {
	var meetup ds.Meetup
	err := r.db.Preload("Speakers").Preload("Creator").Preload("Moderator").First(&meetup, id).Error
	if err != nil {
		return nil, err
	}
	return &meetup, nil
}

func (r *Repository) UpdateMeetup(meetup *ds.Meetup) error {
	return r.db.Save(meetup).Error
}

func (r *Repository) FormMeetup(id, userID uint) (*ds.Meetup, error) {
	now := time.Now()
	var meetup ds.Meetup

	// 1. Обновляем статус
	result := r.db.Model(&meetup).
		Where("id = ? ", id).
		Updates(map[string]interface{}{
			"status":    "formed",
			"formed_at": &now,
		})

	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, errors.New("meetup not found or you are not the creator")
	}

	// 2. Загружаем обновленный объект со всеми связями
	err := r.db.Preload("Creator").Preload("Moderator").First(&meetup, id).Error
	return &meetup, err
}

func (r *Repository) CompleteOrRejectMeetup(id uint, action string, userID uint) (*ds.Meetup, error) {
	var updates map[string]interface{}
	now := time.Now()
	var meetup ds.Meetup

	if action == "complete" {
		// Логика расчета стоимости (как в твоем примере)
		var speakers []ds.Speaker
		r.db.Table("speakers").
			Joins("JOIN speaker_meetups ON speakers.id = speaker_meetups.speaker_id").
			Where("speaker_meetups.meetup_id = ?", id).
			Find(&speakers)

		updates = map[string]interface{}{
			"status":       "completed",
			"completed_at": &now,
			"moderator_id": userID,
			"is_finished":  true,
		}

		if len(speakers) > 0 {
			totalCost := 1000.0 // Пример стоимости
			r.db.Model(&ds.SpeakerMeetup{}).
				Where("meetup_id = ?", id).
				Update("value", totalCost/float64(len(speakers)))
		}
	} else {
		updates = map[string]interface{}{
			"status":       "rejected",
			"moderator_id": userID,
		}
	}

	// 1. Выполняем обновление
	if err := r.db.Model(&meetup).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, err
	}

	// 2. Подгружаем свежие данные для ответа
	err := r.db.Preload("Creator").Preload("Moderator").First(&meetup, id).Error
	return &meetup, err
}

func (r *Repository) DeleteMeetup(id uint) error {
	return r.db.Delete(&ds.Meetup{}, id).Error
}

// Новые методы для SpeakerMeetup
func (r *Repository) DeleteFromMeetup(speakerID, meetupID uint) error {
	return r.db.Where("speaker_id = ? AND meetup_id = ?", speakerID, meetupID).Delete(&ds.SpeakerMeetup{}).Error
}

func (r *Repository) UpdateSpeakerMeetup(speakerID, meetupID uint, updates map[string]interface{}) error {
	return r.db.Model(&ds.SpeakerMeetup{}).
		Where("speaker_id = ? AND meetup_id = ?", speakerID, meetupID).
		Updates(updates).Error
}

// In your repository file (e.g., repository.go)
func (r *Repository) DeleteSpeakerMeetupsByMeetupID(meetupID uint) error {
	return r.db.Where("meetup_id = ?", meetupID).Delete(&ds.SpeakerMeetup{}).Error
}
func (r *Repository) CreateDraftMeetup(userID uint) (*ds.Meetup, error) {
	draft := ds.Meetup{
		CreatorID:   userID,
		Status:      "draft",    // Статус по умолчанию
		Date:        time.Now(), // Заглушка даты (потом изменится через PUT)
		Title:       "Черновик", // Минимально допустимые данные для NOT NULL колонок
		Location:    "Не указано",
		IS_FINISHED: false,
	}

	if err := r.db.Create(&draft).Error; err != nil {
		return nil, err
	}

	return &draft, nil
}
func (r *Repository) GetCartIconData(userID uint) (uint, int64, error) {
	var meetup ds.Meetup
	var count int64

	// 1. Ищем черновик (статус draft) конкретного пользователя
	err := r.db.Model(&ds.Meetup{}).
		Where("creator_id = ? AND status = ?", userID, "draft").
		First(&meetup).Error

	if err != nil {
		// Если заявка не найдена, возвращаем 0, 0 без ошибки (или специфичный флаг)
		return 0, 0, err
	}

	// 2. Считаем количество спикеров в этой заявке
	// Если у тебя нет поля is_delete, убери это условие
	err = r.db.Model(&ds.SpeakerMeetup{}).
		Where("meetup_id = ?", meetup.ID).
		Count(&count).Error

	return meetup.ID, count, err
}
