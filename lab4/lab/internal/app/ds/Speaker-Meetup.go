package ds

type SpeakerMeetup struct {
	ID uint `gorm:"primaryKey"`
	// здесь создаем Unique key, указывая общий uniqueIndex
	SpeakerID uint `gorm:"not null;uniqueIndex:idx_speaker_meetup"`
	MeetupID  uint `gorm:"not null;uniqueIndex:idx_speaker_meetup"`

	Speaker Speaker `gorm:"foreignKey:SpeakerID"`
	Meetup  Meetup  `gorm:"foreignKey:MeetupID"`
}
