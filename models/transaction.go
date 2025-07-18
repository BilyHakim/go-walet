package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TransactionType string

const (
	TransactionTypeTopUp    TransactionType = "CREDIT"
	TransactionTypePayment  TransactionType = "DEBIT"
	TransactionTypeTransfer TransactionType = "DEBIT"
	TransactionTypeReceive  TransactionType = "CREDIT"
)

type Transaction struct {
	ID            string          `json:"transaction_id" gorm:"primaryKey;type:uuid"`
	UserID        string          `json:"user_id" gorm:"type:uuid;not null"`
	TargetUserID  *string         `json:"target_user_id,omitempty" gorm:"type:uuid"`
	Type          TransactionType `json:"transaction_type" gorm:"not null"`
	Amount        float64         `json:"amount" gorm:"not null"`
	Remarks       string          `json:"remarks"`
	BalanceBefore float64         `json:"balance_before"`
	BalanceAfter  float64         `json:"balance_after"`
	Status        string          `json:"status" gorm:"default:PENDING"`
	CreatedAt     time.Time       `json:"created_date" gorm:"autoCreateTime"`
}

func (t *Transaction) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}

	return nil
}

type TopUp struct {
	Transaction
}

type Payment struct {
	Transaction
}

type Transfer struct {
	Transaction
	Status string `json:"status" gorm:"default:PENDING"`
}
