package models
	
import (
	"time"
)

type Users struct {
	Id	string	`json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Username	string	`json:"username" gorm:"not null"`
	Password	string	`json:"password" gorm:"not null"`
	IsVerified	bool	`json:"is_verified" gorm:"not null"`
	CreatedAt	time.Time	`json:"created_at" gorm:"not null"`
	LastUpdatedAt	time.Time	`json:"last_updated_at" gorm:"not null"`
}

func (Users) TableName() string {
	return "users"
}

type Sessions struct {
	Id	string	`json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ExpiresAt	time.Time	`json:"expires_at" gorm:"not null"`
	UserId	string	`json:"user_id" gorm:"type:uuid;not null"`
	CreatedAt	time.Time	`json:"created_at" gorm:"not null"`
	LastUpdatedAt	time.Time	`json:"last_updated_at" gorm:"not null"`
}

func (Sessions) TableName() string {
	return "sessions"
}

type Logs struct {
	Id	int	`json:"id" gorm:"primaryKey"`
	Message	string	`json:"message" gorm:"not null"`
	Level	int16	`json:"level" gorm:"not null"`
	CreatedAt	time.Time	`json:"created_at" gorm:"not null"`
}

func (Logs) TableName() string {
	return "logs"
}

