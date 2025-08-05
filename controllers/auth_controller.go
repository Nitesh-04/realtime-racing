package controllers

import (
	"fmt"
	"os"
	"time"

	"github.com/Nitesh-04/realtime-racing/config"
	"github.com/Nitesh-04/realtime-racing/models"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// generateJWT creates a JWT token for the user

func generateJWT(userID string) (string,error) {
	
	secretKey := os.Getenv("JWT_SECRET_KEY")

	if secretKey == "" {
		return "", fmt.Errorf("JWT_SECRET_KEY is not set")
	}

	claims := jwt.MapClaims{
		"userId": userID,
		"exp":    time.Now().Add(time.Hour * 24 * 30).Unix(),
		"iat":    time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(secretKey))

	if err != nil {
		return "", fmt.Errorf("error signing token: %v", err)
	}
	return signedToken, nil
}

// validateRegistrationInput checks if the registration input is valid

func validateRegistrationInput(name, email, username, password string) (string, bool) {
	if name == "" || email == "" || username == "" || password == "" {
		return "name, email, username and password are required", false
	}
	if len(password) < 6 {
		return "password must be at least 6 characters long", false
	}
	return "", true
}

// checkUniqueUsername checks if the username is unique in the database

func checkUniqueUsername(username string, db *gorm.DB) bool {
	var user models.User

	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		return false
	}
	return true
}


func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func RegisterUser(c *fiber.Ctx) error {

	db := config.DB

	var body struct {
		Name     string `json:"name"`
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if errMsg, valid := validateRegistrationInput(body.Name, body.Email, body.Username, body.Password); !valid {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": errMsg,
		})
	}

	if checkUniqueUsername(body.Username, db) {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Username already exists",
		})
	}

	hashedPassword, err := HashPassword(body.Password)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to hash password",
		})
	}

	newUser := models.User{
		Name:     body.Name,
		Username: body.Username,
		Email:    body.Email,
		Password: hashedPassword,
	}

	if err := db.Create(&newUser).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create user",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User registered successfully",
		"user":    newUser,
	})
}

func LoginUser(c *fiber.Ctx) error {

	db := config.DB

	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	user := models.User{}

	if err := db.Where("username = ?", body.Username).First(&user).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	if !VerifyPassword(body.Password, user.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid password",
		})
	}

	token, err := generateJWT(user.ID.String()) // Generate JWT token for the user

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate token",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Login successful",
		"token":   token,
	})
}


func Me(c *fiber.Ctx) error{
	db := config.DB

	userId := c.Locals("userId").(string)

	user := models.User{}

	if err := db.Where("id = ?", userId).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"user": user,
	})

}