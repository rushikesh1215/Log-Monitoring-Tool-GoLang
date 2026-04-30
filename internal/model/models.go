package model

import (
	"time"
	"github.com/google/uuid"
)


type User struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email     string    `gorm:"uniqueIndex;not null"`
	Password  string    `gorm:"not null"` 
	CreatedAt time.Time
}


type Service struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name      string    `gorm:"uniqueIndex;not null"` 
	CreatedAt time.Time
}



type Log struct {
	ID         int64     `gorm:"primaryKey;autoIncrement"`
	ServiceID  uuid.UUID `gorm:"type:uuid;index;not null"`
	RawMessage string    `gorm:"type:text;not null"` 
	CreatedAt  time.Time `gorm:"index"`              
}