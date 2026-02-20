package ds

import "lab1/internal/app/role"

type User struct {
	UUID     uint      `gorm:"column:id;primaryKey"`
	Login    string    `gorm:"type:varchar(25);unique;not null" json:"login"`
	Password string    `gorm:"type:varchar(100);not null" json:"-"`
	Role     role.Role `sql:"type:string;"`
}
