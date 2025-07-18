package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/BilyHakim/go-walet/config"
	"github.com/BilyHakim/go-walet/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
)

type WalletController struct {
	DB       *gorm.DB
	RabbitMQ *config.RabbitMQ
}

type TopUpRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

type PaymentRequest struct {
	Amount  float64 `json:"amount" binding:"required,gt=0"`
	Remarks string  `json:"remarks" binding:"required"`
}

type TransferRequest struct {
	TargetUser string  `json:"target_user" binding:"required"`
	Amount     float64 `json:"amount" binding:"required,gt=0"`
	Remarks    string  `json:"remarks"`
}

func NewWalletController(db *gorm.DB, rmq *config.RabbitMQ) *WalletController {
	return &WalletController{
		DB:       db,
		RabbitMQ: rmq,
	}
}

func (wc *WalletController) TopUp(c *gin.Context) {
	var req TopUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Unauthorized",
		})
		return
	}

	tx := wc.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var user models.User
	if err := tx.First(&user, "id = ?", userID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{
			"message": "User not found",
		})
		return
	}

	balanceBefore := user.Balance

	user.Balance += req.Amount
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to update user balance",
		})
		return
	}

	topUpID := uuid.New().String()
	transaction := models.Transaction{
		ID:            topUpID,
		UserID:        user.ID,
		Type:          models.TransactionTypeTopUp,
		Amount:        req.Amount,
		Remarks:       "Top Up",
		BalanceBefore: balanceBefore,
		BalanceAfter:  user.Balance,
		Status:        "SUCCESS",
	}

	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to create transaction record",
		})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to process top-up",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "SUCCESS",
		"result": gin.H{
			"top_up_id":      topUpID,
			"amount_top_up":  req.Amount,
			"balance_before": balanceBefore,
			"balance_after":  user.Balance,
			"created_date":   transaction.CreatedAt.Format("2006-01-02 15:04:05"),
		},
	})
}

func (wc *WalletController) Payment(c *gin.Context) {
	var req PaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Unauthorized",
		})
		return
	}

	tx := wc.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var user models.User
	if err := tx.First(&user, "id = ?", userID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{
			"message": "User not found",
		})
		return
	}

	if user.Balance < req.Amount {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Saldo tidak cukup",
		})
		return
	}

	balanceBefore := user.Balance

	user.Balance -= req.Amount
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to update balance",
		})
		return
	}

	paymentID := uuid.New().String()
	transaction := models.Transaction{
		ID:            paymentID,
		UserID:        user.ID,
		Type:          models.TransactionTypePayment,
		Amount:        req.Amount,
		Remarks:       req.Remarks,
		BalanceBefore: balanceBefore,
		BalanceAfter:  user.Balance,
		Status:        "SUCCESS",
	}

	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to create transaction record",
		})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to process payment",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "SUCCESS",
		"result": gin.H{
			"payment_id":     paymentID,
			"amount":         req.Amount,
			"remarks":        req.Remarks,
			"balance_before": balanceBefore,
			"balance_after":  user.Balance,
			"created_date":   transaction.CreatedAt.Format("2006-01-02 15:04:05"),
		},
	})
}

func (wc *WalletController) Transfer(c *gin.Context) {
	var req TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Unauthorized",
		})
		return
	}

	tx := wc.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var user models.User
	if err := tx.First(&user, "id = ?", userID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{
			"message": "User not found",
		})
		return
	}

	if user.Balance < req.Amount {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Saldo tidak cukup",
		})
		return
	}

	var targetUser models.User
	if err := tx.First(&targetUser, "id = ?", req.TargetUser).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Target user not found",
		})
		return
	}

	balanceBefore := user.Balance

	user.Balance -= req.Amount
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to update balance",
		})
		return
	}

	transferID := uuid.New().String()
	transfer := models.Transaction{
		ID:            transferID,
		UserID:        user.ID,
		TargetUserID:  &req.TargetUser,
		Type:          models.TransactionTypeTransfer,
		Amount:        req.Amount,
		Remarks:       req.Remarks,
		BalanceBefore: balanceBefore,
		BalanceAfter:  user.Balance,
		Status:        "PENDING",
	}

	if err := tx.Create(&transfer).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to create transfer record",
		})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to process transfer",
		})
		return
	}

	transferData := map[string]interface{}{
		"transfer_id": transferID,
		"source_id":   user.ID,
		"target_id":   req.TargetUser,
		"amount":      req.Amount,
		"remarks":     req.Remarks,
	}

	transferBytes, err := json.Marshal(transferData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to prepare transfer data",
		})
		return
	}

	err = wc.RabbitMQ.Channel.Publish(
		"",               // exchange
		"transfer_queue", // routing key
		false,            // mandatory
		false,            // immediate
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        transferBytes,
		})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to queue transfer",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "SUCCESS",
		"result": gin.H{
			"transfer_id":    transferID,
			"amount":         req.Amount,
			"remarks":        req.Remarks,
			"balance_before": balanceBefore,
			"balance_after":  user.Balance,
			"created_date":   transfer.CreatedAt.Format("2006-01-02 15:04:05"),
		},
	})
}

func (wc *WalletController) GetTransactions(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Unauthorized",
		})
		return
	}

	var transactions []models.Transaction
	if err := wc.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to retrieve transactions",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "SUCCESS",
		"result": transactions,
	})
}
