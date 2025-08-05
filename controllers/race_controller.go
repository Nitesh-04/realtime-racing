package controllers

import (
	"github.com/Nitesh-04/realtime-racing/config"
	"github.com/Nitesh-04/realtime-racing/constants"
	"github.com/Nitesh-04/realtime-racing/models"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)


func LoadFullRoom(db *gorm.DB, roomCode string) (models.Room, error) {
	var room models.Room
	err := db.Preload("Creator").
		Preload("Opponent").
		Preload("Winner").
		First(&room, "room_code = ?", roomCode).Error
	return room, err
}


func CreateRoom(c *fiber.Ctx) error {

	db := config.DB

	userId := c.Locals("userId").(string)

	if userId == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
			"details": "User ID is required to create a room",
		})
	}

	roomCode := constants.GenerateRoomCode()

	// Ensure the room code is unique

	for {
		var existingRoom models.Room
		result := db.Where("room_code = ? AND room_status != ?", roomCode, models.RoomStatusCompleted).First(&existingRoom)
		if result.RowsAffected == 0 {
			break
		}
		roomCode = constants.GenerateRoomCode()
	}

	// parse the user ID to uuid.UUID

	creatorUUID, err := uuid.Parse(userId)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
			"details": err.Error(),
		})
	}

	prompt := constants.GetRandomPrompt() // Get a random prompt for the room

	// Create a new room with the generated room code and the creator's user ID
	// creator is the user who created the room

	room := models.Room{
		RoomCode:   roomCode,
		CreatorID:  creatorUUID,
		RoomStatus: models.RoomStatusWaiting,
		Prompt:     prompt,
	}


	if err := db.Create(&room).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create room",
			"details": err.Error(),
		})
	}

	// Preload the creator, opponent, and winner for the room
	room, err = LoadFullRoom(db, room.RoomCode)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to load room",
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Room created successfully",
		"room":    room,
	})
}

func JoinRoom(c *fiber.Ctx) error {
	db := config.DB

	userId := c.Locals("userId").(string)

	if userId == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
			"details": "User ID is required to join a room",
		})
	}

	roomCode := c.Params("roomCode")

	if roomCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Room code is required",
			"details": "Please provide a valid room code to join",
		})
	}

	var room models.Room

	if err := db.Where("room_code = ? AND room_status = ?", roomCode, models.RoomStatusWaiting).First(&room).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Room not found or already in progress",
			"details": "The room may not exist or is already occupied by another player",
		})
	}

	opponentUUID, err := uuid.Parse(userId)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid user ID",
			"error":   err.Error(),
		})
	}

	// Add user as opponent in the room

	room.OpponentID = &opponentUUID
	room.RoomStatus = models.RoomStatusReady // Update the room status to ready to start

	if err := db.Save(&room).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to join room",
			"details": err.Error(),
		})
	}

	room, err = LoadFullRoom(db, room.RoomCode)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to load room",
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Joined room successfully",
		"room":    room,
	})
}

func LeaveRoom(c *fiber.Ctx) error {
	db := config.DB

	userId := c.Locals("userId").(string)

	if userId == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
			"details": "User ID is required to leave a room",
		})
	}

	roomCode := c.Params("roomCode")

	if roomCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Room code is required",
			"details": "Please provide a valid room code to leave",
		})
	}

	var room models.Room

	if err := db.Where("room_code = ?", roomCode).First(&room).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Room not found",
			"details": "The room you are trying to leave does not exist",
		})
	}

	userUUID, err := uuid.Parse(userId)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
			"details": err.Error(),
		})
	}

	if room.CreatorID == userUUID {
		// If the user is the creator, delete the room
		if err := db.Delete(&room).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to delete room",
				"details": err.Error(),
			})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Room deleted as host as left room",
		})
	} else if room.OpponentID != nil && *room.OpponentID == userUUID {
		room.OpponentID = nil // Remove the opponent from the room
		room.RoomStatus = models.RoomStatusWaiting
	} else {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You are not part of this room",
			"details": "You can only leave a room you are currently in",
		})
	}

	if err := db.Save(&room).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to leave room",
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Left room successfully",
	})
}

