package repository

import (
	"lab1/internal/app/ds"

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

/*
func (r *Repository) GetSpeakersIDs(ids int) ([]ds.SpeakerMeetup, error) {
	var Speakers []ds.Speaker

	for i, SpeakerMeetupRecord := range SpeakerMeetupRecords {
	speaker, err := h.Repository.GetSpeaker(SpeakerMeetupRecord.speaker_id)
	if err != nil {
		logrus.Error(err)
	}
}
})*/

/*
func (r *Repository) GetAllMeetups() ([]ds.Meetup, error) {
	// тут мы пользуемся ORM
	var meetups []ds.Meetup
	err := r.db.Find(&meetups).Error
	if err != nil {
		return nil, err
	}
	return meetups, nil
}

func (r *Repository) SearchMeetupsByName(title string) ([]ds.Meetup, error) {
	var meetups []ds.Meetup
	err := r.db.Where("title ILIKE ?", "%"+title+"%").Find(&meetups).Error
	if err != nil {
		return nil, err
	}
	return meetups, nil
}
*/
