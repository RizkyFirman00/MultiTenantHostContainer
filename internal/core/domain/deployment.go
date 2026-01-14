package domain

import (
	"time"

	"github.com/google/uuid"
)

// Deployment mencatat riwayat container yang berjalan
type Deployment struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProjectID   uuid.UUID `gorm:"type:uuid;not null;index"`
	ContainerID string    `gorm:"type:varchar(64);index"` // Docker Container ID
	Status      string    `gorm:"type:varchar(20)"`       // running, exited, failed
	DeployedAt  time.Time `gorm:"autoCreateTime"`
}
