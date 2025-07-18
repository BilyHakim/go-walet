package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
)

func (s *IntegrationTestSuite) TestRegister() {
	registerBody := map[string]string{
		"first_name":   "John",
		"last_name":    "Doe",
		"phone_number": "08123456780",
		"address":      "Test Address",
		"pin":          "123456",
	}

	w := s.makeAuthRequest("POST", "/api/register", registerBody)
	s.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	s.Equal("SUCCESS", response["status"])

	w = s.makeAuthRequest("POST", "/api/register", registerBody)
	s.Equal(http.StatusBadRequest, w.Code)
	json.Unmarshal(w.Body.Bytes(), &response)
	s.Equal("ERROR", response["status"])

	invalidBody := map[string]string{
		"first_name": "John",
	}
	w = s.makeAuthRequest("POST", "/api/register", invalidBody)
	s.Equal(http.StatusBadRequest, w.Code)
}

func (s *IntegrationTestSuite) TestLogin() {
	loginBody := map[string]string{
		"phone_number": s.TestUser.PhoneNumber,
		"pin":          "123456",
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/login", nil)
	s.Router.ServeHTTP(w, req)

	w = s.makeAuthRequest("POST", "/api/login", loginBody)
	s.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	s.Equal("SUCCESS", response["status"])
	result := response["result"].(map[string]interface{})
	s.NotEmpty(result["access_token"])

	invalidLoginBody := map[string]string{
		"phone_number": "0000000000",
		"pin":          "123456",
	}
	w = s.makeAuthRequest("POST", "/api/login", invalidLoginBody)
	s.Equal(http.StatusUnauthorized, w.Code)

	invalidPinBody := map[string]string{
		"phone_number": s.TestUser.PhoneNumber,
		"pin":          "000000",
	}
	w = s.makeAuthRequest("POST", "/api/login", invalidPinBody)
	s.Equal(http.StatusUnauthorized, w.Code)
}