func GetRoomDetails(c *fiber.Ctx) error {
	db := config.DB

	roomCode := c.Params("roomCode")

	if roomCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Room code is required",
			"details": "Please provide a valid room code to get details",
		})
	}

	var room models.Room

	if err := db.Where("room_code = ?", roomCode).First(&room).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Room not found",
			"details": "The room you are trying to access does not exist",
		})
	}

	room, err := LoadFullRoom(db, room.RoomCode)
	
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to load room",
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"room": room,
	})
}

func GameOver(c *fiber.Ctx) error {
	db := config.DB

	userId := c.Locals("userId").(string)

	if userId == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
			"details": "User ID is required to mark game as over",
		})
	}

	// Get the room code from the URL parameters

	roomCode := c.Params("roomCode")

	// Get the winner ID and user stats from the request body

	var body struct {
		WinnerID string  `json:"winner_id"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
			"details": err.Error(),
		})
	}

	if roomCode == "" || body.WinnerID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Room code and winner ID are required",
			"details": "Please provide both room code and winner ID to mark game as over",
		})
	}

	// Parse the winner ID to uuid.UUID

	winnerUUID, err := uuid.Parse(body.WinnerID)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid winner ID",
			"details": err.Error(),
		})
	}

	// Find the room by room code

	var room models.Room

	if err := db.Where("room_code = ?", roomCode).First(&room).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Room not found",
			"details": err.Error(),
		})
	}

	// Update the room with the winner ID and change the status to completed

	room.WinnerID = &winnerUUID
	room.RoomStatus = models.RoomStatusCompleted

	if err := db.Save(&room).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update room status",
			"details": err.Error(),
		})
	}

	room, err = LoadFullRoom(db, room.RoomCode)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to load room",
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Game over, winner updated successfully",
		"room":    room,
	})
}

func UpdateUserResult(c *fiber.Ctx) error {
	db := config.DB

	userId := c.Locals("userId").(string)

	if userId == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
			"details": "User ID is required to update user result",
		})
	}

	roomCode := c.Params("roomCode")

	if roomCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Room code is required",
			"details": "Please provide a valid room code to update user result",
		})
	}

	var body struct {
		RoomID   string  `json:"room_id"`
		WinnerID string  `json:"winner_id"`
		OpponentID string `json:"opponent_id"`
		WPM      int     `json:"wpm"`
		Accuracy float64 `json:"accuracy"`
		Error    float64 `json:"error"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
			"details": err.Error(),
		})
	}

	if roomCode == "" || body.WinnerID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Room code and winner ID are required",
			"details": "Please provide both room code and winner ID to update user result",
		})
	}

	// Parse the user ID to uuid.UUID
	userUUID, err := uuid.Parse(userId)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
			"details": err.Error(),
		})
	}

	winnerUUID, err := uuid.Parse(body.WinnerID)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid winner ID",
			"details": err.Error(),
		})
	}

	opponentUUID, err := uuid.Parse(body.OpponentID)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid opponent ID",
			"details": err.Error(),
		})
	}

	roomUUID, err := uuid.Parse(body.RoomID)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID or winner ID",
			"details": err.Error(),
		})
	}

	// Update the result for user

	result := models.Results{
		RoomID:     roomUUID,
		UserID:     userUUID,
		OpponentID: opponentUUID,
		Won:        winnerUUID == userUUID,
		WPM:        body.WPM,
		Accuracy:   body.Accuracy,
		Error:      body.Error,
	}

	if err := db.Create(&result).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update user result",
			"details": err.Error(),
		})
	}

	db.Preload("Room").
	Preload("User").
	Preload("Opponent").
	First(&result, "user_id = ? AND room_id = ?", userUUID, roomUUID)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "User result updated successfully",
		"result":  result,
	})
}