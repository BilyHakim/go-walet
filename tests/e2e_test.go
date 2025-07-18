package tests

import (
	"encoding/json"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/BilyHakim/go-walet/models"
	"github.com/BilyHakim/go-walet/worker"
	"github.com/stretchr/testify/suite"
)

type E2ETestSuite struct {
	IntegrationTestSuite
	Worker *worker.TransferWorker
}

func (s *E2ETestSuite) SetupSuite() {
	s.IntegrationTestSuite.SetupSuite()
	s.Worker = worker.NewTransferWorker(s.DB, s.RabbitMQ)
	s.Worker.Start()
}

func (s *E2ETestSuite) TearDownSuite() {
	if s.Worker != nil {
		s.Worker.Stop()
	}

	s.IntegrationTestSuite.TearDownSuite()
}

func (s *E2ETestSuite) TestAsyncTransfer() {
	s.DB.Exec("DELETE FROM transactions WHERE remarks LIKE '%E2E Test%'")

	secondUser := models.User{
		FirstName:   "Jane",
		LastName:    "Doe",
		PhoneNumber: "08223456789",
		Address:     "Recipient Address",
		PIN:         "654321",
		Balance:     0,
	}

	s.DB.Where("phone_number = ?", secondUser.PhoneNumber).Delete(&models.User{})

	result := s.DB.Create(&secondUser)
	s.Nil(result.Error)

	topupBody := map[string]interface{}{
		"amount": 100000,
	}
	w := s.makeAuthRequest("POST", "/api/topup", topupBody)
	s.Equal(http.StatusOK, w.Code)

	var user models.User
	s.DB.First(&user, "id = ?", s.TestUser.ID)
	s.Equal(float64(100000), user.Balance)

	transferBody := map[string]interface{}{
		"target_user": secondUser.ID,
		"amount":      50000,
		"remarks":     "E2E Test Transfer",
	}
	w = s.makeAuthRequest("POST", "/api/transfers", transferBody)
	s.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	s.Equal("SUCCESS", response["status"])

	respBytes, _ := json.MarshalIndent(response, "", "  ")
	log.Printf("Transfer API Response: %s", string(respBytes))

	s.NotNil(response["result"], "Response should contain result field")
	var transactionID string

	if result, ok := response["result"].(map[string]interface{}); ok {
		if id, ok := result["transfer_id"].(string); ok {
			transactionID = id
		} else if id, ok := result["id"].(string); ok {
			transactionID = id
		} else if id, ok := result["transaction_id"].(string); ok {
			transactionID = id
		}
	} else if resultArray, ok := response["result"].([]interface{}); ok && len(resultArray) > 0 {
		if resultMap, ok := resultArray[0].(map[string]interface{}); ok {
			if id, ok := resultMap["transfer_id"].(string); ok {
				transactionID = id
			} else if id, ok := resultMap["id"].(string); ok {
				transactionID = id
			} else if id, ok := resultMap["transaction_id"].(string); ok {
				transactionID = id
			}
		}
	} else if id, ok := response["transfer_id"].(string); ok {
		transactionID = id
	} else if id, ok := response["transaction_id"].(string); ok {
		transactionID = id
	} else if id, ok := response["id"].(string); ok {
		transactionID = id
	}

	log.Printf("Extracted transaction ID: %s", transactionID)

	s.NotEmpty(transactionID, "Transaction ID should not be empty")

	s.DB.Exec("COMMIT")

	var transaction models.Transaction
	err := s.DB.First(&transaction, "id = ?", transactionID).Error
	s.Nil(err, "Should find transaction with ID: "+transactionID)

	s.Contains([]string{"PENDING", "SUCCESS"}, transaction.Status, "Status should be either PENDING or SUCCESS")

	s.DB.First(&user, "id = ?", s.TestUser.ID)
	s.Equal(float64(50000), user.Balance)

	var recipient models.User
	s.DB.First(&recipient, "id = ?", secondUser.ID)
	s.Equal(float64(0), recipient.Balance)

	log.Printf("Waiting for async processing of transfer ID: %s", transactionID)
	log.Printf("Source user ID: %s, Target user ID: %s", s.TestUser.ID, secondUser.ID)

	maxWait := 10 * time.Second
	waitInterval := 200 * time.Millisecond
	timeout := time.After(maxWait)
	ticker := time.NewTicker(waitInterval)
	defer ticker.Stop()

	success := false
	for !success {
		select {
		case <-timeout:
			log.Printf("Transfer processing timed out. Current status: %s", transaction.Status)
			log.Printf("Recipient balance: %f", recipient.Balance)

			// Check messages in queue
			var msgCount int
			q, err := s.RabbitMQ.Channel.QueueInspect("transfer_queue")
			if err == nil {
				msgCount = q.Messages
				log.Printf("Messages remaining in queue: %d", msgCount)
			}

			s.Fail("Timed out waiting for async transfer to complete")
			return
		case <-ticker.C:
			s.DB.First(&transaction, "id = ?", transactionID)
			s.DB.First(&recipient, "id = ?", secondUser.ID)

			log.Printf("Current transfer status: %s, recipient balance: %f", transaction.Status, recipient.Balance)

			if transaction.Status == "SUCCESS" && recipient.Balance == 50000 {
				success = true
				break
			}
		}
	}

	s.Equal("SUCCESS", transaction.Status)

	s.Equal(float64(50000), recipient.Balance)

	var receiveTransaction models.Transaction
	s.DB.Where("target_user_id = ? AND type = ? AND remarks LIKE ?",
		s.TestUser.ID, models.TransactionTypeReceive, "%E2E Test Transfer%").First(&receiveTransaction)
	s.Equal(models.TransactionTypeReceive, receiveTransaction.Type)
	s.Equal("SUCCESS", receiveTransaction.Status)
	s.Equal(float64(50000), receiveTransaction.Amount)
	s.Equal(secondUser.ID, receiveTransaction.UserID)
}

