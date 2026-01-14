package repository

import (
	"context"

	"github.com/damantine/multi-tenant-hosting/internal/core/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormProjectRepository struct {
	db *gorm.DB
}

func NewGormProjectRepository(db *gorm.DB) *GormProjectRepository {
	return &GormProjectRepository{db: db}
}

func (r *GormProjectRepository) Create(ctx context.Context, project *domain.Project) error {
	return r.db.WithContext(ctx).Create(project).Error
}

func (r *GormProjectRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
	var p domain.Project
	if err := r.db.WithContext(ctx).Preload("EnvVars").Preload("Deployments").First(&p, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *GormProjectRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Project, error) {
	var projects []domain.Project
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&projects).Error; err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *GormProjectRepository) Update(ctx context.Context, project *domain.Project) error {
	return r.db.WithContext(ctx).Save(project).Error
}

func (r *GormProjectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Project{}, "id = ?", id).Error
}
