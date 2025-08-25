package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Results struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;" json:"id"`

	UserID uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`

	OpponentID uuid.UUID `gorm:"type:uuid;not null" json:"opponent_id"`
	Opponent User `gorm:"foreignKey:OpponentID;constraint:OnDelete:CASCADE"`

	Won  bool `json:"won"`

	WPM int `gorm:"not null" json:"wpm"`
	Accuracy float64 `gorm:"not null" json:"accuracy"`
	Error float64 `gorm:"not null" json:"error"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (r *Results) BeforeCreate(tx *gorm.DB) (err error) {
    r.ID = uuid.New()
    return
}