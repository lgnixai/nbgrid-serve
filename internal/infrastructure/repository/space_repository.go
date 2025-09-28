package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"teable-go-backend/internal/domain/space"
	"teable-go-backend/internal/infrastructure/database/models"
	"teable-go-backend/pkg/errors"
)

type SpaceRepository struct { db *gorm.DB }

func NewSpaceRepository(db *gorm.DB) space.Repository { return &SpaceRepository{db: db} }

func (r *SpaceRepository) Create(ctx context.Context, s *space.Space) error {
	model := r.domainToModel(s)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil { return r.handleDBError(err) }
	return nil
}

func (r *SpaceRepository) GetByID(ctx context.Context, id string) (*space.Space, error) {
	var m models.Space
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound { return nil, nil }
		return nil, r.handleDBError(err)
	}
	return r.modelToDomain(&m), nil
}

func (r *SpaceRepository) Update(ctx context.Context, s *space.Space) error {
	model := r.domainToModel(s)
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil { return r.handleDBError(err) }
	return nil
}

func (r *SpaceRepository) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&models.Space{}, "id = ?", id).Error; err != nil { return r.handleDBError(err) }
	return nil
}

func (r *SpaceRepository) List(ctx context.Context, filter space.ListFilter) ([]*space.Space, error) {
	var rows []models.Space
	q := r.db.WithContext(ctx).Model(&models.Space{})
	if filter.Name != nil { q = q.Where("name LIKE ?", "%"+*filter.Name+"%") }
	if filter.CreatedBy != nil { q = q.Where("created_by = ?", *filter.CreatedBy) }
	if filter.Search != "" {
		like := "%" + strings.ToLower(filter.Search) + "%"
		q = q.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?", like, like)
	}
	if filter.OrderBy != "" && filter.Order != "" { q = q.Order(fmt.Sprintf("%s %s", filter.OrderBy, filter.Order)) }
	if filter.Limit > 0 { q = q.Limit(filter.Limit) }
	if filter.Offset > 0 { q = q.Offset(filter.Offset) }
	if err := q.Find(&rows).Error; err != nil { return nil, r.handleDBError(err) }
	items := make([]*space.Space, len(rows))
	for i := range rows { items[i] = r.modelToDomain(&rows[i]) }
	return items, nil
}

func (r *SpaceRepository) Count(ctx context.Context, filter space.CountFilter) (int64, error) {
	var count int64
	q := r.db.WithContext(ctx).Model(&models.Space{})
	if filter.Name != nil { q = q.Where("name LIKE ?", "%"+*filter.Name+"%") }
	if filter.CreatedBy != nil { q = q.Where("created_by = ?", *filter.CreatedBy) }
	if filter.Search != "" {
		like := "%" + strings.ToLower(filter.Search) + "%"
		q = q.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?", like, like)
	}
	if err := q.Count(&count).Error; err != nil { return 0, r.handleDBError(err) }
	return count, nil
}

func (r *SpaceRepository) AddCollaborator(ctx context.Context, collab *space.SpaceCollaborator) error {
	m := &models.SpaceCollaborator{ ID: collab.ID, SpaceID: collab.SpaceID, UserID: collab.UserID, Role: collab.Role, CreatedTime: collab.CreatedTime }
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil { return r.handleDBError(err) }
	return nil
}

func (r *SpaceRepository) RemoveCollaborator(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&models.SpaceCollaborator{}, "id = ?", id).Error; err != nil { return r.handleDBError(err) }
	return nil
}

func (r *SpaceRepository) ListCollaborators(ctx context.Context, spaceID string) ([]*space.SpaceCollaborator, error) {
	var rows []models.SpaceCollaborator
	if err := r.db.WithContext(ctx).Where("space_id = ?", spaceID).Find(&rows).Error; err != nil { return nil, r.handleDBError(err) }
	items := make([]*space.SpaceCollaborator, len(rows))
	for i := range rows {
		items[i] = &space.SpaceCollaborator{ ID: rows[i].ID, SpaceID: rows[i].SpaceID, UserID: rows[i].UserID, Role: rows[i].Role, CreatedTime: rows[i].CreatedTime }
	}
	return items, nil
}

// 转换
func (r *SpaceRepository) domainToModel(s *space.Space) *models.Space {
	return &models.Space{
		ID: s.ID, Name: s.Name, Description: s.Description, Icon: s.Icon, CreatedBy: s.CreatedBy,
		CreatedTime: s.CreatedTime, LastModifiedTime: s.LastModifiedTime,
	}
}

func (r *SpaceRepository) modelToDomain(m *models.Space) *space.Space {
	var deleted *time.Time
	if m.DeletedTime.Valid { deleted = &m.DeletedTime.Time }
	return &space.Space{
		ID: m.ID, Name: m.Name, Description: m.Description, Icon: m.Icon, CreatedBy: m.CreatedBy,
		CreatedTime: m.CreatedTime, DeletedTime: deleted, LastModifiedTime: m.LastModifiedTime,
	}
}

func (r *SpaceRepository) handleDBError(err error) error {
	if strings.Contains(err.Error(), "duplicate key") { return errors.ErrConflict }
	return errors.ErrDatabaseOperation.WithDetails(err.Error())
}


