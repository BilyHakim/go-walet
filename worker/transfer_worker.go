package worker

import (
	"encoding/json"
	"log"

	"github.com/BilyHakim/go-walet/config"
	"github.com/BilyHakim/go-walet/models"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
)

type TransferWorker struct {
	DB       *gorm.DB
	RabbitMQ *config.RabbitMQ
	QuitChan chan bool
}

type TransferMessage struct {
	TransferID string  `json:"transfer_id"`
	SourceID   string  `json:"source_id"`
	TargetID   string  `json:"target_id"`
	Amount     float64 `json:"amount"`
	Remarks    string  `json:"remarks"`
}

func NewTransferWorker(db *gorm.DB, rmq *config.RabbitMQ) *TransferWorker {
	return &TransferWorker{
		DB:       db,
		RabbitMQ: rmq,
		QuitChan: make(chan bool),
	}
}

func (w *TransferWorker) Start() {
	go w.process()
}

func (w *TransferWorker) Stop() {
	w.QuitChan <- true
}

func (w *TransferWorker) process() {
	log.Println("Transfer worker started")

	msgs, err := w.RabbitMQ.Channel.Consume(
		"transfer_queue", // queue
		"",               // consumer
		false,            // auto-ack
		false,            // exclusive
		false,            // no-local
		false,            // no-wait
		nil,              // args
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	for {
		select {
		case <-w.QuitChan:
			log.Println("Transfer worker stopped")
			return
		case msg := <-msgs:
			w.processMessage(msg)
		}
	}
}

func (w *TransferWorker) processMessage(msg amqp.Delivery) {
	var transferMsg TransferMessage
	err := json.Unmarshal(msg.Body, &transferMsg)
	if err != nil {
		log.Printf("Error parsing transfer message: %v", err)
		msg.Reject(false)
		return
	}

	err = w.DB.Transaction(func(tx *gorm.DB) error {
		var transfer models.Transaction
		if err := tx.First(&transfer, "id = ?", transferMsg.TransferID).Error; err != nil {
			return err
		}

		if transfer.Status != "PENDING" {
			return nil
		}

		var targetUser models.User
		if err := tx.First(&targetUser, "id = ?", transferMsg.TargetID).Error; err != nil {
			transfer.Status = "FAILED"
			tx.Save(&transfer)
			return err
		}

		balanceBefore := targetUser.Balance

		targetUser.Balance += transferMsg.Amount
		if err := tx.Save(&targetUser).Error; err != nil {
			transfer.Status = "FAILED"
			tx.Save(&transfer)
			return err
		}

		// Generate a new valid UUID for the receive transaction
		receiveID := uuid.New().String()
		
		receiveTransaction := models.Transaction{
			ID:            receiveID, // Use a new valid UUID
			UserID:        targetUser.ID,
			TargetUserID:  &transferMsg.SourceID,
			Type:          models.TransactionTypeReceive,
			Amount:        transferMsg.Amount,
			Remarks:       transferMsg.Remarks + " (from transfer ID: " + transfer.ID + ")",
			BalanceBefore: balanceBefore,
			BalanceAfter:  targetUser.Balance,
			Status:        "SUCCESS",
		}

		if err := tx.Create(&receiveTransaction).Error; err != nil {
			transfer.Status = "FAILED"
			tx.Save(&transfer)
			return err
		}

		transfer.Status = "SUCCESS"
		if err := tx.Save(&transfer).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		log.Printf("Error processing transfer %s: %v", transferMsg.TransferID, err)
		msg.Reject(true)
	} else {
		msg.Ack(false)
	}
}
