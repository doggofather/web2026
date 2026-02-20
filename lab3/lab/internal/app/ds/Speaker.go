package ds

type Speaker struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"type:varchar(25);not null"`
	Title       string `gorm:"type:varchar(15);not null"`
	Format      string `gorm:"type:varchar(15);not null"`
	IS_FINISHED bool   `gorm:"type:boolean not null;default:false"`
	Description string `gorm:"type:varchar(100)"`
	Img         string `json:"pic"`
}
