package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/BilyHakim/go-walet/config"
	"github.com/BilyHakim/go-walet/models"
	"github.com/BilyHakim/go-walet/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type IntegrationTestSuite struct {
	suite.Suite
	DB        *gorm.DB
	Router    *gin.Engine
	RabbitMQ  *config.RabbitMQ
	TestUser  models.User
	AuthToken string
}

func (s *IntegrationTestSuite) SetupSuite() {
	_ = godotenv.Load("../.env")

	// Gunakan database testing
	os.Setenv("DB_NAME", "ewallet_test")

	originalDbName := os.Getenv("DB_NAME")
	os.Setenv("DB_NAME", "postgres")
	db := config.InitDB()

	db.Exec("CREATE DATABASE ewallet_test;")

	sqlDB, _ := db.DB()
	sqlDB.Close()

	os.Setenv("DB_NAME", originalDbName)
	s.DB = config.InitDB()
	s.DB.Exec("DROP SCHEMA public CASCADE")
	s.DB.Exec("CREATE SCHEMA public")
	s.DB.AutoMigrate(&models.User{}, &models.Transaction{})

	s.RabbitMQ = config.InitRabbitMQ()

	gin.SetMode(gin.TestMode)
	s.Router = gin.New()
	routes.SetupRoutes(s.Router, s.DB, s.RabbitMQ)

	s.createTestUser()

	s.getAuthToken()
}

func (s *IntegrationTestSuite) SetupTest() {
	s.DB.Model(&s.TestUser).Update("balance", 0)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.RabbitMQ.Close()

	db, _ := s.DB.DB()
	db.Close()
}

func (s *IntegrationTestSuite) createTestUser() {
	testUser := models.User{
		FirstName:   "Test",
		LastName:    "User",
		PhoneNumber: "08123456789",
		Address:     "Test Address",
		Balance:     0,
	}
	testUser.SetPin("123456")

	s.DB.Create(&testUser)
	s.TestUser = testUser
}

func (s *IntegrationTestSuite) getAuthToken() {
	loginBody, _ := json.Marshal(map[string]string{
		"phone_number": s.TestUser.PhoneNumber,
		"pin":          "123456",
	})

	req, _ := http.NewRequest("POST", "/api/login", bytes.NewBuffer(loginBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Router.ServeHTTP(w, req)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if result, ok := response["result"].(map[string]interface{}); ok {
		s.AuthToken = fmt.Sprintf("%v", result["access_token"])
	}
}

func (s *IntegrationTestSuite) makeAuthRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			s.FailNow("Failed to marshal request body")
		}
	}

	req, _ := http.NewRequest(method, path, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	if s.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+s.AuthToken)
	}

	w := httptest.NewRecorder()
	s.Router.ServeHTTP(w, req)
	return w
}

func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
