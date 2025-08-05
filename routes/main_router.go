package routes

import (
	"github.com/Nitesh-04/realtime-racing/controllers"
	"github.com/gofiber/fiber/v2"
)

func MainRouter(api fiber.Router) {

	api.Get("/health", controllers.HealthCheck)
	AuthRouter(api)
	RaceRouter(api)
	UserRouter(api)
}