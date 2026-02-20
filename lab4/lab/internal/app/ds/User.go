package ds

import (
	"lab1/internal/app/role"

	"github.com/google/uuid"
)

type User struct {
	ID       uuid.UUID `gorm:"primary_key" json:"id"`
	Login    string    `gorm:"type:varchar(25);unique;not null" json:"login"`
	Password string    `gorm:"type:varchar(100);not null" json:"-"`
	Role     role.Role `sql:"type:string;"`
}
