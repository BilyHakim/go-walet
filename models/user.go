package models

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID          string    `json:"user_id" gorm:"primaryKey;type:uuid"`
	FirstName   string    `json:"first_name" gorm:"not null"`
	LastName    string    `json:"last_name" gorm:"not null"`
	PhoneNumber string    `json:"phone_number" gorm:"unique;not null"`
	Address     string    `json:"address" gorm:"not null"`
	PIN         string    `json:"-" gorm:"not null"`
	Balance     float64   `json:"balance" gorm:"default:0"`
	CreatedAt   time.Time `json:"created_date" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_date" gorm:"autoUpdateTime"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}

	if len(u.PIN) != 60 {
		hashedPIN, err := bcrypt.GenerateFromPassword([]byte(u.PIN), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		u.PIN = string(hashedPIN)
	}

	return nil
}

func (u *User) ValidatePIN(pin string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PIN), []byte(pin))
	return err == nil
}

func (u *User) SetPin(pin string) error {
	hashedPIN, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PIN = string(hashedPIN)
	return nil
}
