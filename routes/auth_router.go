package routes

import (
	"github.com/Nitesh-04/realtime-racing/controllers"
	"github.com/gofiber/fiber/v2"
)


func AuthRouter(api fiber.Router){
	api.Post("/register", controllers.RegisterUser)
	api.Post("/login", controllers.LoginUser)
	api.Get("/me", controllers.Me)
} 