package domain

import (
	"time"

	"github.com/google/uuid"
)

// Project merepresentasikan aplikasi web yang dimiliki user
type Project struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID        uuid.UUID `gorm:"type:uuid;not null;index"`
	Name          string    `gorm:"type:varchar(100);not null"`
	Subdomain     string    `gorm:"type:varchar(63);uniqueIndex;not null"` // e.g., "blog" -> blog.domain.com
	ImageName     string    `gorm:"type:varchar(255);not null"`            // e.g., "nginx:alpine"
	ContainerPort int       `gorm:"not null"`                              // e.g., 80
	Status        string    `gorm:"type:varchar(20);default:'stopped'"`    // active, stopped
	CreatedAt     time.Time
	UpdatedAt     time.Time

	// Relations
	Deployments []Deployment `gorm:"foreignKey:ProjectID"`
	EnvVars     []EnvVar     `gorm:"foreignKey:ProjectID"`
}

// EnvVar menyimpan konfigurasi environment variable untuk container
type EnvVar struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProjectID uuid.UUID `gorm:"type:uuid;not null;index"`
	Key       string    `gorm:"type:varchar(255);not null"`
	Value     string    `gorm:"type:text;not null"`
}
