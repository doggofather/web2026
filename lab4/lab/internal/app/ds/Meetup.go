package ds

import "time"

type Meetup struct {
	ID          int       `gorm:"primaryKey"`
	Title       string    `gorm:"type:varchar(15);not null"`
	Date        time.Time `gorm:"type:timestamp;not null"`
	Description string    `gorm:"type:varchar(100)"`
	Location    string    `gorm:"type:varchar(25);not null"`
	IS_FINISHED bool      `gorm:"type:boolean not null;default:false"`
	CreatorID   uint      `gorm:"not null"`
	ModeratorID uint

	Creator   User `gorm:"foreignKey:CreatorID"`
	Moderator User `gorm:"foreignKey:ModeratorID"`
}
