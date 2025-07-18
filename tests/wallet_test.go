package tests

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/BilyHakim/go-walet/models"
)

func (s *IntegrationTestSuite) TestTopup() {
	topupBody := map[string]interface{}{
		"amount": 100000,
	}

	w := s.makeAuthRequest("POST", "/api/topup", topupBody)
	s.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	s.Equal("SUCCESS", response["status"])

	var user models.User
	s.DB.First(&user, "id = ?", s.TestUser.ID)
	s.Equal(float64(100000), user.Balance)

	invalidBody := map[string]interface{}{
		"amount": -1000,
	}
	w = s.makeAuthRequest("POST", "/api/topup", invalidBody)
	s.Equal(http.StatusBadRequest, w.Code)
}

func (s *IntegrationTestSuite) TestPayment() {
	s.DB.Model(&s.TestUser).Update("balance", 100000)

	paymentBody := map[string]interface{}{
		"amount":  25000,
		"remarks": "Test Payment",
	}

	w := s.makeAuthRequest("POST", "/api/payments", paymentBody)
	s.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	s.Equal("SUCCESS", response["status"])

	var user models.User
	s.DB.First(&user, "id = ?", s.TestUser.ID)
	s.Equal(float64(75000), user.Balance)

	largePaymentBody := map[string]interface{}{
		"amount":  100000,
		"remarks": "Too Large Payment",
	}
	w = s.makeAuthRequest("POST", "/api/payments", largePaymentBody)
	s.Equal(http.StatusBadRequest, w.Code)
}

func (s *IntegrationTestSuite) TestTransfer() {
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

	var transaction models.Transaction
	result := s.DB.Where("user_id = ? AND type = ?", s.TestUser.ID, models.TransactionTypeTransfer).First(&transaction)
	if result.Error == nil {
		s.Equal("PENDING", transaction.Status)
	}

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

func (s *IntegrationTestSuite) TestTransactions() {
	for i := 1; i <= 5; i++ {
		s.DB.Create(&models.Transaction{
			UserID:        s.TestUser.ID,
			Type:          models.TransactionTypeTopUp,
			Amount:        float64(1000 * i),
			Remarks:       fmt.Sprintf("Test Transaction %d", i),
			BalanceBefore: float64(1000 * (i - 1)),
			BalanceAfter:  float64(1000 * i),
			Status:        "SUCCESS",
		})
	}

	w := s.makeAuthRequest("GET", "/api/transactions", nil)
	s.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	s.Equal("SUCCESS", response["status"])

	transactions := response["result"].([]interface{})
	s.GreaterOrEqual(len(transactions), 5)
}
