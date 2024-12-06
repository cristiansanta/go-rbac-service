package models

import "time"

type TokenBlacklist struct {
	ID        int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Token     string    `json:"token" gorm:"type:text;not null;unique"`
	ExpiresAt time.Time `json:"expires_at" gorm:"type:timestamp;not null"`
}

func (TokenBlacklist) TableName() string {
	return "token_blacklist"
}