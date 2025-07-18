package controllers

import (
	"log"
	"net/http"

	"github.com/BilyHakim/go-walet/middleware"
	"github.com/BilyHakim/go-walet/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserController struct {
	DB *gorm.DB
}

type RegisterRequest struct {
	FirstName   string `json:"first_name" binding:"required"`
	LastName    string `json:"last_name" binding:"required"`
	PhoneNumber string `json:"phone_number" binding:"required"`
	Address     string `json:"address" binding:"required"`
	PIN         string `json:"pin" binding:"required,len=6"`
}

type LoginRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
	PIN         string `json:"pin" binding:"required"`
}

type UpdateProfileRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Address   string `json:"address"`
}

type GetUserRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
}

func NewUserController(db *gorm.DB) *UserController {
	return &UserController{
		DB: db,
	}
}

func (uc *UserController) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	var existingUser models.User
	result := uc.DB.Where("phone_number = ?", req.PhoneNumber).First(&existingUser)
	if result.RowsAffected > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "ERROR",
			"message": "Phone number already exists",
		})
		return
	}

	user := models.User{
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		PhoneNumber: req.PhoneNumber,
		Address:     req.Address,
		PIN:         req.PIN,
		Balance:     0,
	}

	if err := uc.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to register user",
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "SUCCESS",
		"result": gin.H{
			"first_name":   user.FirstName,
			"last_name":    user.LastName,
			"phone_number": user.PhoneNumber,
			"address":      user.Address,
			"created_date": user.CreatedAt.Format("2006-01-02 15:04:05"),
		},
	})
}

func (uc *UserController) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	var user models.User
	result := uc.DB.Where("phone_number = ?", req.PhoneNumber).First(&user)
	if result.RowsAffected == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Invalid phone number or PIN",
		})
		return
	}

	if !user.ValidatePIN(req.PIN) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Invalid phone number or PIN",
		})
		return
	}

	accessToken, refreshToken, err := middleware.GenerateJWT(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to generate JWT",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "SUCCESS",
		"result": gin.H{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		},
	})
}

func (uc *UserController) UpdateProfile(c *gin.Context) {
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	// Debug logging
	log.Printf("UpdateProfile request: FirstName='%s', LastName='%s', Address='%s'", req.FirstName, req.LastName, req.Address)

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Unauthorized",
		})
		return
	}

	if req.FirstName == "" && req.LastName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "First name or last name is required",
		})
		return
	}

	var user models.User
	if err := uc.DB.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "User not found",
		})
		return
	}

	updates := map[string]interface{}{
		"first_name": req.FirstName,
		"last_name":  req.LastName,
	}

	if req.Address != "" {
		updates["address"] = req.Address
	}

	if err := uc.DB.Model(&user).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to update profile",
		})
		return
	}

	uc.DB.First(&user, "id = ?", userID)

	c.JSON(http.StatusOK, gin.H{
		"status": "SUCCESS",
		"result": gin.H{
			"first_name":   user.FirstName,
			"last_name":    user.LastName,
			"phone_number": user.PhoneNumber,
			"address":      user.Address,
			"created_date": user.CreatedAt.Format("2006-01-02 15:04:05"),
		},
	})
}

func (uc *UserController) GetUserByPhone(c *gin.Context) {
	var req GetUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	var user models.User
	result := uc.DB.Where("phone_number = ?", req.PhoneNumber).First(&user)
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Target user not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "SUCCESS",
		"result": gin.H{
			"user_id":      user.ID,
			"first_name":   user.FirstName,
			"last_name":    user.LastName,
			"phone_number": user.PhoneNumber,
		},
	})
}