func (s *E2ETestSuite) TestTransfer() {
	secondUser := models.User{
		FirstName:   "Second",
		LastName:    "User",
		PhoneNumber: "08123456790",
		Address:     "Second Address",
		Balance:     0,
	}
	secondUser.SetPin("123456")
	s.DB.Create(&secondUser)

	s.DB.Model(&s.TestUser).Update("balance", 100000)

	transferBody := map[string]interface{}{
		"target_user": secondUser.ID,
		"amount":      50000,
		"remarks":     "Test Transfer",
	}

	w := s.makeAuthRequest("POST", "/api/transfers", transferBody)
	s.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	s.Equal("SUCCESS", response["status"])

	var firstUser models.User
	s.DB.First(&firstUser, "id = ?", s.TestUser.ID)
	s.Equal(float64(50000), firstUser.Balance)
	time.Sleep(500 * time.Millisecond)

	var transaction models.Transaction
	result := s.DB.Where("user_id = ? AND type = ?", s.TestUser.ID, models.TransactionTypeTransfer).First(&transaction)
	if result.Error == nil {
		s.Equal("SUCCESS", transaction.Status)
	}

	var updatedSecondUser models.User
	s.DB.First(&updatedSecondUser, "id = ?", secondUser.ID)
	s.Equal(float64(50000), updatedSecondUser.Balance)

	largeTransferBody := map[string]interface{}{
		"target_user": secondUser.ID,
		"amount":      1000000,
		"remarks":     "Large Transfer",
	}
	w = s.makeAuthRequest("POST", "/api/transfers", largeTransferBody)
	s.Equal(http.StatusBadRequest, w.Code)

	invalidTargetBody := map[string]interface{}{
		"target_user": "invalid-uuid",
		"amount":      10000,
		"remarks":     "Invalid Target",
	}
	w = s.makeAuthRequest("POST", "/api/transfers", invalidTargetBody)
	s.Equal(http.StatusNotFound, w.Code)
}

func TestE2ETestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}
	suite.Run(t, new(E2ETestSuite))
}
