package controllers

import (
	"fmt"

	"github.com/Nitesh-04/realtime-racing/config"
	"github.com/Nitesh-04/realtime-racing/models"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)


func GetUserResults(c *fiber.Ctx) error {
	db := config.DB

	userId := c.Locals("userId").(string)

	if userId == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
			"details": "User ID is required to fetch results",
		})
	}

	userUUID, err := uuid.Parse(userId)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
			"details": err.Error(),
		})
	}

	limit := 20
	page := 1
	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	offset := (page - 1) * limit

	var results []models.Results
	
	err = db.
		Preload("User").
		Preload("Opponent").
		Where("user_id = ?", userUUID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&results).Error

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch user results",
			"details": err.Error(),
		})
	}

	type resultResponse struct {
		ID         uuid.UUID `json:"id"`
		UserID     uuid.UUID `json:"user_id"`
		User 	  models.User `json:"user"`
		OpponentID *uuid.UUID `json:"opponent_id"`
		Opponent   models.User `json:"opponent"`
		Won        bool      `json:"won"`
		WPM        int   `json:"wpm"`
		Accuracy   float64   `json:"accuracy"`
		Error      float64   `json:"error"`
	}

	var response []resultResponse
	for _, result := range results {
		resp := resultResponse{
			ID:         result.ID,
			UserID:     result.UserID,
			User:       result.User,
			OpponentID: &result.OpponentID,
			Opponent:   result.Opponent,
			Won:        result.Won,
			WPM:        result.WPM,
			Accuracy:   result.Accuracy,
			Error:      result.Error,
		}
		response = append(response, resp)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"results": response,
	})
}

func GetUserStats(c *fiber.Ctx) error {
	db := config.DB

	userId := c.Locals("userId").(string)
	if userId == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
			"details": "User ID is required to fetch stats",
		})
	}

	userUUID, err := uuid.Parse(userId)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
			"details": err.Error(),
		})
	}

	type Stats struct {
		AvgWPM      float64
		AvgAccuracy float64
		AvgError    float64
		TotalRaces  int64
		Wins        int64
		Losses      int64
	}

	var stats Stats

	// Calculate averages and counts
	if err := db.Model(&models.Results{}).
		Select("AVG(wpm) as avg_wpm, AVG(accuracy) as avg_accuracy, AVG(error) as avg_error, COUNT(*) as total_races").
		Where("user_id = ?", userUUID).
		Scan(&stats).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to calculate stats",
			"details": err.Error(),
		})
	}

	// Count wins
	if err := db.Model(&models.Results{}).
		Where("user_id = ? AND won = true", userUUID).
		Count(&stats.Wins).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to count wins",
			"details": err.Error(),
		})
	}

	// Count losses (not first position)
	if err := db.Model(&models.Results{}).
		Where("user_id = ? AND won = false", userUUID).
		Count(&stats.Losses).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to count losses",
			"details": err.Error(),
		})
	}


	stats.TotalRaces = stats.Wins + stats.Losses

	// return the stats

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"avg_wpm":      stats.AvgWPM,
		"avg_accuracy": stats.AvgAccuracy,
		"avg_error":    stats.AvgError,
		"total_races":  stats.TotalRaces,
		"wins":         stats.Wins,
		"losses":       stats.Losses,
	})
}