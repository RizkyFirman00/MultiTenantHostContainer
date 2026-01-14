package domain

import (
	"time"

	"github.com/google/uuid"
)

// User merepresentasikan pengguna sistem (tenant)
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Username     string    `gorm:"type:varchar(50);uniqueIndex;not null"`
	Email        string    `gorm:"type:varchar(100);uniqueIndex;not null"`
	PasswordHash string    `gorm:"type:text;not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	
	// Relations
	Projects []Project `gorm:"foreignKey:UserID"`
}
