package space

import (
	"time"

	"teable-go-backend/pkg/utils"
)

// Space 领域实体
type Space struct {
	ID               string     `json:"id"`
	Name             string     `json:"name"`
	Description      *string    `json:"description"`
	Icon             *string    `json:"icon"`
	CreatedBy        string     `json:"created_by"`
	CreatedTime      time.Time  `json:"created_at"`
	DeletedTime      *time.Time `json:"deleted_time"`
	LastModifiedTime *time.Time `json:"updated_at"`
}

func NewSpace(name string, createdBy string) *Space {
	return &Space{
		ID:          utils.GenerateSpaceID(),
		Name:        name,
		CreatedBy:   createdBy,
		CreatedTime: time.Now(),
	}
}

func (s *Space) Update(name *string, description *string, icon *string) {
	if name != nil {
		s.Name = *name
	}
	if description != nil {
		s.Description = description
	}
	if icon != nil {
		s.Icon = icon
	}
	now := time.Now()
	s.LastModifiedTime = &now
}

func (s *Space) SoftDelete() {
	now := time.Now()
	s.DeletedTime = &now
}

// SpaceCollaborator 协作者
type SpaceCollaborator struct {
	ID          string
	SpaceID     string
	UserID      string
	Role        string
	CreatedTime time.Time
}

func NewSpaceCollaborator(spaceID, userID, role string) *SpaceCollaborator {
	return &SpaceCollaborator{
		ID:          utils.GenerateIDWithPrefix("spcusr"),
		SpaceID:     spaceID,
		UserID:      userID,
		Role:        role,
		CreatedTime: time.Now(),
	}
}


