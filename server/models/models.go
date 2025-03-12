package models

import (
	pb "messages/msgspb"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID       string `json:"id" gorm:"type:uuid;primaryKey"`
	Email    string `json:"email" gorm:"unique"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

func (base *User) BeforeCreate(tx *gorm.DB) (err error) {
	base.ID = uuid.NewString()
	return
}

type Camera struct {
	ID       string `json:"id" gorm:"type:uuid;primaryKey"`
	UserID   string `json:"userID"`
	User     User
	Name     string        `json:"name"`
	LastSeen *time.Time    `json:"lastSeen"`
	IsOnline bool          `json:"isOnline" gorm:"default:false"`
	Config   pb.UserConfig `json:"config" gorm:"serializer:json"`
}
