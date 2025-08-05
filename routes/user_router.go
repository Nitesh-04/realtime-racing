package routes

import (
	"github.com/Nitesh-04/realtime-racing/controllers"
	"github.com/gofiber/fiber/v2"
)

func UserRouter(api fiber.Router) {
	api.Get("/user/results", controllers.GetUserResults)
	api.Get("/user/stats", controllers.GetUserStats)
}