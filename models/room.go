package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Room struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;" json:"id"`

	RoomCode string `gorm:"not null" json:"room_code"`

	CreatorID uuid.UUID `gorm:"type:uuid;not null" json:"creator_id"`
	Creator User `gorm:"foreignKey:CreatorID;constraint:OnDelete:CASCADE"`

	OpponentID *uuid.UUID `gorm:"type:uuid;" json:"opponent_id"`
	Opponent User `gorm:"foreignKey:OpponentID;constraint:OnDelete:CASCADE"`

	Prompt string `gorm:"not null" json:"prompt"`

	RoomStatus RoomStatus `gorm:"not null;default:'waiting'" json:"status"`

	WinnerID *uuid.UUID `gorm:"type:uuid" json:"winner_id"`
	Winner User `gorm:"foreignKey:WinnerID;constraint:OnDelete:SET NULL"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type RoomStatus string
const (
	RoomStatusWaiting    RoomStatus = "waiting"
	RoomStatusReady      RoomStatus = "ready"
	RoomStatusInProgress RoomStatus = "in_progress"
	RoomStatusCompleted  RoomStatus = "completed"
)

func (r *Room) BeforeCreate(tx *gorm.DB) (err error) {
	r.ID = uuid.New()
	return
}